package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/arrinal/paraphrase-saas/internal/config"
	"github.com/arrinal/paraphrase-saas/internal/models"
)

type IAPReceipt struct {
	TransactionID string `json:"transaction_id"`
	ProductID     string `json:"product_id"`
	Receipt       string `json:"receipt"`
	Environment   string `json:"environment"` // sandbox or production
}

type IAPVerifyResponse struct {
	Status            int `json:"status"`
	LatestReceiptInfo []struct {
		ProductID                   string `json:"product_id"`
		TransactionID               string `json:"transaction_id"`
		ExpiresDateMS               string `json:"expires_date_ms"`
		PurchaseDateMS              string `json:"purchase_date_ms"`
		SubscriptionGroupIdentifier string `json:"subscription_group_identifier"`
	} `json:"latest_receipt_info"`
}

type IOSPaymentService struct {
	cfg *config.Config
}

func NewIOSPaymentService(cfg *config.Config) *IOSPaymentService {
	return &IOSPaymentService{cfg: cfg}
}

func (s *IOSPaymentService) VerifyReceipt(receipt IAPReceipt) (*models.Subscription, error) {
	if s.cfg.Environment == "development" {
		return s.mockVerifyReceipt(receipt)
	}

	verifyURL := "https://buy.itunes.apple.com/verifyReceipt"
	if receipt.Environment == "sandbox" {
		verifyURL = "https://sandbox.itunes.apple.com/verifyReceipt"
	}

	// Make request to Apple's verification server
	response, err := http.Post(verifyURL, "application/json", bytes.NewBuffer([]byte(receipt.Receipt)))
	if err != nil {
		return nil, fmt.Errorf("failed to verify receipt: %v", err)
	}
	defer response.Body.Close()

	var verifyResponse IAPVerifyResponse
	if err := json.NewDecoder(response.Body).Decode(&verifyResponse); err != nil {
		return nil, fmt.Errorf("failed to decode verify response: %v", err)
	}

	// Process verification response
	if verifyResponse.Status != 0 {
		return nil, fmt.Errorf("receipt verification failed with status: %d", verifyResponse.Status)
	}

	// Create subscription from receipt info
	// ... implementation continues
	return nil, nil
}

func (s *IOSPaymentService) mockVerifyReceipt(receipt IAPReceipt) (*models.Subscription, error) {
	// Map Apple product IDs to our plan IDs
	planMap := map[string]string{
		"com.frazai.basic": "basic",
		"com.frazai.pro":   "pro",
	}

	planID, ok := planMap[receipt.ProductID]
	if !ok {
		return nil, fmt.Errorf("invalid product ID: %s", receipt.ProductID)
	}

	// Create mock subscription
	subscription := &models.Subscription{
		PlanID:           planID,
		Status:           "active",
		PaddleSubID:      receipt.TransactionID,       // Use transaction ID as subscription ID
		CurrentPeriodEnd: time.Now().AddDate(0, 1, 0), // 1 month from now
	}

	return subscription, nil
}
