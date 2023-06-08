package xredis

import (
	"context"
	"time"

	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/redis/go-redis/v9"
)

type Client interface {
	Exist(ctx context.Context, key string) (bool, error)
	ZAdd(ctx context.Context, key string, z redis.Z) error
	ZIncrBy(ctx context.Context, key string, incr int64, member string) error
	ZRevRangeWithScores(ctx context.Context, key string, offset, limit int) ([]redis.Z, error)
	ZRevRank(ctx context.Context, key string, member string) (uint64, error)
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}

type client struct {
	redisClient *redis.Client
}

func NewClient(ctx context.Context) (*client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:            xcontext.Configs(ctx).Redis.Addr,
		MaxRetries:      5,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		PoolFIFO:        false,
		PoolSize:        5,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &client{redisClient: redisClient}, nil
}

func (c *client) Exist(ctx context.Context, key string) (bool, error) {
	n, err := c.redisClient.Exists(ctx, key).Uint64()
	if err != nil {
		return false, err
	}

	if n != 1 {
		return false, nil
	}

	return true, nil
}

func (c *client) ZAdd(ctx context.Context, key string, z redis.Z) error {
	_, err := c.redisClient.ZAdd(ctx, key, z).Uint64()
	if err != nil {
		return err
	}

	return nil
}

func (c *client) ZIncrBy(ctx context.Context, key string, incr int64, member string) error {
	_, err := c.redisClient.ZIncrBy(ctx, key, float64(incr), member).Result()
	if err != nil {
		return err
	}

	return nil
}

func (c *client) ZRevRangeWithScores(
	ctx context.Context, key string, offset, limit int,
) ([]redis.Z, error) {
	result := c.redisClient.ZRevRangeWithScores(ctx, key, int64(offset), int64(offset+limit-1))
	return result.Result()
}

func (c *client) ZRevRank(
	ctx context.Context, key string, member string,
) (uint64, error) {
	result := c.redisClient.ZRevRank(ctx, key, member)
	return result.Uint64()
}

func (c *client) Get(ctx context.Context, key string) (string, error) {
	return c.redisClient.Get(ctx, key).Result()
}

func (c *client) Set(ctx context.Context, key, value string) error {
	return c.redisClient.Set(ctx, key, value, -1).Err()
}
