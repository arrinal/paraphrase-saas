package middleware

import (
	"net/http"
	"strings"

	"github.com/arrinal/paraphrase-saas/internal/auth"
	"github.com/arrinal/paraphrase-saas/internal/config"
	"github.com/gin-gonic/gin"
)

func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no authorization header"})
			return
		}

		tokenString := strings.Replace(header, "Bearer ", "", 1)
		claims, err := auth.ValidateToken(tokenString, cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}
