package controller

import (
	"context"
	"log/slog"

	commonhttp "github.com/notopia-uit/notopia/pkg/common/controller/http"
	"github.com/notopia-uit/notopia/pkg/note/controller/grpc"
	"github.com/notopia-uit/notopia/pkg/note/controller/http"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	http          *http.Server
	grpc          *grpc.Server
	healthManager *commonhttp.HealthManager
}

func NewServer(
	httpServer *http.Server,
	grpcServer *grpc.Server,
	healthManager *commonhttp.HealthManager,
) *Server {
	return &Server{
		http:          httpServer,
		grpc:          grpcServer,
		healthManager: healthManager,
	}
}

func (s *Server) Run(ctx context.Context) error {
	// Run migration here

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
