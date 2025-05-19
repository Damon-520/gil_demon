package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	ctx         context.Context
	redisClient *redis.Client
}

func NewRedisCache(ctx context.Context, redisClient *redis.Client) DictCache {
	return &RedisCache{
		ctx:         ctx,
		redisClient: redisClient,
	}
}

func (c *RedisCache) Set(ctx context.Context, key string, value string) error {
	err := c.redisClient.Set(c.ctx, key, value, 0).Err()
	return err
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.redisClient.Get(c.ctx, key).Result()
}
