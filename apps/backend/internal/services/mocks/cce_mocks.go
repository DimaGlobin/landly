package mocks

import (
	"context"

	domain "github.com/landly/backend/internal/models"
	"github.com/stretchr/testify/mock"
)

type BrandProfileRepositoryMock struct{ mock.Mock }

type ProductProfileRepositoryMock struct{ mock.Mock }

type ContentSnippetRepositoryMock struct{ mock.Mock }

func (m *BrandProfileRepositoryMock) Upsert(ctx context.Context, profile *domain.BrandProfile) (*domain.BrandProfile, error) {
	args := m.Called(ctx, profile)
	if p, ok := args.Get(0).(*domain.BrandProfile); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *BrandProfileRepositoryMock) GetByProjectID(ctx context.Context, projectID string) (*domain.BrandProfile, error) {
	args := m.Called(ctx, projectID)
	if p, ok := args.Get(0).(*domain.BrandProfile); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ProductProfileRepositoryMock) Upsert(ctx context.Context, profile *domain.ProductProfile) (*domain.ProductProfile, error) {
	args := m.Called(ctx, profile)
	if p, ok := args.Get(0).(*domain.ProductProfile); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ProductProfileRepositoryMock) GetByProjectID(ctx context.Context, projectID string) (*domain.ProductProfile, error) {
	args := m.Called(ctx, projectID)
	if p, ok := args.Get(0).(*domain.ProductProfile); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ContentSnippetRepositoryMock) Create(ctx context.Context, snippet *domain.ContentSnippet) (*domain.ContentSnippet, error) {
	args := m.Called(ctx, snippet)
	if s, ok := args.Get(0).(*domain.ContentSnippet); ok {
		return s, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ContentSnippetRepositoryMock) ListByProject(ctx context.Context, projectID string) ([]*domain.ContentSnippet, error) {
	args := m.Called(ctx, projectID)
	if snippets, ok := args.Get(0).([]*domain.ContentSnippet); ok {
		return snippets, args.Error(1)
	}
	return nil, args.Error(1)
}
