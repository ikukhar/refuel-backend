package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/service"
)

type MealPeriodHandler struct {
	svc *service.MealPeriodService
}

func NewMealPeriodHandler(svc *service.MealPeriodService) *MealPeriodHandler {
	return &MealPeriodHandler{svc: svc}
}

func GetDefaultMealPeriods(c *gin.Context) {
	c.JSON(http.StatusOK, model.DefaultMealPeriods)
}

type createMealPeriodItem struct {
	MealType    string `json:"meal_type" binding:"required"`
	Name        string `json:"name"`
	StartHour   int    `json:"start_hour" binding:"required"`
	StartMinute int    `json:"start_minute"`
}

func (h *MealPeriodHandler) Create(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		abortUnauthorized(c)
		return
	}

	var req []createMealPeriodItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty periods"})
		return
	}

	items := make([]service.UpsertMealPeriodItem, len(req))
	for i, item := range req {
		items[i] = service.UpsertMealPeriodItem{
			MealType:    item.MealType,
			Name:        item.Name,
			StartHour:   item.StartHour,
			StartMinute: item.StartMinute,
		}
	}

	periods, err := h.svc.Upsert(userID, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, periods)
}
