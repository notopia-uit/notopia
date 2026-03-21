package note

import (
	"context"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/controller/grpc"
	"github.com/notopia-uit/notopia/internal/note/controller/http"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	http        *http.Server
	grpc        *grpc.Server
	application *app.App
	logger      *slog.Logger
}

func NewServer(
	httpServer *http.Server,
	grpcServer *grpc.Server,
	application *app.App,
	logger *slog.Logger,
) *Server {
	slog.SetDefault(logger)

	return &Server{
		http:        httpServer,
		grpc:        grpcServer,
		application: application,
		logger:      logger,
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
		<-ctx.Done()
		return s.application.Stop(context.Background())
	})

	return g.Wait()
}

var ProvideServer = NewServer
