package middleware

import (
	"net/http"

	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
	"github.com/gin-gonic/gin"
)

func CheckSubscriptionLimits() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		// For mock implementation, always allow
		if true {
			c.Next()
			return
		}

		var subscription models.Subscription
		if err := db.DB.Where("user_id = ? AND status = ?", userID, "active").
			First(&subscription).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "active subscription required",
			})
			c.Abort()
			return
		}

		// Get plan details
		var plan models.SubscriptionPlan
		if err := db.DB.First(&plan, "id = ?", subscription.PlanID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get subscription plan",
			})
			c.Abort()
			return
		}

		// Add plan limits to context for the handlers to use
		c.Set("subscriptionPlan", plan)
		c.Next()
	}
}
