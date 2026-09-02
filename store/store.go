package store

import (
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

func (s *Store) Create(p Paste) error {
	return nil
}

func (s *Store) Get(id string) (Paste, bool) {
	var zero Paste
	return zero, false
}

func (s *Store) List() []Paste {
	return nil
}

func (s *Store) Delete(id string) bool {
	return false
}
