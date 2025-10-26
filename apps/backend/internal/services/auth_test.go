package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/services/mocks"
)

func TestAuthService_SignUp_CreateUserAndGenerateTokens(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mocks.UserRepositoryMock)
	authService := NewAuthService(userRepo, "super-secret")

	userRepo.On("GetByEmail", ctx, "user@example.com").Return(nil, errors.New("not found"))
	userRepo.On("Create", ctx, mock.MatchedBy(func(user *domain.User) bool {
		require.Equal(t, "user@example.com", user.Email)
		require.NotEmpty(t, user.PasswordHash)
		require.NotZero(t, user.ID)
		return true
	})).Return(nil)

	tokens, err := authService.SignUp(ctx, "user@example.com", "password123")
	require.NoError(t, err)
	require.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), tokens.ExpiresAt, 5*time.Second)

	userRepo.AssertExpectations(t)
}

func TestAuthService_SignIn_InvalidPassword(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mocks.UserRepositoryMock)
	authService := NewAuthService(userRepo, "secret")

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)
	stored := &domain.User{ID: uuid.New(), Email: "user@example.com", PasswordHash: string(hash)}
	userRepo.On("GetByEmail", ctx, stored.Email).Return(stored, nil)

	resp, err := authService.SignIn(ctx, stored.Email, "wrong")
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mocks.UserRepositoryMock)
	authService := NewAuthService(userRepo, "secret")

	user := &domain.User{ID: uuid.New(), Email: "user@example.com"}
	tokens, err := authService.generateTokens(user.ID)
	require.NoError(t, err)

	userRepo.On("GetByID", ctx, user.ID.String()).Return(user, nil)

	result, err := authService.ValidateToken(ctx, tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestAuthService_RefreshToken_Invalid(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mocks.UserRepositoryMock)
	authService := NewAuthService(userRepo, "secret")

	resp, err := authService.RefreshToken(ctx, "bad-token")
	assert.Error(t, err)
	assert.Nil(t, resp)
}
