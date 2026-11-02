package cron

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	pkgcron "github.com/notopia-uit/notopia/pkg/cron"
)

// NOTE: move this controller to cmd/note-worker when crons are split out of the API process.
type Cron struct {
	scheduler *pkgcron.Scheduler
}

func New(a *app.Server, locker gocron.Locker, cfg *config.Cron, logger *slog.Logger) (*Cron, error) {
	scheduler, err := pkgcron.NewScheduler(locker, logger, time.Local)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron scheduler: %w", err)
	}
	if err := scheduler.Register(
		pkgcron.JobFromConfig(cfg.PurgeExpiredTrash, func(ctx context.Context) error {
			return a.Cmds.PurgeExpiredTrash.Handle(ctx, &app.PurgeExpiredTrash{})
		}),
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
