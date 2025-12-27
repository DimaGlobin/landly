//go:build integration
// +build integration

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/landly/backend/internal/handlers/dto"
	"github.com/landly/backend/internal/repositories"
	"github.com/landly/backend/internal/services"
	testhelpers "github.com/landly/backend/internal/testing"
)

const testJWTSecret = "test-e2e-jwt-secret"

func setupTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	qb := testhelpers.SetupTestDB(t)

	// Create repositories
	userRepo := repositories.NewUserRepository(qb)
	projectRepo := repositories.NewProjectRepository(qb)
	publishTargetRepo := repositories.NewPublishTargetRepository(qb)

	// Create services
	authService := services.NewAuthService(userRepo, testJWTSecret, 15*time.Minute, 7*24*time.Hour)
	// For e2e tests, we can use nil publisher since project deletion is not tested here
	projectService := services.NewProjectService(projectRepo, publishTargetRepo, nil)

	// Create handlers
	authHandler := NewAuthHandler(authService)
	publicBaseProvider := services.NewStaticPublicBaseProvider("http://localhost:8080")
	projectHandler := NewProjectHandler(projectService, nil, publicBaseProvider)

	// Create router
	r := gin.New()

	// Auth routes (no middleware)
	auth := r.Group("/v1/auth")
	{
		auth.POST("/signup", authHandler.SignUp)
		auth.POST("/login", authHandler.SignIn)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected routes
	protected := r.Group("/v1")
	protected.Use(AuthMiddleware(testJWTSecret))
	{
		protected.GET("/projects", projectHandler.GetProjects)
		protected.POST("/projects", projectHandler.CreateProject)
		protected.GET("/projects/:id", projectHandler.GetProject)
		protected.DELETE("/projects/:id", projectHandler.DeleteProject)
	}

	return r
}

func generateToken(userID uuid.UUID) string {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testJWTSecret))
	return tokenString
}

func TestE2E_AuthFlow(t *testing.T) {
	router := setupTestRouter(t)

	email := "e2e-test-" + uuid.New().String()[:8] + "@example.com"
	password := "SecurePass123!"

	// Sign up
	signupReq := dto.SignUpRequest{Email: email, Password: password}
	body, _ := json.Marshal(signupReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Signup should return 200")

	var signupResp dto.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &signupResp)
	require.NoError(t, err)
	assert.NotEmpty(t, signupResp.AccessToken)
	assert.NotEmpty(t, signupResp.RefreshToken)

	// Sign in with same credentials
	signinReq := dto.SignInRequest{Email: email, Password: password}
	body, _ = json.Marshal(signinReq)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Login should return 200")

	var signinResp dto.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &signinResp)
	require.NoError(t, err)
	assert.NotEmpty(t, signinResp.AccessToken)

	// Refresh token
	refreshReq := dto.RefreshTokenRequest{RefreshToken: signinResp.RefreshToken}
	body, _ = json.Marshal(refreshReq)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Refresh should return 200")
}

func TestE2E_AuthErrors(t *testing.T) {
	router := setupTestRouter(t)

	t.Run("signup with existing email", func(t *testing.T) {
		email := "e2e-dup-" + uuid.New().String()[:8] + "@example.com"

		// First signup
		signupReq := dto.SignUpRequest{Email: email, Password: "Pass123!"}
		body, _ := json.Marshal(signupReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// Second signup with same email
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("login with wrong password", func(t *testing.T) {
		email := "e2e-wrong-" + uuid.New().String()[:8] + "@example.com"

		// Signup
		signupReq := dto.SignUpRequest{Email: email, Password: "CorrectPass123!"}
		body, _ := json.Marshal(signupReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// Login with wrong password
		signinReq := dto.SignInRequest{Email: email, Password: "WrongPass123!"}
		body, _ = json.Marshal(signinReq)
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errResp ErrorResponseBody
		json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "INVALID_CREDENTIALS", errResp.Error.Code)
	})
}

func TestE2E_ProjectCRUD(t *testing.T) {
	router := setupTestRouter(t)

	// First create a user and get token
	email := "e2e-project-" + uuid.New().String()[:8] + "@example.com"
	signupReq := dto.SignUpRequest{Email: email, Password: "Pass123!"}
	body, _ := json.Marshal(signupReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var authResp dto.AuthResponse
	json.Unmarshal(w.Body.Bytes(), &authResp)
	token := authResp.AccessToken

	// Create project
	createReq := dto.CreateProjectRequest{Name: "E2E Test Project", Niche: "SaaS"}
	body, _ = json.Marshal(createReq)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Create project should return 200")

	var projectResp dto.ProjectResponse
	err := json.Unmarshal(w.Body.Bytes(), &projectResp)
	require.NoError(t, err)
	assert.Equal(t, "E2E Test Project", projectResp.Name)
	assert.Equal(t, "SaaS", projectResp.Niche)

	projectID := projectResp.ID.String()

	// Get project
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/projects/"+projectID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Get project should return 200")

	// List projects
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "List projects should return 200")

	var listResp dto.ProjectsListResponse
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	require.NoError(t, err)
	assert.Equal(t, 1, listResp.Total)

	// Delete project
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/v1/projects/"+projectID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Accept 200 or 204
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent)
}

func TestE2E_UnauthorizedAccess(t *testing.T) {
	router := setupTestRouter(t)

	// Try to access protected endpoint without token
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/projects", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp ErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "UNAUTHORIZED", errResp.Error.Code)
}

func TestE2E_ExpiredToken(t *testing.T) {
	router := setupTestRouter(t)

	// Generate expired token
	userID := uuid.New()
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		"iat":     time.Now().Add(-2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testJWTSecret))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/projects", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp ErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "TOKEN_EXPIRED", errResp.Error.Code)
}

