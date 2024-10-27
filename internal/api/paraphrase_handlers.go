package api

import (
	"log"
	"net/http"
	"time"

	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
	"github.com/arrinal/paraphrase-saas/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ParaphraseRequest struct {
	Text     string `json:"text" binding:"required"`
	Language string `json:"language" binding:"required"`
	Style    string `json:"style" binding:"required"`
}

type ParaphraseResponse struct {
	Paraphrased string `json:"paraphrased"`
	Language    string `json:"language"` // Add this field
}

func HandleParaphrase(openAI *services.OpenAIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ParaphraseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, _ := c.Get("userID")

		// If language is auto, detect it first
		actualLanguage := req.Language
		if req.Language == "auto" {
			detected, err := openAI.GetDetectedLanguage(req.Text)
			if err != nil {
				log.Printf("Language detection error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to detect language"})
				return
			}
			actualLanguage = detected
		}

		paraphrased, err := openAI.Paraphrase(req.Text, req.Language, req.Style)
		if err != nil {
			log.Printf("Paraphrase error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Save to history with the actual detected language
		history := models.ParaphraseHistory{
			UserID:          userID.(uint),
			OriginalText:    req.Text,
			ParaphrasedText: paraphrased,
			Language:        actualLanguage,
			Style:           req.Style,
		}

		if err := db.DB.Create(&history).Error; err != nil {
			log.Printf("Error saving history: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save history"})
			return
		}

		// Track usage after successful paraphrase
		if err := db.DB.Model(&models.UserStats{}).
			Where("user_id = ?", userID).
			UpdateColumn("total_paraphrases", gorm.Expr("total_paraphrases + ?", 1)).
			Error; err != nil {
			log.Printf("Failed to update usage stats: %v", err)
		}

		// Update daily usage
		today := time.Now().UTC().Truncate(24 * time.Hour)
		var dailyUsage models.DailyUsage
		result := db.DB.Where("user_id = ? AND date = ?", userID, today).First(&dailyUsage)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				dailyUsage = models.DailyUsage{
					UserID: userID.(uint),
					Date:   today,
					Count:  1,
				}
				db.DB.Create(&dailyUsage)
			}
		} else {
			db.DB.Model(&dailyUsage).UpdateColumn("count", gorm.Expr("count + ?", 1))
		}

		c.JSON(http.StatusOK, gin.H{
			"paraphrased": paraphrased,
			"language":    actualLanguage,
			"history_id":  history.ID,
		})
	}
}
