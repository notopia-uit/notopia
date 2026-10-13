package billing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/controller/event"
	"github.com/notopia-uit/notopia/internal/billing/controller/grpc"
	"github.com/notopia-uit/notopia/internal/billing/controller/health"
	"github.com/notopia-uit/notopia/internal/billing/controller/http"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence"
	"github.com/notopia-uit/notopia/pkg/otel"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	app         *app.Server
	persistence *persistence.Pg
	http        *http.HTTP
	grpc        *grpc.GRPC
	event       *event.Event
	health      *health.Health
	logger      *slog.Logger
}

func NewServer(
	app *app.Server,
	persistence *persistence.Pg,
	http *http.HTTP,
	grpc *grpc.GRPC,
	event *event.Event,
	health *health.Health,
	logger *slog.Logger,
	_ otel.Global,
) *Server {
	slog.SetDefault(logger)

	return &Server{
		app:         app,
		persistence: persistence,
		http:        http,
		grpc:        grpc,
		event:       event,
		health:      health,
		logger:      logger,
	}
}

var ProvideServer = NewServer

func (s *Server) Run(ctx context.Context) error {
	if err := s.persistence.RunMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	if err := s.app.Cmds.EnsureDefaults.Handle(ctx, &app.EnsureDefaults{}); err != nil {
		return fmt.Errorf("failed to ensure billing defaults: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	httpShutdownTimeout := 10 * time.Second
	grpcShutdownTimeout := 10 * time.Second
	healthShutdownTimeout := 5 * time.Second

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
			defer cancel()
			if err := s.http.Shutdown(shutdownCtx); err != nil {
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
			done := make(chan struct{})
			go func() {
				s.grpc.GracefulStop()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(grpcShutdownTimeout):
				s.logger.WarnContext(ctx, "grpc shutdown timed out")
				s.grpc.Stop()
			}
		}()
		if err := s.grpc.Run(); err != nil {
			return fmt.Errorf("failed to run grpc server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := s.event.Run(ctx); err != nil {
			return fmt.Errorf("failed to run event router: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), healthShutdownTimeout)
			defer cancel()
			if err := s.health.Shutdown(shutdownCtx); err != nil {
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
