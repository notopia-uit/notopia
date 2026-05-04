package authorization

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/notopia-uit/notopia/internal/authorization/controller/grpc"
	"github.com/notopia-uit/notopia/internal/authorization/controller/health"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/otel"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	grpc   *grpc.Server
	health *health.Health
	logger *slog.Logger
	app    *app.App
	appEnv commonconfig.AppEnv
}

func NewServer(
	ctx context.Context,
	grpc *grpc.Server,
	health *health.Health,
	logger *slog.Logger,
	globalOtel otel.Global,
	generalCfg *commonconfig.General,
	authApp *app.App,
) (*Server, error) {
	slog.SetDefault(logger)

	if err := authApp.BootStrapPolicies(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap policies: %w", err)
	}
	if generalCfg.AppEnv == commonconfig.AppEnvDevelopment {
		if err := authApp.SeedDev(ctx); err != nil {
			return nil, fmt.Errorf("failed to seed dev policies: %w", err)
		}
	}

	return &Server{
		grpc:   grpc,
		health: health,
		logger: logger,
		app:    authApp,
		appEnv: generalCfg.AppEnv,
	}, nil
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
