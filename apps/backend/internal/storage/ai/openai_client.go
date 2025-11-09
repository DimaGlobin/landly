package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const defaultSystemPrompt = `You are Landly AI. Generate a concise JSON object describing a marketing landing page structure.
The JSON MUST be valid and parsable. Include hero, features, CTA and footer blocks. Avoid HTML tags; provide plain text only.`

type chatCompletionClient struct {
	httpClient     *http.Client
	baseURL        string
	apiKey         string
	model          string
	maxTokens      int
	temperature    float64
	responseFormat string
	maxRetries     int
	logger         *zap.Logger
	extraHeaders   map[string]string
	provider       string
}

func newChatCompletionClient(opts chatClientOptions) (Client, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("%s api key is empty", opts.Provider)
	}
	if opts.Model == "" {
		return nil, fmt.Errorf("%s model is empty", opts.Provider)
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 45 * time.Second
	}

	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1200
	}

	temperature := opts.Temperature
	if temperature <= 0 {
		temperature = 0.6
	}

	client := &chatCompletionClient{
		httpClient:     &http.Client{Timeout: timeout},
		baseURL:        strings.TrimRight(opts.BaseURL, "/"),
		apiKey:         opts.APIKey,
		model:          opts.Model,
		maxTokens:      maxTokens,
		temperature:    temperature,
		responseFormat: opts.ResponseFormat,
		maxRetries:     opts.MaxRetries,
		logger:         opts.Logger,
		extraHeaders:   opts.ExtraHeaders,
		provider:       opts.Provider,
	}

	return client, nil
}

// GenerateLandingSchema вызывает OpenAI-совместимый API и возвращает JSON схему.
func (c *chatCompletionClient) GenerateLandingSchema(ctx context.Context, prompt, paymentURL string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	request := chatCompletionRequest{
		Model:       c.model,
		Messages:    c.buildMessages(prompt, paymentURL),
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
	}

	if rf := c.buildResponseFormat(); rf != nil {
		request.ResponseFormat = rf
	}

	var lastErr error
	attempts := c.maxRetries + 1
	for attempt := 1; attempt <= attempts; attempt++ {
		schema, usage, err := c.sendRequest(ctx, request)
		if err == nil {
			if c.logger != nil {
				c.logger.Info("ai completion success",
					zap.String("provider", c.provider),
					zap.Int("attempt", attempt),
					zap.Int("prompt_tokens", usage.PromptTokens),
					zap.Int("completion_tokens", usage.CompletionTokens),
					zap.Int("total_tokens", usage.TotalTokens),
				)
			}
			return schema, nil
		}

		lastErr = err
		if c.logger != nil {
			c.logger.Warn("ai completion failed",
				zap.String("provider", c.provider),
				zap.Int("attempt", attempt),
				zap.Error(err),
			)
		}

		// Грейсфул backoff
		if attempt < attempts {
			wait := time.Duration(attempt) * 500 * time.Millisecond
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}

	return "", fmt.Errorf("AI generation failed after %d attempts: %w", attempts, lastErr)
}

func (c *chatCompletionClient) sendRequest(ctx context.Context, payload chatCompletionRequest) (string, usageMetrics, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", usageMetrics{}, fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", usageMetrics{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	for k, v := range c.extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", usageMetrics{}, fmt.Errorf("failed to call provider: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", usageMetrics{}, fmt.Errorf("failed to read provider response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", usageMetrics{}, fmt.Errorf("provider returned %d: %s", resp.StatusCode, truncate(string(respBody), 512))
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(respBody, &completion); err != nil {
		return "", usageMetrics{}, fmt.Errorf("failed to decode provider response: %w", err)
	}

	if completion.Error != nil {
		return "", usageMetrics{}, fmt.Errorf("provider error (%s): %s", completion.Error.Type, completion.Error.Message)
	}

	if len(completion.Choices) == 0 {
		return "", usageMetrics{}, fmt.Errorf("provider returned no choices")
	}

	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	if content == "" {
		return "", usageMetrics{}, fmt.Errorf("provider returned empty content")
	}

	return content, completion.Usage, nil
}

func (c *chatCompletionClient) buildMessages(prompt, paymentURL string) []chatMessage {
	userPrompt := strings.Builder{}
	userPrompt.WriteString("User prompt:\n")
	userPrompt.WriteString(prompt)
	if paymentURL != "" {
		userPrompt.WriteString("\nPayment URL: ")
		userPrompt.WriteString(paymentURL)
	}
	userPrompt.WriteString("\nReturn only JSON.")

	return []chatMessage{
		{Role: "system", Content: defaultSystemPrompt},
		{Role: "user", Content: userPrompt.String()},
	}
}

func (c *chatCompletionClient) buildResponseFormat() any {
	format := strings.ToLower(c.responseFormat)
	switch format {
	case "json_schema", "json_object":
		return map[string]any{
			"type": "json_object",
		}
	default:
		return nil
	}
}

type chatCompletionRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	MaxTokens      int           `json:"max_tokens,omitempty"`
	Temperature    float64       `json:"temperature,omitempty"`
	ResponseFormat any           `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Usage usageMetrics `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type usageMetrics struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
