package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	domain "github.com/landly/backend/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ErrorResponseBody represents the expected error response structure
type ErrorResponseBody struct {
	Error ErrorResponse `json:"error"`
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name           string
		err            *domain.Error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "unauthorized error",
			err:            domain.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
			expectedMsg:    "Authorization required",
		},
		{
			name:           "invalid credentials error",
			err:            domain.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "INVALID_CREDENTIALS",
			expectedMsg:    "Invalid email or password",
		},
		{
			name:           "project not found error",
			err:            domain.ErrProjectNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "PROJECT_NOT_FOUND",
			expectedMsg:    "Project not found",
		},
		{
			name:           "user already exists error",
			err:            domain.ErrUserAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedCode:   "USER_ALREADY_EXISTS",
			expectedMsg:    "User with this email already exists",
		},
		{
			name:           "validation error",
			err:            domain.ErrValidationError,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_ERROR",
			expectedMsg:    "Validation error",
		},
		{
			name:           "internal error",
			err:            domain.ErrInternal,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
			expectedMsg:    "Internal server error",
		},
		{
			name:           "generation failed error",
			err:            domain.ErrGenerationFailed,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "GENERATION_FAILED",
			expectedMsg:    "AI generation failed",
		},
		{
			name:           "custom message error",
			err:            domain.ErrValidationError.WithMessage("Email is required"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_ERROR",
			expectedMsg:    "Email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			RespondError(c, tt.err)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var body ErrorResponseBody
			err := json.Unmarshal(w.Body.Bytes(), &body)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, body.Error.Code)
			assert.Equal(t, tt.expectedMsg, body.Error.Message)
		})
	}
}

func TestRespondErrorWithStatus(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondErrorWithStatus(c, http.StatusTeapot, "CUSTOM_ERROR", "I'm a teapot")

	assert.Equal(t, http.StatusTeapot, w.Code)

	var body ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "CUSTOM_ERROR", body.Error.Code)
	assert.Equal(t, "I'm a teapot", body.Error.Message)
}

func TestRespondValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondValidationError(c, "Field 'email' is required")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", body.Error.Code)
	assert.Equal(t, "Field 'email' is required", body.Error.Message)
}

func TestRespondUnauthorized(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
	}{
		{"token expired", "TOKEN_EXPIRED", "Token has expired"},
		{"invalid token", "INVALID_TOKEN", "Invalid token"},
		{"unauthorized", "UNAUTHORIZED", "Authorization required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			RespondUnauthorized(c, tt.code, tt.message)

			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var body ErrorResponseBody
			err := json.Unmarshal(w.Body.Bytes(), &body)
			assert.NoError(t, err)
			assert.Equal(t, tt.code, body.Error.Code)
			assert.Equal(t, tt.message, body.Error.Message)
		})
	}
}

func TestRespondInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondInternalError(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", body.Error.Code)
	assert.Equal(t, "Internal server error", body.Error.Message)
}

// Test that error response has correct JSON structure
func TestErrorResponseFormat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondError(c, domain.ErrUnauthorized)

	// Verify exact JSON structure
	expected := `{"error":{"code":"UNAUTHORIZED","message":"Authorization required"}}`
	assert.JSONEq(t, expected, w.Body.String())
}

