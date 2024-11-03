package api

import (
	"net/http"

	"github.com/arrinal/paraphrase-saas/internal/config"
	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/services"
	"github.com/gin-gonic/gin"
)

func HandleVerifyIOSReceipt(cfg *config.Config) gin.HandlerFunc {
	iosService := services.NewIOSPaymentService(cfg)

	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var receipt services.IAPReceipt
		if err := c.ShouldBindJSON(&receipt); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		subscription, err := iosService.VerifyReceipt(receipt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Assign user ID to subscription
		subscription.UserID = userID.(uint)

		// Save subscription to database
		if err := db.DB.Create(subscription).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save subscription"})
			return
		}

		c.JSON(http.StatusOK, subscription)
	}
}
