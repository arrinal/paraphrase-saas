package models

import (
	"time"

	"gorm.io/gorm"
)

type Subscription struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	UserID            uint           `gorm:"index" json:"user_id"`
	PaddleSubID       string         `gorm:"uniqueIndex" json:"paddle_subscription_id"`
	PlanID            string         `json:"plan_id"`
	Status            string         `json:"status"`
	CurrentPeriodEnd  time.Time      `json:"current_period_end"`
	CancelAtPeriodEnd bool           `json:"cancel_at_period_end"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

type SubscriptionPlan struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Price        int64     `json:"price"` // in cents
	Currency     string    `json:"currency"`
	Interval     string    `json:"interval"`
	PaddlePlanID string    `json:"paddle_plan_id"`
	Features     JSON      `gorm:"type:jsonb" json:"features"`
	Limits       JSON      `gorm:"type:jsonb" json:"limits"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
