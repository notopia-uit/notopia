package redis

import (
	"context"
	"fmt"

	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	goredis "github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, cfg *commonconfig.Redis) (*goredis.Client, func(), error) {
	client := goredis.NewClient(&goredis.Options{Addr: cfg.Addr})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to ping redis: %w", err)
	}
	return client, func() { _ = client.Close() }, nil
}

var ProvideClient = NewClient
