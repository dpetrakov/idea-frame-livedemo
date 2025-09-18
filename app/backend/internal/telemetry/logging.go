package telemetry

import (
	"log/slog"
	"os"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/config"
)

// NewLogger создает новый структурированный логгер
func NewLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	
	switch cfg.Env {
	case "dev":
		level = slog.LevelDebug
	case "prod":
		level = slog.LevelInfo
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}