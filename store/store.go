package store

import (
	"errors"
	"sync"
	"time"
)

type Paste struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Language  string     `json:"language,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type Store struct {
	mu     sync.RWMutex
	pastes map[string]Paste
}

func NewStore() *Store {
	return &Store{
		pastes: make(map[string]Paste),
	}
}

var ErrAlreadyExists = errors.New("paste with this id already exists")

func (s *Store) Create(p Paste) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.pastes[p.ID]; ok {
		return ErrAlreadyExists
	}
	s.pastes[p.ID] = p
	return nil
}

func (s *Store) Get(id string) (Paste, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.pastes[id]
	if !ok {
		var zero Paste
		return zero, false
	}
	if expired(p) {
		delete(s.pastes, id)
		var zero Paste
		return zero, false
	}
	return p, true
}

func (s *Store) List() []Paste {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Paste, 0, len(s.pastes))
	for id, p := range s.pastes {
		if expired(p) {
			delete(s.pastes, id)
			continue
		}
		out = append(out, p)
	}
	return out
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.pastes[id]
	if !ok {
		return false
	}
	if expired(p) {
		delete(s.pastes, id)
		return false
	}
	delete(s.pastes, id)
	return true
}

func expired(p Paste) bool {
	return p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt)
}
