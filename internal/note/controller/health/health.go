package health

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/alexliesenfeld/health"
	httpCheck "github.com/hellofresh/health-go/v5/checks/http"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence"
	"github.com/notopia-uit/notopia/internal/note/infra/service"
	"github.com/notopia-uit/notopia/internal/note/infra/workspaceevent"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

type Health struct {
	*http.Server
}

func New(
	persistence *persistence.Pg,
	serverCfg *config.Server,
	kafkaCfg *commonconfig.Kafka,
	workspaceEventHub app.WorkspaceEventHub,
	redisClient *workspaceevent.RedisClient,
	authorizationSvc *service.Authorization,
	authentikCfg *commonconfig.Authentik,
) *Health {
	startupChecker := health.NewChecker(
		health.WithCheck(
			health.Check{
				Name: "persistenceMigration",
				Check: func(ctx context.Context) error {
					ok, err := persistence.IsMigrationDone(ctx)
					if err != nil {
						return fmt.Errorf("failed to check migration status: %w", err)
					}
					if !ok {
						return errors.New("database migration is not completed yet")
					}
					return nil
				},
			},
		),
	)

	readyChecker := health.NewChecker(
		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "persistenceConnection",
				Check: func(ctx context.Context) error {
					if err := persistence.Ping(ctx); err != nil {
						return fmt.Errorf("failed to ping persistence: %w", err)
					}
					return nil
				},
			},
		),

		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "http",
				Check: httpCheck.New(httpCheck.Config{
					URL: "http://" + serverCfg.HTTP.Address(),
				}),
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

		// TODO: this have to check kafka, not the pub sub
		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "workspaceEventHubRedisConnection",
				Check: func(ctx context.Context) error {
					if err := redisClient.Ping(ctx).Err(); err != nil {
						return fmt.Errorf("failed to ping Redis: %w", err)
					}
					return nil
				},
			},
		),

		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "authorizationService",
				Check: func(ctx context.Context) error {
					if err := authorizationSvc.CheckHealth(ctx); err != nil {
						return fmt.Errorf("failed to check authorization service: %w", err)
					}
					return nil
				},
			},
		),

		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "authentik",
				Check: httpCheck.New(httpCheck.Config{
					URL: authentikCfg.HealthLiveURL(),
				}),
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
	mux.Handle("/note/health/startup", health.NewHandler(startupChecker))
	mux.Handle("/note/health/ready", health.NewHandler(readyChecker))
	mux.Handle("/note/health/live", health.NewHandler(liveCheck))
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
