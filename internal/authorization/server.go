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
		s.grpc.Stop()
	}()
	return s.grpc.Run()
}
