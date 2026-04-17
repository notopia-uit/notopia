package logging

import (
	"log/slog"
	"os"

	"github.com/notopia-uit/notopia/pkg/helper"
)

func FirstStart(envVar string) error {
	var level slog.Level
	var err error
	level, err = helper.GetLogLevelFromString(os.Getenv(envVar))
	if err != nil {
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
	return nil
}
