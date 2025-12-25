package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/landly/backend/internal/handlers/dto"
	"github.com/landly/backend/internal/handlers/mocks"
	domain "github.com/landly/backend/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthHandler_SignUp_Success(t *testing.T) {
	g := gin.Default()
	service := new(mocks.AuthServiceMock)
	handler := NewAuthHandler(service)

	request := dto.SignUpRequest{Email: "user@example.com", Password: "password123"}
	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.On("Register", mock.Anything, mock.MatchedBy(func(r *domain.RegisterRequest) bool {
		return r.Email == request.Email && r.Password == request.Password
	})).Return(&domain.AuthResponse{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil)

	ctx := gin.CreateTestContextOnly(w, g)
	ctx.Request = req

	handler.SignUp(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
	service.AssertExpectations(t)
}

func TestAuthHandler_SignUp_DomainError(t *testing.T) {
	g := gin.Default()
	service := new(mocks.AuthServiceMock)
	handler := NewAuthHandler(service)

	request := dto.SignUpRequest{Email: "user@example.com", Password: "password123"}
	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	domainErr := domain.ErrUserAlreadyExists
	service.On("Register", mock.Anything, mock.Anything).Return(nil, domainErr)

	ctx := gin.CreateTestContextOnly(w, g)
	ctx.Request = req

	handler.SignUp(ctx)

	assert.Equal(t, http.StatusConflict, w.Code)

	// Verify new error response format
	var errResp ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(t, err)
	assert.Equal(t, "USER_ALREADY_EXISTS", errResp.Error.Code)
	assert.Equal(t, "User with this email already exists", errResp.Error.Message)

	service.AssertExpectations(t)
}

func TestAuthHandler_SignUp_ValidationError(t *testing.T) {
	g := gin.Default()
	service := new(mocks.AuthServiceMock)
	handler := NewAuthHandler(service)

	// Empty body should trigger validation error
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := gin.CreateTestContextOnly(w, g)
	ctx.Request = req

	handler.SignUp(ctx)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", errResp.Error.Code)
}

func TestAuthHandler_SignIn_Success(t *testing.T) {
	g := gin.Default()
	service := new(mocks.AuthServiceMock)
	handler := NewAuthHandler(service)

	request := dto.SignInRequest{Email: "user@example.com", Password: "password123"}
	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.On("Login", mock.Anything, mock.MatchedBy(func(r *domain.LoginRequest) bool {
		return r.Email == request.Email && r.Password == request.Password
	})).Return(&domain.AuthResponse{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil)

	ctx := gin.CreateTestContextOnly(w, g)
	ctx.Request = req

	handler.SignIn(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
	service.AssertExpectations(t)
}

func TestAuthHandler_SignIn_InvalidCredentials(t *testing.T) {
	g := gin.Default()
	service := new(mocks.AuthServiceMock)
	handler := NewAuthHandler(service)

	request := dto.SignInRequest{Email: "user@example.com", Password: "wrongpassword"}
	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.On("Login", mock.Anything, mock.Anything).Return(nil, domain.ErrInvalidCredentials)

	ctx := gin.CreateTestContextOnly(w, g)
	ctx.Request = req

	handler.SignIn(ctx)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_CREDENTIALS", errResp.Error.Code)
	assert.Equal(t, "Invalid email or password", errResp.Error.Message)

	service.AssertExpectations(t)
}

func TestAuthHandler_SignIn_InternalError(t *testing.T) {
	g := gin.Default()
	service := new(mocks.AuthServiceMock)
	handler := NewAuthHandler(service)

	request := dto.SignInRequest{Email: "user@example.com", Password: "password123"}
	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.On("Login", mock.Anything, mock.Anything).Return(nil, errors.New("boom"))

	ctx := gin.CreateTestContextOnly(w, g)
	ctx.Request = req

	handler.SignIn(ctx)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify internal error doesn't expose details
	var errResp ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", errResp.Error.Code)
	assert.Equal(t, "Internal server error", errResp.Error.Message)

	service.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	g := gin.Default()
	service := new(mocks.AuthServiceMock)
	handler := NewAuthHandler(service)

	request := dto.RefreshTokenRequest{RefreshToken: "refresh"}
	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.On("RefreshToken", mock.Anything, request.RefreshToken).Return(&domain.AuthResponse{
		AccessToken:  "new-access",
		RefreshToken: "new-refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil)

	ctx := gin.CreateTestContextOnly(w, g)
	ctx.Request = req

	handler.RefreshToken(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
	service.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_TokenExpired(t *testing.T) {
	g := gin.Default()
	service := new(mocks.AuthServiceMock)
	handler := NewAuthHandler(service)

	request := dto.RefreshTokenRequest{RefreshToken: "expired-token"}
	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.On("RefreshToken", mock.Anything, request.RefreshToken).Return(nil, domain.ErrTokenExpired)

	ctx := gin.CreateTestContextOnly(w, g)
	ctx.Request = req

	handler.RefreshToken(ctx)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(t, err)
	assert.Equal(t, "TOKEN_EXPIRED", errResp.Error.Code)

	service.AssertExpectations(t)
}
