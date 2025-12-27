package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/landly/backend/internal/logger"
	domain "github.com/landly/backend/internal/models"
	"go.uber.org/zap"
)

// ErrorResponse represents a standardized API error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespondError sends a standardized error response and logs it
// Format: {"error": {"code": "ERROR_CODE", "message": "Human readable message"}}
// Внутренние детали ошибки логируются, но не отправляются клиенту
func RespondError(c *gin.Context, err *domain.Error) {
	// Логируем доменную ошибку с контекстом запроса (если Request доступен)
	if c.Request != nil {
		logger.LogDomainError(c.Request.Context(), "handler error",
			err,
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.ClientIP()),
		)
	} else {
		// Для тестов без Request используем Background контекст
		logger.LogDomainError(context.Background(), "handler error", err)
	}
	
	// Отправляем клиенту только безопасное сообщение
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
// Логирует ошибку с контекстом запроса
func RespondInternalError(c *gin.Context, err error) {
	// Логируем внутреннюю ошибку с максимальным контекстом (если Request доступен)
	if c.Request != nil {
		logger.LogInternalError(c.Request.Context(), "internal server error",
			err,
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.ClientIP()),
		)
	} else {
		// Для тестов без Request используем Background контекст
		logger.LogInternalError(context.Background(), "internal server error", err)
	}
	
	// Отправляем клиенту только общее сообщение без деталей
	c.JSON(500, gin.H{
		"error": ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Internal server error",
		},
	})
}

// RespondInternalErrorWithoutLogging sends an internal server error response without logging
// Используется только в случаях, когда ошибка уже была залогирована ранее
func RespondInternalErrorWithoutLogging(c *gin.Context) {
	c.JSON(500, gin.H{
		"error": ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Internal server error",
		},
	})
}

