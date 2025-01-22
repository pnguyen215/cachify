<h1>cachify</h1>

`cachify` is a lightweight, high-performance, thread-safe **Least Recently Used (LRU)** cache library for Go. It is designed for in-memory caching with optional support for expiration, eviction callbacks, and dynamic capacity adjustment.

Whether you're optimizing resource usage, caching frequently accessed data, or adding session management, `cachify` simplifies your task with an intuitive and flexible API.

---

# Getting Started

## Requirements

Go version **1.19** or higher

## Installation

To start using `cachify`, run `go get`:

- For a specific version:

  ```bash
  go get https://github.com/pnguyen215/cachify@v0.0.1
  ```

- For the latest version (globally):
  ```bash
  go get -u https://github.com/pnguyen215/cachify@latest
  ```

## Key Features

- **LRU Eviction Policy:** Automatically removes the least recently used (LRU) items when the cache reaches capacity.
- **Expiration Support:** Set time-to-live (TTL) for cache entries for automatic expiration.
- **Custom Eviction Callback:** Trigger custom logic whenever an item is evicted.
- **Thread-safe Design:** Built for concurrent access with `sync.RWMutex`.
- **Dynamic Capacity Adjustment:** Adjust the cache capacity on the fly.
- **Comprehensive API:** Includes utility methods to inspect and manage cache state.

## API Documentation

### Cache Initialization

- `NewLRU(capacity int)`: Create an LRU cache with a fixed capacity.
- `NewLRUCallback(capacity int, callback OnCallback)`: Add a callback for evictions.
- `NewLRUExpires(capacity int, expiry time.Duration)`: Add entry expiration.

### Cache Operations

- `Get(key string) (value interface{}, ok bool)`: Retrieve an entry by key.
- `GetAll() map[string]interface{}`: Retrieve all key-value pairs.
- `Set(key string, value interface{})`: Add or update an entry.
- `Update(key string, value interface{})`: Update the value associated with a specific key in the cache.
- `Remove(key string)`: Remove a specific entry.
- `Clear()`: Clear all entries.
- `Len() int`: Get the number of entries in the cache.
- `IsEmpty() bool`: Check if the cache is empty.
- `IsExpired(key string) bool`: Check if a specific key has expired.
- `Contains(key string) bool`: Check if a key exists.
- `Pairs() (key string, value interface{}, ok bool)`: Get the least recently used pair.

### Advanced Features

- `SetCapacity(capacity int)`: Dynamically adjust the capacity.
- `SetCallback(callback OnCallback)`: Set the eviction callback function.
- `SetExpiry(expiry time.Duration)`: Update the expiration time for cache entries.
- `GetStates() []state`: Get metadata for all entries.
- `GetState() (m *state, ok bool)`: Get state returns the metadata of the least recently used item without removing it from the cache.
- `IsMostRecentlyUsed(key string) bool`: Check if a key is the most recently used.
- `GetMostRecentlyUsed() (state *state, ok bool)`: Retrieve the most recently used item.
- `ExpandExpiry(key string, expiry time.Duration)`: Extend the expiration time for a key.
- `PersistExpiry(key string) (remain time.Duration, ok bool)`: PersistExpiry returns the remaining time until expiration for a specific key.
