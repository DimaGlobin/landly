//go:build integration
// +build integration

package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/repositories"
	testhelpers "github.com/landly/backend/internal/testing"
)

func TestProjectService_Integration_CRUD(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	projectRepo := repositories.NewProjectRepository(qb)
	projectService := NewProjectService(projectRepo)

	ctx := context.Background()
	// Create a real user first (foreign key requirement)
	user, _ := testhelpers.CreateTestUser(t, qb, "", "")
	userID := user.ID

	// Create a project
	project, err := projectService.CreateProject(ctx, userID.String(), &domain.CreateProjectRequest{
		Name:  "Integration Test Project",
		Niche: "E-commerce",
	})
	require.NoError(t, err, "CreateProject should succeed")
	assert.NotEqual(t, uuid.Nil, project.ID)
	assert.Equal(t, "Integration Test Project", project.Name)
	assert.Equal(t, "E-commerce", project.Niche)
	assert.Equal(t, userID, project.UserID)
	assert.Equal(t, domain.ProjectStatusDraft, project.Status)

	projectID := project.ID.String()

	// Get the project
	fetched, err := projectService.GetProject(ctx, userID.String(), projectID)
	require.NoError(t, err, "GetProject should succeed")
	assert.Equal(t, project.ID, fetched.ID)
	assert.Equal(t, project.Name, fetched.Name)

	// List user's projects
	projects, err := projectService.ListProjects(ctx, userID.String())
	require.NoError(t, err, "ListProjects should succeed")
	assert.Len(t, projects, 1)
	assert.Equal(t, project.ID, projects[0].ID)

	// Update the project
	updated, err := projectService.UpdateProject(ctx, userID.String(), projectID, &domain.UpdateProjectRequest{
		Name:  "Updated Project Name",
		Niche: "SaaS",
	})
	require.NoError(t, err, "UpdateProject should succeed")
	assert.Equal(t, "Updated Project Name", updated.Name)
	assert.Equal(t, "SaaS", updated.Niche)

	// Verify update persisted
	fetched2, err := projectService.GetProject(ctx, userID.String(), projectID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Project Name", fetched2.Name)

	// Delete the project
	err = projectService.DeleteProject(ctx, userID.String(), projectID)
	require.NoError(t, err, "DeleteProject should succeed")

	// Verify project is deleted
	_, err = projectService.GetProject(ctx, userID.String(), projectID)
	require.Error(t, err, "GetProject should fail after delete")
}

func TestProjectService_Integration_AccessControl(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	projectRepo := repositories.NewProjectRepository(qb)
	projectService := NewProjectService(projectRepo)

	ctx := context.Background()
	// Create real users (foreign key requirement)
	owner, _ := testhelpers.CreateTestUser(t, qb, "", "")
	otherUser, _ := testhelpers.CreateTestUser(t, qb, "", "")
	ownerID := owner.ID
	otherUserID := otherUser.ID

	// Create a project as owner
	project, err := projectService.CreateProject(ctx, ownerID.String(), &domain.CreateProjectRequest{
		Name:  "Owner's Project",
		Niche: "SaaS",
	})
	require.NoError(t, err)
	projectID := project.ID.String()

	// Other user trying to get the project should fail
	_, err = projectService.GetProject(ctx, otherUserID.String(), projectID)
	require.Error(t, err, "Other user should not access project")
	var domainErr *domain.Error
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, "FORBIDDEN", domainErr.Code)

	// Other user trying to update should fail
	_, err = projectService.UpdateProject(ctx, otherUserID.String(), projectID, &domain.UpdateProjectRequest{
		Name: "Hacked Name",
	})
	require.Error(t, err, "Other user should not update project")
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, "FORBIDDEN", domainErr.Code)

	// Other user trying to delete should fail
	err = projectService.DeleteProject(ctx, otherUserID.String(), projectID)
	require.Error(t, err, "Other user should not delete project")
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, "FORBIDDEN", domainErr.Code)

	// Owner can still access
	_, err = projectService.GetProject(ctx, ownerID.String(), projectID)
	require.NoError(t, err, "Owner should still access project")
}

func TestProjectService_Integration_MultipleProjects(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	projectRepo := repositories.NewProjectRepository(qb)
	projectService := NewProjectService(projectRepo)

	ctx := context.Background()
	// Create a real user first (foreign key requirement)
	user, _ := testhelpers.CreateTestUser(t, qb, "", "")
	userID := user.ID

	// Create multiple projects
	projectNames := []string{"Project A", "Project B", "Project C"}
	for _, name := range projectNames {
		_, err := projectService.CreateProject(ctx, userID.String(), &domain.CreateProjectRequest{
			Name:  name,
			Niche: "Test",
		})
		require.NoError(t, err)
	}

	// List should return all projects
	projects, err := projectService.ListProjects(ctx, userID.String())
	require.NoError(t, err)
	assert.Len(t, projects, 3)
}

func TestProjectService_Integration_ValidationErrors(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	projectRepo := repositories.NewProjectRepository(qb)
	projectService := NewProjectService(projectRepo)

	ctx := context.Background()
	// Create a real user first (foreign key requirement)
	user, _ := testhelpers.CreateTestUser(t, qb, "", "")
	userID := user.ID

	// Empty name should fail
	_, err := projectService.CreateProject(ctx, userID.String(), &domain.CreateProjectRequest{
		Name:  "",
		Niche: "SaaS",
	})
	require.Error(t, err)
	var domainErr *domain.Error
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, "BAD_REQUEST", domainErr.Code)
	assert.Contains(t, domainErr.Message, "name")

	// Empty niche should fail
	_, err = projectService.CreateProject(ctx, userID.String(), &domain.CreateProjectRequest{
		Name:  "Valid Name",
		Niche: "",
	})
	require.Error(t, err)
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, "BAD_REQUEST", domainErr.Code)
	assert.Contains(t, domainErr.Message, "niche")

	// Invalid user ID should fail
	_, err = projectService.CreateProject(ctx, "invalid-uuid", &domain.CreateProjectRequest{
		Name:  "Valid Name",
		Niche: "Valid Niche",
	})
	require.Error(t, err)
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, "BAD_REQUEST", domainErr.Code)
}

