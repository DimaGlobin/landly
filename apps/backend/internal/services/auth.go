package services

import (
	"context"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository интерфейс для репозитория пользователей
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

// AuthService сервис для аутентификации
type AuthService struct {
	userRepo   UserRepository
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// AuthTokens токены аутентификации
type AuthTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// NewAuthService создаёт новый auth service
func NewAuthService(userRepo UserRepository, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}

	return &AuthService{
		userRepo:   userRepo,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Register регистрация нового пользователя (новый интерфейс)
func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	tokens, err := s.SignUp(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

// Login вход пользователя (новый интерфейс)
func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	tokens, err := s.SignIn(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

// SignUp регистрация нового пользователя
func (s *AuthService) SignUp(ctx context.Context, email, password string) (*AuthTokens, error) {
	// Проверяем, существует ли пользователь
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return nil, domain.ErrConflict.WithMessage("пользователь уже существует")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInternal.WithMessage("failed to hash password")
	}

	// Создаём пользователя
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, domain.ErrInternal.WithMessage("failed to create user").WithError(err)
	}

	// Генерируем токены
	return s.generateTokens(user.ID)
}

// SignIn вход пользователя
func (s *AuthService) SignIn(ctx context.Context, email, password string) (*AuthTokens, error) {
	// Получаем пользователя
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrUnauthorized.WithMessage("invalid credentials")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrUnauthorized.WithMessage("invalid credentials")
	}

	// Генерируем токены
	return s.generateTokens(user.ID)
}

// ValidateToken валидирует токен (новый интерфейс)
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	// Парсим токен
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !parsedToken.Valid {
		return nil, domain.ErrUnauthorized.WithMessage("invalid token")
	}

	// Получаем claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrUnauthorized.WithMessage("invalid token claims")
	}

	// Проверяем тип токена
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return nil, domain.ErrUnauthorized.WithMessage("invalid token type")
	}

	// Получаем user_id
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, domain.ErrUnauthorized.WithMessage("invalid user_id in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, domain.ErrUnauthorized.WithMessage("invalid user_id format")
	}

	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, userID.String())
	if err != nil {
		return nil, domain.ErrUnauthorized.WithMessage("user not found")
	}

	return user, nil
}

// RefreshToken обновление токена
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	// Парсим токен
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized.WithMessage("invalid refresh token")
	}

	// Получаем claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrUnauthorized.WithMessage("invalid token claims")
	}

	// Проверяем тип токена
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, domain.ErrUnauthorized.WithMessage("invalid token type")
	}

	// Получаем user ID
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, domain.ErrUnauthorized.WithMessage("invalid user id")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, domain.ErrUnauthorized.WithMessage("invalid user id format")
	}

	// Проверяем, существует ли пользователь
	_, err = s.userRepo.GetByID(ctx, userID.String())
	if err != nil {
		return nil, domain.ErrUnauthorized.WithMessage("user not found")
	}

	// Генерируем новые токены
	tokens, err := s.generateTokens(userID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

// generateTokens генерирует access и refresh токены
func (s *AuthService) generateTokens(userID uuid.UUID) (*AuthTokens, error) {
	now := time.Now()
	accessExp := now.Add(s.accessTTL)
	refreshExp := now.Add(s.refreshTTL)

	// Access token
	accessClaims := jwt.MapClaims{
		"user_id": userID.String(),
		"type":    "access",
		"exp":     accessExp.Unix(),
		"iat":     now.Unix(),
		"jti":     uuid.NewString(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, domain.ErrInternal.WithMessage("failed to generate access token")
	}

	// Refresh token
	refreshClaims := jwt.MapClaims{
		"user_id": userID.String(),
		"type":    "refresh",
		"exp":     refreshExp.Unix(),
		"iat":     now.Unix(),
		"jti":     uuid.NewString(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, domain.ErrInternal.WithMessage("failed to generate refresh token")
	}

	return &AuthTokens{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExp,
	}, nil
}
