package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
)

type StatsResponse struct {
	TotalParaphrases  int                  `json:"totalParaphrases"`
	LanguageBreakdown map[string]int       `json:"languageBreakdown"`
	StyleBreakdown    map[string]int       `json:"styleBreakdown"`
	DailyUsage        []DailyUsageResponse `json:"dailyUsage"`
}

type DailyUsageResponse struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

func HandleGetUserStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		// Get total paraphrases
		var totalParaphrases int64
		if err := db.DB.Model(&models.ParaphraseHistory{}).
			Where("user_id = ?", userID).
			Count(&totalParaphrases).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stats"})
			return
		}

		// Get language breakdown
		var languageStats []struct {
			Language string
			Count    int
		}
		if err := db.DB.Model(&models.ParaphraseHistory{}).
			Select("language, count(*) as count").
			Where("user_id = ?", userID).
			Group("language").
			Find(&languageStats).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch language stats"})
			return
		}

		// Get style breakdown
		var styleStats []struct {
			Style string
			Count int
		}
		if err := db.DB.Model(&models.ParaphraseHistory{}).
			Select("style, count(*) as count").
			Where("user_id = ?", userID).
			Group("style").
			Find(&styleStats).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch style stats"})
			return
		}

		// Get daily usage for the last 30 days
		var dailyUsage []struct {
			Date  time.Time
			Count int
		}
		if err := db.DB.Model(&models.ParaphraseHistory{}).
			Select("DATE(created_at) as date, count(*) as count").
			Where("user_id = ? AND created_at >= ?", userID, time.Now().AddDate(0, 0, -30)).
			Group("DATE(created_at)").
			Order("date DESC").
			Find(&dailyUsage).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch daily usage"})
			return
		}

		// Format response
		languageBreakdown := make(map[string]int)
		for _, stat := range languageStats {
			languageBreakdown[stat.Language] = stat.Count
		}

		styleBreakdown := make(map[string]int)
		for _, stat := range styleStats {
			styleBreakdown[stat.Style] = stat.Count
		}

		dailyUsageResponse := make([]DailyUsageResponse, len(dailyUsage))
		for i, usage := range dailyUsage {
			dailyUsageResponse[i] = DailyUsageResponse{
				Date:  usage.Date.Format("2006-01-02"),
				Count: usage.Count,
			}
		}

		c.JSON(http.StatusOK, StatsResponse{
			TotalParaphrases:  int(totalParaphrases),
			LanguageBreakdown: languageBreakdown,
			StyleBreakdown:    styleBreakdown,
			DailyUsage:        dailyUsageResponse,
		})
	}
}
