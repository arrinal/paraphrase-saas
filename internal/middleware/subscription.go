package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
	"github.com/gin-gonic/gin"
)

type PlanLimits struct {
	CharactersPerRequest int  `json:"charactersPerRequest"`
	RequestsPerDay       int  `json:"requestsPerDay"`
	BulkParaphrase       bool `json:"bulkParaphrase"`
}

func CheckSubscriptionLimits() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var subscription models.Subscription
		if err := db.DB.Where("user_id = ? AND status = ?", userID, "active").
			First(&subscription).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "active subscription required",
				"code":  "SUBSCRIPTION_REQUIRED",
			})
			c.Abort()
			return
		}

		var plan models.SubscriptionPlan
		if err := db.DB.First(&plan, "id = ?", subscription.PlanID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get subscription plan",
			})
			c.Abort()
			return
		}

		var limits PlanLimits
		if err := json.Unmarshal(plan.Limits, &limits); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to parse plan limits",
			})
			c.Abort()
			return
		}

		c.Set("planLimits", limits)
		c.Next()
	}
}
