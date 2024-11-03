package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/arrinal/paraphrase-saas/internal/config"
	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
)

type PaddleService struct {
	cfg *config.Config
}

type PaddleSubscriptionEvent struct {
	SubscriptionID string `json:"subscription_id"`
	Status         string `json:"status"`
	UserID         string `json:"user_id"`
	PlanID         string `json:"plan_id"`
	NextBillDate   string `json:"next_bill_date"`
	CancelURL      string `json:"cancel_url"`
	UpdateURL      string `json:"update_url"`
}

func NewRealPaddleService(cfg *config.Config) *PaddleService {
	return &PaddleService{cfg: cfg}
}

func (s *PaddleService) CreateCheckoutSession(userID uint, planID string) (string, error) {
	priceID, ok := s.cfg.PaddlePriceIDs[planID]
	if !ok {
		return "", fmt.Errorf("invalid plan ID")
	}

	// In production, you would make an API call to Paddle to create a checkout session
	// For now, we'll return the direct checkout URL
	checkoutURL := fmt.Sprintf(
		"https://checkout.paddle.com/checkout/%s?user_id=%d&vendor=%s",
		priceID,
		userID,
		s.cfg.PaddleVendorID,
	)

	return checkoutURL, nil
}

func (s *PaddleService) VerifyWebhookSignature(payload []byte, signature string) bool {
	if s.cfg.Environment == "development" {
		return true // Skip verification in development
	}

	// Create HMAC SHA256 hash using your public key
	mac := hmac.New(sha256.New, []byte(s.cfg.PaddlePublicKey))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (s *PaddleService) HandleWebhookEvent(payload []byte) error {
	var event struct {
		AlertName string          `json:"alert_name"`
		Data      json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %v", err)
	}

	switch event.AlertName {
	case "subscription_created":
		return s.handleSubscriptionCreated(event.Data)
	case "subscription_updated":
		return s.handleSubscriptionUpdated(event.Data)
	case "subscription_cancelled":
		return s.handleSubscriptionCancelled(event.Data)
	case "subscription_payment_succeeded":
		return s.handlePaymentSucceeded(event.Data)
	case "subscription_payment_failed":
		return s.handlePaymentFailed(event.Data)
	}

	return nil
}

func (s *PaddleService) handleSubscriptionCreated(data json.RawMessage) error {
	var subEvent PaddleSubscriptionEvent
	if err := json.Unmarshal(data, &subEvent); err != nil {
		return fmt.Errorf("failed to parse subscription data: %v", err)
	}

	// Parse user ID from string to uint
	var userID uint
	fmt.Sscanf(subEvent.UserID, "%d", &userID)

	// Create new subscription record
	subscription := models.Subscription{
		UserID:           userID,
		PaddleSubID:      subEvent.SubscriptionID,
		PlanID:           subEvent.PlanID,
		Status:           "active",
		CurrentPeriodEnd: parseDate(subEvent.NextBillDate),
	}

	return db.DB.Create(&subscription).Error
}

func (s *PaddleService) handleSubscriptionUpdated(data json.RawMessage) error {
	var subEvent PaddleSubscriptionEvent
	if err := json.Unmarshal(data, &subEvent); err != nil {
		return fmt.Errorf("failed to parse subscription data: %v", err)
	}

	return db.DB.Model(&models.Subscription{}).
		Where("paddle_sub_id = ?", subEvent.SubscriptionID).
		Updates(map[string]interface{}{
			"status":             subEvent.Status,
			"current_period_end": parseDate(subEvent.NextBillDate),
		}).Error
}

func (s *PaddleService) handleSubscriptionCancelled(data json.RawMessage) error {
	var subEvent PaddleSubscriptionEvent
	if err := json.Unmarshal(data, &subEvent); err != nil {
		return fmt.Errorf("failed to parse subscription data: %v", err)
	}

	return db.DB.Model(&models.Subscription{}).
		Where("paddle_sub_id = ?", subEvent.SubscriptionID).
		Update("status", "cancelled").Error
}

func (s *PaddleService) handlePaymentSucceeded(data json.RawMessage) error {
	var subEvent PaddleSubscriptionEvent
	if err := json.Unmarshal(data, &subEvent); err != nil {
		return fmt.Errorf("failed to parse payment data: %v", err)
	}

	return db.DB.Model(&models.Subscription{}).
		Where("paddle_sub_id = ?", subEvent.SubscriptionID).
		Updates(map[string]interface{}{
			"status":             "active",
			"current_period_end": parseDate(subEvent.NextBillDate),
		}).Error
}

func (s *PaddleService) handlePaymentFailed(data json.RawMessage) error {
	var subEvent PaddleSubscriptionEvent
	if err := json.Unmarshal(data, &subEvent); err != nil {
		return fmt.Errorf("failed to parse payment data: %v", err)
	}

	return db.DB.Model(&models.Subscription{}).
		Where("paddle_sub_id = ?", subEvent.SubscriptionID).
		Update("status", "past_due").Error
}

func (s *PaddleService) CancelSubscription(subscriptionID string) error {
	if s.cfg.Environment == "development" {
		// In development, just update the database
		return db.DB.Model(&models.Subscription{}).
			Where("paddle_sub_id = ?", subscriptionID).
			Update("status", "cancelled").Error
	}

	// In production, make API call to Paddle
	// TODO: Implement real API call
	return nil
}

func parseDate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Now().AddDate(0, 1, 0) // Default to 1 month from now
	}
	return t
}
