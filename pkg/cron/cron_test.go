package cron_test

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/notopia-uit/notopia/pkg/cron"
	"github.com/redis/go-redis/v9"
)

func TestJobRunsOnceAcrossSchedulers(t *testing.T) {
	mr := miniredis.RunT(t)
	var runs atomic.Int32
	job := cron.Job{
		Name:     "test:job",
		Schedule: "0 0 1 1 *",
		Timeout:  time.Second,
		Run: func(ctx context.Context) error {
			runs.Add(1)
			time.Sleep(300 * time.Millisecond)
			return nil
		},
	}

	schedulers := make([]*cron.Scheduler, 2)
	for i := range schedulers {
		locker, err := cron.NewRedisLocker(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
		if err != nil {
			t.Fatal(err)
		}
		s, err := cron.NewScheduler(locker, slog.Default(), time.UTC)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Register(job); err != nil {
			t.Fatal(err)
		}
		s.Start()
		t.Cleanup(func() { _ = s.Shutdown() })
		schedulers[i] = s
	}

	var wg sync.WaitGroup
	for _, s := range schedulers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.RunNow("test:job")
		}()
	}
	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	if got := runs.Load(); got != 1 {
		t.Fatalf("runs = %d, want 1", got)
	}
}
