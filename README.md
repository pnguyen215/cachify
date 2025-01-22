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

## Usage

### Cache Initialization

`NewLRU(capacity int)`: Create an LRU cache with a fixed capacity.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	// Create an LRU cache with a capacity of 3
	cache := cachify.NewLRU(3)

	// Add entries to the cache
	cache.Set("a", "alpha")
	cache.Set("b", "beta")
	cache.Set("c", "gamma")

	// Adding another entry will evict the least recently used ('a')
	cache.Set("d", "delta")

	// Inspect the cache state
	fmt.Println(cache.GetAll())
	// Output: map[b:beta c:gamma d:delta]
}
```

`NewLRUCallback(capacity int, callback OnCallback)`: Create an LRU cache with a custom eviction callback.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func onEviction(key string, value interface{}) {
	fmt.Printf("Evicted: Key=%s, Value=%v\n", key, value)
}

func main() {
	// Create an LRU cache with a capacity of 2 and a custom callback
	cache := cachify.NewLRUCallback(2, onEviction)

	cache.Set("x", "X-ray")
	cache.Set("y", "Yankee")

	// Adding a third entry triggers eviction of the least recently used ('x')
	cache.Set("z", "Zulu")
	// Output: Evicted: Key=x, Value=X-ray
}
```

`NewLRUExpires(capacity int, expiry time.Duration)`: Create an LRU cache with expiration.

eg.

```go
package main

import (
	"fmt"
	"time"
	"github.com/pnguyen215/cachify"
)

func main() {
	// Create an LRU cache with expiration (5 seconds)
	cache := cachify.NewLRUExpires(2, 5*time.Second)

	cache.Set("k", "keep")

	// Retrieve the key before expiration
	if val, ok := cache.Get("k"); ok {
		fmt.Println("Value:", val) // Output: Value: keep
	}

	// Wait for the key to expire
	time.Sleep(6 * time.Second)

	// Attempt to retrieve the key after expiration
	if _, ok := cache.Get("k"); !ok {
		fmt.Println("Key expired.") // Output: Key expired.
	}
}
```

### Cache Operations

`Get(key string)`: Retrieve an entry by key.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)
	cache.Set("a", "alpha")

	// Retrieve an existing key
	if value, ok := cache.Get("a"); ok {
		fmt.Println("Value:", value) // Output: Value: alpha
	}

	// Attempt to retrieve a non-existing key
	if _, ok := cache.Get("b"); !ok {
		fmt.Println("Key not found.") // Output: Key not found.
	}
}
```

`GetAll()`: Retrieve all key-value pairs.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)
	cache.Set("a", "alpha")
	cache.Set("b", "beta")

	// Get all key-value pairs
	fmt.Println("Cache state:", cache.GetAll())
	// Output: Cache state: map[a:alpha b:beta]
}
```

`Set(key string, value interface{})`: Add or update an entry.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(2)

	// Add entries to the cache
	cache.Set("a", "alpha")
	cache.Set("b", "beta")

	// Update an existing key
	cache.Set("a", "updated_alpha")

	// Inspect the cache state
	fmt.Println(cache.GetAll())
	// Output: map[a:updated_alpha b:beta]
}
```

`Update(key string, value interface{})`: Update an existing entry without altering order.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)

	cache.Set("x", "X-ray")
	cache.Update("x", "Updated X-ray")

	// Confirm the update
	if value, ok := cache.Get("x"); ok {
		fmt.Println("Updated value:", value) // Output: Updated value: Updated X-ray
	}
}
```

`Remove(key string)`: Remove a specific entry.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)

	cache.Set("a", "alpha")
	cache.Set("b", "beta")

	// Remove an entry
	cache.Remove("a")

	// Attempt to retrieve the removed entry
	if _, ok := cache.Get("a"); !ok {
		fmt.Println("Key 'a' removed.") // Output: Key 'a' removed.
	}
}
```

`Clear()`: Clear all entries from the cache.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)

	cache.Set("x", "X-ray")
	cache.Set("y", "Yankee")

	// Clear the cache
	cache.Clear()

	// Confirm the cache is empty
	if cache.Len() == 0 {
		fmt.Println("Cache cleared.") // Output: Cache cleared.
	}
}
```

`Len()`: Get the number of entries in the cache.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)

	cache.Set("a", "alpha")
	cache.Set("b", "beta")

	fmt.Println("Cache length:", cache.Len())
	// Output: Cache length: 2
}
```

`IsEmpty()`: Check if the cache is empty.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)

	fmt.Println("Is cache empty?", cache.IsEmpty())
	// Output: Is cache empty? true
}
```

