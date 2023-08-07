package xredis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/redis/go-redis/v9"
)

type Client interface {
	Exist(ctx context.Context, key string) (bool, error)
	Del(ctx context.Context, key ...string) error
	Keys(ctx context.Context, pattern string) ([]string, error)

	// Sorted list
	ZAdd(ctx context.Context, key string, z redis.Z) error
	ZIncrBy(ctx context.Context, key string, incr int64, member string) error
	ZRevRangeWithScores(ctx context.Context, key string, offset, limit int) ([]redis.Z, error)
	ZRevRank(ctx context.Context, key string, member string) (uint64, error)

	// Set
	SAdd(ctx context.Context, key string, members ...string) error
	SRem(ctx context.Context, key string, members ...string) error
	SMembers(ctx context.Context, key string, count int) ([]string, error)
	SScan(ctx context.Context, key, pattern string, cursor uint64, limit int) ([]string, uint64, error)
	SCard(ctx context.Context, key string) (uint64, error)

	// Single object
	Set(ctx context.Context, key, value string) error
	SetObj(ctx context.Context, key string, obj any, ttl time.Duration) error
	MSet(ctx context.Context, kv map[string]any) error
	Get(ctx context.Context, key string) (string, error)
	GetObj(ctx context.Context, key string, v any) error
	MGet(ctx context.Context, keys ...string) ([]any, error)
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

///// COMMON FEATURE
func (c *client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.redisClient.Keys(ctx, pattern).Result()
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

func (c *client) Del(ctx context.Context, key ...string) error {
	err := c.redisClient.Del(ctx, key...).Err()
	if err == nil || err == redis.Nil {
		return nil
	}

	return err
}

///// SORTED LIST
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

///// SET
func (c *client) SAdd(ctx context.Context, key string, members ...string) error {
	return c.redisClient.SAdd(ctx, key, members).Err()
}

func (c *client) SRem(ctx context.Context, key string, members ...string) error {
	return c.redisClient.SRem(ctx, key, members).Err()
}

func (c *client) SMembers(ctx context.Context, key string, count int) ([]string, error) {
	if count == 0 {
		return c.redisClient.SMembers(ctx, key).Result()
	} else {
		return c.redisClient.SRandMemberN(ctx, key, int64(count)).Result()
	}
}

func (c *client) SScan(
	ctx context.Context, key, pattern string, cursor uint64, limit int,
) ([]string, uint64, error) {
	return c.redisClient.SScan(ctx, key, cursor, pattern, int64(limit)).Result()
}

func (c *client) SCard(ctx context.Context, key string) (uint64, error) {
	n, err := c.redisClient.SCard(ctx, key).Result()
	return uint64(n), err
}

///// SINGLE OBJECT
func (c *client) Set(ctx context.Context, key, value string) error {
	return c.redisClient.Set(ctx, key, value, -1).Err()
}

func (c *client) SetObj(ctx context.Context, key string, obj any, ttl time.Duration) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return c.redisClient.Set(ctx, key, b, ttl).Err()
}

func (c *client) MSet(ctx context.Context, kv map[string]any) error {
	newKV := map[string]string{}
	for k, v := range kv {
		if s, ok := v.(string); ok {
			newKV[k] = s
		} else {
			b, err := json.Marshal(v)
			if err != nil {
				return err
			}

			newKV[k] = string(b)
		}
	}

	return c.redisClient.MSet(ctx, newKV).Err()
}

func (c *client) Get(ctx context.Context, key string) (string, error) {
	return c.redisClient.Get(ctx, key).Result()
}

func (c *client) GetObj(ctx context.Context, key string, v any) error {
	s, err := c.redisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(s), v)
}

func (c *client) MGet(ctx context.Context, keys ...string) ([]any, error) {
	return c.redisClient.MGet(ctx, keys...).Result()
}
