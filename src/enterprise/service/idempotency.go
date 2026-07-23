package service

import "sync"

type IdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]bool
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{entries: make(map[string]bool)}
}

func (s *IdempotencyStore) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entries[key]
}

func (s *IdempotencyStore) Put(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = true
}
