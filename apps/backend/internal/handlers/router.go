package handlers

import (
	"net/http"

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
	jwtSecret             string
	allowedOrigins        []string
	allowedMethods        []string
	allowedHeaders        []string
	logger                *zap.Logger
}

func NewRouter(
	authHandler *AuthHandler,
	projectHandler *ProjectHandler,
	generateHandler *GenerateHandler,
	simpleGenerateHandler *SimpleGenerateHandler,
	analyticsHandler *AnalyticsHandler,
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
		jwtSecret:             jwtSecret,
		allowedOrigins:        allowedOrigins,
		allowedMethods:        allowedMethods,
		allowedHeaders:        allowedHeaders,
		logger:                logger,
	}
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
		// Auth (публичные)
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", r.authHandler.SignUp)
			auth.POST("/login", r.authHandler.SignIn)
			auth.POST("/refresh", r.authHandler.RefreshToken)
		}

		// Projects (требуют авторизацию)
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
		}

		// Analytics
		analytics := v1.Group("/analytics")
		{
			// Публичный эндпойнт для трекинга (с опубликованных сайтов)
			analytics.POST("/:id/event", r.analyticsHandler.TrackEvent)

			// Приватный эндпойнт для получения статистики
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

func (r *Router) readinessCheck(c *gin.Context) {
	// TODO: проверить подключение к БД, Redis и т.д.
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
