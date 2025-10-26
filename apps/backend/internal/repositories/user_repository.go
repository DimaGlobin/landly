package repositories

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/query"
)

// UserRepository интерфейс репозитория пользователей
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

// userRepository реализация репозитория пользователей
type userRepository struct {
	qb *query.Builder
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(qb *query.Builder) UserRepository {
	return &userRepository{qb: qb}
}

// Create создает пользователя
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := r.qb.Insert("users").
		Columns("id", "email", "password_hash", "created_at", "updated_at").
		Values(user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)

	_, err := r.qb.Execute(query)
	return err
}

// GetByID получает пользователя по ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid user ID format")
	}

	query := r.qb.Select("id", "email", "password_hash", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"id": userID})

	row := r.qb.QueryRow(query)

	var user domain.User
	err = row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("user not found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &user, nil
}

// GetByEmail получает пользователя по email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := r.qb.Select("id", "email", "password_hash", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"email": email})

	row := r.qb.QueryRow(query)

	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("user not found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &user, nil
}

// Update обновляет пользователя
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := r.qb.Update("users").
		Set("email", user.Email).
		Set("password_hash", user.PasswordHash).
		Set("updated_at", user.UpdatedAt).
		Where(squirrel.Eq{"id": user.ID})

	_, err := r.qb.Execute(query)
	return err
}

// Delete удаляет пользователя
func (r *userRepository) Delete(ctx context.Context, id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrBadRequest.WithMessage("invalid user ID format")
	}

	query := r.qb.Delete("users").
		Where(squirrel.Eq{"id": userID})

	_, err = r.qb.Execute(query)
	return err
}
