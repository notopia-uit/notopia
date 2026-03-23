package health

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/alexliesenfeld/health"
	httpCheck "github.com/hellofresh/health-go/v4/checks/http"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
)

type Health struct {
	*http.Server
}

func New(
	persistence app.Persistence,
	serverCfg *config.Server,
	workspaceEventPubSub app.WorkspaceEventPubSub,
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
		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "workspaceEventPubSub",
				Check: func(ctx context.Context) error {
					return workspaceEventPubSub.Check(ctx)
				},
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
