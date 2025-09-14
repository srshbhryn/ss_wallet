package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Auth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != fmt.Sprintf("Bearer %s", token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
