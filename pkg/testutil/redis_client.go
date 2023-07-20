package testutil

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type MockRedisClient struct {
	ExistFunc               func(ctx context.Context, key string) (bool, error)
	ZAddFunc                func(ctx context.Context, key string, z redis.Z) error
	ZIncrByFunc             func(ctx context.Context, key string, incr int64, member string) error
	ZRevRangeWithScoresFunc func(ctx context.Context, key string, offset, limit int) ([]redis.Z, error)
	ZRevRankFunc            func(ctx context.Context, key string, member string) (uint64, error)
	GetFunc                 func(ctx context.Context, key string) (string, error)
	SetFunc                 func(ctx context.Context, key string, value string) error
	DelFunc                 func(ctx context.Context, key string) error
	SetObjFunc              func(ctx context.Context, key string, obj any, ttl time.Duration) error
	GetObjFunc              func(ctx context.Context, key string, v any) error
}

func (m *MockRedisClient) Exist(ctx context.Context, key string) (bool, error) {
	if m.ExistFunc != nil {
		return m.ExistFunc(ctx, key)
	}

	return false, nil
}

func (m *MockRedisClient) ZAdd(ctx context.Context, key string, z redis.Z) error {
	if m.ZAddFunc != nil {
		return m.ZAddFunc(ctx, key, z)
	}

	return nil
}

func (m *MockRedisClient) ZIncrBy(ctx context.Context, key string, incr int64, member string) error {
	if m.ZIncrByFunc != nil {
		return m.ZIncrByFunc(ctx, key, incr, member)
	}

	return nil
}

func (m *MockRedisClient) ZRevRangeWithScores(ctx context.Context, key string, offset, limit int) ([]redis.Z, error) {
	if m.ZRevRangeWithScoresFunc != nil {
		return m.ZRevRangeWithScoresFunc(ctx, key, offset, limit)
	}

	return nil, nil
}

func (m *MockRedisClient) ZRevRank(ctx context.Context, key string, member string) (uint64, error) {
	if m.ZRevRankFunc != nil {
		return m.ZRevRankFunc(ctx, key, member)
	}

	return 0, nil
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}

	return "", nil
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value string) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value)
	}

	return nil
}

func (m *MockRedisClient) Del(ctx context.Context, key string) error {
	if m.DelFunc != nil {
		return m.DelFunc(ctx, key)
	}

	return nil
}

func (m *MockRedisClient) SetObj(ctx context.Context, key string, obj any, ttl time.Duration) error {
	if m.SetObjFunc != nil {
		return m.SetObjFunc(ctx, key, obj, ttl)
	}

	return nil
}

func (m *MockRedisClient) GetObj(ctx context.Context, key string, v any) error {
	if m.GetObjFunc != nil {
		return m.GetObjFunc(ctx, key, v)
	}

	return errors.New("not found")
}
