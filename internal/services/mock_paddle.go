package services

import (
	"fmt"
	"time"

	"github.com/arrinal/paraphrase-saas/internal/config"
	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
)

type MockPaddleService struct {
	cfg *config.Config
}

func NewMockPaddleService(cfg *config.Config) *MockPaddleService {
	return &MockPaddleService{cfg: cfg}
}

func (s *MockPaddleService) CreateCheckoutSession(userID uint, planID string) (string, error) {
	// Create mock subscription in database immediately
	subscription := models.Subscription{
		UserID:           userID,
		PlanID:           planID,
		Status:           "active",
		PaddleSubID:      fmt.Sprintf("mock_sub_%d_%s", userID, planID),
		CurrentPeriodEnd: time.Now().AddDate(0, 1, 0), // 1 month from now
	}

	// Save to database
	if err := db.DB.Create(&subscription).Error; err != nil {
		return "", fmt.Errorf("failed to create subscription: %v", err)
	}

	// Return mock checkout URL
	return fmt.Sprintf("%s/checkout/success?session_id=mock_session_%d_%s",
		s.cfg.FrontendURL, userID, planID), nil
}

func (s *MockPaddleService) VerifyWebhookSignature(payload []byte, signature string) bool {
	// Always return true for mock implementation
	return true
}

func (s *MockPaddleService) HandleWebhookEvent(payload []byte) error {
	// No need to handle webhooks in mock implementation since we create the subscription directly
	return nil
}

func (s *MockPaddleService) CancelSubscription(subscriptionID string) error {
	// Update subscription status in database
	if err := db.DB.Model(&models.Subscription{}).
		Where("paddle_sub_id = ?", subscriptionID).
		Update("status", "cancelled").Error; err != nil {
		return fmt.Errorf("failed to cancel subscription: %v", err)
	}
	return nil
}
