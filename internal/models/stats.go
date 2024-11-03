package models

import (
	"time"

	"gorm.io/gorm"
)

type UserStats struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	UserID           uint           `gorm:"index" json:"user_id"`
	TotalParaphrases int            `json:"total_paraphrases"`
	LastUsedAt       time.Time      `json:"last_used_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type DailyUsage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Date      time.Time `gorm:"index" json:"date"`
	Count     int       `json:"count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
