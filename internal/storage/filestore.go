package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type FileStore struct {
	mu    sync.RWMutex
	path  string
	ready bool
}

func NewFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "items.json")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			return nil, err
		}
	}

	return &FileStore{
		path:  path,
		ready: true,
	}, nil
}

func (f *FileStore) readAll() (map[string]Item, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return nil, err
	}

	var items map[string]Item
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (f *FileStore) writeAll(items map[string]Item) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(f.path, data, 0644)
}

func (f *FileStore) Ready() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	_, err := os.Stat(f.path)
	return err == nil && f.ready
}

func (f *FileStore) Create(name string, ttl time.Duration) Item {
	f.mu.Lock()
	defer f.mu.Unlock()

	items, _ := f.readAll()

	item := Item{
		ID:        uuid.NewString(),
		Name:      name,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	items[item.ID] = item
	_ = f.writeAll(items)

	return item
}

func (f *FileStore) Get(id string) (Item, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	items, err := f.readAll()
	if err != nil {
		return Item{}, err
	}

	item, ok := items[id]
	if !ok {
		return Item{}, ErrNotFound
	}

	return item, nil
}

func (f *FileStore) List() []Item {
	f.mu.RLock()
	defer f.mu.RUnlock()

	items, err := f.readAll()
	if err != nil {
		return nil
	}

	result := make([]Item, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result
}

func (f *FileStore) Delete(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	items, err := f.readAll()
	if err != nil {
		return err
	}

	if _, ok := items[id]; !ok {
		return ErrNotFound
	}

	delete(items, id)
	return f.writeAll(items)
}

func (f *FileStore) CleanupExpired() int {
	/* f.mu.Lock()
	defer f.mu.Unlock()

	items, err := f.readAll()
	if err != nil {
		return 0
	}

	now := time.Now()
	count := 0

	for id, item := range items {
		if item.ExpiresAt.Before(now) {
			delete(items, id)
			count++
		}
	}

	_ = f.writeAll(items)
	return count */
	return 0
}
