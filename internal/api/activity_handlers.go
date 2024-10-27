package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
	"github.com/arrinal/paraphrase-saas/internal/websocket"
	"github.com/gin-gonic/gin"
)

type ActivityRequest struct {
	Action   string         `json:"action" binding:"required"`
	Details  string         `json:"details"`
	Metadata map[string]any `json:"metadata"`
}

func HandleTrackActivity(hub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		var req ActivityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		metadataJSON, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metadata"})
			return
		}

		activity := models.Activity{
			UserID:    userID.(uint),
			Action:    req.Action,
			Details:   req.Details,
			Metadata:  models.JSON(metadataJSON),
			CreatedAt: time.Now(),
		}

		if err := db.DB.Create(&activity).Error; err != nil {
			log.Printf("Failed to save activity: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save activity"})
			return
		}

		// Broadcast activity to connected clients
		hub.BroadcastToUser(userID.(uint), "activity_update", activity)

		c.JSON(http.StatusOK, gin.H{"message": "activity tracked"})
	}
}

func HandleGetActivities() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var activities []models.Activity
		if err := db.DB.Where("user_id = ?", userID).
			Order("created_at desc").
			Limit(100).
			Find(&activities).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch activities"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"activities": activities})
	}
}
