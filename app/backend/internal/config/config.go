package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	Environment string
	LogLevel    string
	
	// Computed values
	JWTExpiration time.Duration
}

// Load читает конфигурацию из переменных окружения
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("BACKEND_PORT", "8080"),
		DatabaseURL: getEnvRequired("DATABASE_URL"),
		JWTSecret:   getEnvRequired("JWT_SECRET"),
		Environment: getEnv("ENV", "dev"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		JWTExpiration: 24 * time.Hour, // 24 часа как требуется в спецификации
	}
	
	// Валидация JWT секрета
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	
	// Валидация порта
	if port, err := strconv.Atoi(cfg.Port); err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid BACKEND_PORT: %s", cfg.Port)
	}
	
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return value
}