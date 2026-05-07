package logging

import (
	"log/slog"
	"os"
)

func FirstStart(envVar string) error {
	var level slog.Level
	err := level.UnmarshalText([]byte(os.Getenv(envVar)))
	if err != nil {
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
	return nil
}
