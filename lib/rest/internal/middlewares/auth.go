package middlewares

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func Auth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		fmt.Println(token)
		if !slices.Contains([]string{
			fmt.Sprintf("Bearer %s", token),
			fmt.Sprintf("bearer %s", token),
		}, header) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
