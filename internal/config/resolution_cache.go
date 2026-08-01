package config

import "sync"

// publishOnceCache publishes only completely initialized values. Every lookup,
// including a losing LoadOrStore path, invokes the stored sync.OnceValue and
// therefore waits for the winning initializer instead of observing its slot.
type publishOnceCache[K comparable, V any] struct {
	entries sync.Map
}

type publishOnceEntry[V any] struct {
	value func() V
}

func (cache *publishOnceCache[K, V]) getOrInit(key K, init func() V) V {
	candidate := &publishOnceEntry[V]{value: sync.OnceValue(init)}
	actual, _ := cache.entries.LoadOrStore(key, candidate)
	// entries is private and every stored value is a publishOnceEntry[V].
	return actual.(*publishOnceEntry[V]).value() //nolint:forcetypeassert
}
