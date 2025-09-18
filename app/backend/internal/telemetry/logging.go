package telemetry

import (
	"log/slog"
	"os"
)

// NewLogger создает новый структурированный логгер
func NewLogger(env string) *slog.Logger {
	var handler slog.Handler

	if env == "prod" {
		// JSON формат для продакшена
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// Текстовый формат для разработки
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return slog.New(handler)
}
