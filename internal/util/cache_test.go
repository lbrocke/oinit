package util

import (
	"testing"
	"time"
)

func TestTimedCache_Get(t *testing.T) {
	cache := NewTimedCache[string, int]()

	t.Run("Non-existing key", func(t *testing.T) {
		value, exists := cache.Get("key1")
		if exists {
			t.Errorf("Expected 'exists' to be false, but got true")
		}
		if value != 0 {
			t.Errorf("Expected value to be 0, but got %d", value)
		}
	})

	cache.Set("key2", 42, time.Duration(1))

	t.Run("Existing key with unexpired value", func(t *testing.T) {
		value, exists := cache.Get("key2")
		if !exists {
			t.Errorf("Expected 'exists' to be true, but got false")
		}
		if value != 42 {
			t.Errorf("Expected value to be 42, but got %d", value)
		}
	})

	t.Run("Existing key with expired value", func(t *testing.T) {
		// Sleep for more than 2 seconds to simulate expiration
		time.Sleep(2 * time.Second)
		value, exists := cache.Get("key2")
		if exists {
			t.Errorf("Expected 'exists' to be false, but got true")
		}
		if value != 0 {
			t.Errorf("Expected value to be 0, but got %d", value)
		}
	})
}

func TestTimedCache_Set(t *testing.T) {
	cache := NewTimedCache[string, int]()
	cache.Set("key1", 42, time.Duration(1))

	// Check that the value is set correctly
	value, exists := cache.Get("key1")
	if !exists {
		t.Errorf("Expected 'exists' to be true, but got false")
	}
	if value != 42 {
		t.Errorf("Expected value to be 42, but got %d", value)
	}

	// Check that the value expires after the specified duration
	time.Sleep(2 * time.Second)
	value, exists = cache.Get("key1")
	if exists {
		t.Errorf("Expected 'exists' to be false, but got true")
	}
	if value != 0 {
		t.Errorf("Expected value to be 0, but got %d", value)
	}
}

func TestTimedCache_SetWithNegativeDuration(t *testing.T) {
	cache := NewTimedCache[string, int]()

	t.Run("Setting with negative duration", func(t *testing.T) {
		cache.Set("key1", 42, time.Duration(-1))

		// Check that the value is not added to the cache
		value, exists := cache.Get("key1")
		if exists {
			t.Errorf("Expected 'exists' to be false, but got true")
		}
		if value != 0 {
			t.Errorf("Expected value to be 0, but got %d", value)
		}
	})
}
