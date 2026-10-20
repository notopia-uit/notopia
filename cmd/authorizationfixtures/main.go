package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/notopia-uit/notopia/pkg/logging"
)

func main() {
	if err := logging.FirstStart("NOTOPIA_AUTHORIZATIONFIXTURES_LOG_LEVEL"); err != nil {
		slog.Error("failed to set up logger", slog.Any("error", err))
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		slog.Error("authorization fixtures failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("authorization fixtures done")
}

func run(ctx context.Context) error {
	fixtures, cleanup, err := InitializeFixtures(ctx)
	if err != nil {
		return err
	}
	defer cleanup()
	return fixtures.Run(ctx)
}
