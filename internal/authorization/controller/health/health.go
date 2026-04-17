package health

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/notopia-uit/notopia/internal/authorization/config"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"gorm.io/gorm"
)

type Health struct {
	*http.Server
}

func New(
	db *gorm.DB,
	serverCfg *config.ServerConfig,
	kafkaCfg *commonconfig.Kafka,
) *Health {
	startupChecker := health.NewChecker(
		health.WithCheck(
			health.Check{
				Name: "database",
				Check: func(ctx context.Context) error {
					if db == nil {
						return errors.New("database not initialized")
					}
					sqlDB, err := db.DB()
					if err != nil {
						return err
					}
					return sqlDB.PingContext(ctx)
				},
			},
		),
	)

	readyChecker := health.NewChecker(
		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "database",
				Check: func(ctx context.Context) error {
					if db == nil {
						return errors.New("database not initialized")
					}
					sqlDB, err := db.DB()
					if err != nil {
						return err
					}
					return sqlDB.PingContext(ctx)
				},
			},
		),

		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "grpc",
				Check: func(ctx context.Context) error {
					conn, err := net.DialTimeout("tcp", serverCfg.GRPC.Address(), 5*time.Second)
					if err != nil {
						return fmt.Errorf("failed to connect to gRPC server: %w", err)
					}
					if err := conn.Close(); err != nil {
						return fmt.Errorf("failed to close gRPC connection: %w", err)
					}
					return nil
				},
			},
		),

		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "kafka",
				Check: func(ctx context.Context) error {
					if kafkaCfg == nil || len(kafkaCfg.Brokers) == 0 {
						return errors.New("kafka not configured")
					}
					// Try to connect to each broker
					for i, broker := range kafkaCfg.Brokers {
						conn, err := net.DialTimeout("tcp", broker, 5*time.Second)
						if err != nil {
							return fmt.Errorf("failed to connect to kafka broker %d (%s): %w", i, broker, err)
						}
						if err := conn.Close(); err != nil {
							return fmt.Errorf("failed to close kafka broker %d connection: %w", i, err)
						}
					}
					return nil
				},
			},
		),
	)

	liveCheck := health.NewChecker()

	mux := http.NewServeMux()
	mux.Handle("/authorization/health/startup", health.NewHandler(startupChecker))
	mux.Handle("/authorization/health/ready", health.NewHandler(readyChecker))
	mux.Handle("/authorization/health/live", health.NewHandler(liveCheck))
	return &Health{
		&http.Server{
			Handler: mux,
			Addr:    serverCfg.Health.Address(),
		},
	}
}

var Provide = New

func (h *Health) Run() error {
	return h.ListenAndServe()
}

func (h *Health) Shutdown(ctx context.Context) error {
	return h.Server.Shutdown(ctx)
}
