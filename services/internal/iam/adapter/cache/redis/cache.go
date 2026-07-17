package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	memcache "github.com/saleforge/pos/services/internal/iam/adapter/cache/memory"
	"github.com/saleforge/pos/services/pkg/logger"
)

type UserCache struct {
	rdb      *redis.Client
	fallback port.UserCache
	ttl      time.Duration
}

func NewUserCache(addr string, ttl time.Duration) port.UserCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("[cache/redis] connection failed, falling back to in-memory", "error", err.Error())
		return memcache.NewUserCache(ttl, 5*time.Minute)
	}

	logger.Info("[cache/redis] connected to redis", "addr", addr)
	return &UserCache{
		rdb:      rdb,
		fallback: memcache.NewUserCache(ttl, 5*time.Minute),
		ttl:      ttl,
	}
}

func (c *UserCache) Get(ctx context.Context, id int64) (*domain.User, bool) {
	if c.rdb == nil {
		return c.fallback.Get(ctx, id)
	}

	data, err := c.rdb.Get(ctx, fmt.Sprintf("user:%d", id)).Bytes()
	if err != nil {
		return c.fallback.Get(ctx, id)
	}

	var u domain.User
	if err := json.Unmarshal(data, &u); err != nil {
		c.fallback.Set(ctx, &u, 0)
		return &u, true
	}
	return nil, false
}

func (c *UserCache) Set(ctx context.Context, u *domain.User, ttl time.Duration) {
	if c.rdb == nil {
		c.fallback.Set(ctx, u, ttl)
		return
	}

	data, err := json.Marshal(u)
	if err != nil {
		return
	}

	if ttl == 0 {
		ttl = c.ttl
	}
	c.rdb.Set(ctx, fmt.Sprintf("user:%d", u.ID), data, ttl)
}

func (c *UserCache) Delete(ctx context.Context, id int64) {
	if c.rdb != nil {
		c.rdb.Del(ctx, fmt.Sprintf("user:%d", id))
	}
	c.fallback.Delete(ctx, id)
}

func (c *UserCache) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}
	return c.fallback.Close()
}
