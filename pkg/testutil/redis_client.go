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
	DelFunc                 func(ctx context.Context, key ...string) error
	SetObjFunc              func(ctx context.Context, key string, obj any, ttl time.Duration) error
	GetObjFunc              func(ctx context.Context, key string, v any) error
	MSetFunc                func(ctx context.Context, kv map[string]any) error
	MGetFunc                func(ctx context.Context, keys ...string) ([]any, error)
	KeysFunc                func(ctx context.Context, pattern string) ([]string, error)
	SAddFunc                func(ctx context.Context, key string, members ...string) error
	SRemFunc                func(ctx context.Context, key string, members ...string) error
	SMembersFunc            func(ctx context.Context, key string, count int) ([]string, error)
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

func (m *MockRedisClient) Del(ctx context.Context, key ...string) error {
	if m.DelFunc != nil {
		return m.DelFunc(ctx, key...)
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

	return redis.Nil
}

func (m *MockRedisClient) MSet(ctx context.Context, kv map[string]any) error {
	if m.MSetFunc != nil {
		return m.MSetFunc(ctx, kv)
	}

	return nil
}

func (m *MockRedisClient) MGet(ctx context.Context, keys ...string) ([]any, error) {
	if m.MGetFunc != nil {
		return m.MGetFunc(ctx, keys...)
	}

	return nil, errors.New("not implemented")
}

func (m *MockRedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	if m.KeysFunc != nil {
		return m.KeysFunc(ctx, pattern)
	}

	return []string{}, nil
}

func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...string) error {
	if m.SAddFunc != nil {
		return m.SAddFunc(ctx, key, members...)
	}

	return nil
}

func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...string) error {
	if m.SRemFunc != nil {
		return m.SRemFunc(ctx, key, members...)
	}

	return nil
}

func (m *MockRedisClient) SMembers(ctx context.Context, key string, count int) ([]string, error) {
	if m.SMembersFunc != nil {
		return m.SMembersFunc(ctx, key, count)
	}

	return []string{}, nil
}
