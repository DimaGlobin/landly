package main

import (
	"log"

	"github.com/landly/backend/config"
	"go.uber.org/zap"
)

func main() {
	// Логгер
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("failed to sync logger: %v", err)
		}
	}()

	// Конфигурация
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	logger.Info("starting worker",
		zap.String("env", cfg.App.Env),
	)

	// TODO: Инициализация Asynq worker для фоновых задач
	// - GENERATE: AI генерация лендингов
	// - RENDER: Рендеринг статических сайтов
	// - PUBLISH: Публикация в S3

	logger.Info("worker started (stub implementation)")

	// Блокируем выполнение
	select {}
}
