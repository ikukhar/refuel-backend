package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/service"
)

type RecipeHandler struct {
	recipeRepo service.RecipeRepository
}

func NewRecipeHandler(recipeRepo service.RecipeRepository) *RecipeHandler {
	return &RecipeHandler{recipeRepo: recipeRepo}
}

func (h *RecipeHandler) ListByIDs(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids query param is required"})
		return
	}

	parts := strings.Split(idsStr, ",")
	ids := make([]uint, 0, len(parts))
	for _, p := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + p})
			return
		}
		ids = append(ids, uint(id))
	}

	recipes, err := h.recipeRepo.FindByIDs(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if recipes == nil {
		recipes = []model.Recipe{}
	}

	c.JSON(http.StatusOK, recipes)
}
