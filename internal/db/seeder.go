package db

import (
	"encoding/json"

	"github.com/arrinal/paraphrase-saas/internal/models"
)

func SeedSubscriptionPlans() error {
	plans := []models.SubscriptionPlan{
		{
			ID:           "pro",
			Name:         "Pro",
			Price:        500, // $5.00
			Currency:     "USD",
			Interval:     "month",
			PaddlePlanID: "pro",
			Features: models.JSON(mustMarshal([]string{
				"Paraphrase in any language (auto-detect)",
				"Paraphrase and translate at the same time",
				"Unlimited paraphrase with AI",
				"All paraphrasing styles",
			})),
			Limits: models.JSON(mustMarshal(map[string]interface{}{
				"charactersPerRequest": 10000, // unlimited
				"requestsPerDay":       -1,    // unlimited
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
