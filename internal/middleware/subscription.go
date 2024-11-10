package middleware

import (
	"net/http"
	"time"

	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
	"github.com/gin-gonic/gin"
)

func CheckSubscriptionLimits() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		// Check subscription status
		var subscription models.Subscription
		if err := db.DB.Where("user_id = ? AND status = ?", userID, "active").
			First(&subscription).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "active subscription required"})
			c.Abort()
			return
		}

		// Only check period end for pro subscriptions
		if subscription.PlanID == "pro" && subscription.CurrentPeriodEnd.Before(time.Now()) {
			// Update subscription status to expired
			subscription.Status = "expired"
			if err := db.DB.Save(&subscription).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription status"})
				c.Abort()
				return
			}

			c.JSON(http.StatusForbidden, gin.H{"error": "subscription has expired"})
			c.Abort()
			return
		}

		// Parse request once and store in context
		var req struct {
			Text     string `json:"text" binding:"required"`
			Language string `json:"language" binding:"required"`
			Style    string `json:"style" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Store parsed request in context for handler to use
		c.Set("parsedRequest", req)

		// Trial plan restrictions - check plan_id instead of status
		if subscription.PlanID == "trial" {
			// Enforce English only
			if req.Language != "English" && req.Language != "english" {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "trial plan only supports English language",
					"code":  "TRIAL_RESTRICTION",
				})
				c.Abort()
				return
			}

			// Enforce standard style only
			if req.Style != "standard" {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "trial plan only supports standard style",
					"code":  "TRIAL_RESTRICTION",
				})
				c.Abort()
				return
			}

			// Check character limit
			if len(req.Text) > 1000 {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "trial plan limited to 1000 characters",
					"code":  "TRIAL_RESTRICTION",
				})
				c.Abort()
				return
			}

			// Check total usage limit for trial account
			var totalUsageCount int64
			if err := db.DB.Model(&models.ParaphraseHistory{}).
				Where("user_id = ?", userID).
				Count(&totalUsageCount).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check usage limit"})
				c.Abort()
				return
			}

			if totalUsageCount >= 5 {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "trial plan limited to 5 paraphrases total",
					"code":  "TRIAL_RESTRICTION",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
