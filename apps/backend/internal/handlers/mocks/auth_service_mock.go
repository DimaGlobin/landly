package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	domain "github.com/landly/backend/internal/models"
)

type AuthServiceMock struct {
	mock.Mock
}

func (m *AuthServiceMock) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	args := m.Called(ctx, req)
	if tokens, ok := args.Get(0).(*domain.AuthResponse); ok {
		return tokens, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *AuthServiceMock) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	args := m.Called(ctx, req)
	if tokens, ok := args.Get(0).(*domain.AuthResponse); ok {
		return tokens, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *AuthServiceMock) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	args := m.Called(ctx, refreshToken)
	if tokens, ok := args.Get(0).(*domain.AuthResponse); ok {
		return tokens, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *AuthServiceMock) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	args := m.Called(ctx, token)
	if user, ok := args.Get(0).(*domain.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

