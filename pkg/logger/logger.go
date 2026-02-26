package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/mal-as/aiqoder/internal/config"
)

func New(cfg config.Log) *slog.Logger {
	envLogLevelMap := map[string]slog.Level{
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
		"debug": slog.LevelDebug,
	}

	level, ok := envLogLevelMap[strings.ToLower(cfg.Level)]
	if !ok {
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})

	return slog.New(handler)
}
