package cachify

import (
	"container/list"
	"time"
)

// NewLRU creates a new LRUCache with the specified capacity.
func NewLRU(capacity int) *LRU {
	return &LRU{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// NewLRUCache creates a new LRUCache with the specified capacity and an optional eviction callback.
func NewLRUCallback(capacity int, callback OnCallback) *LRU {
	c := NewLRU(capacity)
	c.onEvict = callback
	return c
}

// NewLRUCache creates a new LRUCache with the specified capacity, an optional eviction callback,
// and an optional time-to-live for cache entries.
func NewLRUExpires(capacity int, expiry time.Duration) *LRU {
	c := NewLRU(capacity)
	c.SetExpiry(expiry)
	c.stopCleanup = make(chan struct{})
	// Start a background goroutine for periodic cache cleanup
	go c.startCleanup()
	return c
}

// Get retrieves a value from the cache based on the key.
func (c *LRU) Get(key string) (value interface{}, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if element, exists := c.cache[key]; exists {
		// Check if the entry has expired
		if c.expiration > 0 && time.Now().After(element.Value.(*entries).expiration) {
			// If the entry has expired, evict it from the cache
			c.evict(element)
			return nil, false
		}
		// Move the accessed element to the front of the list (most recently used)
		c.list.MoveToFront(element)
		return element.Value.(*entries).value, true
	}
	return nil, false
}

// GetAll returns all key-value pairs in the cache.
func (c *LRU) GetAll() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	allEntries := make(map[string]interface{})
	for _, element := range c.cache {
		entry := element.Value.(*entries)
		allEntries[entry.key] = entry.value
	}
	return allEntries
}

// Pairs returns the least recently used key-value pair without removing it from the cache.
func (c *LRU) Pairs() (key string, value interface{}, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	oldest := c.list.Back()
	if oldest != nil {
		entry := oldest.Value.(*entries)
		return entry.key, entry.value, true
	}
	return "", nil, false
}

// Set adds a key-value pair to the cache. If the cache is full, it removes the least recently used item.
func (c *LRU) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if element, exists := c.cache[key]; exists {
		// Update the value and move the element to the front (most recently used)
		entry := element.Value.(*entries)
		entry.value = value
		entry.expiration = c.calculateExpiry()
		c.list.MoveToFront(element)
	} else {
		// Add a new element to the cache
		entry := &entries{
			key:        key,
			value:      value,
			expiration: c.calculateExpiry(),
		}
		element := c.list.PushFront(entry)
		c.cache[key] = element

		// If the cache is full, remove the least recently used item
		if len(c.cache) > c.capacity {
			oldest := c.list.Back()
			if oldest != nil {
				c.evict(oldest)
			}
		}
	}
}

// Update updates the value associated with a specific key in the cache.
func (c *LRU) Update(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if element, exists := c.cache[key]; exists {
		entry := element.Value.(*entries)
		entry.value = value
		entry.expiration = c.calculateExpiry()
		c.list.MoveToFront(element)
	}
}

// Remove removes a specific key from the cache.
func (c *LRU) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if element, exists := c.cache[key]; exists {
		c.evict(element)
	}
}

// Clear removes all items from the cache.
func (c *LRU) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string]*list.Element)
	c.list.Init()
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

// IsEmpty checks if the cache is empty.
func (c *LRU) IsEmpty() bool {
	return c.Len() == 0
}

// IsExpired checks if a specific key has expired without updating its access time.
func (c *LRU) IsExpired(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if element, exists := c.cache[key]; exists {
		entry := element.Value.(*entries)
		return c.expiration > 0 && time.Now().After(entry.expiration)
	}
	return false
}

// Contains checks if a key exists in the cache without updating its access time.
func (c *LRU) Contains(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, exists := c.cache[key]
	return exists
}

// SetCapacity updates the capacity of the cache.
// Allows you to dynamically update the capacity of the cache.
// If the new capacity is less than the current number of items, it removes the excess items from the cache.
func (c *LRU) SetCapacity(capacity int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.capacity = capacity
	// If the new capacity is less than the current number of items, remove the excess items
	for len(c.cache) > c.capacity {
		oldest := c.list.Back()
		if oldest != nil {
			c.evict(oldest)
		}
	}
}

