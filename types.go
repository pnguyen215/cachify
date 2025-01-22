package cachify

import (
	"container/list"
	"sync"
	"time"
)

// OnCallback is a callback function type that gets called when an item is evicted from the cache.
// Parameters:
//   - key: The key of the item being evicted.
//   - value: The value associated with the key.
type OnCallback func(key string, value interface{})

// LRU represents an implementation of a Least Recently Used (LRU) cache.
// It provides thread-safe operations, optional entry expiration, and an eviction callback.
//
// Fields:
//   - capacity: The maximum number of items the cache can hold.
//   - cache: A map for quick access to cache entries by key.
//   - list: A doubly linked list for maintaining access order.
//   - mutex: A read-write lock to ensure thread-safe operations.
//   - onEvict: An optional callback function invoked when an item is evicted.
//   - expiration: The duration for which entries are valid in the cache. Zero means no expiration.
//   - stopCleanup: A channel used to signal stopping of the background cleanup goroutine.
type LRU struct {
	capacity    int
	cache       map[string]*list.Element
	list        *list.List
	mutex       sync.RWMutex
	onEvict     OnCallback
	expiration  time.Duration
	stopCleanup chan struct{}
}

// state represents metadata about the least recently used item.
// Fields:
//   - key: The key of the cache entry.
//   - value: The value associated with the key.
//   - accessTime: The last time the entry was accessed.
//   - expiration: The expiration time of the entry.
type state struct {
	key        string
	value      interface{}
	accessTime time.Time
	expiration time.Time
}

// entries represents a cache entry with associated metadata.
// Fields:
//   - key: The key of the entry.
//   - value: The value associated with the key.
//   - expiration: The expiration time of the entry.
type entries struct {
	key        string
	value      interface{}
	expiration time.Time
}
