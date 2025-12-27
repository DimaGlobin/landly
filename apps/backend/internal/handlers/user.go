package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/landly/backend/internal/handlers/dto"
	domain "github.com/landly/backend/internal/models"
)

// UserService интерфейс для сервиса пользователей
type UserService interface {
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID string, req *domain.UpdateProfileRequest) (*domain.User, error)
}

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Tags user
// @Produce json
// @Success 200 {object} dto.UserProfileResponse
// @Router /v1/user/profile [get]
// @Security BearerAuth
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), userID.String())
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			RespondError(c, domainErr)
			return
		}
		RespondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.UserProfileResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Tags user
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} dto.UserProfileResponse
// @Router /v1/user/profile [put]
// @Security BearerAuth
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID.String(), &domain.UpdateProfileRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			RespondError(c, domainErr)
			return
		}
		RespondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.UserProfileResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

