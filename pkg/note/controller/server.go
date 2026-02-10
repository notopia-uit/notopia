package controller

import (
	"context"
	"log/slog"

	commonhttp "github.com/notopia-uit/notopia/pkg/common/controller/http"
	"github.com/notopia-uit/notopia/pkg/note/controller/grpc"
	"github.com/notopia-uit/notopia/pkg/note/controller/http"
	"github.com/notopia-uit/notopia/pkg/note/infra/persistence"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	http          *http.Server
	grpc          *grpc.Server
	healthManager *commonhttp.HealthManager
	persistence   persistence.Persistence
}

func NewServer(
	httpServer *http.Server,
	grpcServer *grpc.Server,
	healthManager *commonhttp.HealthManager,
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
	if err := s.persistence.RunMigrations(); err != nil {
		return err
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
