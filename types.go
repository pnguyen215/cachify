package cachify

import (
	"container/list"
	"sync"
	"time"
)

// OnCallback is a callback function type that gets called when an item is evicted from the cache.
type OnCallback func(key string, value interface{})

// LRU represents the LRU cache.
type LRU struct {
	capacity    int
	cache       map[string]*list.Element
	list        *list.List
	mutex       sync.RWMutex
	onEvict     OnCallback
	expiration  time.Duration
	stopCleanup chan struct{}
}

// state represents the state of the least recently used item.
type state struct {
	key        string
	value      interface{}
	accessTime time.Time
	expiration time.Time
}

// entries represents a key-value pair in the cache.
type entries struct {
	key        string
	value      interface{}
	expiration time.Time
}
