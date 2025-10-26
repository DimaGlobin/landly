package logger

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TraceIDKey ключ для trace ID в контексте
const TraceIDKey = "trace_id"

// UserIDKey ключ для user ID в контексте
const UserIDKey = "user_id"

type contextKey string

const (
	traceIDContextKey contextKey = "trace_id"
	userIDContextKey  contextKey = "user_id"
)

// TraceMiddleware добавляет trace ID к каждому запросу
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Генерируем trace ID
		traceID := uuid.New().String()

		// Добавляем в заголовки ответа
		c.Header("X-Trace-ID", traceID)

		// Добавляем в контекст
		ctx := context.WithValue(c.Request.Context(), traceIDContextKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// LoggingMiddleware логирует HTTP запросы
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Обрабатываем запрос
		c.Next()

		// Логируем после обработки
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		// Получаем trace ID из контекста
		traceID := ""
		if traceIDValue := c.Request.Context().Value(traceIDContextKey); traceIDValue != nil {
			if id, ok := traceIDValue.(string); ok {
				traceID = id
			}
		}

		// Получаем user ID из контекста (если есть)
		userID := ""
		if userIDValue := c.Request.Context().Value(userIDContextKey); userIDValue != nil {
			if id, ok := userIDValue.(string); ok {
				userID = id
			}
		}

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.Int("body_size", bodySize),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if traceID != "" {
			fields = append(fields, zap.String("trace_id", traceID))
		}

		if userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		// Логируем в зависимости от статуса
		logger := Get().WithContext(c.Request.Context())
		if statusCode >= 500 {
			logger.Error("HTTP request", fields...)
		} else if statusCode >= 400 {
			logger.Warn("HTTP request", fields...)
		} else {
			logger.Info("HTTP request", fields...)
		}
	}
}

// AddUserToContext добавляет user ID в контекст
func AddUserToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}
