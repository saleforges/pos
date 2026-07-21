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

type SessionStore struct {
	rdb *redis.Client
}

func NewSessionStore(addr string) *SessionStore {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("[session/redis] connection failed", "error", err.Error())
		return &SessionStore{rdb: nil}
	}

	logger.Info("[session/redis] connected", "addr", addr)
	return &SessionStore{rdb: rdb}
}

func sessionKey(id string) string {
	return fmt.Sprintf("session:%s", id)
}

func (s *SessionStore) Create(ctx context.Context, session *domain.Session) error {
	if s.rdb == nil {
		return fmt.Errorf("redis session store not available")
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
		return nil, fmt.Errorf("redis session store not available")
	}

	data, err := s.rdb.Get(ctx, sessionKey(id)).Bytes()
	if err != nil {
		return nil, err
	}

	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &session, nil
}

func (s *SessionStore) Update(ctx context.Context, session *domain.Session) error {
	if s.rdb == nil {
		return fmt.Errorf("redis session store not available")
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
		return nil
	}
	return s.rdb.Del(ctx, sessionKey(id)).Err()
}
