package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	domain "github.com/landly/backend/internal/models"
)

type ProjectRepositoryMock struct {
	mock.Mock
}

func (m *ProjectRepositoryMock) Create(ctx context.Context, project *domain.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *ProjectRepositoryMock) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	args := m.Called(ctx, id)
	if project, ok := args.Get(0).(*domain.Project); ok {
		return project, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ProjectRepositoryMock) GetByUserID(ctx context.Context, userID string) ([]*domain.Project, error) {
	args := m.Called(ctx, userID)
	if projects, ok := args.Get(0).([]*domain.Project); ok {
		return projects, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ProjectRepositoryMock) Update(ctx context.Context, project *domain.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *ProjectRepositoryMock) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *ProjectRepositoryMock) UpdateSchema(ctx context.Context, projectID string, schemaJSON string) error {
	args := m.Called(ctx, projectID, schemaJSON)
	return args.Error(0)
}

type GenerationSessionRepositoryMock struct {
	mock.Mock
}

func (m *GenerationSessionRepositoryMock) Create(ctx context.Context, session *domain.GenerationSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *GenerationSessionRepositoryMock) GetByID(ctx context.Context, id string) (*domain.GenerationSession, error) {
	args := m.Called(ctx, id)
	if session, ok := args.Get(0).(*domain.GenerationSession); ok {
		return session, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *GenerationSessionRepositoryMock) GetByProjectID(ctx context.Context, projectID string) ([]*domain.GenerationSession, error) {
	args := m.Called(ctx, projectID)
	if sessions, ok := args.Get(0).([]*domain.GenerationSession); ok {
		return sessions, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *GenerationSessionRepositoryMock) Update(ctx context.Context, session *domain.GenerationSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *GenerationSessionRepositoryMock) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type IntegrationRepositoryMock struct {
	mock.Mock
}

func (m *IntegrationRepositoryMock) Create(ctx context.Context, integration *domain.Integration) error {
	args := m.Called(ctx, integration)
	return args.Error(0)
}

func (m *IntegrationRepositoryMock) GetByID(ctx context.Context, id string) (*domain.Integration, error) {
	args := m.Called(ctx, id)
	if integration, ok := args.Get(0).(*domain.Integration); ok {
		return integration, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *IntegrationRepositoryMock) GetByProjectID(ctx context.Context, projectID string) ([]*domain.Integration, error) {
	args := m.Called(ctx, projectID)
	if integrations, ok := args.Get(0).([]*domain.Integration); ok {
		return integrations, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *IntegrationRepositoryMock) GetByProjectIDAndType(ctx context.Context, projectID string, integrationType domain.IntegrationType) (*domain.Integration, error) {
	args := m.Called(ctx, projectID, integrationType)
	if integration, ok := args.Get(0).(*domain.Integration); ok {
		return integration, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *IntegrationRepositoryMock) Update(ctx context.Context, integration *domain.Integration) error {
	args := m.Called(ctx, integration)
	return args.Error(0)
}

func (m *IntegrationRepositoryMock) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type AIClientMock struct {
	mock.Mock
}

func (m *AIClientMock) GenerateLandingSchema(ctx context.Context, prompt, paymentURL string) (string, error) {
	args := m.Called(ctx, prompt, paymentURL)
	return args.String(0), args.Error(1)
}

