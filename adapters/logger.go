package adapters

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "DEV" {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}
