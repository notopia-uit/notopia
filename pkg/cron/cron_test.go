package cron_test

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/notopia-uit/notopia/pkg/cron"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobRunsOnceAcrossSchedulers(t *testing.T) {
	mr := miniredis.RunT(t)
	var runs atomic.Int32
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	job := cron.Job{
		Name:     "test:job",
		Schedule: "0 0 1 1 *",
		Timeout:  5 * time.Second,
		Run: func(ctx context.Context) error {
			runs.Add(1)
			started <- struct{}{}
			<-release
			return nil
		},
	}

	schedulers := make([]*cron.Scheduler, 2)
	for i := range schedulers {
		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		t.Cleanup(func() { _ = client.Close() })
		locker, err := cron.NewRedisLocker(client)
		require.NoError(t, err)
		s, err := cron.NewScheduler(locker, slog.Default(), time.UTC)
		require.NoError(t, err)
		require.NoError(t, s.Register(job))
		s.Start()
		t.Cleanup(func() { _ = s.Shutdown() })
		schedulers[i] = s
	}

	require.NoError(t, schedulers[0].RunNow("test:job"))
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		require.FailNow(t, "first run did not start")
	}

	require.NoError(t, schedulers[1].RunNow("test:job"))
	select {
	case <-started:
		require.FailNow(t, "second scheduler ran the job while the lock was held")
	case <-time.After(300 * time.Millisecond):
	}

	close(release)
	time.Sleep(200 * time.Millisecond)

	assert.EqualValues(t, 1, runs.Load())
}

func TestRegisterRejectsInvalidJob(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	locker, err := cron.NewRedisLocker(client)
	require.NoError(t, err)
	s, err := cron.NewScheduler(locker, slog.Default(), time.UTC)
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Shutdown() })

	run := func(context.Context) error { return nil }
	tests := map[string]cron.Job{
		"empty name":   {Schedule: "* * * * *", Timeout: time.Second, Run: run},
		"nil run":      {Name: "a", Schedule: "* * * * *", Timeout: time.Second},
		"zero timeout": {Name: "b", Schedule: "* * * * *", Run: run},
		"neg timeout":  {Name: "c", Schedule: "* * * * *", Timeout: -time.Second, Run: run},
	}
	for name, job := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, s.Register(job))
		})
	}

	_, err = cron.NewScheduler(locker, nil, time.UTC)
	assert.Error(t, err, "nil logger")
}
