package users

import (
	"errors"
	"sync"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string // e.g., "user"
}

type Store struct {
	mu      sync.RWMutex
	seq     int64
	byEmail map[string]*User
}

func NewStore() *Store {
	return &Store{
		byEmail: make(map[string]*User),
	}
}

var (
	ErrEmailExists = errors.New("email already exists")
	ErrNotFound    = errors.New("user not found")
)

func (s *Store) Create(email, passwordHash string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byEmail[email]; ok {
		return nil, ErrEmailExists
	}
	s.seq++
	u := &User{
		ID:           s.seq,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         "user",
	}
	s.byEmail[email] = u
	return u, nil
}

func (s *Store) GetByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.byEmail[email]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}
