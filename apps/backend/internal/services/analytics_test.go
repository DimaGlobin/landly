package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/landly/backend/internal/models"
)

func TestAnalyticsService_TrackEvent_RequiresProjectID(t *testing.T) {
	// Mock repo that records events
	var capturedEvent *domain.AnalyticsEvent
	repo := &mockAnalyticsRepo{
		trackFunc: func(ctx context.Context, e *domain.AnalyticsEvent) error {
			capturedEvent = e
			return nil
		},
	}
	projectRepo := &mockProjectRepo{}
	svc := NewAnalyticsService(projectRepo, repo)

	projectID := uuid.New()

	err := svc.TrackEvent(context.Background(), &domain.TrackEventRequest{
		ProjectID: projectID.String(),
		EventType: "pageview",
		Path:      "/",
		Referrer:  "",
	})
	require.NoError(t, err)
	require.NotNil(t, capturedEvent)
	assert.Equal(t, projectID, capturedEvent.ProjectID)
	assert.Equal(t, "pageview", capturedEvent.EventType)
	assert.Equal(t, "/", capturedEvent.Path)
}

func TestAnalyticsService_TrackEvent_RejectsEmptyProjectID(t *testing.T) {
	repo := &mockAnalyticsRepo{}
	projectRepo := &mockProjectRepo{}
	svc := NewAnalyticsService(projectRepo, repo)

	err := svc.TrackEvent(context.Background(), &domain.TrackEventRequest{
		ProjectID: "",
		EventType: "pageview",
		Path:      "/",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project ID")
}

func TestAnalyticsService_TrackEvent_RejectsInvalidProjectID(t *testing.T) {
	repo := &mockAnalyticsRepo{}
	projectRepo := &mockProjectRepo{}
	svc := NewAnalyticsService(projectRepo, repo)

	err := svc.TrackEvent(context.Background(), &domain.TrackEventRequest{
		ProjectID: "not-a-uuid",
		EventType: "pageview",
		Path:      "/",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

type mockAnalyticsRepo struct {
	trackFunc func(ctx context.Context, e *domain.AnalyticsEvent) error
}

func (m *mockAnalyticsRepo) TrackEvent(ctx context.Context, e *domain.AnalyticsEvent) error {
	if m.trackFunc != nil {
		return m.trackFunc(ctx, e)
	}
	return nil
}

func (m *mockAnalyticsRepo) GetStats(ctx context.Context, projectID uuid.UUID) (*domain.AnalyticsStats, error) {
	return &domain.AnalyticsStats{}, nil
}

func (m *mockAnalyticsRepo) GetEvents(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*domain.AnalyticsEvent, error) {
	return nil, nil
}

type mockProjectRepo struct{}

func (m *mockProjectRepo) Create(ctx context.Context, project *domain.Project) error   { return nil }
func (m *mockProjectRepo) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	return nil, domain.ErrNotFound
}
func (m *mockProjectRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Project, error) {
	return nil, nil
}
func (m *mockProjectRepo) Update(ctx context.Context, project *domain.Project) error   { return nil }
func (m *mockProjectRepo) Delete(ctx context.Context, id string) error                 { return nil }
func (m *mockProjectRepo) UpdateSchema(ctx context.Context, projectID, schemaJSON string) error {
	return nil
}
