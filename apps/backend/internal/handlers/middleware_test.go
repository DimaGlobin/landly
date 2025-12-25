package handlers

import (
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
)

func init() {
	gin.SetMode(gin.TestMode)
}

func generateTestToken(userID uuid.UUID, secret string, expiresIn time.Duration) string {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(expiresIn).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	token := generateTestToken(userID, secret, time.Hour)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := AuthMiddleware(secret)
	middleware(c)

	// Check that user_id was set
	storedUserID, exists := c.Get("user_id")
	require.True(t, exists)
	assert.Equal(t, userID, storedUserID)

	// Should not abort
	assert.False(t, c.IsAborted())
}

func TestAuthMiddleware_XLandlyAuthorizationHeader(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	token := generateTestToken(userID, secret, time.Hour)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("X-Landly-Authorization", "Bearer "+token)

	middleware := AuthMiddleware(secret)
	middleware(c)

	// Check that user_id was set
	storedUserID, exists := c.Get("user_id")
	require.True(t, exists)
	assert.Equal(t, userID, storedUserID)
}

func TestAuthMiddleware_NoAuthHeader(t *testing.T) {
	secret := "test-secret"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	middleware := AuthMiddleware(secret)
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp ErrorResponseBody
	_ = parseJSON(w.Body.Bytes(), &errResp)
	assert.Equal(t, "UNAUTHORIZED", errResp.Error.Code)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name       string
		authHeader string
	}{
		{"missing bearer", "token-without-bearer"},
		{"wrong prefix", "Basic token"},
		{"too many parts", "Bearer token extra"},
		{"empty token", "Bearer "},
	}

	secret := "test-secret"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			c.Request.Header.Set("Authorization", tc.authHeader)

			middleware := AuthMiddleware(secret)
			middleware(c)

			assert.True(t, c.IsAborted())
			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var errResp ErrorResponseBody
			_ = parseJSON(w.Body.Bytes(), &errResp)
			assert.Equal(t, "INVALID_TOKEN", errResp.Error.Code)
		})
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	token := generateTestToken(userID, secret, -time.Hour) // Expired 1 hour ago

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := AuthMiddleware(secret)
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp ErrorResponseBody
	_ = parseJSON(w.Body.Bytes(), &errResp)
	assert.Equal(t, "TOKEN_EXPIRED", errResp.Error.Code)
}

func TestAuthMiddleware_InvalidSignature(t *testing.T) {
	userID := uuid.New()
	token := generateTestToken(userID, "secret-1", time.Hour)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := AuthMiddleware("secret-2") // Different secret
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp ErrorResponseBody
	_ = parseJSON(w.Body.Bytes(), &errResp)
	assert.Equal(t, "INVALID_TOKEN", errResp.Error.Code)
}

func TestAuthMiddleware_MalformedToken(t *testing.T) {
	secret := "test-secret"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer not.a.valid.jwt.token")

	middleware := AuthMiddleware(secret)
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	middleware := RequestIDMiddleware()
	middleware(c)

	// Check that request_id was set
	requestID, exists := c.Get("request_id")
	require.True(t, exists)
	assert.NotEmpty(t, requestID)

	// Check header was set
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestRequestIDMiddleware_UsesExistingID(t *testing.T) {
	existingID := "existing-request-id-123"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("X-Request-ID", existingID)

	middleware := RequestIDMiddleware()
	middleware(c)

	// Check that existing ID was used
	requestID, exists := c.Get("request_id")
	require.True(t, exists)
	assert.Equal(t, existingID, requestID)

	// Check header was set with same ID
	assert.Equal(t, existingID, w.Header().Get("X-Request-ID"))
}

func TestCORSMiddleware_AllowAll(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Origin", "https://example.com")

	middleware := CORSMiddleware([]string{"*"}, nil, nil)
	middleware(c)

	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_SpecificOrigin(t *testing.T) {
	allowedOrigins := []string{"https://allowed.com", "https://also-allowed.com"}

	t.Run("allowed origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Request.Header.Set("Origin", "https://allowed.com")

		middleware := CORSMiddleware(allowedOrigins, nil, nil)
		middleware(c)

		assert.Equal(t, "https://allowed.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("not allowed origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Request.Header.Set("Origin", "https://not-allowed.com")

		middleware := CORSMiddleware(allowedOrigins, nil, nil)
		middleware(c)

		// Should not set the origin header for disallowed origins
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodOptions, "/test", nil)
	c.Request.Header.Set("Origin", "https://example.com")

	middleware := CORSMiddleware([]string{"*"}, nil, nil)
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCORSMiddleware_CustomHeaders(t *testing.T) {
	customHeaders := []string{"X-Custom-Header", "X-Another-Header"}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Origin", "https://example.com")

	middleware := CORSMiddleware([]string{"*"}, nil, customHeaders)
	middleware(c)

	headersHeader := w.Header().Get("Access-Control-Allow-Headers")
	assert.Contains(t, headersHeader, "X-Custom-Header")
	assert.Contains(t, headersHeader, "X-Another-Header")
}

func TestGetUserID_Exists(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	userID := uuid.New()
	c.Set("user_id", userID)

	result, ok := GetUserID(c)
	assert.True(t, ok)
	assert.Equal(t, userID, result)
}

func TestGetUserID_NotExists(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	result, ok := GetUserID(c)
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, result)
}

func TestGetUserID_WrongType(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "not-a-uuid")

	result, ok := GetUserID(c)
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, result)
}

// Helper to parse JSON
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

