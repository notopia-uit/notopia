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
	if err := logging.FirstStart("NOTOPIA_BILLINGSEED_LOG_LEVEL"); err != nil {
		slog.Error("failed to set up logger", slog.Any("error", err))
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		slog.Error("billing seed failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("billing seed done")
}

func run(ctx context.Context) error {
	seeder, cleanup, err := InitializeSeeder(ctx)
	if err != nil {
		return err
	}
	defer cleanup()
	return seeder.Run(ctx)
}
