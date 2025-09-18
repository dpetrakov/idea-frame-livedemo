package config

import (
	"log"
	"os"
)

// Config содержит настройки приложения
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	Env         string
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		Env:         getEnv("ENV", "dev"),
	}

	// Критические перемены - fail fast
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	return cfg
}

// getEnv получает переменную окружения с fallback значением
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
