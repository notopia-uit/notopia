package main

import (
	"context"
	"log/slog"

	"github.com/notopia-uit/notopia/pkg/logging"
)

func main() {
	if err := logging.FirstStart("NOTOPIA_NOTE_LOG_LEVEL"); err != nil {
		slog.Error("failed to set up logger", slog.Any("error", err))
		return
	}
	ctx := context.Background()
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
