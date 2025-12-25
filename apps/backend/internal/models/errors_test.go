package domain

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_HTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected int
	}{
		// Authentication errors
		{"UNAUTHORIZED", ErrUnauthorized, http.StatusUnauthorized},
		{"INVALID_TOKEN", ErrInvalidToken, http.StatusUnauthorized},
		{"TOKEN_EXPIRED", ErrTokenExpired, http.StatusUnauthorized},
		{"INVALID_CREDENTIALS", ErrInvalidCredentials, http.StatusUnauthorized},
		{"FORBIDDEN", ErrForbidden, http.StatusForbidden},

		// Not found errors
		{"NOT_FOUND", ErrNotFound, http.StatusNotFound},
		{"PROJECT_NOT_FOUND", ErrProjectNotFound, http.StatusNotFound},
		{"USER_NOT_FOUND", ErrUserNotFound, http.StatusNotFound},

		// Conflict errors
		{"ALREADY_EXISTS", ErrAlreadyExists, http.StatusConflict},
		{"USER_ALREADY_EXISTS", ErrUserAlreadyExists, http.StatusConflict},
		{"CONFLICT", ErrConflict, http.StatusConflict},

		// Validation errors
		{"VALIDATION_ERROR", ErrValidationError, http.StatusBadRequest},
		{"INVALID_INPUT", ErrInvalidInput, http.StatusBadRequest},
		{"BAD_REQUEST", ErrBadRequest, http.StatusBadRequest},

		// Internal errors
		{"INTERNAL_ERROR", ErrInternal, http.StatusInternalServerError},
		{"GENERATION_FAILED", ErrGenerationFailed, http.StatusInternalServerError},
		{"RENDER_FAILED", ErrRenderFailed, http.StatusInternalServerError},
		{"PUBLISH_FAILED", ErrPublishFailed, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.HTTPStatus())
		})
	}
}

func TestError_Error(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		err := ErrUnauthorized
		assert.Equal(t, "UNAUTHORIZED: Authorization required", err.Error())
	})

	t.Run("with wrapped error", func(t *testing.T) {
		wrapped := errors.New("connection refused")
		err := ErrInternal.WithError(wrapped)
		assert.Equal(t, "INTERNAL_ERROR: Internal server error: connection refused", err.Error())
	})
}

func TestError_WithMessage(t *testing.T) {
	customErr := ErrValidationError.WithMessage("Email is required")

	assert.Equal(t, "VALIDATION_ERROR", customErr.Code)
	assert.Equal(t, "Email is required", customErr.Message)
	assert.Equal(t, http.StatusBadRequest, customErr.HTTPStatus())
}

func TestError_WithError(t *testing.T) {
	wrapped := errors.New("database connection failed")
	customErr := ErrInternal.WithError(wrapped)

	assert.Equal(t, "INTERNAL_ERROR", customErr.Code)
	assert.Equal(t, "Internal server error", customErr.Message)
	assert.Equal(t, wrapped, customErr.Unwrap())
}

func TestError_Unwrap(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		assert.Nil(t, ErrUnauthorized.Unwrap())
	})

	t.Run("with wrapped error", func(t *testing.T) {
		wrapped := errors.New("original")
		err := ErrInternal.WithError(wrapped)
		assert.Equal(t, wrapped, err.Unwrap())
	})
}

func TestError_ChainedMethods(t *testing.T) {
	wrapped := errors.New("db error")
	err := ErrInternal.WithMessage("Failed to save user").WithError(wrapped)

	assert.Equal(t, "INTERNAL_ERROR", err.Code)
	assert.Equal(t, "Failed to save user", err.Message)
	assert.Equal(t, wrapped, err.Unwrap())
}

func TestError_IsImmutable(t *testing.T) {
	// Verify that WithMessage doesn't modify original
	original := ErrValidationError
	modified := original.WithMessage("Custom message")

	assert.NotEqual(t, original.Message, modified.Message)
	assert.Equal(t, "Validation error", original.Message)
	assert.Equal(t, "Custom message", modified.Message)
}

// Test all predefined errors have correct codes
func TestPredefinedErrors(t *testing.T) {
	allErrors := []*Error{
		ErrUnauthorized,
		ErrInvalidToken,
		ErrTokenExpired,
		ErrInvalidCredentials,
		ErrForbidden,
		ErrNotFound,
		ErrProjectNotFound,
		ErrUserNotFound,
		ErrAlreadyExists,
		ErrUserAlreadyExists,
		ErrConflict,
		ErrValidationError,
		ErrInvalidInput,
		ErrBadRequest,
		ErrInternal,
		ErrGenerationFailed,
		ErrRenderFailed,
		ErrPublishFailed,
	}

	for _, err := range allErrors {
		t.Run(err.Code, func(t *testing.T) {
			assert.NotEmpty(t, err.Code, "Error code should not be empty")
			assert.NotEmpty(t, err.Message, "Error message should not be empty")
			assert.Greater(t, err.HTTPStatus(), 0, "HTTP status should be positive")
			assert.Less(t, err.HTTPStatus(), 600, "HTTP status should be valid")
		})
	}
}

