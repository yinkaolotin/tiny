package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("item not found")

type Item struct {
	ID        string
	Name      string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]Item
	ready bool
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items: make(map[string]Item),
		ready: true,
	}
}

func (s *MemoryStore) Ready() bool {
	return s.ready
}

func (s *MemoryStore) Create(name string, ttl time.Duration) Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := Item{
		ID:        uuid.NewString(),
		Name:      name,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	s.items[item.ID] = item
	return item
}

func (s *MemoryStore) Get(id string) (Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	if !ok {
		return Item{}, ErrNotFound
	}
	return item, nil
}

func (s *MemoryStore) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Item, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, item)
	}
	return result
}

func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}

func (s *MemoryStore) CleanupExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0
	for id, item := range s.items {
		if item.ExpiresAt.Before(now) {
			delete(s.items, id)
			count++
		}
	}
	return count
}
