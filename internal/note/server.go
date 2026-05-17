package note

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/notopia-uit/notopia/internal/note/controller/event"
	"github.com/notopia-uit/notopia/internal/note/controller/grpc"
	"github.com/notopia-uit/notopia/internal/note/controller/health"
	"github.com/notopia-uit/notopia/internal/note/controller/http"
	"github.com/notopia-uit/notopia/internal/note/infra/outbox"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence"
	"github.com/notopia-uit/notopia/internal/note/infra/workspaceevent"
	"github.com/notopia-uit/notopia/pkg/otel"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	persistence       *persistence.Pg
	http              *http.HTTP
	grpc              *grpc.GRPC
	event             *event.Event
	workspaceEventHub *workspaceevent.WorkspaceEventHub
	outbox            *outbox.Outbox
	health            *health.Health
	logger            *slog.Logger
}

// TODO: we have to start the workspace event also
func NewServer(
	persistence *persistence.Pg,
	http *http.HTTP,
	grpc *grpc.GRPC,
	event *event.Event,
	workspaceEventHub *workspaceevent.WorkspaceEventHub,
	outbox *outbox.Outbox,
	health *health.Health,
	logger *slog.Logger,
	globalOtel otel.Global, // This have to be here for deps
) *Server {
	slog.SetDefault(logger)

	return &Server{
		persistence:       persistence,
		http:              http,
		grpc:              grpc,
		event:             event,
		workspaceEventHub: workspaceEventHub,
		outbox:            outbox,
		health:            health,
		logger:            logger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	if err := s.persistence.RunMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
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

	g.Go(func() error {
		// This has context passed down, so we don't really to close/stop it
		if err := s.event.Run(ctx); err != nil {
			return fmt.Errorf("failed to run integration event listener: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := s.workspaceEventHub.Run(ctx); err != nil {
			return fmt.Errorf("failed to run workspace event hub: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		// This has context passed down, so we don't really to close/stop it
		if err := s.outbox.Run(ctx); err != nil {
			return fmt.Errorf("failed to run outbox forwarder: %w", err)
		}
		return nil
	})

	return g.Wait()
}

var ProvideServer = NewServer
