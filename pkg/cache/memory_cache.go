package cache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value     any
	expiresAt time.Time
}

// MemoryCache is a thread-safe in-process key/value store with optional TTL per entry.
type MemoryCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

// NewMemoryCache creates an empty MemoryCache ready for use.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		entries: make(map[string]cacheEntry),
	}
}

// Get returns the cached value and true if the key exists and has not expired.
func (c *MemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.value, true
}

// Set stores value under key. A zero ttl means the entry never expires.
func (c *MemoryCache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := cacheEntry{value: value}
	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	}
	c.entries[key] = entry
}

// Delete removes key from the cache. No-op if the key does not exist.
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Flush removes all entries from the cache.
func (c *MemoryCache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]cacheEntry)
}
