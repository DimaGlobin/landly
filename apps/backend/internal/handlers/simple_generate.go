package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SimpleGenerateRequest простая структура для генерации
type SimpleGenerateRequest struct {
	Prompt     string `json:"prompt" binding:"required"`
	PaymentURL string `json:"payment_url"`
}

// SimpleGenerateResponse простой ответ генерации
type SimpleGenerateResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Schema  map[string]interface{} `json:"schema,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// SimpleGenerateService простой интерфейс для генерации
type SimpleGenerateService interface {
	GenerateSimple(ctx context.Context, userID, projectID string, prompt, paymentURL string) (map[string]interface{}, error)
}

type SimpleGenerateHandler struct {
	generateService SimpleGenerateService
}

func NewSimpleGenerateHandler(generateService SimpleGenerateService) *SimpleGenerateHandler {
	return &SimpleGenerateHandler{
		generateService: generateService,
	}
}

// GenerateSimple простой эндпоинт генерации
func (h *SimpleGenerateHandler) GenerateSimple(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, SimpleGenerateResponse{
			Success: false,
			Error:   "unauthorized",
		})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, SimpleGenerateResponse{
			Success: false,
			Error:   "invalid project id",
		})
		return
	}

	var req SimpleGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SimpleGenerateResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Генерируем лендинг
	schema, err := h.generateService.GenerateSimple(c.Request.Context(), userID.String(), projectID.String(), req.Prompt, req.PaymentURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, SimpleGenerateResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SimpleGenerateResponse{
		Success: true,
		Message: "Лендинг успешно сгенерирован",
		Schema:  schema,
	})
}
