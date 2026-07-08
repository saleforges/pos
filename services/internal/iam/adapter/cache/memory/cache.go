package memory

import (
	"context"
	"sync"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type item struct {
	user     *domain.User
	expireAt time.Time
}

type UserCache struct {
	mu       sync.RWMutex
	items    map[string]*item
	ttl      time.Duration
	stopCh   chan struct{}
}

func NewUserCache(ttl time.Duration, cleanupInterval time.Duration) *UserCache {
	c := &UserCache{
		items:  make(map[string]*item),
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}
	go c.cleanup(cleanupInterval)
	return c
}

func (c *UserCache) Get(_ context.Context, id string) (*domain.User, bool) {
	c.mu.RLock()
	it, ok := c.items[id]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}
	if time.Now().After(it.expireAt) {
		c.Delete(nil, id)
		return nil, false
	}
	return it.user, true
}

func (c *UserCache) Set(_ context.Context, u *domain.User, ttl time.Duration) {
	if ttl == 0 {
		ttl = c.ttl
	}
	c.mu.Lock()
	c.items[u.ID] = &item{
		user:     u,
		expireAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

func (c *UserCache) Delete(_ context.Context, id string) {
	c.mu.Lock()
	delete(c.items, id)
	c.mu.Unlock()
}

func (c *UserCache) Close() error {
	close(c.stopCh)
	return nil
}

func (c *UserCache) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			for id, it := range c.items {
				if now.After(it.expireAt) {
					delete(c.items, id)
				}
			}
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}
