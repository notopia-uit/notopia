package authorization

import (
	"context"
	"log/slog"
	"time"
)

type Server struct {
	grpc   *GRPCServer
	logger *slog.Logger
}

func NewServer(
	grpc *GRPCServer,
	logger *slog.Logger,
) *Server {
	slog.SetDefault(logger)

	return &Server{
		grpc:   grpc,
		logger: logger,
	}
}

var ProvideServer = NewServer

func (s *Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.grpc.Shutdown(shutdownCtx); err != nil {
			s.logger.ErrorContext(ctx, "failed to shutdown grpc server", slog.String("error", err.Error()))
		}
	}()
	return s.grpc.Run()
}
