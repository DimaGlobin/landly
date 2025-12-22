package domain

import (
	"time"

	"github.com/google/uuid"
)

// SchemaVersionSource источник версии схемы
type SchemaVersionSource string

const (
	SchemaVersionSourceChat     SchemaVersionSource = "chat"
	SchemaVersionSourceGenerate SchemaVersionSource = "generate"
	SchemaVersionSourceRevert   SchemaVersionSource = "revert"
	SchemaVersionSourceSystem   SchemaVersionSource = "system"
)

// SchemaVersion версия схемы проекта
type SchemaVersion struct {
	ID         uuid.UUID          `db:"id" json:"id"`
	ProjectID  uuid.UUID          `db:"project_id" json:"project_id"`
	SchemaJSON string             `db:"schema_json" json:"schema_json"`
	CreatedAt  time.Time          `db:"created_at" json:"created_at"`
	CreatedBy  uuid.UUID          `db:"created_by" json:"created_by"`
	Source     SchemaVersionSource `db:"source" json:"source"`
	TokensUsed int                `db:"tokens_used" json:"tokens_used"`
}

// NewSchemaVersion создаёт новую версию схемы
func NewSchemaVersion(projectID, userID uuid.UUID, schemaJSON string, source SchemaVersionSource) *SchemaVersion {
	return &SchemaVersion{
		ID:         uuid.New(),
		ProjectID:  projectID,
		SchemaJSON: schemaJSON,
		CreatedAt:  time.Now(),
		CreatedBy:  userID,
		Source:     source,
		TokensUsed: 0,
	}
}

