package main

import (
	"context"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()
	server, err := InitializeServer(ctx)
	if err != nil {
		slog.Error("failed to initialize server", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := server.Run(); err != nil {
		slog.Error("server encountered an error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
