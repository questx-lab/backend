package testutil

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type MockRedisClient struct {
	ExistFunc               func(ctx context.Context, key string) (bool, error)
	ZAddFunc                func(ctx context.Context, key string, z redis.Z) error
	ZIncrByFunc             func(ctx context.Context, key string, incr int64, member string) error
	ZRevRangeWithScoresFunc func(ctx context.Context, key string, offset, limit int) ([]redis.Z, error)
	ZRevRankFunc            func(ctx context.Context, key string, member string) (uint64, error)
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
