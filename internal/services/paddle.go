package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/arrinal/paraphrase-saas/internal/config"
)

type PaddleService struct {
	cfg *config.Config
}

func NewPaddleService(cfg *config.Config) *PaddleService {
	return &PaddleService{cfg: cfg}
}

func (s *PaddleService) CreateCheckoutSession(userID uint, planID string) (string, error) {
	// Get Paddle price ID for the plan
	priceID, ok := s.cfg.PaddlePriceIDs[planID]
	if !ok {
		return "", fmt.Errorf("invalid plan ID")
	}

	// In production, you would make an API call to Paddle to create a checkout
	// For now, return a mock checkout URL
	checkoutURL := fmt.Sprintf(
		"https://checkout.paddle.com/checkout/%s?user_id=%d",
		priceID,
		userID,
	)

	return checkoutURL, nil
}

func (s *PaddleService) VerifyWebhookSignature(payload []byte, signature string) bool {
	// Implement Paddle's webhook signature verification
	// https://developer.paddle.com/webhook-reference/verifying-webhooks
	mac := hmac.New(sha256.New, []byte(s.cfg.PaddlePublicKey))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

func (s *PaddleService) HandleWebhookEvent(payload []byte) error {
	var event struct {
		EventType string         `json:"event_type"`
		Data      map[string]any `json:"data"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %v", err)
	}

	switch event.EventType {
	case "subscription.created":
		return s.handleSubscriptionCreated(event.Data)
	case "subscription.updated":
		return s.handleSubscriptionUpdated(event.Data)
	case "subscription.cancelled":
		return s.handleSubscriptionCancelled(event.Data)
	}

	return nil
}

func (s *PaddleService) handleSubscriptionCreated(data map[string]any) error {
	// Handle subscription creation
	// You'll implement this based on your database schema
	return nil
}

func (s *PaddleService) handleSubscriptionUpdated(data map[string]any) error {
	// Handle subscription updates
	return nil
}

func (s *PaddleService) handleSubscriptionCancelled(data map[string]any) error {
	// Handle subscription cancellation
	return nil
}
