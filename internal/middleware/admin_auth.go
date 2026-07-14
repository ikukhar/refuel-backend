package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminAuth(user, pass string) gin.HandlerFunc {
	userHash := sha256.Sum256([]byte(user))
	passHash := sha256.Sum256([]byte(pass))

	return func(c *gin.Context) {
		u, p, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="admin"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		uHash := sha256.Sum256([]byte(u))
		pHash := sha256.Sum256([]byte(p))

		userOk := subtle.ConstantTimeCompare(userHash[:], uHash[:]) == 1
		passOk := subtle.ConstantTimeCompare(passHash[:], pHash[:]) == 1

		if !userOk || !passOk {
			c.Header("WWW-Authenticate", `Basic realm="admin"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}
