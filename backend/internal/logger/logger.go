package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New builds the structured JSON logger. Production uses info level,
// every other environment keeps debug output for local development.
func New(environment string) *slog.Logger {
	level := slog.LevelDebug
	if strings.EqualFold(environment, "production") {
		level = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
