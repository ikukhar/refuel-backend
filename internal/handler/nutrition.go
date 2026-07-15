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
	userID, ok := userIDFromContext(c)
	if !ok {
		abortUnauthorized(c)
		return
	}

	refresh := c.Query("refresh") == "true"
	meal := c.Query("meal")

	if refresh && meal != "" {
		nutrition, err := h.nutritionService.RefreshMeal(c.Request.Context(), userID, meal)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, nutrition)
		return
	}

	if refresh {
		nutrition, err := h.nutritionService.RefreshToday(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, nutrition)
		return
	}

	if meal != "" {
		mealResp, err := h.nutritionService.GetMeal(c.Request.Context(), userID, meal)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{meal: mealResp})
		return
	}

	nutrition, err := h.nutritionService.GetToday(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nutrition)
}
