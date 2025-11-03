//go:build integration
// +build integration

package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/repositories"
	"github.com/landly/backend/internal/storage/ai"
	testhelpers "github.com/landly/backend/internal/testing"
)

func TestGenerateService_Integration_GenerateSite(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	projectRepo := repositories.NewProjectRepository(qb)
	sessionRepo := repositories.NewGenerationSessionRepository(qb)
	integrationRepo := repositories.NewIntegrationRepository(qb)
	messageRepo := repositories.NewGenerationMessageRepository(qb)

	user, _ := testhelpers.CreateTestUser(t, qb, "", "")
	project := testhelpers.CreateTestProject(t, qb, user.ID, "Integration Test Project", "SaaS")

	aiClient := ai.NewMockClient()
	generateService := NewGenerateService(projectRepo, integrationRepo, sessionRepo, messageRepo, aiClient)

	ctx := context.Background()
	req := &domain.GenerateRequest{
		Prompt:     "Create a landing page for a SaaS product",
		PaymentURL: "https://example.com/pay",
	}

	session, err := generateService.GenerateSite(ctx, user.ID.String(), project.ID.String(), req)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, domain.GenerationStatusCompleted, session.Status)
	assert.NotEmpty(t, session.SchemaJSON)
	assert.NotNil(t, session.CompletedAt)

	updatedProject, err := projectRepo.GetByID(ctx, project.ID.String())
	require.NoError(t, err)
	assert.Equal(t, session.SchemaJSON, updatedProject.SchemaJSON)

	retrievedSession, err := sessionRepo.GetByID(ctx, session.ID.String())
	require.NoError(t, err)
	assert.Equal(t, domain.GenerationStatusCompleted, retrievedSession.Status)
	assert.Equal(t, session.SchemaJSON, retrievedSession.SchemaJSON)
}

func TestGenerateService_Integration_GetGenerationStatusAndResult(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	projectRepo := repositories.NewProjectRepository(qb)
	sessionRepo := repositories.NewGenerationSessionRepository(qb)
	integrationRepo := repositories.NewIntegrationRepository(qb)
	messageRepo := repositories.NewGenerationMessageRepository(qb)

	user, _ := testhelpers.CreateTestUser(t, qb, "", "")
	project := testhelpers.CreateTestProject(t, qb, user.ID, "Status Test Project", "Analytics")

	aiClient := ai.NewMockClient()
	generateService := NewGenerateService(projectRepo, integrationRepo, sessionRepo, messageRepo, aiClient)

	ctx := context.Background()
	req := &domain.GenerateRequest{
		Prompt:     "Another landing page",
		PaymentURL: "https://example.com/buy",
	}

	session, err := generateService.GenerateSite(ctx, user.ID.String(), project.ID.String(), req)
	require.NoError(t, err)
	require.NotNil(t, session)

	statusSession, err := generateService.GetGenerationStatus(ctx, user.ID.String(), session.ID.String())
	require.NoError(t, err)
	assert.Equal(t, session.ID, statusSession.ID)
	assert.Equal(t, domain.GenerationStatusCompleted, statusSession.Status)

	result, err := generateService.GetGenerationResult(ctx, user.ID.String(), session.ID.String())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Schema)

	var expectedSchema domain.PageSchema
	require.NoError(t, json.Unmarshal([]byte(session.SchemaJSON), &expectedSchema))
	assert.Equal(t, expectedSchema.Title, result.Schema.Title)
}

func TestGenerateService_Integration_GenerateSiteAIError(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	projectRepo := repositories.NewProjectRepository(qb)
	sessionRepo := repositories.NewGenerationSessionRepository(qb)
	integrationRepo := repositories.NewIntegrationRepository(qb)
	messageRepo := repositories.NewGenerationMessageRepository(qb)

	user, _ := testhelpers.CreateTestUser(t, qb, "", "")
	project := testhelpers.CreateTestProject(t, qb, user.ID, "Error Test Project", "Marketing")

	ctx := context.Background()
	req := &domain.GenerateRequest{
		Prompt:     "Prompt that causes AI error",
		PaymentURL: "https://example.com/fail",
	}

	generateService := NewGenerateService(projectRepo, integrationRepo, sessionRepo, messageRepo, failingAIClient{})

	session, err := generateService.GenerateSite(ctx, user.ID.String(), project.ID.String(), req)
	require.Error(t, err)
	require.NotNil(t, session)
	assert.Equal(t, domain.GenerationStatusFailed, session.Status)
	assert.NotNil(t, session.CompletedAt)

	retrievedSession, getErr := sessionRepo.GetByID(ctx, session.ID.String())
	require.NoError(t, getErr)
	assert.Equal(t, domain.GenerationStatusFailed, retrievedSession.Status)
}

type failingAIClient struct{}

func (failingAIClient) GenerateLandingSchema(ctx context.Context, prompt, paymentURL string) (string, error) {
	return "", errors.New("ai generation failed")
}
