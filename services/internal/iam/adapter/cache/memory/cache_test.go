package memory

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

func TestUserCache(t *testing.T) {
	t.Parallel()

	cache := NewUserCache(100*time.Millisecond, 50*time.Millisecond)
	defer cache.Close()

	ctx := context.Background()
	user := &domain.User{ID: 1, Username: "testuser", Email: "test@t.com"}

	t.Run("set and get", func(t *testing.T) {
		cache.Set(ctx, user, 0)
		got, ok := cache.Get(ctx, 1)
		if !ok { t.Fatal("expected cached user") }
		if got.Username != "testuser" { t.Errorf("expected 'testuser', got %q", got.Username) }
	})

	t.Run("get non-existent returns false", func(t *testing.T) {
		_, ok := cache.Get(ctx, 999)
		if ok { t.Error("expected false for non-existent key") }
	})

	t.Run("delete removes from cache", func(t *testing.T) {
		cache.Set(ctx, user, time.Minute)
		cache.Delete(ctx, 1)
		_, ok := cache.Get(ctx, 1)
		if ok { t.Error("expected false after delete") }
	})

	t.Run("expired item is evicted on get", func(t *testing.T) {
		cache.Set(ctx, user, 1*time.Millisecond)
		time.Sleep(5 * time.Millisecond)
		_, ok := cache.Get(ctx, 1)
		if ok { t.Error("expected false after expiry") }
	})

	t.Run("custom ttl overrides default", func(t *testing.T) {
		short := &domain.User{ID: 2, Username: "short"}
		cache.Set(ctx, short, 1*time.Millisecond)
		long := &domain.User{ID: 3, Username: "long"}
		cache.Set(ctx, long, time.Hour)

		time.Sleep(5 * time.Millisecond)
		_, ok1 := cache.Get(ctx, 2)
		_, ok2 := cache.Get(ctx, 3)
		if ok1 { t.Error("expected 'short' to expire") }
		if !ok2 { t.Error("expected 'long' to still be cached") }
	})
}
