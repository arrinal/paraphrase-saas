package services

import (
	"github.com/arrinal/paraphrase-saas/internal/config"
)

// PaddleServiceInterface defines methods that both mock and real implementations must have
type PaddleServiceInterface interface {
	CreateCheckoutSession(userID uint, planID string) (string, error)
	VerifyWebhookSignature(payload []byte, signature string) bool
	HandleWebhookEvent(payload []byte) error
	CancelSubscription(subscriptionID string) error
}

// NewPaddleService returns either mock or real Paddle service based on environment
func NewPaddleService(cfg *config.Config) PaddleServiceInterface {
	if cfg.Environment == "development" {
		return NewMockPaddleService(cfg)
	}
	return NewRealPaddleService(cfg)
}