`Pairs()`: Retrieve the least recently used pair.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)

	cache.Set("a", "alpha")
	cache.Set("b", "beta")

	// Retrieve the least recently used pair
	if key, value, ok := cache.Pairs(); ok {
		fmt.Printf("Least Recently Used: Key=%s, Value=%v\n", key, value)
		// Output: Least Recently Used: Key=a, Value=alpha
	}
}
```

### Advanced Features

`SetCapacity(capacity int)`: Dynamically adjust the cache capacity.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(2)

	cache.Set("a", "alpha")
	cache.Set("b", "beta")

	// Increase capacity to 3
	cache.SetCapacity(3)
	cache.Set("c", "gamma")

	fmt.Println("Cache state after capacity increase:", cache.GetAll())
	// Output: Cache state after capacity increase: map[a:alpha b:beta c:gamma]

	// Decrease capacity back to 2, evicting the least recently used ('a')
	cache.SetCapacity(2)
	fmt.Println("Cache state after capacity decrease:", cache.GetAll())
	// Output: Cache state after capacity decrease: map[b:beta c:gamma]
}
```

`SetCallback(callback OnCallback)`: Set or update the eviction callback function.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func evictionLogger(key string, value interface{}) {
	fmt.Printf("Evicted: Key=%s, Value=%v\n", key, value)
}

func main() {
	cache := cachify.NewLRU(2)

	// Set a callback for evictions
	cache.SetCallback(evictionLogger)

	cache.Set("x", "X-ray")
	cache.Set("y", "Yankee")
	cache.Set("z", "Zulu")
	// Output: Evicted: Key=x, Value=X-ray
}
```

`SetExpiry(expiry time.Duration)`: Update the expiration time for all cache entries.

eg.

```go
package main

import (
	"fmt"
	"time"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRUExpires(2, 5*time.Second)

	cache.Set("a", "alpha")

	// Update the expiration time to 10 seconds
	cache.SetExpiry(10 * time.Second)

	time.Sleep(2 * time.Second)
	if val, ok := cache.Get("a"); ok {
		fmt.Println("Value still exists:", val) // Output: Value still exists: alpha
	}
}
```

`GetStates()`: Get metadata for all entries.

eg.

```go
package main

import (
	"fmt"

	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)

	cache.Set("x", "X-ray")
	cache.Set("y", "Yankee")
	cache.Set("z", "Zulu")

	// Retrieve metadata for all entries
	for _, state := range cache.GetStates() {
		fmt.Printf("Key: %v, Value: %v, LastAccess: %v\n", state.Key(), state.Value(), state.AccessTime())
	}
}
```

`GetState()`: Get metadata for the least recently used item.

eg.

```go
package main

import (
	"fmt"

	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)

	cache.Set("x", "X-ray")
	cache.Set("y", "Yankee")

	// Retrieve metadata for the least recently used item
	if state, ok := cache.GetState(); ok {
		fmt.Printf("LRU State: Key=%s, Value=%v, LastAccess=%v\n", state.Key(), state.Value(), state.AccessTime())
	}
}
```

`IsMostRecentlyUsed(key string)`: Check if a key is the most recently used.

eg.

```go
package main

import (
	"fmt"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)

	cache.Set("a", "alpha")
	cache.Set("b", "beta")
	cache.Set("c", "gamma")

	if cache.IsMostRecentlyUsed("c") {
		fmt.Println("Key 'c' is the most recently used.") // Output: Key 'c' is the most recently used.
	}
}
```

`GetMostRecentlyUsed()`: Retrieve the most recently used item.

eg.

```go
package main

import (
	"fmt"

	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRU(3)

	cache.Set("a", "alpha")
	cache.Set("b", "beta")
	cache.Set("c", "gamma")

	// Retrieve the most recently used item
	if state, ok := cache.GetMostRecentlyUsed(); ok {
		fmt.Printf("Most Recently Used: Key=%s, Value=%v\n", state.Key(), state.Value())
		// Output: Most Recently Used: Key=c, Value=gamma
	}
}
```

`ExpandExpiry(key string, expiry time.Duration)`: Extend the expiration time for a specific key.

eg.

```go
package main

import (
	"fmt"
	"time"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRUExpires(2, 5*time.Second)

	cache.Set("a", "alpha")

	// Extend the expiration time by 10 seconds
	cache.ExpandExpiry("a", 10*time.Second)

	time.Sleep(2 * time.Second)
	if val, ok := cache.Get("a"); ok {
		fmt.Println("Value after extended expiry:", val) // Output: Value after extended expiry: alpha
	}
}
```

`PersistExpiry(key string)`: Retrieve the remaining time until expiration for a key.

eg.

```go
package main

import (
	"fmt"
	"time"
	"github.com/pnguyen215/cachify"
)

func main() {
	cache := cachify.NewLRUExpires(2, 10*time.Second)

	cache.Set("a", "alpha")

	// Check remaining time for expiration
	if remain, ok := cache.PersistExpiry("a"); ok {
		fmt.Printf("Time until expiration: %v\n", remain)
		// Output: Time until expiration: ~9s (actual value may vary slightly)
	}
}
```
