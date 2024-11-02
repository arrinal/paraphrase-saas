package api

import (
	"log"
	"net/http"
	"time"

	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
	"github.com/gin-gonic/gin"
)

type IAPVerifyRequest struct {
	Receipt     string `json:"receipt" binding:"required"`
	ProductID   string `json:"product_id" binding:"required"`
	Transaction string `json:"transaction_id" binding:"required"`
}

func HandleVerifyIAPReceipt() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		var req IAPVerifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Verify receipt with Apple's servers
		// This is a placeholder for the actual verification
		// In production, you should verify the receipt with Apple's servers
		isValid := true // Replace with actual verification

		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receipt"})
			return
		}

		// Create or update subscription
		subscription := models.Subscription{
			UserID:           userID.(uint),
			PlanID:           req.ProductID,
			Status:           "active",
			CurrentPeriodEnd: time.Now().AddDate(0, 1, 0), // 1 month from now
		}

		if err := db.DB.Save(&subscription).Error; err != nil {
			log.Printf("Error saving subscription: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save subscription"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "subscription activated",
			"subscription": subscription,
		})
	}
}
