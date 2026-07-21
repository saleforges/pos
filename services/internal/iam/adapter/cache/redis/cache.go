package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/pkg/logger"
)

type UserCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewUserCache(addr string, ttl time.Duration) *UserCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("[cache/redis] connection failed", "error", err.Error())
		return &UserCache{rdb: nil, ttl: ttl}
	}

	logger.Info("[cache/redis] connected", "addr", addr)
	return &UserCache{rdb: rdb, ttl: ttl}
}

func (c *UserCache) Get(ctx context.Context, id int64) (*domain.User, error) {
	if c.rdb == nil {
		return nil, fmt.Errorf("redis cache not available")
	}

	data, err := c.rdb.Get(ctx, fmt.Sprintf("user:%d", id)).Bytes()
	if err != nil {
		return nil, err
	}

	var u domain.User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, fmt.Errorf("unmarshal user cache: %w", err)
	}
	return &u, nil
}

func (c *UserCache) Set(ctx context.Context, u *domain.User, ttl time.Duration) error {
	if c.rdb == nil {
		return fmt.Errorf("redis cache not available")
	}

	data, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("marshal user cache: %w", err)
	}

	if ttl == 0 {
		ttl = c.ttl
	}
	return c.rdb.Set(ctx, fmt.Sprintf("user:%d", u.ID), data, ttl).Err()
}

func (c *UserCache) Delete(ctx context.Context, id int64) error {
	if c.rdb == nil {
		return nil // not available, nothing to delete
	}
	return c.rdb.Del(ctx, fmt.Sprintf("user:%d", id)).Err()
}

func (c *UserCache) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}
	return nil
}
