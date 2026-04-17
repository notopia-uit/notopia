package authorization

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/authorization/controller/grpc"
	"github.com/notopia-uit/notopia/internal/authorization/controller/health"
	"github.com/notopia-uit/notopia/pkg/otel"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	grpc   *grpc.Server
	health *health.Health
	logger *slog.Logger
}

func NewServer(
	grpc *grpc.Server,
	health *health.Health,
	logger *slog.Logger,
	globalOtel otel.Global,
) *Server {
	slog.SetDefault(logger)

	return &Server{
		grpc:   grpc,
		health: health,
		logger: logger,
	}
}

var ProvideServer = NewServer

func (s *Server) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			s.grpc.Stop()
		}()
		if err := s.grpc.Run(); err != nil {
			return fmt.Errorf("failed to run grpc server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			if err := s.health.Shutdown(context.Background()); err != nil {
				s.logger.ErrorContext(ctx, "failed to shutdown health server", slog.Any("error", err))
			}
		}()
		if err := s.health.Run(); err != nil {
			return fmt.Errorf("failed to run health server: %w", err)
		}
		return nil
	})

	return g.Wait()
}
