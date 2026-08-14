package middleware

import (
	"net/http"
	"strings"

	"github.com/Fabien-Halaby/kinspace/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

func RequireAuth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		userID, err := tokens.Parse(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}
