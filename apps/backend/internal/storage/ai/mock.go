package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/landly/backend/internal/logger"
	"go.uber.org/zap"
)

// Client интерфейс для AI клиента
type Client interface {
	GenerateLandingSchema(ctx context.Context, prompt, paymentURL string) (string, error)
}

// MockClient мок-реализация AI клиента для разработки и тестирования
// PLUGGABLE: замените на реальную реализацию (OpenAI, Claude, YandexGPT)
type MockClient struct{}

// NewMockClient создаёт новый мок-клиент
func NewMockClient() *MockClient {
	return &MockClient{}
}

// GenerateLandingSchema генерирует предсказуемую JSON-схему лендинга
func (c *MockClient) GenerateLandingSchema(ctx context.Context, prompt, paymentURL string) (string, error) {
	log := logger.WithContext(ctx)
	log.Info("generating landing schema",
		zap.String("prompt", prompt),
		zap.String("payment_url", paymentURL),
	)

	// Генерируем простую схему на основе промпта
	schema := map[string]interface{}{
		"version": "1.0",
		"pages": []map[string]interface{}{
			{
				"path":        "/",
				"title":       extractTitle(prompt),
				"description": "Сгенерированный лендинг с помощью AI",
				"blocks": []map[string]interface{}{
					{
						"type":  "hero",
						"order": 0,
						"props": map[string]interface{}{
							"headline":    extractTitle(prompt),
							"subheadline": "Превратите свою идею в реальность",
							"ctaText":     "Начать сейчас",
							"image":       "https://images.unsplash.com/photo-1498050108023-c5249f4df085?w=1200",
						},
					},
					{
						"type":  "features",
						"order": 1,
						"props": map[string]interface{}{
							"title": "Наши преимущества",
							"items": []map[string]interface{}{
								{
									"icon":        "⚡",
									"title":       "Быстро",
									"description": "Запустите ваш проект за считанные минуты",
								},
								{
									"icon":        "🎯",
									"title":       "Эффективно",
									"description": "Проверенные решения для вашего бизнеса",
								},
								{
									"icon":        "💡",
									"title":       "Инновационно",
									"description": "Современные технологии и подходы",
								},
							},
						},
					},
					{
						"type":  "pricing",
						"order": 2,
						"props": map[string]interface{}{
							"title": "Тарифы",
							"plans": []map[string]interface{}{
								{
									"name":     "Базовый",
									"price":    "9990",
									"currency": "₽",
									"period":   "месяц",
									"features": []string{
										"Все базовые функции",
										"Email поддержка",
										"1 пользователь",
									},
								},
								{
									"name":     "Премиум",
									"price":    "19990",
									"currency": "₽",
									"period":   "месяц",
									"featured": true,
									"features": []string{
										"Все функции Базового",
										"Приоритетная поддержка",
										"До 10 пользователей",
										"Аналитика",
									},
								},
								{
									"name":     "Бизнес",
									"price":    "49990",
									"currency": "₽",
									"period":   "месяц",
									"features": []string{
										"Все функции Премиум",
										"Персональный менеджер",
										"Безлимитные пользователи",
										"API доступ",
									},
								},
							},
						},
					},
					{
						"type":  "testimonials",
						"order": 3,
						"props": map[string]interface{}{
							"title": "Отзывы клиентов",
							"items": []map[string]interface{}{
								{
									"author": "Иван Петров",
									"role":   "CEO, Startup Inc",
									"text":   "Отличный сервис! Помог нам быстро запустить MVP.",
									"rating": 5,
								},
								{
									"author": "Мария Сидорова",
									"role":   "Маркетолог",
									"text":   "Интуитивно понятный интерфейс и быстрая поддержка.",
									"rating": 5,
								},
							},
						},
					},
					{
						"type":  "faq",
						"order": 4,
						"props": map[string]interface{}{
							"title": "Частые вопросы",
							"items": []map[string]interface{}{
								{
									"question": "Как быстро я могу начать?",
									"answer":   "Регистрация занимает меньше минуты, и вы сразу получаете доступ ко всем функциям.",
								},
								{
									"question": "Можно ли отменить подписку?",
									"answer":   "Да, вы можете отменить подписку в любой момент без дополнительных комиссий.",
								},
								{
									"question": "Какая поддержка доступна?",
									"answer":   "Мы предлагаем email поддержку на всех тарифах и приоритетную поддержку на премиум-тарифах.",
								},
							},
						},
					},
					{
						"type":  "cta",
						"order": 5,
						"props": map[string]interface{}{
							"title":       "Готовы начать?",
							"description": "Присоединяйтесь к тысячам довольных клиентов",
							"buttonText":  "Попробовать бесплатно",
						},
					},
				},
			},
		},
		"theme": map[string]interface{}{
			"palette": map[string]interface{}{
				"primary":    "#2563EB", // Синхронизировано с дефолтными цветами рендерера
				"secondary":  "#7C3AED", // Синхронизировано с дефолтными цветами рендерера
				"accent":     "#F97316", // Синхронизировано с дефолтными цветами рендерера
				"background": "#FFFFFF",
				"text":       "#1F2937",
			},
			"font":         "inter",
			"borderRadius": "lg",
		},
	}

	// Добавляем payment URL если указан
	if paymentURL != "" {
		schema["payment"] = map[string]interface{}{
			"url":        paymentURL,
			"buttonText": "Оплатить",
		}
	}

	// Сериализуем в JSON
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Error("failed to serialize schema", zap.Error(err))
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	log.Info("schema generated successfully",
		zap.Int("schema_length", len(schemaJSON)),
	)
	return string(schemaJSON), nil
}

// extractTitle извлекает заголовок из промпта (упрощённая логика)
func extractTitle(prompt string) string {
	if prompt == "" {
		return "Ваш новый проект"
	}

	// Убираем символы новой строки и заменяем на пробелы
	title := strings.ReplaceAll(prompt, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")

	// Убираем лишние пробелы
	title = strings.TrimSpace(title)

	// Обрезаем до 50 символов с учетом UTF-8
	if len(title) > 50 {
		// Преобразуем в руны для корректной обрезки
		runes := []rune(title)
		if len(runes) > 50 {
			title = string(runes[:50]) + "..."
		}
	}

	return title
}
