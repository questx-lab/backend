package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewClient(addr string) *redis.Client {
	options := &redis.Options{
		Addr:            addr,
		MaxRetries:      5,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		PoolFIFO:        false,
		PoolSize:        5,
	}
	client := redis.NewClient(options)
	ctx := context.Background()
	if cmd := client.Ping(ctx); cmd.Err() != nil {
		panic(cmd.Err())
	}
	return client
}
