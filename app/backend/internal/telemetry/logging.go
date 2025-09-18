package telemetry

import (
	"context"
	"log/slog"
	"os"
)

// NewLogger создаёт новый структурированный логгер
func NewLogger(level string, env string) *slog.Logger {
	var logLevel slog.Level
	
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	
	var handler slog.Handler
	if env == "prod" || env == "production" {
		// JSON логи для продакшена
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Читаемые логи для разработки
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	
	return slog.New(handler)
}

// ContextKey тип для ключей контекста
type ContextKey string

const (
	// LoggerKey ключ для логгера в контексте
	LoggerKey ContextKey = "logger"
)

// LoggerFromContext получает логгер из контекста
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	}
	// Возвращаем стандартный логгер если не найден
	return slog.Default()
}
