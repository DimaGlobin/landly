package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware returns a Gin middleware that records HTTP metrics
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint to avoid recursion
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		HTTPRequestsInFlight.Inc()
		defer HTTPRequestsInFlight.Dec()

		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// Normalize path to avoid high cardinality
		path := normalizePath(c)

		RecordHTTPRequest(c.Request.Method, path, status, duration)
	}
}

// normalizePath normalizes the request path to avoid high cardinality metrics
// Replaces dynamic segments like UUIDs with placeholders
func normalizePath(c *gin.Context) string {
	// Use the matched route pattern if available (Gin provides this)
	if c.FullPath() != "" {
		return c.FullPath()
	}
	// Fallback to actual path for unmatched routes
	return c.Request.URL.Path
}

