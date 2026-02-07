package main

import (
	"context"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()
	server, cleanup, err := InitializeServer(ctx)
	defer cleanup()
	if err != nil {
		slog.Error("failed to initialize server", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := server.Run(); err != nil {
		slog.Error("server encountered an error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
