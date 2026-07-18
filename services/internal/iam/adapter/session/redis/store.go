package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/saleforge/pos/services/internal/iam/domain"
	sessionmem "github.com/saleforge/pos/services/internal/iam/adapter/session/memory"
	"github.com/saleforge/pos/services/pkg/logger"
)

type SessionStore struct {
	rdb      *redis.Client
	fallback *sessionmem.SessionStore
}

func NewSessionStore(addr string) *SessionStore {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("[session/redis] connection failed, falling back to in-memory", "error", err.Error())
		return &SessionStore{
			fallback: sessionmem.NewSessionStore(),
		}
	}

	logger.Info("[session/redis] connected to redis", "addr", addr)
	return &SessionStore{
		rdb: rdb,
	}
}

func sessionKey(id string) string {
	return fmt.Sprintf("session:%s", id)
}

func (s *SessionStore) Create(ctx context.Context, session *domain.Session) error {
	if s.rdb == nil {
		return s.fallback.Create(ctx, session)
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl < 0 {
		ttl = 0
	}

	return s.rdb.Set(ctx, sessionKey(session.ID), data, ttl).Err()
}

func (s *SessionStore) Get(ctx context.Context, id string) (*domain.Session, error) {
	if s.rdb == nil {
		return s.fallback.Get(ctx, id)
	}

	data, err := s.rdb.Get(ctx, sessionKey(id)).Bytes()
	if err != nil {
		return s.fallback.Get(ctx, id)
	}

	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &session, nil
}

func (s *SessionStore) Update(ctx context.Context, session *domain.Session) error {
	if s.rdb == nil {
		return s.fallback.Update(ctx, session)
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl < 0 {
		ttl = 0
	}

	return s.rdb.Set(ctx, sessionKey(session.ID), data, ttl).Err()
}

func (s *SessionStore) Delete(ctx context.Context, id string) error {
	if s.rdb == nil {
		return s.fallback.Delete(ctx, id)
	}

	return s.rdb.Del(ctx, sessionKey(id)).Err()
}
