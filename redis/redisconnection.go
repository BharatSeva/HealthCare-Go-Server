package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redisconn struct {
	ctx  context.Context
	conn *redis.Client

	// this one is for rate_limiting
	limit       int64
	window      time.Duration
	lastchecked time.Time
	refill      time.Duration
}

func Connect2Redis(addr string, limit int64, window time.Duration) (*Redisconn, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Redisconn{
		ctx:    ctx,
		conn:   client,
		limit:  limit,
		window: window,
	}, nil
}
