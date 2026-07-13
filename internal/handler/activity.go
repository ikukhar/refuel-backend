package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/service"
)

type ActivityHandler struct {
	activityService *service.ActivityService
}

func NewActivityHandler(activityService *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: activityService}
}

type CreateActivityRequest struct {
	Type      string    `json:"type" binding:"required"`
	Distance  *float64  `json:"distance,omitempty"`
	Duration  *int      `json:"duration,omitempty"`
	Elevation *float64  `json:"elevation,omitempty"`
	Calories  *int      `json:"calories,omitempty"`
	StartedAt time.Time `json:"started_at" binding:"required"`
	Source    string    `json:"source,omitempty"`
	SourceID  string    `json:"source_id" binding:"required"`
}

func (h *ActivityHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := service.CreateActivityInput{
		Type:      req.Type,
		Distance:  req.Distance,
		Duration:  req.Duration,
		Elevation: req.Elevation,
		Calories:  req.Calories,
		StartedAt: req.StartedAt,
		Source:    req.Source,
		SourceID:  req.SourceID,
	}

	resp, created, err := h.activityService.Create(c.Request.Context(), userID.(uint), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if created {
		c.JSON(http.StatusCreated, resp)
	} else {
		c.JSON(http.StatusOK, resp)
	}
}

func (h *ActivityHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")

	from, to := parseDateRange(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	activities, err := h.activityService.List(c.Request.Context(), userID.(uint), from, to, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, activities)
}

func parseDateRange(c *gin.Context) (*time.Time, *time.Time) {
	var from, to *time.Time
	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = &t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = &t
		}
	}
	return from, to
}
