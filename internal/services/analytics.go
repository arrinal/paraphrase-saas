package services

import (
	"encoding/json"
	"time"

	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
	"github.com/arrinal/paraphrase-saas/internal/websocket"
)

type AnalyticsService struct {
	hub *websocket.Hub
}

func NewAnalyticsService(hub *websocket.Hub) *AnalyticsService {
	return &AnalyticsService{hub: hub}
}

func (s *AnalyticsService) TrackEvent(userID uint, eventType string, metadata interface{}) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	event := models.AnalyticsEvent{
		UserID:    userID,
		EventType: eventType,
		Metadata:  models.JSON(metadataJSON),
		CreatedAt: time.Now(),
	}

	if err := db.DB.Create(&event).Error; err != nil {
		return err
	}

	// Broadcast event to connected clients
	s.hub.BroadcastToUser(userID, "analytics_update", event)

	return nil
}

func (s *AnalyticsService) GetUserAnalytics(userID uint) (*models.UserAnalytics, error) {
	var analytics models.UserAnalytics
	err := db.DB.Where("user_id = ?", userID).First(&analytics).Error
	if err != nil {
		return nil, err
	}
	return &analytics, nil
}

func (s *AnalyticsService) UpdateUserAnalytics(userID uint) error {
	// Calculate analytics
	var totalParaphrases int64
	db.DB.Model(&models.ParaphraseHistory{}).Where("user_id = ?", userID).Count(&totalParaphrases)

	var mostUsedLanguage string
	db.DB.Model(&models.ParaphraseHistory{}).
		Select("language").
		Where("user_id = ?", userID).
		Group("language").
		Order("count(*) desc").
		Limit(1).
		Pluck("language", &mostUsedLanguage)

	analytics := models.UserAnalytics{
		UserID:           userID,
		TotalParaphrases: int(totalParaphrases),
		LastActivityAt:   time.Now(),
		MostUsedLanguage: mostUsedLanguage,
		UpdatedAt:        time.Now(),
	}

	return db.DB.Save(&analytics).Error
}
