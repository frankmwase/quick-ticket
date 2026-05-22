package security

import (
	"context"
	"sync"
	"time"

	"quick-ticket/app/ports"
)

// ──────────────────────────────────────────────────────────────────────────────
// InMemoryIdempotencyStore — Thread-safe in-memory implementation of the
// IdempotencyStore port. Suitable for single-instance deployments or tests.
// For multi-instance production, swap with a Redis-backed implementation.
// ──────────────────────────────────────────────────────────────────────────────

type idempotencyEntry struct {
	statusCode int
	body       []byte
	expiresAt  time.Time
}

type InMemoryIdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]idempotencyEntry
}

func NewInMemoryIdempotencyStore() ports.IdempotencyStore {
	store := &InMemoryIdempotencyStore{
		entries: make(map[string]idempotencyEntry),
	}
	// Start background cleanup goroutine
	go store.cleanup()
	return store
}

func (s *InMemoryIdempotencyStore) Exists(_ context.Context, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[key]
	if !ok {
		return false, nil
	}
	if time.Now().After(entry.expiresAt) {
		return false, nil
	}
	return true, nil
}

func (s *InMemoryIdempotencyStore) Get(_ context.Context, key string) (int, []byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[key]
	if !ok {
		return 0, nil, nil
	}
	if time.Now().After(entry.expiresAt) {
		return 0, nil, nil
	}
	return entry.statusCode, entry.body, nil
}

func (s *InMemoryIdempotencyStore) Set(_ context.Context, key string, statusCode int, body []byte, ttlSeconds int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[key] = idempotencyEntry{
		statusCode: statusCode,
		body:       body,
		expiresAt:  time.Now().Add(time.Duration(ttlSeconds) * time.Second),
	}
	return nil
}

// cleanup periodically removes expired entries to prevent unbounded memory growth.
func (s *InMemoryIdempotencyStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.entries {
			if now.After(entry.expiresAt) {
				delete(s.entries, key)
			}
		}
		s.mu.Unlock()
	}
}
