package server

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/landly/backend/config"
	"github.com/landly/backend/internal/database/postgres"
	"github.com/landly/backend/internal/handlers"
	"github.com/landly/backend/internal/repositories"
	"github.com/landly/backend/internal/services"
	"github.com/landly/backend/internal/storage/ai"
	"github.com/landly/backend/internal/storage/render"
	"github.com/landly/backend/internal/storage/s3"
	"go.uber.org/zap"
)

// Server представляет HTTP сервер приложения
type Server struct {
	engine *gin.Engine
	config *config.Config
	logger *zap.Logger
}

// NewServer создает новый сервер с инициализированными зависимостями
func NewServer(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Подключение к базе данных (Query Builder)
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
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Репозитории (Query Builder)
	userRepo := repositories.NewUserRepository(qb)
	projectRepo := repositories.NewProjectRepository(qb)
	analyticsRepo := repositories.NewAnalyticsRepository(qb)
	integrationRepo := repositories.NewIntegrationRepository(qb)
	publishTargetRepo := repositories.NewPublishTargetRepository(qb)
	sessionRepo := repositories.NewGenerationSessionRepository(qb)
	messageRepo := repositories.NewGenerationMessageRepository(qb)
	brandRepo := repositories.NewBrandProfileRepository(qb)
	productProfileRepo := repositories.NewProductProfileRepository(qb)
	contentSnippetRepo := repositories.NewContentSnippetRepository(qb)

	// S3 клиент
	s3Client, err := s3.NewClient(s3.Config{
		Endpoint:        cfg.Storage.S3.Endpoint,
		AccessKeyID:     cfg.Storage.S3.AccessKey,
		SecretAccessKey: cfg.Storage.S3.SecretKey,
		BucketName:      cfg.Storage.S3.Bucket,
		UseSSL:          cfg.Storage.S3.UseSSL,
		CDNBase:         cfg.Storage.CDN.BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client: %w", err)
	}

	// AI клиент
	aiClient, err := ai.NewProviderClient(cfg.AI, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init AI client: %w", err)
	}

	// Renderer
	renderer := render.NewStaticRenderer(cfg.Render.TmpDir)

	// Services
	authService := services.NewAuthService(userRepo, cfg.Auth.JWT.Secret, cfg.Auth.JWT.AccessTokenTTL, cfg.Auth.JWT.RefreshTokenTTL)
	projectService := services.NewProjectService(projectRepo)
	generateService := services.NewGenerateService(projectRepo, integrationRepo, sessionRepo, messageRepo, aiClient)
	publishService := services.NewPublishService(projectRepo, publishTargetRepo, userRepo, renderer, s3Client, cfg.App.BaseURL)
	simpleGenerateService := services.NewSimpleGenerateService(projectRepo, aiClient)
	analyticsService := services.NewAnalyticsService(projectRepo, analyticsRepo)
	cceService := services.NewCCEService(projectRepo, brandRepo, productProfileRepo, contentSnippetRepo)

	// HTTP handlers
	authHandler := handlers.NewAuthHandler(authService)
	projectHandler := handlers.NewProjectHandler(projectService, publishTargetRepo, cfg.App.BaseURL)
	generateHandler := handlers.NewGenerateHandler(generateService, publishService, cfg.App.BaseURL)
	simpleGenerateHandler := handlers.NewSimpleGenerateHandler(simpleGenerateService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	cceHandler := handlers.NewCCEHandler(cceService)

	// Router
	router := handlers.NewRouter(
		authHandler,
		projectHandler,
		generateHandler,
		simpleGenerateHandler,
		analyticsHandler,
		cceHandler,
		cfg.Auth.JWT.Secret,
		cfg.Server.CORS.AllowedOrigins,
		cfg.Server.CORS.AllowedMethods,
		cfg.Server.CORS.AllowedHeaders,
		logger,
	)

	engine := router.Setup()

	return &Server{
		engine: engine,
		config: cfg,
		logger: logger,
	}, nil
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	addr := s.config.Server.HTTP.Addr
	s.logger.Info("starting HTTP server", zap.String("addr", addr))

	return s.engine.Run(addr)
}

// Shutdown корректно останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	// Здесь можно добавить graceful shutdown логику
	return nil
}