// SetCallback sets the eviction callback function.
func (c *LRU) SetCallback(callback OnCallback) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.onEvict = callback
}

// SetExpiry updates the expiration time for cache entries.
func (c *LRU) SetExpiry(expiry time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.expiration = expiry
}

// GetStates returns a snapshot of the current cache state.
func (c *LRU) GetStates() []state {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	snapshot := make([]state, 0, len(c.cache))
	now := time.Now()
	for _, element := range c.cache {
		entry := element.Value.(*entries)
		l := NewState().
			WithKey(entry.key).
			WithValue(entry.value).
			WithAccessTime(now).
			WithExpiration(entry.expiration)
		snapshot = append(snapshot, *l)
	}
	return snapshot
}

// GetState returns the metadata of the least recently used item without removing it from the cache.
func (c *LRU) GetState() (m *state, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	oldest := c.list.Back()
	if oldest != nil {
		entry := oldest.Value.(*entries)
		l := NewState().
			WithKey(entry.key).
			WithValue(entry.value).
			WithExpiration(entry.expiration).
			WithAccessTime(time.Now())
		return l, true
	}
	return nil, false
}

// IsMostRecentlyUsed checks if a specific key is the most recently used item in the cache.
func (c *LRU) IsMostRecentlyUsed(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if e := c.list.Front(); e != nil {
		entry := e.Value.(*entries)
		return entry.key == key
	}
	return false
}

// GetMostRecentlyUsed returns the most recently used key-value pair without removing it from the cache.
func (c *LRU) GetMostRecentlyUsed() (m *state, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	newest := c.list.Front()
	if newest != nil {
		entry := newest.Value.(*entries)
		l := NewState().
			WithKey(entry.key).
			WithValue(entry.value).
			WithExpiration(entry.expiration).
			WithAccessTime(time.Now())
		return l, true
	}
	return nil, false
}

// ExpandExpiry extends the expiration time of a specific key in the cache.
func (c *LRU) ExpandExpiry(key string, expiry time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if element, exists := c.cache[key]; exists {
		entry := element.Value.(*entries)
		entry.expiration = entry.expiration.Add(expiry)
		c.list.MoveToFront(element)
	}
}

// PersistExpiry returns the remaining time until expiration for a specific key.
func (c *LRU) PersistExpiry(key string) (remain time.Duration, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if element, exists := c.cache[key]; exists {
		entry := element.Value.(*entries)
		if c.expiration > 0 {
			// remain = entry.expiration.Sub(time.Now())
			remain = time.Until(entry.expiration)
			return remain, true
		}
	}
	return 0, false
}

// DestroyCleanup stops the background goroutine for periodic cache cleanup.
func (c *LRU) DestroyCleanup() {
	close(c.stopCleanup)
}

// evict evicts an element from the cache.
func (c *LRU) evict(element *list.Element) {
	// Invoke the eviction callback before removing the item
	if c.onEvict != nil {
		entry := element.Value.(*entries)
		c.onEvict(entry.key, entry.value)
	}
	delete(c.cache, element.Value.(*entries).key)
	c.list.Remove(element)
}

// cleanupExpired removes expired entries from the cache.
func (c *LRU) cleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for _, element := range c.cache {
		entry := element.Value.(*entries)
		if entry.expiration.After(now) {
			// Entry has expired, evict it from the cache
			c.evict(element)
		}
	}
}

// startCleanup starts a background goroutine for periodic cache cleanup.
func (c *LRU) startCleanup() {
	ticker := time.NewTicker(c.expiration / 2) // Run cleanup at half the expiration interval
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// calculateExpiry calculates the expiration time for a cache entry.
func (c *LRU) calculateExpiry() time.Time {
	if c.expiration > 0 {
		return time.Now().Add(c.expiration)
	}
	return time.Time{}
}
