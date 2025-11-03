package repositories

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/query"
)

// GenerationMessageRepository интерфейс для работы с сообщениями генерации
type GenerationMessageRepository interface {
	Create(ctx context.Context, message *domain.GenerationMessage) error
	ListBySession(ctx context.Context, sessionID string) ([]*domain.GenerationMessage, error)
	DeleteBySession(ctx context.Context, sessionID string) error
}

type generationMessageRepository struct {
	qb *query.Builder
}

// NewGenerationMessageRepository создаёт репозиторий сообщений генерации
func NewGenerationMessageRepository(qb *query.Builder) GenerationMessageRepository {
	return &generationMessageRepository{qb: qb}
}

func (r *generationMessageRepository) Create(ctx context.Context, message *domain.GenerationMessage) error {
	query := r.qb.Insert("generation_messages").
		Columns("id", "session_id", "role", "content", "metadata", "tokens_used", "created_at").
		Values(message.ID, message.SessionID, message.Role, message.Content, message.Metadata, message.TokensUsed, message.CreatedAt)

	_, err := r.qb.Execute(query)
	return err
}

func (r *generationMessageRepository) ListBySession(ctx context.Context, sessionID string) ([]*domain.GenerationMessage, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid session ID format")
	}

	query := r.qb.Select("id", "session_id", "role", "content", "metadata", "tokens_used", "created_at").
		From("generation_messages").
		Where(squirrel.Eq{"session_id": sessionUUID}).
		OrderBy("created_at ASC", "id ASC")

	rows, err := r.qb.Query(query)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}
	defer rows.Close()

	var messages []*domain.GenerationMessage
	for rows.Next() {
		var msg domain.GenerationMessage
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.Metadata, &msg.TokensUsed, &msg.CreatedAt); err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

func (r *generationMessageRepository) DeleteBySession(ctx context.Context, sessionID string) error {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return domain.ErrBadRequest.WithMessage("invalid session ID format")
	}

	query := r.qb.Delete("generation_messages").
		Where(squirrel.Eq{"session_id": sessionUUID})

	_, err = r.qb.Execute(query)
	return err
}

// Ensure interface compliance at compile time
var _ GenerationMessageRepository = (*generationMessageRepository)(nil)
