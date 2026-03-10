package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	CORSOrigins []string

	UploadDir string

	// External services
	AnthropicAPIKey string
	OCRProvider     string
	OCRAPIKey       string
	MISAApiURL      string
	MISAApiKey      string
	MISAAppID       string
	SentryDSN       string
	ResendAPIKey    string
}

func Load() (*Config, error) {
	// Load .env file if exists (ignored in production)
	_ = godotenv.Load()

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		CORSOrigins:     []string{getEnv("CORS_ORIGINS", "http://localhost:3000")},
		UploadDir:       getEnv("UPLOAD_DIR", "./data/uploads"),
		AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
		OCRProvider:     getEnv("OCR_PROVIDER", "fpt"),
		OCRAPIKey:       getEnv("OCR_API_KEY", ""),
		MISAApiURL:      getEnv("MISA_API_URL", ""),
		MISAApiKey:      getEnv("MISA_API_KEY", ""),
		MISAAppID:       getEnv("MISA_APP_ID", ""),
		SentryDSN:       getEnv("SENTRY_DSN", ""),
		ResendAPIKey:    getEnv("RESEND_API_KEY", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
