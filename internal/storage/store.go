package storage

import "time"

type Store interface {
	Ready() bool
	Create(name string, ttl time.Duration) Item
	Get(id string) (Item, error)
	List() []Item
	Delete(id string) error
	CleanupExpired() int
}
