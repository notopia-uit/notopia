package main

import (
	"context"
	"log/slog"
)

func main() {
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
