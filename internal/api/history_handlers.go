package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
)

func HandleGetHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var history []models.ParaphraseHistory
		if err := db.DB.Where("user_id = ?", userID).
			Order("created_at desc").
			Find(&history).Error; err != nil {
			log.Printf("Error fetching history: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch history"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"history": history,
		})
	}
}

// Add this new handler
func HandleGetUsedLanguages() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var languages []string
		if err := db.DB.Model(&models.ParaphraseHistory{}).
			Where("user_id = ?", userID).
			Distinct().
			Pluck("language", &languages).
			Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch languages"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"languages": languages,
		})
	}
}
