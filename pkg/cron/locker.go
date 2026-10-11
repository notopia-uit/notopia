package cron

import (
	"fmt"

	redislock "github.com/go-co-op/gocron-redis-lock/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
)

func NewRedisLocker(client redis.UniversalClient) (gocron.Locker, error) {
	locker, err := redislock.NewRedisLocker(client, redislock.WithTries(1))
	if err != nil {
		return nil, fmt.Errorf("failed to create redis cron locker: %w", err)
	}
	return locker, nil
}
