package config

import (
	"log/slog"
	"os"
)

// NewLogger builds the logger for the given environment: structured JSON at info
// level in production, human-readable text at debug level everywhere else.
func NewLogger(cfg BaseConfig) *slog.Logger {
	if cfg.Env.Is(Production) {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}
