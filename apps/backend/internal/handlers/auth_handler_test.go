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

	domainErr := domain.ErrConflict
	service.On("Register", mock.Anything, mock.Anything).Return(nil, domainErr)

	ctx := gin.CreateTestContextOnly(w, g)
	ctx.Request = req

	handler.SignUp(ctx)

	assert.Equal(t, http.StatusConflict, w.Code)
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
