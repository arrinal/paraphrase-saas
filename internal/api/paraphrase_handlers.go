package api

import (
	"net/http"

	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
	"github.com/arrinal/paraphrase-saas/internal/services"
	"github.com/gin-gonic/gin"
)

type ParaphraseRequest struct {
	Text     string `json:"text" binding:"required"`
	Language string `json:"language" binding:"required"`
	Style    string `json:"style" binding:"required"`
}

type ParaphraseResponse struct {
	Paraphrased string `json:"paraphrased"`
	Language    string `json:"language"`
}

func HandleParaphrase(openAIService *services.OpenAIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		// Get the parsed request from context
		reqValue, exists := c.Get("parsedRequest")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		req := reqValue.(struct {
			Text     string `json:"text" binding:"required"`
			Language string `json:"language" binding:"required"`
			Style    string `json:"style" binding:"required"`
		})

		// Paraphrase the text
		paraphrasedResp, err := openAIService.Paraphrase(req.Text, req.Language, req.Style)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to paraphrase text"})
			return
		}

		// Create history entry
		history := models.ParaphraseHistory{
			UserID:          userID.(uint),
			OriginalText:    req.Text,
			ParaphrasedText: paraphrasedResp.Paraphrased,
			Language:        paraphrasedResp.DetectedLanguage,
			Style:           req.Style,
		}

		if err := db.DB.Create(&history).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save history"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"paraphrased": paraphrasedResp.Paraphrased,
			"language":    paraphrasedResp.DetectedLanguage,
			"history_id":  history.ID,
		})
	}
}
