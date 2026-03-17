package authorization

import (
	"context"
	"log/slog"
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
		if err := s.grpc.Shutdown(context.Background()); err != nil {
			s.logger.ErrorContext(ctx, "failed to shutdown grpc server", slog.String("error", err.Error()))
		}
	}()
	return s.grpc.Run()
}
