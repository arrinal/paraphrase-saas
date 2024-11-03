package api

import (
	"net/http"

	"github.com/arrinal/paraphrase-saas/internal/config"
	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
	"github.com/arrinal/paraphrase-saas/internal/services"
	"github.com/gin-gonic/gin"
)

type CreateCheckoutSessionRequest struct {
	PlanID   string `json:"planId" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=web ios"`
}

func HandleCreateCheckoutSession(cfg *config.Config) gin.HandlerFunc {
	paddleService := services.NewPaddleService(cfg)

	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		var req CreateCheckoutSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var existingSub models.Subscription
		if err := db.DB.Where("user_id = ? AND status = ?", userID, "active").
			First(&existingSub).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already has an active subscription"})
			return
		}

		checkoutURL, err := paddleService.CreateCheckoutSession(userID.(uint), req.PlanID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create checkout session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": checkoutURL})
	}
}

func HandleWebhook(cfg *config.Config) gin.HandlerFunc {
	paddleService := services.NewPaddleService(cfg)

	return func(c *gin.Context) {
		payload, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read payload"})
			return
		}

		// Verify webhook signature
		if !paddleService.VerifyWebhookSignature(payload, c.GetHeader("Paddle-Signature")) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook signature"})
			return
		}

		if err := paddleService.HandleWebhookEvent(payload); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to handle webhook"})
			return
		}

		c.Status(http.StatusOK)
	}
}

func HandleGetSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var subscription models.Subscription
		if err := db.DB.Where("user_id = ? AND status = ?", userID, "active").
			First(&subscription).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active subscription found"})
			return
		}

		c.JSON(http.StatusOK, subscription)
	}
}

func HandleCancelSubscription(cfg *config.Config) gin.HandlerFunc {
	paddleService := services.NewPaddleService(cfg)

	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var subscription models.Subscription
		if err := db.DB.Where("user_id = ? AND status = ?", userID, "active").
			First(&subscription).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active subscription found"})
			return
		}

		if err := paddleService.CancelSubscription(subscription.PaddleSubID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel subscription"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "subscription cancelled successfully"})
	}
}
