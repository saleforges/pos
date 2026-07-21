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
		if err := cache.Set(ctx, user, 0); err != nil {
			t.Fatalf("set failed: %v", err)
		}
		got, err := cache.Get(ctx, 1)
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected cached user, got nil")
		}
		if got.Username != "testuser" {
			t.Errorf("expected 'testuser', got %q", got.Username)
		}
	})

	t.Run("get non-existent returns nil error", func(t *testing.T) {
		got, err := cache.Get(ctx, 999)
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if got != nil {
			t.Error("expected nil for non-existent key")
		}
	})

	t.Run("delete removes from cache", func(t *testing.T) {
		if err := cache.Set(ctx, user, time.Minute); err != nil {
			t.Fatalf("set failed: %v", err)
		}
		if err := cache.Delete(ctx, 1); err != nil {
			t.Fatalf("delete failed: %v", err)
		}
		got, err := cache.Get(ctx, 1)
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if got != nil {
			t.Error("expected nil after delete")
		}
	})

	t.Run("expired item is evicted on get", func(t *testing.T) {
		if err := cache.Set(ctx, user, 1*time.Millisecond); err != nil {
			t.Fatalf("set failed: %v", err)
		}
		time.Sleep(5 * time.Millisecond)
		got, err := cache.Get(ctx, 1)
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if got != nil {
			t.Error("expected nil after expiry")
		}
	})

	t.Run("custom ttl overrides default", func(t *testing.T) {
		short := &domain.User{ID: 2, Username: "short"}
		long := &domain.User{ID: 3, Username: "long"}
		if err := cache.Set(ctx, short, 1*time.Millisecond); err != nil {
			t.Fatalf("set failed: %v", err)
		}
		if err := cache.Set(ctx, long, time.Hour); err != nil {
			t.Fatalf("set failed: %v", err)
		}

		time.Sleep(5 * time.Millisecond)
		g1, err1 := cache.Get(ctx, 2)
		if err1 != nil {
			t.Fatalf("get failed: %v", err1)
		}
		g2, err2 := cache.Get(ctx, 3)
		if err2 != nil {
			t.Fatalf("get failed: %v", err2)
		}
		if g1 != nil {
			t.Error("expected 'short' to expire")
		}
		if g2 == nil {
			t.Error("expected 'long' to still be cached")
		}
	})
}
