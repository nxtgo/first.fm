package cache

import (
	"time"

	"github.com/nxtgo/gache"
)

// FuncCache is a generic wrapper around gache that caches function results.
type FuncCache[K comparable, V any] struct {
	cache *gache.Cache[K, V]
	ttl   time.Duration
}

// NewFuncCache creates a new FuncCache with a given TTL.
func NewFuncCache[K comparable, V any](ttl time.Duration) *FuncCache[K, V] {
	return &FuncCache[K, V]{
		cache: gache.New[K, V](ttl),
		ttl:   ttl,
	}
}

// GetOrFetch tries to get a value from cache, or calls fetcher if not present.
func (fc *FuncCache[K, V]) GetOrFetch(key K, fetcher func(K) (V, error)) (V, error) {
	if val, ok := fc.cache.Get(key); ok {
		return val, nil
	}

	val, err := fetcher(key)
	if err != nil {
		var zero V
		return zero, err
	}

	fc.cache.Set(key, val, fc.ttl)
	return val, nil
}
