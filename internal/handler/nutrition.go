package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/service"
)

type NutritionHandler struct {
	nutritionService *service.NutritionService
}

func NewNutritionHandler(nutritionService *service.NutritionService) *NutritionHandler {
	return &NutritionHandler{nutritionService: nutritionService}
}

func (h *NutritionHandler) GetToday(c *gin.Context) {
	userID, _ := c.Get("user_id")

	nutrition, err := h.nutritionService.GetToday(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nutrition)
}
