package api

import (
	"net/http"
	"time"

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

		var currentSubscription models.Subscription
		hasCurrentSub := db.DB.Where("user_id = ? AND status = ?", userID, "active").
			First(&currentSubscription).Error == nil

		var subscriptionHistory models.Subscription
		hasProHistory := db.DB.Unscoped().Where("user_id = ? AND plan_id = ?", userID, "pro").
			First(&subscriptionHistory).Error == nil

		if req.PlanID == "trial" {
			if hasProHistory {
				c.JSON(http.StatusBadRequest, gin.H{"error": "trial plan is not available after having a pro subscription"})
				return
			}
			if hasCurrentSub {
				c.JSON(http.StatusBadRequest, gin.H{"error": "user already has an active subscription"})
				return
			}
		} else if req.PlanID == "pro" {
			if hasCurrentSub && currentSubscription.PlanID == "trial" {
				if err := db.DB.Unscoped().Where("user_id = ?", userID).
					Delete(&models.Subscription{}).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clean up trial subscription"})
					return
				}
			} else if hasCurrentSub {
				c.JSON(http.StatusBadRequest, gin.H{"error": "user already has an active subscription"})
				return
			}
		}

		checkoutURL, err := paddleService.CreateCheckoutSession(userID.(uint), req.PlanID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create checkout session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": checkoutURL})
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

		// Only check expiration for pro subscriptions
		if subscription.PlanID == "pro" && subscription.CurrentPeriodEnd.Before(time.Now()) {
			subscription.Status = "expired"
			if err := db.DB.Save(&subscription).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription has expired"})
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
		if err := db.DB.Where("user_id = ? AND status IN (?)", userID, []string{"active", "trial"}).
			First(&subscription).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active subscription found"})
			return
		}

		if subscription.PlanID == "pro" {
			if err := paddleService.CancelSubscription(subscription.PaddleSubID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel subscription"})
				return
			}

			subscription.CancelAtPeriodEnd = true
			if err := db.DB.Save(&subscription).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
				return
			}
		} else {
			subscription.Status = "cancelled"
			if err := db.DB.Save(&subscription).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "subscription cancelled successfully"})
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

func HandleCheckSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uint)

		var count int64
		db.DB.Model(&models.Subscription{}).Where("user_id = ? AND plan_id = ?", userID, "pro").Count(&count)

		c.JSON(http.StatusOK, gin.H{"hasPro": count > 0})
	}
}
