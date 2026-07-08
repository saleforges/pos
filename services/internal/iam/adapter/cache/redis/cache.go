package redis

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	memcache "github.com/saleforge/pos/services/internal/iam/adapter/cache/memory"
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
		log.Printf("[cache/redis] connection failed (%v), falling back to in-memory", err)
		return memcache.NewUserCache(ttl, 5*time.Minute)
	}

	log.Printf("[cache/redis] connected to %s", addr)
	return &UserCache{
		rdb:      rdb,
		fallback: memcache.NewUserCache(ttl, 5*time.Minute),
		ttl:      ttl,
	}
}

func (c *UserCache) Get(ctx context.Context, id string) (*domain.User, bool) {
	if c.rdb == nil {
		return c.fallback.Get(ctx, id)
	}

	data, err := c.rdb.Get(ctx, "user:"+id).Bytes()
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
	c.rdb.Set(ctx, "user:"+u.ID, data, ttl)
}

func (c *UserCache) Delete(ctx context.Context, id string) {
	if c.rdb != nil {
		c.rdb.Del(ctx, "user:"+id)
	}
	c.fallback.Delete(ctx, id)
}

func (c *UserCache) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}
	return c.fallback.Close()
}
