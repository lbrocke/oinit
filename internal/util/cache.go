package util

import (
	"time"
)

func NewTimedCache[K comparable, E any](expires time.Duration) *TimedCache[K, E] {
	return &TimedCache[K, E]{
		entries: make(map[K]timedCacheEntry[E]),
		expires: expires,
	}
}

type TimedCache[K comparable, E any] struct {
	entries map[K]timedCacheEntry[E]
	expires time.Duration
}

type timedCacheEntry[E any] struct {
	content E
	expires time.Time
}

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

func (c TimedCache[K, E]) Set(key K, content E) {
	c.entries[key] = timedCacheEntry[E]{
		content: content,
		expires: time.Now().Add(c.expires),
	}
}
