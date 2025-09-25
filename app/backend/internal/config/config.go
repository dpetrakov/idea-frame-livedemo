package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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

	// Email verification
	AxenixEmailDomain    string
	EmailCodesTTLMinutes int

	// Admin roles
	AdminEmails []string

	// SMTP settings (preferred)
	SMTPHost                  string
	SMTPPort                  int
	SMTPUsername              string
	SMTPPassword              string
	SMTPFrom                  string
	SMTPTLSServerName         string
	SMTPTLSInsecureSkipVerify bool
	SMTPUseSSL                bool
	SMTPEhloDomain            string
}

// Load читает конфигурацию из переменных окружения
func Load() (*Config, error) {
	cfg := &Config{
		Port:                 getEnv("BACKEND_PORT", "8080"),
		DatabaseURL:          getEnvRequired("DATABASE_URL"),
		JWTSecret:            getEnvRequired("JWT_SECRET"),
		Environment:          getEnv("ENV", "dev"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		JWTExpiration:        24 * time.Hour, // 24 часа как требуется в спецификации
		AxenixEmailDomain:    getEnv("AXENIX_EMAIL_DOMAIN", "axenix.pro"),
		EmailCodesTTLMinutes: getEnvInt("EMAIL_CODES_TTL_MINUTES", 10),

		// SMTP
		SMTPHost:                  getEnv("SMTP_HOST", ""),
		SMTPPort:                  getEnvInt("SMTP_PORT", 587),
		SMTPUsername:              getEnv("SMTP_USERNAME", ""),
		SMTPPassword:              getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:                  getEnv("SMTP_FROM", ""),
		SMTPTLSServerName:         getEnv("SMTP_TLS_SERVER_NAME", ""),
		SMTPTLSInsecureSkipVerify: getEnvBool("SMTP_TLS_INSECURE_SKIP_VERIFY", false),
		SMTPUseSSL:                getEnvBool("SMTP_USE_SSL", false),
		SMTPEhloDomain:            getEnv("SMTP_EHLO_DOMAIN", "localhost"),
	}

	// Валидация JWT секрета
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	// Валидация порта
	if port, err := strconv.Atoi(cfg.Port); err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid BACKEND_PORT: %s", cfg.Port)
	}

	// Нормализуем домен в нижний регистр
	cfg.AxenixEmailDomain = strings.ToLower(cfg.AxenixEmailDomain)

	// Парсим ADMIN_EMAILS: запятая-разделитель, без пробелов, в нижний регистр
	adminCSV := getEnv("ADMIN_EMAILS", "")
	if adminCSV != "" {
		parts := strings.Split(adminCSV, ",")
		cfg.AdminEmails = make([]string, 0, len(parts))
		for _, p := range parts {
			e := strings.ToLower(strings.TrimSpace(p))
			if e != "" {
				cfg.AdminEmails = append(cfg.AdminEmails, e)
			}
		}
	} else {
		cfg.AdminEmails = []string{}
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

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if i, err := strconv.Atoi(v); err == nil {
		return i
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}
