package cron

import (
	"fmt"
	"time"

	redislock "github.com/go-co-op/gocron-redis-lock/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
)

const lockAutoExtend = 3 * time.Second

func NewRedisLocker(client redis.UniversalClient) (gocron.Locker, error) {
	locker, err := redislock.NewRedisLockerWithOptions(
		client,
		redislock.WithRedsyncOptions(redislock.WithTries(1)),
		redislock.WithAutoExtendDuration(lockAutoExtend),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis cron locker: %w", err)
	}
	return locker, nil
}
