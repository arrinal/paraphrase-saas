package db

import (
	"log"

	"github.com/arrinal/paraphrase-saas/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Initialize(dbURL string) error {
	log.Printf("Connecting to database: %s", dbURL)

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return err
	}

	// Auto migrate the schemas
	log.Println("Running database migrations...")
	err = db.AutoMigrate(
		&models.User{},
		&models.ParaphraseHistory{},
		&models.Subscription{},
		&models.SubscriptionPlan{},
		&models.UserStats{},
		&models.DailyUsage{},
	)
	if err != nil {
		return err
	}

	DB = db
	log.Println("Database initialized successfully")
	return nil
}
