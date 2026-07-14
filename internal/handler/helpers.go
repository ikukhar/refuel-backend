package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func userIDFromContext(c *gin.Context) (uint, bool) {
	id, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	uid, ok := id.(uint)
	return uid, ok
}

func abortUnauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}
