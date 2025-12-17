package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env  string
	DSN  string
}

func ConfigLoad() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("ENV", "development"),
		DSN:  getEnv("DB_DSN", ""),
	}

	if cfg.DSN == "" {
		return nil, errors.New("DB_DSN is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return fallback
}
