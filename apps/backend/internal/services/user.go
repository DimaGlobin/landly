package services

import (
	"context"
	"time"

	domain "github.com/landly/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserService сервис для управления пользователями
type UserService struct {
	userRepo UserRepository
}

// NewUserService создаёт новый user service
func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetProfile получает профиль пользователя
func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Не возвращаем password_hash в профиле
	user.PasswordHash = ""
	return user, nil
}

// UpdateProfile обновляет профиль пользователя
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *domain.UpdateProfileRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Обновляем email если указан
	if req.Email != "" {
		// Проверяем, что email не занят другим пользователем
		existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
		if err == nil && existingUser.ID != user.ID {
			return nil, domain.ErrConflict.WithMessage("email already taken")
		}
		user.Email = req.Email
	}

	// Обновляем пароль если указан
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	// Не возвращаем password_hash
	user.PasswordHash = ""
	return user, nil
}

