package logging

import (
	"log/slog"
	"os"

	"github.com/notopia-uit/notopia/pkg/helper"
)

func FirstStart() error {
	level, err := helper.GetLogLevelFromString(os.Getenv("LOG_LEVEL"))
	if err != nil {
		return err
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
	return nil
}
