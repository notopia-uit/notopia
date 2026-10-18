package common

import (
	"context"
	"log/slog"

	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	pkgcron "github.com/notopia-uit/notopia/pkg/cron"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
}

func NewRedisClient(
	ctx context.Context,
	cfg *commonconfig.Redis,
	logger *slog.Logger,
) (*RedisClient, func()) {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
	})
	cleanup := func() {
		if err := client.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown Redis client", slog.Any("error", err))
		}
	}
	return &RedisClient{client}, cleanup
}

var ProvideRedisClient = NewRedisClient

func NewCronLocker(client *RedisClient) (pkgcron.Locker, error) {
	return pkgcron.NewRedisLocker(client.Client)
}

var ProvideCronLocker = NewCronLocker
