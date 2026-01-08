package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL string

	// DingTalk
	DingTalkAppKey      string
	DingTalkAppSecret   string
	ApprovalProcessCode string

	// Auth API (external)
	AuthAPIBaseURL  string
	JWTSecret       string
	JWTAccessSecret string

	// Ollama (Local LLM)
	OllamaBaseURL string
	OllamaModel   string

	// Timezone
	Location *time.Location
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	// Parse timezone
	loc, err := time.LoadLocation(getEnv("TZ", "Asia/Jakarta"))
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60) // UTC+7 fallback
	}

	return &Config{
		Port:                getEnv("PORT", "8087"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:allure2025@localhost:5434/ncr_dashboard?sslmode=disable"),
		DingTalkAppKey:      os.Getenv("DINGTALK_APP_KEY"),
		DingTalkAppSecret:   os.Getenv("DINGTALK_APP_SECRET"),
		ApprovalProcessCode: os.Getenv("APPROVAL_PROCESS_CODE"),
		AuthAPIBaseURL:      getEnv("AUTH_API_BASE_URL", "https://api-incoming.ws-allure.com"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		JWTAccessSecret:     os.Getenv("JWT_ACCESS_SECRET"),
		OllamaBaseURL:       getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaModel:         getEnv("OLLAMA_MODEL", "llama3.2:3b"),
		Location:            loc,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
