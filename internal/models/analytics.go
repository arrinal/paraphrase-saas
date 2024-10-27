package models

import (
	"time"

	"gorm.io/gorm"
)

type AnalyticsEvent struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index" json:"user_id"`
	EventType string         `json:"event_type"`
	Metadata  JSON           `gorm:"type:jsonb" json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserAnalytics struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserID           uint      `gorm:"index" json:"user_id"`
	TotalParaphrases int       `json:"total_paraphrases"`
	LastActivityAt   time.Time `json:"last_activity_at"`
	AverageWordCount float64   `json:"average_word_count"`
	MostUsedLanguage string    `json:"most_used_language"`
	MostUsedStyle    string    `json:"most_used_style"`
	UpdatedAt        time.Time `json:"updated_at"`
}
