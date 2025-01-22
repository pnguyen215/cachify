package cachify

import (
	"container/list"
	"time"
)

// NewLRU creates a new LRU cache with the specified capacity.
//
// Parameters:
//   - capacity: The maximum number of items the cache can hold.
//
// Returns:
//   - A pointer to an initialized LRU cache.
//
// Details:
//   - The cache uses a combination of a map and a doubly linked list for efficient
//     O(1) insertion, deletion, and lookup operations.
//   - Items are evicted based on the "least recently used" policy when the capacity is exceeded.
func NewLRU(capacity int) *LRU {
	return &LRU{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// NewLRUCallback creates a new LRU cache with the specified capacity and eviction callback.
//
// Parameters:
//   - capacity: The maximum number of items the cache can hold.
//   - callback: A function of type `OnCallback` that gets invoked when an item is evicted.
//
// Returns:
//   - A pointer to an initialized LRU cache.
//
// Details:
//   - The callback function is executed before an item is removed from the cache.
func NewLRUCallback(capacity int, callback OnCallback) *LRU {
	c := NewLRU(capacity)
	c.onEvict = callback
	return c
}

// NewLRUExpires creates a new LRU cache with a time-to-live for entries.
//
// Parameters:
//   - capacity: The maximum number of items the cache can hold.
//   - expiry: The expiration duration for each cache entry.
//
// Returns:
//   - A pointer to an initialized LRU cache.
//
// Details:
//   - Starts a background goroutine to periodically remove expired items.
func NewLRUExpires(capacity int, expiry time.Duration) *LRU {
	c := NewLRU(capacity)
	c.SetExpiry(expiry)
	c.stopCleanup = make(chan struct{})
	// Start a background goroutine for periodic cache cleanup
	go c.startCleanup()
	return c
}

// Get retrieves the value associated with a given key from the cache.
//
// Parameters:
//   - key: The key whose value is to be retrieved.
//
// Returns:
//   - The value associated with the key, or nil if the key is not found.
//   - A boolean indicating whether the key exists.
//
// Details:
//   - Moves the accessed item to the front of the list, marking it as most recently used.
//   - Evicts the item if it is expired (when expiration is enabled).
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

// GetAll retrieves all key-value pairs currently in the cache.
//
// Returns:
//   - A map containing all key-value pairs in the cache.
//
// Details:
//   - Does not modify the order of items in the cache.
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

// Pairs retrieves the least recently used key-value pair without removing it.
//
// Returns:
//   - The key and value of the least recently used item.
//   - A boolean indicating whether such an item exists.
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

// Set inserts or updates a key-value pair in the cache.
//
// Parameters:
//   - key: The key to be added or updated.
//   - value: The value to be associated with the key.
//
// Details:
//   - If the key exists, updates its value and moves it to the front of the list.
//   - If the key does not exist and the cache is full, evicts the least recently used item.
//   - The expiration time is reset or initialized based on the cache's expiration setting.
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

// Update updates the value associated with a key in the cache.
// Parameters:
//   - key: The key to update.
//   - value: The new value to associate with the key.
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

// Remove deletes a specific key-value pair from the cache.
//
// Parameters:
//   - key: The key to be removed.
//
// Details:
//   - If the key does not exist, the method does nothing.
func (c *LRU) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if element, exists := c.cache[key]; exists {
		c.evict(element)
	}
}

// Clear removes all key-value pairs from the cache.
//
// Details:
//   - Resets the internal data structures to their initial state.
func (c *LRU) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string]*list.Element)
	c.list.Init()
}

// Len returns the current number of items in the cache.
//
// Returns:
//   - The number of items in the cache.
func (c *LRU) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

// IsEmpty checks if the cache is empty.
//
// Returns:
//   - A boolean indicating whether the cache contains no items.
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
//
// Parameters:
//   - callback: A function of type `OnCallback` to be invoked when an item is evicted from the cache.
//
// Details:
//   - Uses write locking to ensure thread-safe updates to the `onEvict` field.
//   - Replaces the existing callback (if any) with the provided one.
//   - This callback function will be triggered during evictions, allowing custom behavior
//     (e.g., logging, cleanup, or persisting evicted data).
//
// Example Usage:
//
//	cache.SetCallback(func(key string, value interface{}) {
//	    fmt.Printf("Evicted: key=%s, value=%v\n", key, value)
//	})
func (c *LRU) SetCallback(callback OnCallback) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.onEvict = callback
}

