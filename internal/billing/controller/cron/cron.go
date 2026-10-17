package cron

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/app"
	pkgcron "github.com/notopia-uit/notopia/pkg/cron"
	goredis "github.com/redis/go-redis/v9"
)

// NOTE: move this controller to cmd/billing-worker when crons are split out of the API process.
type Cron struct {
	scheduler *pkgcron.Scheduler
}

func New(a *app.Server, redis *goredis.Client, logger *slog.Logger) (*Cron, error) {
	locker, err := pkgcron.NewRedisLocker(redis)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron locker: %w", err)
	}
	scheduler, err := pkgcron.NewScheduler(locker, logger, time.Local)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron scheduler: %w", err)
	}
	if err := scheduler.Register(
		pkgcron.Job{
			Name:     "billing:reconcile-subscriptions",
			Schedule: "0 * * * *",
			Timeout:  15 * time.Minute,
			Run: func(ctx context.Context) error {
				return a.Cmds.ReconcileSubscriptions.Handle(ctx, &app.ReconcileSubscriptions{})
			},
		},
		pkgcron.Job{
			Name:     "billing:prune-provider-events",
			Schedule: "0 4 * * *",
			Timeout:  5 * time.Minute,
			Run: func(ctx context.Context) error {
				return a.Cmds.PruneProviderEvents.Handle(ctx, &app.PruneProviderEvents{})
			},
		},
	); err != nil {
		return nil, fmt.Errorf("failed to register cron jobs: %w", err)
	}
	return &Cron{scheduler: scheduler}, nil
}

var Provide = New

func (c *Cron) Run(ctx context.Context) error {
	c.scheduler.Start()
	<-ctx.Done()
	return c.scheduler.Shutdown()
}
