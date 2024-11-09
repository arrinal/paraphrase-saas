package config

import (
	"os"
)

type Config struct {
	JWTSecret        string
	DatabaseURL      string
	ServerPort       string
	OpenAIKey        string
	FrontendURL      string
	Environment      string // 'development' or 'production'
	PaddleVendorID   string
	PaddlePublicKey  string
	PaddleProPriceID string
}

func LoadConfig() (*Config, error) {
	return &Config{
		JWTSecret:        getEnvOrDefault("JWT_SECRET", "your-default-secret"),
		DatabaseURL:      getEnvOrDefault("DATABASE_URL", "postgresql://postgres@localhost:5432/frazai_db"),
		ServerPort:       getEnvOrDefault("PORT", "8080"),
		OpenAIKey:        getEnvOrDefault("OPENAI_API_KEY", ""),
		FrontendURL:      getEnvOrDefault("FRONTEND_URL", "http://localhost:3000"),
		Environment:      getEnvOrDefault("ENVIRONMENT", "development"),
		PaddleVendorID:   getEnvOrDefault("PADDLE_VENDOR_ID", ""),
		PaddlePublicKey:  getEnvOrDefault("PADDLE_PUBLIC_KEY", ""),
		PaddleProPriceID: getEnvOrDefault("PADDLE_PRO_PRICE_ID", ""),
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
