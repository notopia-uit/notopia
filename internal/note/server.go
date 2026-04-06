package note

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/controller/grpc"
	"github.com/notopia-uit/notopia/internal/note/controller/health"
	"github.com/notopia-uit/notopia/internal/note/controller/http"
	"github.com/notopia-uit/notopia/internal/note/controller/integrationevent"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence"
	"github.com/notopia-uit/notopia/pkg/otel"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	persistence      *persistence.Pg
	http             *http.HTTP
	grpc             *grpc.GRPC
	integrationEvent *integrationevent.IntegrationEvent
	health           *health.Health
	app              *app.Server
	logger           *slog.Logger
}

func NewServer(
	persistence *persistence.Pg,
	http *http.HTTP,
	grpc *grpc.GRPC,
	integrationEvent *integrationevent.IntegrationEvent,
	health *health.Health,
	app *app.Server,
	logger *slog.Logger,
	globalOtel otel.Global, // This have to be here for deps
) *Server {
	slog.SetDefault(logger)

	return &Server{
		persistence:      persistence,
		http:             http,
		grpc:             grpc,
		integrationEvent: integrationEvent,
		health:           health,
		app:              app,
		logger:           logger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	if err := s.persistence.RunMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			if err := s.http.Shutdown(context.Background()); err != nil {
				s.logger.ErrorContext(ctx, "failed to shutdown http server", slog.Any("error", err))
			}
		}()
		if err := s.http.Run(); err != nil {
			return fmt.Errorf("failed to run http server: %w", err)
		}
		return nil
	})

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

	g.Go(func() error {
		if err := s.integrationEvent.Run(ctx); err != nil {
			return fmt.Errorf("failed to run integration event listener: %w", err)
		}
		return nil
	})

	return g.Wait()
}

var ProvideServer = NewServer
