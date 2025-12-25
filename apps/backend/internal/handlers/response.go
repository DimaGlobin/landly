package handlers

import (
	"github.com/gin-gonic/gin"
	domain "github.com/landly/backend/internal/models"
)

// ErrorResponse represents a standardized API error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespondError sends a standardized error response
// Format: {"error": {"code": "ERROR_CODE", "message": "Human readable message"}}
func RespondError(c *gin.Context, err *domain.Error) {
	c.JSON(err.HTTPStatus(), gin.H{
		"error": ErrorResponse{
			Code:    err.Code,
			Message: err.Message,
		},
	})
}

// RespondErrorWithStatus sends an error with a custom HTTP status
func RespondErrorWithStatus(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": ErrorResponse{
			Code:    code,
			Message: message,
		},
	})
}

// RespondValidationError sends a validation error response
func RespondValidationError(c *gin.Context, message string) {
	c.JSON(400, gin.H{
		"error": ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: message,
		},
	})
}

// RespondUnauthorized sends an unauthorized error response
func RespondUnauthorized(c *gin.Context, code, message string) {
	c.JSON(401, gin.H{
		"error": ErrorResponse{
			Code:    code,
			Message: message,
		},
	})
}

// RespondInternalError sends an internal server error response
// Note: does not expose internal error details to client
func RespondInternalError(c *gin.Context) {
	c.JSON(500, gin.H{
		"error": ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Internal server error",
		},
	})
}

