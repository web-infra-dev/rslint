package utils

import "sync"

// LazyMap memoizes one value per key. Its zero value is ready to use.
//
// A missing value is built outside the lock, so two goroutines racing on the
// same key cost one redundant build at worst rather than serializing every key
// in the map behind one mutex. Whichever build finishes first is the one every
// caller gets, so the build must be pure and its result must be treated as
// read-only: one value is shared with everyone who asks for that key.
type LazyMap[K comparable, V any] struct {
	mu     sync.Mutex
	values map[K]V
}

// Get returns the value stored under key, calling build for the first caller
// to ask for it.
func (m *LazyMap[K, V]) Get(key K, build func() V) V {
	m.mu.Lock()
	value, ok := m.values[key]
	m.mu.Unlock()
	if ok {
		return value
	}

	value = build()

	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.values[key]; ok {
		return existing
	}
	if m.values == nil {
		m.values = make(map[K]V)
	}
	m.values[key] = value
	return value
}