// SetExpiry sets the default expiration duration for cache entries.
//
// Parameters:
//   - expiry: The duration after which a cache entry should expire.
//
// Details:
//   - This affects only new entries or updated entries after the call to SetExpiry.
//   - Existing entries retain their current expiration times until updated.
func (c *LRU) SetExpiry(expiry time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.expiration = expiry
}

// GetStates returns a snapshot of the current cache state.
//
// Returns:
//   - A slice of `state` objects representing all the items in the cache.
//   - Each `state` includes the key, value, access time, and expiration time.
//
// Details:
//   - Uses read locking to ensure safe concurrent access.
//   - Iterates through all cache entries, capturing their metadata.
//   - Creates a new `state` object for each entry using a builder-like pattern.
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

// GetState returns the metadata of the least recently used (LRU) item without removing it.
//
// Returns:
//   - A pointer to a `state` object representing the LRU item, or nil if the cache is empty.
//   - A boolean indicating whether a valid state was retrieved.
//
// Details:
//   - Uses read locking to safely access the cache state.
//   - Retrieves the least recently used item from the tail of the doubly-linked list.
//   - Constructs a `state` object to represent the item's metadata.
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
//
// Parameters:
//   - key: The key to check.
//
// Returns:
//   - true if the specified key is the most recently used item.
//   - false otherwise or if the cache is empty.
//
// Details:
//   - Uses read locking to safely access the cache state.
//   - Compares the provided key with the key of the item at the front of the list (MRU).
func (c *LRU) IsMostRecentlyUsed(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if e := c.list.Front(); e != nil {
		entry := e.Value.(*entries)
		return entry.key == key
	}
	return false
}

// GetMostRecentlyUsed returns the most recently used (MRU) key-value pair without removing it.
//
// Returns:
//   - A pointer to a `state` object representing the MRU item, or nil if the cache is empty.
//   - A boolean indicating whether a valid state was retrieved.
//
// Details:
//   - Uses read locking to safely access the cache state.
//   - Retrieves the most recently used item from the head of the doubly-linked list.
//   - Constructs a `state` object to represent the item's metadata.
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
//
// Parameters:
//   - key: The key of the cache entry to extend the expiration for.
//   - expiry: The duration by which to extend the expiration time.
//
// Details:
//   - Uses write locking to ensure safe updates.
//   - If the key exists, updates its expiration time and moves it to the front of the list.
//   - Does nothing if the key does not exist in the cache.
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
//
// Parameters:
//   - key: The key of the cache entry to check.
//
// Returns:
//   - The remaining time until the entry expires.
//   - A boolean indicating whether the key exists in the cache.
//
// Details:
//   - Uses read locking to safely access the cache state.
//   - If the key exists, calculates the time remaining until expiration.
//   - Returns 0 and false if the key does not exist.
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

// DestroyCleanup stops the background cleanup process.
//
// Details:
//   - Should be called when the cache is no longer needed to prevent goroutine leaks.
func (c *LRU) DestroyCleanup() {
	close(c.stopCleanup)
}

// evict removes a given element from the cache.
//
// Parameters:
//   - element: The list element to be removed.
//
// Details:
//   - Executes the eviction callback (if any) before removal.
func (c *LRU) evict(element *list.Element) {
	// Invoke the eviction callback before removing the item
	if c.onEvict != nil {
		entry := element.Value.(*entries)
		c.onEvict(entry.key, entry.value)
	}
	delete(c.cache, element.Value.(*entries).key)
	c.list.Remove(element)
}

// cleanupExpired removes all expired entries from the cache.
//
// Details:
//   - Iterates through all items and evicts those that have exceeded their expiration time.
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

// startCleanup starts a background goroutine to periodically remove expired entries.
//
// Details:
//   - Runs a cleanup operation at regular intervals to evict expired items.
//   - Stops when the `stopCleanup` channel is closed.
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

// calculateExpiry calculates the expiration time for a new cache entry.
//
// Returns:
//   - A time.Time value representing the expiration time.
//
// Details:
//   - If no expiration is set, returns the zero value for time.Time.
func (c *LRU) calculateExpiry() time.Time {
	if c.expiration > 0 {
		return time.Now().Add(c.expiration)
	}
	return time.Time{}
}
