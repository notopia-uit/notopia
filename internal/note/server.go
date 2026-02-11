package note

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/infra/persistence"
	"github.com/notopia-uit/notopia/internal/note/transport/grpc"
	"github.com/notopia-uit/notopia/internal/note/transport/http"
	"github.com/notopia-uit/notopia/pkg/healthmanager"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	http          *http.Server
	grpc          *grpc.Server
	healthManager *healthmanager.HealthManager
	persistence   persistence.Persistence
}

func NewServer(
	httpServer *http.Server,
	grpcServer *grpc.Server,
	healthManager *healthmanager.HealthManager,
	persistence persistence.Persistence,
	logger *slog.Logger,
) *Server {
	slog.SetDefault(logger)

	return &Server{
		http:          httpServer,
		grpc:          grpcServer,
		healthManager: healthManager,
		persistence:   persistence,
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
				slog.Error("failed to shutdown http server", slog.String("error", err.Error()))
			}
		}()
		return s.http.Run()
	})

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			if err := s.grpc.Shutdown(context.Background()); err != nil {
				slog.Error("failed to shutdown grpc server", slog.String("error", err.Error()))
			}
		}()
		return s.grpc.Run()
	})

	s.healthManager.SetStartedUp()
	go s.healthManager.StartMonitoring(ctx)

	return g.Wait()
}

var ProvideServer = NewServer
