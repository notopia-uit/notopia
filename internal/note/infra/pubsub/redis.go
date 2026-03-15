package pubsub

import (
	"context"
	"log/slog"

	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient redis.Client

func NewRedisClient(
	ctx context.Context,
	cfg *commonconfig.Redis,
	logger *slog.Logger,
) (*RedisClient, func()) {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
		DB:   1,
	})
	cleanup := func() {
		if err := client.Close(); err != nil {
			logger.Error("failed to shutdown Redis client", slog.String("error", err.Error()))
		}
	}
	return (*RedisClient)(client), cleanup
}

var ProvideRedisClient = NewRedisClient
