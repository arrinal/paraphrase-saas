package models

import (
	"time"

	"gorm.io/gorm"
)

type Activity struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index" json:"user_id"`
	Action    string         `json:"action"`
	Details   string         `json:"details"`
	Metadata  JSON           `gorm:"type:jsonb" json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type JSON []byte

func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}
