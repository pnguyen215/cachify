package test

import (
	"testing"
	"time"

	"github.com/pnguyen215/cachify"
	"github.com/stretchr/testify/assert"
)

// Test cache creation and basic  functionality
func TestLRU_SetAndGet(t *testing.T) {
	cache := cachify.NewLRU(2)

	cache.Set("a", "alpha")
	cache.Set("b", "beta")

	// Verify the values
	val, ok := cache.Get("a")
	assert.True(t, ok)
	assert.Equal(t, "alpha", val)

	val, ok = cache.Get("b")
	assert.True(t, ok)
	assert.Equal(t, "beta", val)

	// Add a new entry to exceed capacity and evict the least recently used
	cache.Set("c", "gamma")
	_, ok = cache.Get("a")
	assert.False(t, ok) // "a" should be evicted
}

// Test eviction callback
func TestLRU_Callback(t *testing.T) {
	evicted := make(map[string]interface{})
	callback := func(key string, value interface{}) {
		evicted[key] = value
	}

	cache := cachify.NewLRUCallback(2, callback)

	cache.Set("x", "X-ray")
	cache.Set("y", "Yankee")
	cache.Set("z", "Zulu")

	// Verify eviction
	assert.Len(t, evicted, 1)
	assert.Equal(t, "X-ray", evicted["x"])
}

// Test expiration functionality
func TestLRU_Expiration(t *testing.T) {
	cache := cachify.NewLRUExpires(2, 12*time.Second)

	cache.Set("key", "value")
	time.Sleep(1 * time.Second)

	// Before expiration
	val, ok := cache.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "value", val)

	// After expiration
	time.Sleep(1 * time.Second)
	_, ok = cache.Get("key")
	assert.True(t, ok)
}

// Test dynamic capacity adjustment
func TestLRU_SetCapacity(t *testing.T) {
	cache := cachify.NewLRU(2)

	cache.Set("a", "alpha")
	cache.Set("b", "beta")

	// Increase capacity
	cache.SetCapacity(3)
	cache.Set("c", "gamma")

	assert.Equal(t, 3, cache.Len())
	assert.Contains(t, cache.GetAll(), "c")

	// Decrease capacity
	cache.SetCapacity(2)
	_, ok := cache.Get("a") // "a" should be evicted
	assert.False(t, ok)
}

// Test IsMostRecentlyUsed and GetMostRecentlyUsed
func TestLRU_MostRecentlyUsed(t *testing.T) {
	cache := cachify.NewLRU(3)

	cache.Set("a", "alpha")
	cache.Set("b", "beta")
	cache.Set("c", "gamma")

	assert.True(t, cache.IsMostRecentlyUsed("c"))

	state, ok := cache.GetMostRecentlyUsed()
	assert.True(t, ok)
	assert.Equal(t, "c", state.Key())
	assert.Equal(t, "gamma", state.Value())
}

// Test Clear and IsEmpty
func TestLRU_Clear(t *testing.T) {
	cache := cachify.NewLRU(2)

	cache.Set("a", "alpha")
	cache.Set("b", "beta")

	cache.Clear()
	assert.Equal(t, 0, cache.Len())
	assert.True(t, cache.IsEmpty())
}

// Test ExpandExpiry
func TestLRU_ExpandExpiry(t *testing.T) {
	cache := cachify.NewLRUExpires(2, 12*time.Second)

	cache.Set("key", "value")
	cache.ExpandExpiry("key", 3*time.Second)

	time.Sleep(1 * time.Second)
	val, ok := cache.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "value", val)
}

// Test PersistExpiry
func TestLRU_PersistExpiry(t *testing.T) {
	cache := cachify.NewLRUExpires(2, 5*time.Second)

	cache.Set("key", "value")
	remain, ok := cache.PersistExpiry("key")
	assert.True(t, ok)
	assert.LessOrEqual(t, remain, 5*time.Second)
}

// Test Update
func TestLRU_Update(t *testing.T) {
	cache := cachify.NewLRU(2)

	cache.Set("key", "old_value")
	cache.Update("key", "new_value")

	val, ok := cache.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "new_value", val)
}

// Test Remove
func TestLRU_Remove(t *testing.T) {
	cache := cachify.NewLRU(2)

	cache.Set("a", "alpha")
	cache.Remove("a")

	_, ok := cache.Get("a")
	assert.False(t, ok)
}
