package ai

import (
	"fmt"
	"strings"
	"time"

	"github.com/landly/backend/config"
	"go.uber.org/zap"
)

const (
	openAIBaseURL    = "https://api.openai.com/v1"
	defaultReferer   = "https://github.com/landly/landly"
	defaultClientTag = "Landly"
)

// NewProviderClient создает AI-клиента в зависимости от конфигурации.
func NewProviderClient(cfg config.AIConfig, logger *zap.Logger) (Client, error) {
	switch strings.ToLower(cfg.Provider) {
	case "mock", "":
		if logger != nil {
			logger.Info("using mock AI client")
		}
		return NewMockClient(), nil
	case "openai":
		return newChatCompletionClient(chatClientOptions{
			Provider:       "openai",
			BaseURL:        openAIBaseURL,
			APIKey:         cfg.OpenAI.APIKey,
			Model:          cfg.OpenAI.Model,
			MaxTokens:      cfg.OpenAI.MaxTokens,
			Temperature:    cfg.OpenAI.Temperature,
			ResponseFormat: cfg.ResponseFormat,
			MaxRetries:     cfg.MaxRetries,
			Timeout:        cfg.Timeout,
			Logger:         logger,
		})
	case "openrouter":
		baseURL := cfg.OpenRouter.BaseURL
		if baseURL == "" {
			baseURL = "https://openrouter.ai/api/v1"
		}
		return newChatCompletionClient(chatClientOptions{
			Provider:       "openrouter",
			BaseURL:        baseURL,
			APIKey:         cfg.OpenRouter.APIKey,
			Model:          cfg.OpenRouter.Model,
			MaxTokens:      cfg.OpenRouter.MaxTokens,
			Temperature:    cfg.OpenRouter.Temperature,
			ResponseFormat: cfg.ResponseFormat,
			MaxRetries:     cfg.MaxRetries,
			Timeout:        cfg.Timeout,
			Logger:         logger,
			ExtraHeaders: map[string]string{
				"HTTP-Referer": defaultReferer,
				"X-Title":      defaultClientTag,
			},
		})
	default:
		return nil, fmt.Errorf("AI provider %q is not implemented yet", cfg.Provider)
	}
}

type chatClientOptions struct {
	Provider       string
	BaseURL        string
	APIKey         string
	Model          string
	MaxTokens      int
	Temperature    float64
	ResponseFormat string
	MaxRetries     int
	Timeout        time.Duration
	Logger         *zap.Logger
	ExtraHeaders   map[string]string
}
