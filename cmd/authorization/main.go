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
	if err := logging.FirstStart("NOTOPIA_AUTHORIZATION_LOG_LEVEL"); err != nil {
		slog.Error("failed to set up logger", slog.Any("error", err))
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server, cleanup, err := InitializeServer(ctx)
	if err != nil {
		slog.Error("failed to initialize server", slog.Any("error", err))
		if cleanup != nil {
			cleanup()
		}
		return
	}
	defer cleanup()
	if err := server.Run(ctx); err != nil {
		slog.Error("server encountered an error", slog.Any("error", err))
	}
}
