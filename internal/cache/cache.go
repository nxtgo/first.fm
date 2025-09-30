package cache

import (
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	mu         sync.RWMutex
	items      map[K]*Item[V]
	defaultTTL time.Duration
	maxSize    int

	hits   uint64
	misses uint64
}

type Item[V any] struct {
	Value      V
	ExpiresAt  time.Time
	LastAccess time.Time
}

func New[K comparable, V any](defaultTTL time.Duration, maxSize int) *Cache[K, V] {
	c := &Cache[K, V]{
		items:      make(map[K]*Item[V]),
		defaultTTL: defaultTTL,
		maxSize:    maxSize,
	}

	if defaultTTL > 0 {
		go c.cleanupLoop()
	}

	return c
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		c.misses++
		var zero V
		return zero, false
	}

	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		c.misses++
		var zero V
		return zero, false
	}

	item.LastAccess = time.Now()
	c.hits++
	return item.Value, true
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.maxSize > 0 && len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	expiresAt := time.Time{}
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.items[key] = &Item[V]{
		Value:      value,
		ExpiresAt:  expiresAt,
		LastAccess: time.Now(),
	}
}

func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]*Item[V])
}

func (c *Cache[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *Cache[K, V]) Stats() (hits, misses uint64, size int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses, len(c.items)
}

func (c *Cache[K, V]) evictOldest() {
	var oldestKey K
	var oldestTime time.Time
	first := true

	for key, item := range c.items {
		if first || item.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.LastAccess
			first = false
		}
	}

	if !first {
		delete(c.items, oldestKey)
	}
}

func (c *Cache[K, V]) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

func (c *Cache[K, V]) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
			delete(c.items, key)
		}
	}
}
