package authorization

import (
	"context"
	"log/slog"

	"github.com/notopia-uit/notopia/pkg/otel"
)

type Server struct {
	grpc   *GRPCServer
	logger *slog.Logger
}

func NewServer(
	grpc *GRPCServer,
	logger *slog.Logger,
	globalOtel otel.Global,
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
