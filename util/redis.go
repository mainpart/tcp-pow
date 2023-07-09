package util

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

// InitRedisCache - create new instance of RedisCache
// host and port - connection to Redis instance
func InitRedisCache(ctx context.Context, host string, port int) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// check connection by setting test value
	err := rdb.Set(ctx, "key", "value", 0).Err()

	return &RedisCache{
		ctx:    ctx,
		client: rdb,
	}, err
}

func (c *RedisCache) Add(key string, expiration time.Duration) error {
	return c.client.Set(c.ctx, key, "value", expiration).Err()
}

func (c *RedisCache) Get(key string) (bool, error) {
	val, err := c.client.Get(c.ctx, key).Result()
	return val != "", err
}

func (c *RedisCache) Delete(key string) {
	c.client.Del(c.ctx, key)
}
