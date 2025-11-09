package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// Инициализируем логгер
	logger.Init()
	log := logger.Get()
	defer func() {
		if err := log.Sync(); err != nil {
			logger.Warn("failed to sync logger", zap.Error(err))
		}
	}()

	// Конфигурация
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	log.Info("starting application",
		zap.String("env", cfg.App.Env),
		zap.String("name", cfg.App.Name),
		zap.String("version", cfg.App.Version),
		zap.String("addr", cfg.Server.HTTP.Addr),
	)

	// База данных (Query Builder)
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

	log.Info("✅ Database connected with Query Builder")

	// Репозитории (Query Builder)
	userRepo := repositories.NewUserRepository(qb)
	projectRepo := repositories.NewProjectRepository(qb)
	analyticsRepo := repositories.NewAnalyticsRepository(qb)
	integrationRepo := repositories.NewIntegrationRepository(qb)
	publishTargetRepo := repositories.NewPublishTargetRepository(qb)
	sessionRepo := repositories.NewGenerationSessionRepository(qb)
	messageRepo := repositories.NewGenerationMessageRepository(qb)

	// S3 клиент
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

	// AI клиент
	aiClient, err := ai.NewProviderClient(cfg.AI, log.GetZapLogger())
	if err != nil {
		log.Fatal("failed to init AI client", zap.Error(err), zap.String("provider", cfg.AI.Provider))
	}

	// Рендерер
	renderer := render.NewStaticRenderer(cfg.Render.TmpDir)

	// Сервисы
	authService := services.NewAuthService(userRepo, cfg.Auth.JWT.Secret, cfg.Auth.JWT.AccessTokenTTL, cfg.Auth.JWT.RefreshTokenTTL)
	projectService := services.NewProjectService(projectRepo)
	generateService := services.NewGenerateService(projectRepo, integrationRepo, sessionRepo, messageRepo, aiClient)
	publishService := services.NewPublishService(projectRepo, publishTargetRepo, userRepo, renderer, s3Client, cfg.App.BaseURL)
	analyticsService := services.NewAnalyticsService(projectRepo, analyticsRepo)

	// HTTP handlers
	authHandler := handlers.NewAuthHandler(authService)
	projectHandler := handlers.NewProjectHandler(projectService, publishTargetRepo, cfg.App.BaseURL)
	generateHandler := handlers.NewGenerateHandler(generateService, publishService, cfg.App.BaseURL)
	simpleGenerateService := services.NewSimpleGenerateService(projectRepo, aiClient)
	simpleGenerateHandler := handlers.NewSimpleGenerateHandler(simpleGenerateService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	// Router
	router := handlers.NewRouter(
		authHandler,
		projectHandler,
		generateHandler,
		simpleGenerateHandler,
		analyticsHandler,
		cfg.Auth.JWT.Secret,
		cfg.Server.CORS.AllowedOrigins,
		cfg.Server.CORS.AllowedMethods,
		cfg.Server.CORS.AllowedHeaders,
		logger.GetZapLogger(),
	)

	engine := router.Setup()

	// HTTP сервер
	srv := &http.Server{
		Addr:         cfg.Server.HTTP.Addr,
		Handler:      engine,
		ReadTimeout:  cfg.Server.HTTP.ReadTimeout,
		WriteTimeout: cfg.Server.HTTP.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера в горутине
	go func() {
		logger.WithContext(context.Background()).Info("http server starting", zap.String("addr", cfg.Server.HTTP.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithContext(context.Background()).Fatal("failed to start http server", zap.Error(err))
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
