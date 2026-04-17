package health

import (
	"context"
	"errors"
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
						return err
					}
					if !ok {
						return errors.New("migration not done")
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
					return persistence.Ping(ctx)
				},
			},
		),

		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "http",
				Check: httpCheck.New(httpCheck.Config{
					URL: serverCfg.HTTP.Address(),
				}),
			},
		),

		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "grpc",
				Check: httpCheck.New(httpCheck.Config{
					URL: serverCfg.GRPC.Address(),
				}),
			},
		),

		// TODO: this have to check kafka, not the pub sub
		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "workspaceEventHub redis connection",
				Check: func(ctx context.Context) error {
					return redisClient.Ping(ctx).Err()
				},
			},
		),

		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "authorization service",
				Check: func(ctx context.Context) error {
					return authorizationSvc.CheckHealth(ctx)
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
	)

	liveCheck := health.NewChecker()

	mux := http.NewServeMux()
	mux.Handle("/startup", health.NewHandler(startupChecker))
	mux.Handle("/ready", health.NewHandler(readyChecker))
	mux.Handle("/live", health.NewHandler(liveCheck))
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
