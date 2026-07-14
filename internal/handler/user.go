package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		abortUnauthorized(c)
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

type UpdateProfileRequest struct {
	Name   *string  `json:"name,omitempty"`
	Weight *float64 `json:"weight,omitempty"`
	Height *float64 `json:"height,omitempty"`
	Age    *int     `json:"age,omitempty"`
	Gender *string  `json:"gender,omitempty"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		abortUnauthorized(c)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateProfile(c.Request.Context(), userID, req.Name, req.Weight, req.Height, req.Age, req.Gender); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}
