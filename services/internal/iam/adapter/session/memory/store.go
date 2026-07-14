package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*domain.Session),
	}
}

func (s *SessionStore) Create(_ context.Context, session *domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.sessions[session.ID]; exists {
		return domain.ErrSessionNotFound
	}
	s.sessions[session.ID] = session
	return nil
}

func (s *SessionStore) Get(_ context.Context, id string) (*domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	return session, nil
}

func (s *SessionStore) Update(_ context.Context, session *domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.sessions[session.ID]; !exists {
		return domain.ErrSessionNotFound
	}
	s.sessions[session.ID] = session
	return nil
}

func (s *SessionStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.sessions[id]; !exists {
		return domain.ErrSessionNotFound
	}
	delete(s.sessions, id)
	return nil
}
