package note

import (
	"context"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/controller/grpc"
	"github.com/notopia-uit/notopia/internal/note/controller/health"
	"github.com/notopia-uit/notopia/internal/note/controller/http"
	"github.com/notopia-uit/notopia/internal/note/controller/integrationevent"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	http             *http.HTTP
	grpc             *grpc.GRPC
	integrationevent *integrationevent.IntegrationEvent
	health           *health.Health
	application      *app.App
	logger           *slog.Logger
}

func NewServer(
	http *http.HTTP,
	grpc *grpc.GRPC,
	integrationevent *integrationevent.IntegrationEvent,
	health *health.Health,
	application *app.App,
	logger *slog.Logger,
) *Server {
	slog.SetDefault(logger)

	return &Server{
		http:             http,
		grpc:             grpc,
		integrationevent: integrationevent,
		health:           health,
		application:      application,
		logger:           logger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	if err := s.application.Start(ctx); err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			if err := s.http.Shutdown(context.Background()); err != nil {
				s.logger.ErrorContext(ctx, "failed to shutdown http server", slog.String("error", err.Error()))
			}
		}()
		return s.http.Run()
	})

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			if err := s.grpc.Shutdown(context.Background()); err != nil {
				s.logger.ErrorContext(ctx, "failed to shutdown grpc server", slog.String("error", err.Error()))
			}
		}()
		return s.grpc.Run()
	})

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			if err := s.health.Shutdown(context.Background()); err != nil {
				s.logger.ErrorContext(ctx, "failed to shutdown health server", slog.String("error", err.Error()))
			}
		}()
		return s.health.Run()
	})

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			if err := s.integrationevent.Close(); err != nil {
				s.logger.ErrorContext(ctx, "failed to close integration event processor", slog.String("error", err.Error()))
			}
		}()
		return s.integrationevent.Run(ctx)
	})

	g.Go(func() error {
		<-ctx.Done()
		return s.application.Stop(context.Background())
	})

	return g.Wait()
}

var ProvideServer = NewServer
