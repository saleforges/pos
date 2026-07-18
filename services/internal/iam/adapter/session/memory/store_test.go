package memory

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

func TestSessionStore(t *testing.T) {
	t.Parallel()

	s := NewSessionStore()
	ctx := context.Background()

	now := time.Now().UTC()
	session := &domain.Session{
		ID:         "sess-1",
		UserID:     42,
		UserAgent:  "test-agent",
		IPAddress:  "127.0.0.1",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	t.Run("create and get session", func(t *testing.T) {
		err := s.Create(ctx, session)
		if err != nil { t.Fatalf("Create failed: %v", err) }

		got, err := s.Get(ctx, "sess-1")
		if err != nil { t.Fatalf("Get failed: %v", err) }
		if got.UserID != 42 { t.Errorf("expected UserID 42, got %d", got.UserID) }
	})

	t.Run("get non-existent session", func(t *testing.T) {
		_, err := s.Get(ctx, "nonexistent")
		if err == nil { t.Error("expected error for non-existent session") }
	})

	t.Run("duplicate create returns error", func(t *testing.T) {
		err := s.Create(ctx, session)
		if err == nil { t.Error("expected error for duplicate create") }
	})

	t.Run("update existing session", func(t *testing.T) {
		session.UserAgent = "updated-agent"
		err := s.Update(ctx, session)
		if err != nil { t.Fatalf("Update failed: %v", err) }

		got, _ := s.Get(ctx, "sess-1")
		if got.UserAgent != "updated-agent" { t.Errorf("expected updated UserAgent, got %q", got.UserAgent) }
	})

	t.Run("update non-existent session returns error", func(t *testing.T) {
		orphan := &domain.Session{ID: "ghost"}
		err := s.Update(ctx, orphan)
		if err == nil { t.Error("expected error for updating non-existent session") }
	})

	t.Run("delete session", func(t *testing.T) {
		err := s.Delete(ctx, "sess-1")
		if err != nil { t.Fatalf("Delete failed: %v", err) }

		_, err = s.Get(ctx, "sess-1")
		if err == nil { t.Error("expected session not found after delete") }
	})

	t.Run("delete non-existent session returns error", func(t *testing.T) {
		err := s.Delete(ctx, "nonexistent")
		if err == nil { t.Error("expected error for deleting non-existent session") }
	})
}
