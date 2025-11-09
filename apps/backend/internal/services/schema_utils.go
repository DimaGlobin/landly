package services

import (
	"fmt"

	"github.com/landly/backend/internal/validation/pageschema"
)

func sanitizeSchema(schemaJSON string) (string, []string, error) {
	result, err := pageschema.Validate([]byte(schemaJSON))
	if err != nil {
		return "", nil, fmt.Errorf("schema validation failed: %w", err)
	}
	return string(result.Normalized), result.AutoFixes, nil
}
