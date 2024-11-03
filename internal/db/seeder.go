package db

import (
	"encoding/json"

	"github.com/arrinal/paraphrase-saas/internal/models"
)

func SeedSubscriptionPlans() error {
	plans := []models.SubscriptionPlan{
		{
			ID:           "basic",
			Name:         "Basic",
			Price:        999, // $9.99
			Currency:     "USD",
			Interval:     "month",
			PaddlePlanID: "pri_basic",
			Features: models.JSON(mustMarshal([]string{
				"Up to 1,000 characters per request",
				"50 requests per day",
				"Standard paraphrasing styles",
			})),
			Limits: models.JSON(mustMarshal(map[string]interface{}{
				"charactersPerRequest": 1000,
				"requestsPerDay":       50,
				"bulkParaphrase":       false,
			})),
		},
		{
			ID:           "pro",
			Name:         "Professional",
			Price:        1999, // $19.99
			Currency:     "USD",
			Interval:     "month",
			PaddlePlanID: "pri_pro",
			Features: models.JSON(mustMarshal([]string{
				"Up to 5,000 characters per request",
				"Unlimited requests",
				"All paraphrasing styles",
				"Bulk paraphrasing",
			})),
			Limits: models.JSON(mustMarshal(map[string]interface{}{
				"charactersPerRequest": 5000,
				"requestsPerDay":       -1,
				"bulkParaphrase":       true,
			})),
		},
	}

	for _, plan := range plans {
		if err := DB.Where(models.SubscriptionPlan{ID: plan.ID}).
			Assign(plan).
			FirstOrCreate(&plan).Error; err != nil {
			return err
		}
	}

	return nil
}

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
