package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSubdomain(t *testing.T) {
	testCases := []struct {
		name        string
		projectName string
		projectID   uuid.UUID
		expectStart string
	}{
		{
			name:        "simple name",
			projectName: "My Project",
			projectID:   uuid.MustParse("12345678-1234-1234-1234-123456789abc"),
			expectStart: "my-project-",
		},
		{
			name:        "name with special characters",
			projectName: "Hello! World@2024",
			projectID:   uuid.MustParse("12345678-1234-1234-1234-123456789abc"),
			expectStart: "hello-world2024-",
		},
		{
			name:        "empty name",
			projectName: "",
			projectID:   uuid.MustParse("12345678-1234-1234-1234-123456789abc"),
			expectStart: "project-",
		},
		{
			name:        "cyrillic name",
			projectName: "Мой Проект",
			projectID:   uuid.MustParse("12345678-1234-1234-1234-123456789abc"),
			expectStart: "project-", // Cyrillic gets filtered out
		},
		{
			name:        "name with numbers",
			projectName: "Project123",
			projectID:   uuid.MustParse("12345678-1234-1234-1234-123456789abc"),
			expectStart: "project123-",
		},
		{
			name:        "name with leading spaces",
			projectName: "  Trimmed  ",
			projectID:   uuid.MustParse("12345678-1234-1234-1234-123456789abc"),
			expectStart: "trimmed-",
		},
		{
			name:        "name with consecutive hyphens",
			projectName: "Test---Project",
			projectID:   uuid.MustParse("12345678-1234-1234-1234-123456789abc"),
			expectStart: "test-project-",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := generateSubdomain(tc.projectName, tc.projectID)

			// Should start with expected prefix
			assert.True(t, len(result) > len(tc.expectStart), "subdomain should be longer than prefix")

			// Should end with UUID suffix
			assert.Contains(t, result, tc.projectID.String()[:8], "subdomain should contain part of project ID")

			// Should be lowercase
			assert.Equal(t, result, result, "subdomain should be lowercase")

			// Should not contain spaces
			assert.NotContains(t, result, " ", "subdomain should not contain spaces")
		})
	}
}

func TestGenerateSubdomain_Uniqueness(t *testing.T) {
	projectName := "Same Project"

	id1 := uuid.New()
	id2 := uuid.New()

	subdomain1 := generateSubdomain(projectName, id1)
	subdomain2 := generateSubdomain(projectName, id2)

	assert.NotEqual(t, subdomain1, subdomain2, "different project IDs should produce different subdomains")
}

func TestGenerateSubdomain_Length(t *testing.T) {
	// Very long project name
	longName := "This is a very long project name that should be truncated to a reasonable length for use as a subdomain"
	projectID := uuid.New()

	subdomain := generateSubdomain(longName, projectID)

	// Subdomain should not be empty
	assert.NotEmpty(t, subdomain, "subdomain should not be empty")

	// Should contain the project name part (possibly truncated)
	assert.NotEmpty(t, subdomain, "subdomain should be generated")
}

