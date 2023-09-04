package util

import (
	"time"
)

// NewTimedCache creates a new instance of a TimedCache with specified key and
// value types and returns a pointer to it.
//
// The TimedCache is a data structure that allows you to store key-value pairs
// with expiration times. You can specify the types of keys and values using
// the 'K' and 'E' type parameters. The TimedCache is initialized as an empty
// cache, ready to be used for caching values with specified expiration times.
//
// Example:
//
//	cache := NewTimedCache[string, int]()
//	// Creates a new TimedCache instance for string keys and int values.
//	// It is initially empty and ready to be used for caching values.
func NewTimedCache[K comparable, E any]() *TimedCache[K, E] {
	return &TimedCache[K, E]{
		entries: make(map[K]timedCacheEntry[E]),
	}
}

type TimedCache[K comparable, E any] struct {
	entries map[K]timedCacheEntry[E]
}

type timedCacheEntry[E any] struct {
	content E
	expires time.Time
}

// Get retrieves a value associated with a specified key from the TimedCache.
//
// If the key exists and has not expired, this function returns the value
// associated with the key and a boolean value 'true' to indicate success.
//
// If the key does not exist or has expired, it returns the zero value of the
// cache's value type and 'false'.
//
// Example:
//
//	cache := NewTimedCache[string, int]()
//	cache.Set("key1", 42, 10*time.Second)
//	value, exists := cache.Get("key1")
//	// 'value' will be 42, and 'exists' will be 'true' within the specified
//	// duration of 10 seconds, otherwise 'value' will be the zero value of int
//	// (0) and 'exists' will be 'false'.
func (c TimedCache[K, E]) Get(key K) (E, bool) {
	var content E

	entry, ok := c.entries[key]
	if !ok {
		return content, false
	}

	if time.Now().After(entry.expires) {
		delete(c.entries, key)

		return content, false
	}

	return entry.content, true
}

// Set adds a key-value pair to the TimedCache with a specified expiration time.
//
// It associates the given key of type K with the provided value of type E and
// sets an expiration time based on the specified duration. After this duration
// has passed, attempting to retrieve the value using the 'Get' method will
// return 'false' to indicate that the key has expired.
//
// Example:
//
//	cache := NewTimedCache[string, int]()
//	cache.Set("key1", 42, 10*time.Second)
//	// The value 42 is associated with "key1" and will be valid for 10 seconds.
//	// After that, using 'cache.Get("key1")' will return 'false'.
func (c TimedCache[K, E]) Set(key K, content E, duration time.Duration) {
	c.entries[key] = timedCacheEntry[E]{
		content: content,
		expires: time.Now().Add(duration * time.Second),
	}
}
