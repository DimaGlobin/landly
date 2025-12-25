package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/landly/backend/internal/logger"
	"go.uber.org/zap"
)

type Router struct {
	engine                *gin.Engine
	authHandler           *AuthHandler
	projectHandler        *ProjectHandler
	generateHandler       *GenerateHandler
	simpleGenerateHandler *SimpleGenerateHandler
	analyticsHandler      *AnalyticsHandler
	schemaHandler         *SchemaHandler
	jwtSecret             string
	allowedOrigins        []string
	allowedMethods        []string
	allowedHeaders        []string
	logger                *zap.Logger
	db                    *sql.DB // For readiness checks
}

func NewRouter(
	authHandler *AuthHandler,
	projectHandler *ProjectHandler,
	generateHandler *GenerateHandler,
	simpleGenerateHandler *SimpleGenerateHandler,
	analyticsHandler *AnalyticsHandler,
	schemaHandler *SchemaHandler,
	jwtSecret string,
	allowedOrigins []string,
	allowedMethods []string,
	allowedHeaders []string,
	logger *zap.Logger,
) *Router {
	return &Router{
		engine:                gin.Default(),
		authHandler:           authHandler,
		projectHandler:        projectHandler,
		generateHandler:       generateHandler,
		simpleGenerateHandler: simpleGenerateHandler,
		analyticsHandler:      analyticsHandler,
		schemaHandler:         schemaHandler,
		jwtSecret:             jwtSecret,
		allowedOrigins:        allowedOrigins,
		allowedMethods:        allowedMethods,
		allowedHeaders:        allowedHeaders,
		logger:                logger,
	}
}

// SetDB sets the database connection for readiness checks.
// Must be called before Setup() for readiness endpoint to work properly.
func (r *Router) SetDB(db *sql.DB) {
	r.db = db
}

func (r *Router) Setup() *gin.Engine {
	// Middleware
	r.engine.Use(CORSMiddleware(r.allowedOrigins, r.allowedMethods, r.allowedHeaders))
	r.engine.Use(logger.TraceMiddleware())
	r.engine.Use(logger.LoggingMiddleware())
	r.engine.Use(RequestIDMiddleware())
	r.engine.Use(LoggerMiddleware(r.logger))

	// Health checks
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/healthz", r.healthCheck)
	r.engine.GET("/readyz", r.readinessCheck)

	// Published static sites (public)
	r.engine.GET("/sites/:slug", r.generateHandler.ServePublished)
	r.engine.GET("/sites/:slug/*path", r.generateHandler.ServePublished)
	r.engine.GET("/:slug", r.generateHandler.ServePublishedLegacy)

	// API v1
	v1 := r.engine.Group("/v1")
	{
		// Auth (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", r.authHandler.SignUp)
			auth.POST("/login", r.authHandler.SignIn)
			auth.POST("/refresh", r.authHandler.RefreshToken)
		}

		// Projects (require auth)
		projects := v1.Group("/projects")
		projects.Use(AuthMiddleware(r.jwtSecret))
		{
			projects.POST("", r.projectHandler.CreateProject)
			projects.GET("", r.projectHandler.GetProjects)
			projects.GET("/:id", r.projectHandler.GetProject)
			projects.DELETE("/:id", r.projectHandler.DeleteProject)

			// Generate & Publish
			projects.POST("/:id/generate", r.generateHandler.Generate)
			projects.POST("/:id/generate-simple", r.simpleGenerateHandler.GenerateSimple)
			projects.GET("/:id/preview", r.generateHandler.GetPreview)
			projects.GET("/:id/chat", r.generateHandler.GetChat)
			projects.POST("/:id/chat", r.generateHandler.SendChat)
			projects.POST("/:id/publish", r.generateHandler.Publish)
			projects.DELETE("/:id/publish", r.generateHandler.Unpublish)

			// Schema versions
			projects.GET("/:id/schema/versions", r.schemaHandler.GetSchemaVersions)
			projects.POST("/:id/schema/revert", r.schemaHandler.RevertSchema)
		}

		// Analytics
		analytics := v1.Group("/analytics")
		{
			// Public endpoint for tracking (from published sites)
			analytics.POST("/:id/event", r.analyticsHandler.TrackEvent)

			// Private endpoint for stats
			analytics.GET("/:id/stats", AuthMiddleware(r.jwtSecret), r.analyticsHandler.GetStats)
		}
	}

	return r.engine
}

func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// readinessCheck verifies the service is ready to handle traffic.
// Returns 200 if DB is reachable, 503 otherwise.
func (r *Router) readinessCheck(c *gin.Context) {
	// If no DB connection provided, return degraded status
	if r.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "database connection not configured",
		})
		return
	}

	// Ping database with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := r.db.PingContext(ctx); err != nil {
		r.logger.Warn("readiness check failed: database ping error", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "database unreachable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
