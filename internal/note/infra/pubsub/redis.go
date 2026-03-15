package pubsub

import (
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient redis.Client

func NewRedisClient(cfg *commonconfig.Redis) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
		DB:   1,
	})
	return (*RedisClient)(client)
}

var ProvideRedisClient = NewRedisClient
