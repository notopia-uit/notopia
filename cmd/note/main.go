package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/notopia-uit/notopia/pkg/common/helper"
)

func main() {
	if err := setupInitLogger(); err != nil {
		slog.Error("failed to set up logger", slog.String("error", err.Error()))
		return
	}
	ctx := context.Background()
	server, cleanup, err := InitializeServer(ctx)
	defer cleanup()
	if err != nil {
		slog.Error("failed to initialize server", slog.String("error", err.Error()))
	}
	if err := server.Run(ctx); err != nil {
		slog.Error("server encountered an error", slog.String("error", err.Error()))
	}
}

func setupInitLogger() error {
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
