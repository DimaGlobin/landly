package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/landly/backend/internal/logger"
	"go.uber.org/zap"
)

// AuthMiddleware проверяет JWT токен
// NOTE: Using X-Landly-Authorization instead of Authorization
// because Yandex Serverless Containers intercept Authorization header
// and return 403 before request reaches our backend
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
	// Try custom header first (for Yandex Serverless compatibility)
	authHeader := c.GetHeader("X-Landly-Authorization")
	// Fallback to standard Authorization for backward compatibility (curl, Postman)
	if authHeader == "" {
		authHeader = c.GetHeader("Authorization")
	}
	if authHeader == "" {
		RespondUnauthorized(c, "UNAUTHORIZED", "Authorization required")
		c.Abort()
		return
	}

	// Format: Bearer <token>
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		RespondUnauthorized(c, "INVALID_TOKEN", "Invalid authorization format")
		c.Abort()
		return
	}

	tokenString := parts[1]

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		// Check if token is expired
		if strings.Contains(err.Error(), "expired") {
			RespondUnauthorized(c, "TOKEN_EXPIRED", "Token has expired")
		} else {
			RespondUnauthorized(c, "INVALID_TOKEN", "Invalid token")
		}
		c.Abort()
		return
	}

	if !token.Valid {
		RespondUnauthorized(c, "INVALID_TOKEN", "Token is not valid")
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		RespondUnauthorized(c, "INVALID_TOKEN", "Invalid token claims")
		c.Abort()
		return
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		RespondUnauthorized(c, "INVALID_TOKEN", "Invalid user_id in token")
		c.Abort()
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		RespondUnauthorized(c, "INVALID_TOKEN", "Invalid user_id format")
		c.Abort()
		return
	}

		// Сохраняем user_id в контексте
		c.Set("user_id", userID)

		// Добавляем user_id в контекст запроса для логгирования
		ctx := logger.AddUserToContext(c.Request.Context(), userID.String())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// LoggerMiddleware логирует HTTP запросы
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		logger.Info("http_request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}

// CORSMiddleware configures CORS headers
func CORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders []string) gin.HandlerFunc {
	// Check if wildcard is allowed
	allowAll := false
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
		}
		originSet[o] = struct{}{}
	}

	methodsHeader := strings.Join(defaultIfEmpty(allowedMethods, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}), ", ")
	// Include X-Landly-Authorization for Yandex Serverless compatibility
	headersHeader := strings.Join(defaultIfEmpty(allowedHeaders, []string{"Content-Type", "X-Landly-Authorization", "Authorization"}), ", ")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			if allowAll {
				// Wildcard: reflect the requesting origin
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			} else if _, ok := originSet[origin]; ok {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		} else if len(allowedOrigins) > 0 && !allowAll {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
		}

		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", headersHeader)
		c.Writer.Header().Set("Access-Control-Allow-Methods", methodsHeader)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func defaultIfEmpty(values, fallback []string) []string {
	if len(values) == 0 {
		return fallback
	}
	return values
}

// RequestIDMiddleware добавляет request ID для трейсинга
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

// GetUserID извлекает user_id из контекста
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := userID.(uuid.UUID)
	return id, ok
}
