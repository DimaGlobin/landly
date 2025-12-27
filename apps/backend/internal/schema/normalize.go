package schema

import (
	"encoding/json"
	"fmt"
	"strings"
)

// NormalizeSchema нормализует JSON-схему лендинга
// Приводит схему к единому формату для использования в preview и publish
func NormalizeSchema(schemaJSON string) (string, error) {
	// Валидируем схему
	if err := ValidateSchema(schemaJSON); err != nil {
		return "", fmt.Errorf("schema validation failed: %w", err)
	}

	// Парсим JSON
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return "", fmt.Errorf("failed to parse schema: %w", err)
	}

	// Нормализуем страницы
	pages, ok := schema["pages"].([]interface{})
	if !ok {
		return "", fmt.Errorf("invalid pages structure")
	}

	normalizedPages := make([]interface{}, 0, len(pages))
	for _, pageRaw := range pages {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}

		normalizedPage := normalizePage(page)
		normalizedPages = append(normalizedPages, normalizedPage)
	}

	schema["pages"] = normalizedPages

	// Нормализуем тему: если темы нет, добавляем дефолтную тему с дефолтными цветами
	normalizeTheme(schema)

	// Сериализуем обратно в JSON
	normalizedJSON, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("failed to marshal normalized schema: %w", err)
	}

	return string(normalizedJSON), nil
}

// normalizePage нормализует страницу
func normalizePage(page map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})

	// Копируем основные поля
	for k, v := range page {
		normalized[k] = v
	}

	// Нормализуем блоки
	blocksRaw, ok := page["blocks"]
	if !ok {
		return normalized
	}

	blocks, ok := blocksRaw.([]interface{})
	if !ok {
		return normalized
	}

	normalizedBlocks := make([]interface{}, 0, len(blocks))
	for i, blockRaw := range blocks {
		block, ok := blockRaw.(map[string]interface{})
		if !ok {
			continue
		}

		normalizedBlock := normalizeBlock(block, i)
		normalizedBlocks = append(normalizedBlocks, normalizedBlock)
	}

	normalized["blocks"] = normalizedBlocks
	return normalized
}

// normalizeTheme нормализует тему схемы
// Если темы нет или палитры нет, добавляет дефолтную тему с дефолтными цветами
func normalizeTheme(schema map[string]interface{}) {
	// Дефолтная палитра (синхронизирована с фронтендом и рендерером)
	defaultPalette := map[string]interface{}{
		"primary":    "#2563EB",
		"secondary":  "#7C3AED",
		"accent":     "#F97316",
		"background": "#FFFFFF",
		"text":       "#1F2937",
	}

	// Проверяем наличие темы
	themeRaw, hasTheme := schema["theme"]
	if !hasTheme {
		// Тема отсутствует, создаем новую
		schema["theme"] = map[string]interface{}{
			"palette": defaultPalette,
		}
		return
	}

	theme, ok := themeRaw.(map[string]interface{})
	if !ok {
		// Тема не является объектом, заменяем на дефолтную
		schema["theme"] = map[string]interface{}{
			"palette": defaultPalette,
		}
		return
	}

	// Проверяем наличие палитры
	paletteRaw, hasPalette := theme["palette"]
	if !hasPalette {
		// Палитра отсутствует, добавляем дефолтную
		theme["palette"] = defaultPalette
		return
	}

	palette, ok := paletteRaw.(map[string]interface{})
	if !ok {
		// Палитра не является объектом, заменяем на дефолтную
		theme["palette"] = defaultPalette
		return
	}

	// Дополняем палитру дефолтными значениями для отсутствующих цветов
	// ВАЖНО: не перезаписываем существующие валидные цвета!
	for key, defaultValue := range defaultPalette {
		if _, exists := palette[key]; !exists {
			// Цвет отсутствует - добавляем дефолтный
			palette[key] = defaultValue
		} else if val, ok := palette[key].(string); !ok {
			// Значение не строка - заменяем на дефолт
			palette[key] = defaultValue
		} else if val == "" {
			// Значение пустое - заменяем на дефолт
			palette[key] = defaultValue
		} else {
			// Цвет существует и валидный - НЕ ТРОГАЕМ!
			// Просто нормализуем формат (убираем пробелы, приводим к верхнему регистру для консистентности)
			normalizedVal := strings.TrimSpace(val)
			if strings.HasPrefix(normalizedVal, "#") {
				// Если это hex цвет, приводим к верхнему регистру для консистентности
				if len(normalizedVal) == 7 {
					normalizedVal = strings.ToUpper(normalizedVal)
				}
			}
			palette[key] = normalizedVal
		}
	}
}

// normalizeBlock нормализует блок
func normalizeBlock(block map[string]interface{}, defaultOrder int) map[string]interface{} {
	normalized := make(map[string]interface{})

	// Копируем все поля
	for k, v := range block {
		normalized[k] = v
	}

	// Нормализуем order: если есть sort, конвертируем в order
	if orderRaw, ok := normalized["order"]; ok {
		// order уже есть, оставляем как есть
		_ = orderRaw
	} else if sortRaw, ok := normalized["sort"]; ok {
		// Конвертируем sort в order
		normalized["order"] = sortRaw
		delete(normalized, "sort")
	} else {
		// Используем индекс как order по умолчанию
		normalized["order"] = float64(defaultOrder)
	}

	// Убеждаемся, что props существует
	if _, ok := normalized["props"]; !ok {
		normalized["props"] = make(map[string]interface{})
	}

	return normalized
}
