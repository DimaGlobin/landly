package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/landly/backend/config"
	"github.com/landly/backend/internal/database/postgres"
	"github.com/landly/backend/internal/handlers"
	"github.com/landly/backend/internal/logger"
	"github.com/landly/backend/internal/repositories"
	"github.com/landly/backend/internal/services"
	"github.com/landly/backend/internal/storage/ai"
	"github.com/landly/backend/internal/storage/render"
	"github.com/landly/backend/internal/storage/s3"
	"go.uber.org/zap"
)

func main() {
	// Load configuration first (needed for logger config)
	cfg, err := config.Load()
	if err != nil {
		// Fallback logger for config errors
		logger.Init()
		logger.Get().Fatal("failed to load config", zap.Error(err))
	}

	// Initialize logger with service metadata
	logger.SetConfig(logger.Config{
		Service: cfg.App.Name,
		Version: cfg.App.Version,
		Env:     cfg.App.Env,
	})
	logger.Init()
	log := logger.Get()
	defer func() {
		if err := log.Sync(); err != nil {
			logger.Warn("failed to sync logger", zap.Error(err))
		}
	}()

	log.Info("starting application",
		zap.String("addr", cfg.Server.HTTP.Addr),
		zap.Bool("bootstrap_mode", config.IsBootstrapMode()),
	)

	// Bootstrap mode: start health-only server without DB/S3
	// This allows Yandex Serverless Container to start and respond to health checks
	// even when required env vars are missing
	if config.IsBootstrapMode() {
		log.Warn("BOOTSTRAP MODE: starting health-only server without DB/S3 connections")
		startBootstrapServer(cfg, log)
		return
	}

	// Full initialization (production mode)
	startFullServer(cfg, log)
}

// startBootstrapServer starts a minimal server with only health endpoints.
// Used for Yandex Serverless Container cold start and health checks.
func startBootstrapServer(cfg *config.Config, log logger.Logger) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	// Health endpoints only
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "bootstrap",
			"mode":    "health-only",
			"message": "Server running in bootstrap mode. Set proper env vars and restart for full functionality.",
		})
	})
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "bootstrap"})
	})
	engine.GET("/readyz", func(c *gin.Context) {
		// In bootstrap mode, we're not ready for real traffic
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "bootstrap mode - missing required configuration",
		})
	})

	// All other routes return 503
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "service unavailable",
			"message": "Server running in bootstrap mode. Configure required env vars.",
		})
	})

	srv := &http.Server{
		Addr:         cfg.Server.HTTP.Addr,
		Handler:      engine,
		ReadTimeout:  cfg.Server.HTTP.ReadTimeout,
		WriteTimeout: cfg.Server.HTTP.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Info("bootstrap server starting", zap.String("addr", cfg.Server.HTTP.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start bootstrap server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down bootstrap server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("bootstrap server forced to shutdown", zap.Error(err))
	}
	log.Info("bootstrap server stopped")
}

// startFullServer starts the complete application with all services.
func startFullServer(cfg *config.Config, log logger.Logger) {
	// Database connection
	qb, err := postgres.NewConnection(postgres.Config{
		Host:            cfg.Database.Postgres.Host,
		Port:            cfg.Database.Postgres.Port,
		User:            cfg.Database.Postgres.User,
		Password:        cfg.Database.Postgres.Password,
		DBName:          cfg.Database.Postgres.DBName,
		SSLMode:         cfg.Database.Postgres.SSLMode,
		MaxOpenConns:    cfg.Database.Postgres.MaxOpenConns,
		MaxIdleConns:    cfg.Database.Postgres.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.Postgres.ConnMaxLifetime,
	})
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	log.Info("database connected")

	// Repositories
	userRepo := repositories.NewUserRepository(qb)
	projectRepo := repositories.NewProjectRepository(qb)
	analyticsRepo := repositories.NewAnalyticsRepository(qb)
	integrationRepo := repositories.NewIntegrationRepository(qb)
	publishTargetRepo := repositories.NewPublishTargetRepository(qb)
	publishJobRepo := repositories.NewPublishJobRepository(qb)
	sessionRepo := repositories.NewGenerationSessionRepository(qb)
	messageRepo := repositories.NewGenerationMessageRepository(qb)
	schemaVersionRepo := repositories.NewSchemaVersionRepository(qb)

	// S3 client
	s3Client, err := s3.NewClient(s3.Config{
		Endpoint:        cfg.Storage.S3.Endpoint,
		AccessKeyID:     cfg.Storage.S3.AccessKey,
		SecretAccessKey: cfg.Storage.S3.SecretKey,
		UseSSL:          cfg.Storage.S3.UseSSL,
		BucketName:      cfg.Storage.S3.Bucket,
		CDNBase:         cfg.Storage.CDN.BaseURL,
	})
	if err != nil {
		log.Fatal("failed to create s3 client", zap.Error(err))
	}
	log.Info("s3 client initialized")

	// AI client
	var aiClient ai.Client
	if cfg.AI.Provider == "mock" {
		aiClient = ai.NewMockClient()
		log.Info("using mock AI client")
	} else {
		log.Fatal("AI provider not implemented", zap.String("provider", cfg.AI.Provider))
	}

	// Renderer
	renderer := render.NewStaticRenderer(cfg.Render.TmpDir)

	// Services
	authService := services.NewAuthService(userRepo, cfg.Auth.JWT.Secret, cfg.Auth.JWT.AccessTokenTTL, cfg.Auth.JWT.RefreshTokenTTL)
	userService := services.NewUserService(userRepo)
	projectService := services.NewProjectService(projectRepo, publishTargetRepo, s3Client)
	generateService := services.NewGenerateService(projectRepo, integrationRepo, sessionRepo, messageRepo, schemaVersionRepo, aiClient)
	publicBaseProvider := services.NewStaticPublicBaseProvider(cfg.App.BaseURL)
	publishService := services.NewPublishService(projectRepo, publishTargetRepo, publishJobRepo, userRepo, renderer, s3Client, publicBaseProvider)
	analyticsService := services.NewAnalyticsService(projectRepo, analyticsRepo)
	schemaVersionService := services.NewSchemaVersionService(schemaVersionRepo, projectRepo)
	simpleGenerateService := services.NewSimpleGenerateService(projectRepo, schemaVersionRepo, aiClient)

	// HTTP handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	projectHandler := handlers.NewProjectHandler(projectService, publishTargetRepo, publicBaseProvider)
	generateHandler := handlers.NewGenerateHandler(generateService, publishService, publicBaseProvider)
	simpleGenerateHandler := handlers.NewSimpleGenerateHandler(simpleGenerateService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	schemaHandler := handlers.NewSchemaHandler(schemaVersionService)

	// Router with DB for readiness checks
	router := handlers.NewRouter(
		authHandler,
		projectHandler,
		generateHandler,
		simpleGenerateHandler,
		analyticsHandler,
		schemaHandler,
		userHandler,
		cfg.Auth.JWT.Secret,
		cfg.Server.CORS.AllowedOrigins,
		cfg.Server.CORS.AllowedMethods,
		cfg.Server.CORS.AllowedHeaders,
		logger.GetZapLogger(),
	)

	// Pass DB to router for readiness checks
	router.SetDB(qb.GetDB())

	engine := router.Setup()

	// HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.HTTP.Addr,
		Handler:      engine,
		ReadTimeout:  cfg.Server.HTTP.ReadTimeout,
		WriteTimeout: cfg.Server.HTTP.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("http server starting", zap.String("addr", cfg.Server.HTTP.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start http server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", zap.Error(err))
	}

	log.Info("server stopped gracefully")
}
