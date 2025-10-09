package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("key not found")

type item struct {
	value      string
	expiration time.Time
}

type MemoryCache struct {
	items map[string]item
	mu    sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]item),
	}
	go cache.cleanupExpired()
	return cache
}

func (m *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.items[key]
	if !exists {
		return "", ErrNotFound
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return "", ErrNotFound
	}

	return item.value, nil
}

func (m *MemoryCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	m.items[key] = item{
		value:      value,
		expiration: exp,
	}

	return nil
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.items, key)
	return nil
}

func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.items[key]
	if !exists {
		return false, nil
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return false, nil
	}

	return true, nil
}

func (m *MemoryCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make(map[string]item)
	return nil
}

func (m *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for key, item := range m.items {
			if !item.expiration.IsZero() && now.After(item.expiration) {
				delete(m.items, key)
			}
		}
		m.mu.Unlock()
	}
}
