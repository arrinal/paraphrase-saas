package services

import (
	"fmt"
	"time"

	"github.com/arrinal/paraphrase-saas/internal/config"
	"github.com/arrinal/paraphrase-saas/internal/models"
)

type MockPaddleService struct {
	cfg *config.Config
}

func NewMockPaddleService(cfg *config.Config) *MockPaddleService {
	return &MockPaddleService{cfg: cfg}
}

func (s *MockPaddleService) CreateCheckoutSession(userID uint, planID string) (string, error) {
	// Return a local mock checkout URL instead of Paddle's URL
	return fmt.Sprintf("%s/checkout/success?session_id=mock_session_%d_%s", s.cfg.FrontendURL, userID, planID), nil
}

func (s *MockPaddleService) VerifyWebhookSignature(payload []byte, signature string) bool {
	// For mock implementation, always return true
	return true
}

func (s *MockPaddleService) HandleWebhookEvent(payload []byte) error {
	// For mock implementation, always succeed
	return nil
}

func (s *MockPaddleService) CancelSubscription(subscriptionID string) error {
	// For mock implementation, always succeed
	return nil
}

func (s *MockPaddleService) ActivateSubscription(userID uint, planID string) (*models.Subscription, error) {
	// Create a mock subscription
	subscription := &models.Subscription{
		UserID:           userID,
		PlanID:           planID,
		Status:           "active",
		PaddleSubID:      fmt.Sprintf("mock_sub_%d_%s", userID, planID),
		CurrentPeriodEnd: time.Now().AddDate(0, 1, 0), // 1 month from now
	}
	return subscription, nil
}
