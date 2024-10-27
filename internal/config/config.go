package config

import (
	"os"
)

type Config struct {
	JWTSecret   string
	DatabaseURL string
	ServerPort  string
	OpenAIKey   string
	// Add other config fields as needed
}

func LoadConfig() (*Config, error) {
	return &Config{
		JWTSecret:   getEnvOrDefault("JWT_SECRET", "your-default-secret"),
		DatabaseURL: getEnvOrDefault("DATABASE_URL", "postgresql://postgres@localhost:5432/paraphrase_db"),
		ServerPort:  getEnvOrDefault("PORT", "8080"),
		OpenAIKey:   getEnvOrDefault("OPENAI_API_KEY", ""),
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
