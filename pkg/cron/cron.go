package cron

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	_ "time/tzdata"

	"github.com/go-co-op/gocron/v2"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type Job struct {
	Name     string
	Schedule string
	Timeout  time.Duration
	Enabled  bool
	Run      func(ctx context.Context) error
}

type Scheduler struct {
	scheduler gocron.Scheduler
	logger    *slog.Logger
	jobs      map[string]gocron.Job
}

type slogAdapter struct{ logger *slog.Logger }

func (a slogAdapter) Debug(msg string, args ...any) { a.logger.Debug(msg, args...) } //nolint:sloglint
func (a slogAdapter) Info(msg string, args ...any)  { a.logger.Info(msg, args...) }  //nolint:sloglint
func (a slogAdapter) Warn(msg string, args ...any)  { a.logger.Warn(msg, args...) }  //nolint:sloglint
func (a slogAdapter) Error(msg string, args ...any) { a.logger.Error(msg, args...) } //nolint:sloglint

func JobFromConfig(cfg commonconfig.CronJob, run func(ctx context.Context) error) Job {
	return Job{
		Name:     cfg.Name,
		Schedule: cfg.Schedule,
		Timeout:  cfg.Timeout,
		Enabled:  cfg.Enabled,
		Run:      run,
	}
}

func NewScheduler(locker Locker, logger *slog.Logger, location *time.Location) (*Scheduler, error) {
	if logger == nil {
		return nil, errors.New("cron logger is required")
	}
	if location == nil {
		return nil, errors.New("cron location is required")
	}
	s, err := gocron.NewScheduler(
		gocron.WithLocation(location),
		gocron.WithDistributedLocker(locker),
		gocron.WithLogger(slogAdapter{logger}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron scheduler: %w", err)
	}
	return &Scheduler{scheduler: s, logger: logger, jobs: map[string]gocron.Job{}}, nil
}

func (s *Scheduler) Register(jobs ...Job) error {
	for _, job := range jobs {
		if job.Name == "" || job.Run == nil || job.Timeout <= 0 {
			return fmt.Errorf("invalid cron job %q: name, run and a positive timeout are required", job.Name)
		}
		if !job.Enabled {
			s.logger.Info("cron job disabled, skipping", slog.String("job", job.Name))
			continue
		}
		j, err := s.scheduler.NewJob(
			gocron.CronJob(job.Schedule, false),
			gocron.NewTask(s.wrap(job)),
			gocron.WithName(job.Name),
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
		)
		if err != nil {
			return fmt.Errorf("failed to register cron job %s: %w", job.Name, err)
		}
		s.jobs[job.Name] = j
	}
	return nil
}

func (s *Scheduler) wrap(job Job) func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
		defer cancel()
		ctx, span := otel.Tracer("github.com/notopia-uit/notopia/pkg/cron").Start(ctx, "Cron."+job.Name)
		defer span.End()
		defer func() {
			if r := recover(); r != nil {
				span.RecordError(fmt.Errorf("panic: %v", r))
				span.SetStatus(codes.Error, fmt.Sprint(r))
				s.logger.ErrorContext(
					ctx,
					"cron job panicked",
					slog.String("job", job.Name),
					slog.Any("panic", r),
				)
			}
		}()
		start := time.Now()
		if err := job.Run(ctx); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			s.logger.ErrorContext(
				ctx,
				"cron job failed",
				slog.String("job", job.Name),
				slog.Any("error", err),
			)
			return
		}
		s.logger.InfoContext(
			ctx,
			"cron job done",
			slog.String("job", job.Name),
			slog.Duration("took", time.Since(start)),
		)
	}
}

func (s *Scheduler) RunNow(name string) error {
	j, ok := s.jobs[name]
	if !ok {
		return fmt.Errorf("unknown cron job %s", name)
	}
	return j.RunNow()
}

func (s *Scheduler) Start() { s.scheduler.Start() }

func (s *Scheduler) Shutdown() error { return s.scheduler.Shutdown() }
