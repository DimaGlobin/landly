package schema

import (
	"encoding/json"
	"fmt"

	domain "github.com/landly/backend/internal/models"
)

// AllowedBlockTypes список разрешённых типов блоков
var AllowedBlockTypes = []domain.BlockType{
	domain.BlockTypeHero,
	domain.BlockTypeFeatures,
	domain.BlockTypePricing,
	domain.BlockTypeTestimonials,
	domain.BlockTypeFAQ,
	domain.BlockTypeCTA,
	domain.BlockTypeGallery,
	domain.BlockTypeAbout,
	domain.BlockTypeContact,
}

// ValidationError ошибка валидации схемы
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidateSchema валидирует JSON-схему лендинга
func ValidateSchema(schemaJSON string) error {
	// Парсим JSON
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return &ValidationError{
			Field:   "schema_json",
			Message: fmt.Sprintf("invalid JSON: %v", err),
		}
	}

	// Проверяем наличие pages
	pagesRaw, ok := schema["pages"]
	if !ok {
		return &ValidationError{
			Field:   "pages",
			Message: "required field 'pages' is missing",
		}
	}

	pages, ok := pagesRaw.([]interface{})
	if !ok {
		return &ValidationError{
			Field:   "pages",
			Message: "field 'pages' must be an array",
		}
	}

	if len(pages) == 0 {
		return &ValidationError{
			Field:   "pages",
			Message: "at least one page is required",
		}
	}

	// Валидируем каждую страницу
	for i, pageRaw := range pages {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			return &ValidationError{
				Field:   fmt.Sprintf("pages[%d]", i),
				Message: "page must be an object",
			}
		}

		// Проверяем обязательные поля страницы
		if _, ok := page["path"]; !ok {
			return &ValidationError{
				Field:   fmt.Sprintf("pages[%d].path", i),
				Message: "required field 'path' is missing",
			}
		}

		if _, ok := page["title"]; !ok {
			return &ValidationError{
				Field:   fmt.Sprintf("pages[%d].title", i),
				Message: "required field 'title' is missing",
			}
		}

		// Валидируем блоки
		blocksRaw, ok := page["blocks"]
		if !ok {
			return &ValidationError{
				Field:   fmt.Sprintf("pages[%d].blocks", i),
				Message: "required field 'blocks' is missing",
			}
		}

		blocks, ok := blocksRaw.([]interface{})
		if !ok {
			return &ValidationError{
				Field:   fmt.Sprintf("pages[%d].blocks", i),
				Message: "field 'blocks' must be an array",
			}
		}

		// Валидируем каждый блок
		for j, blockRaw := range blocks {
			block, ok := blockRaw.(map[string]interface{})
			if !ok {
				return &ValidationError{
					Field:   fmt.Sprintf("pages[%d].blocks[%d]", i, j),
					Message: "block must be an object",
				}
			}

			// Проверяем type
			typeRaw, ok := block["type"]
			if !ok {
				return &ValidationError{
					Field:   fmt.Sprintf("pages[%d].blocks[%d].type", i, j),
					Message: "required field 'type' is missing",
				}
			}

			blockType, ok := typeRaw.(string)
			if !ok {
				return &ValidationError{
					Field:   fmt.Sprintf("pages[%d].blocks[%d].type", i, j),
					Message: "field 'type' must be a string",
				}
			}

			// Проверяем, что тип разрешён
			isAllowed := false
			for _, allowedType := range AllowedBlockTypes {
				if string(allowedType) == blockType {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				return &ValidationError{
					Field:   fmt.Sprintf("pages[%d].blocks[%d].type", i, j),
					Message: fmt.Sprintf("block type '%s' is not allowed", blockType),
				}
			}

			// Проверяем order или sort
			hasOrder := false
			if orderRaw, ok := block["order"]; ok {
				if _, ok := orderRaw.(float64); ok {
					hasOrder = true
				}
			}
			if sortRaw, ok := block["sort"]; ok {
				if _, ok := sortRaw.(float64); ok {
					hasOrder = true
				}
			}
			if !hasOrder {
				return &ValidationError{
					Field:   fmt.Sprintf("pages[%d].blocks[%d]", i, j),
					Message: "block must have 'order' or 'sort' field",
				}
			}

			// Проверяем props
			if _, ok := block["props"]; !ok {
				return &ValidationError{
					Field:   fmt.Sprintf("pages[%d].blocks[%d].props", i, j),
					Message: "required field 'props' is missing",
				}
			}

			if _, ok := block["props"].(map[string]interface{}); !ok {
				return &ValidationError{
					Field:   fmt.Sprintf("pages[%d].blocks[%d].props", i, j),
					Message: "field 'props' must be an object",
				}
			}

			// Проверяем наличие id (опционально, но если есть - должен быть строкой)
			if idRaw, ok := block["id"]; ok {
				if _, ok := idRaw.(string); !ok {
					return &ValidationError{
						Field:   fmt.Sprintf("pages[%d].blocks[%d].id", i, j),
						Message: "field 'id' must be a string if present",
					}
				}
			}
		}
	}

	return nil
}
