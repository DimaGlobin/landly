//go:build integration
// +build integration

package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/repositories"
	testhelpers "github.com/landly/backend/internal/testing"
)

func TestAuthService_Integration_SignUpSignInFlow(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	userRepo := repositories.NewUserRepository(qb)
	authService := NewAuthService(userRepo, "test-secret-key-integration", 15*time.Minute, 7*24*time.Hour)

	ctx := context.Background()
	email := "integration-test@example.com"
	password := "SecurePassword123!"

	// Sign up a new user
	tokens, err := authService.SignUp(ctx, email, password)
	require.NoError(t, err, "SignUp should succeed")
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)

	// Ensure user persisted in DB
	user, err := userRepo.GetByEmail(ctx, email)
	require.NoError(t, err, "user should be stored")
	assert.Equal(t, email, user.Email)

	// Sign in with the same credentials
	loginTokens, err := authService.SignIn(ctx, email, password)
	require.NoError(t, err, "SignIn should succeed")
	assert.NotEmpty(t, loginTokens.AccessToken)
	assert.NotEmpty(t, loginTokens.RefreshToken)

	// Validate access token
	validatedUser, err := authService.ValidateToken(ctx, loginTokens.AccessToken)
	require.NoError(t, err, "ValidateToken should succeed")
	assert.Equal(t, user.ID, validatedUser.ID)
	assert.Equal(t, email, validatedUser.Email)
}

func TestAuthService_Integration_DuplicateEmail(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	userRepo := repositories.NewUserRepository(qb)
	authService := NewAuthService(userRepo, "test-secret-key-integration", 15*time.Minute, 7*24*time.Hour)

	ctx := context.Background()
	email := "duplicate@example.com"
	password := "SecurePassword123!"

	// First signup should succeed
	_, err := authService.SignUp(ctx, email, password)
	require.NoError(t, err, "First signup should succeed")

	// Second signup with same email should fail
	_, err = authService.SignUp(ctx, email, password)
	require.Error(t, err, "Duplicate email should fail")
	var domainErr *domain.Error
	require.True(t, errors.As(err, &domainErr), "error should be a domain error")
	assert.Equal(t, domain.ErrConflict.Code, domainErr.Code)
	assert.Contains(t, domainErr.Message, "уже существует")
}

func TestAuthService_Integration_InvalidCredentials(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	userRepo := repositories.NewUserRepository(qb)
	authService := NewAuthService(userRepo, "test-secret-key-integration", 15*time.Minute, 7*24*time.Hour)

	ctx := context.Background()
	email := "password-test@example.com"
	password := "CorrectPassword123!"

	// Create user
	_, err := authService.SignUp(ctx, email, password)
	require.NoError(t, err, "SignUp should succeed")

	// Try to sign in with wrong password
	_, err = authService.SignIn(ctx, email, "WrongPassword123!")
	assert.Error(t, err, "Wrong password should fail")
	assert.Contains(t, err.Error(), "invalid credentials", "Error should mention invalid credentials")

	// Try to sign in with non-existent email
	_, err = authService.SignIn(ctx, "nonexistent@example.com", password)
	assert.Error(t, err, "Non-existent email should fail")
}

func TestAuthService_Integration_InvalidToken(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	userRepo := repositories.NewUserRepository(qb)
	authService := NewAuthService(userRepo, "test-secret-key-integration", 15*time.Minute, 7*24*time.Hour)

	// Test with invalid token
	_, err := authService.ValidateToken(context.Background(), "invalid.token.string")
	assert.Error(t, err, "Invalid token should fail validation")

	// Test with empty token
	_, err = authService.ValidateToken(context.Background(), "")
	assert.Error(t, err, "Empty token should fail validation")
}

func TestAuthService_Integration_RefreshToken(t *testing.T) {
	qb := testhelpers.SetupTestDB(t)
	userRepo := repositories.NewUserRepository(qb)
	authService := NewAuthService(userRepo, "test-secret-key-integration", 15*time.Minute, 7*24*time.Hour)

	ctx := context.Background()
	email := "refresh-test@example.com"
	password := "SecurePassword123!"

	// Create user and get initial token
	tokens, err := authService.SignUp(ctx, email, password)
	require.NoError(t, err, "SignUp should succeed")

	// Refresh token
	resp, err := authService.RefreshToken(ctx, tokens.RefreshToken)
	require.NoError(t, err, "Token refresh should succeed")
	assert.NotEmpty(t, resp.AccessToken, "Access token should be returned")
	assert.NotEmpty(t, resp.RefreshToken, "Refresh token should be returned")
	assert.NotEqual(t, tokens.AccessToken, resp.AccessToken, "New access token should differ")

	// Validate new token
	validatedUser, err := authService.ValidateToken(ctx, resp.AccessToken)
	require.NoError(t, err, "New token should be valid")
	assert.Equal(t, email, validatedUser.Email, "User email should match")
}
