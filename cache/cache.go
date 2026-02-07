package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

const cacheTTL = 7 * 24 * time.Hour // 7 days

var errCacheMiss = errors.New("cache: key not found")

type entry struct {
	value     string
	expiresAt time.Time
}

type MemoryCache struct {
	mu      sync.RWMutex
	items   map[string]entry
	closeCh chan struct{}
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		items:   make(map[string]entry),
		closeCh: make(chan struct{}),
	}
	go c.cleanup()
	return c
}

func (c *MemoryCache) Get(_ context.Context, key string) (string, error) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return "", errCacheMiss
	}
	return e.value, nil
}

func (c *MemoryCache) Set(_ context.Context, key, value string) error {
	c.mu.Lock()
	c.items[key] = entry{value: value, expiresAt: time.Now().Add(cacheTTL)}
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Close() error {
	close(c.closeCh)
	return nil
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			for k, e := range c.items {
				if now.After(e.expiresAt) {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		case <-c.closeCh:
			return
		}
	}
}
