package models

import (
	"time"

	"gorm.io/gorm"
)

type ParaphraseHistory struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `gorm:"index" json:"user_id"`
	OriginalText    string         `gorm:"type:text" json:"original_text"`
	ParaphrasedText string         `gorm:"type:text" json:"paraphrased_text"`
	Language        string         `json:"language"`
	Style           string         `json:"style"`
	CreatedAt       time.Time      `json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
