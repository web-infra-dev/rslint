package utils

import (
	"runtime"
	"sync"
	"weak"

	"github.com/microsoft/typescript-go/shim/compiler"
)

// programCaches holds each Program's derived import state for as long as that
// Program lives. The module graph the core linter owns covers what a file
// imports, which every rule may want; what stays here is the part only these
// rules know how to read — a file's own exports, the merged export maps built
// from them, and whatever a single rule derives on top.
//
// The key is a weak pointer, so an entry never keeps its Program alive, and a
// cleanup drops the entry once the Program is collected. Nothing reachable
// from an entry points back at the Program: source files do not, and neither
// the index nor the values callers store hold a reference of their own.
var programCaches sync.Map // weak.Pointer[compiler.Program] -> *programCache

type programCache struct {
	mu     sync.Mutex
	values map[any]*cachedValue
}

type cachedValue struct {
	once  sync.Once
	value any
}

func cacheFor(program *compiler.Program) *programCache {
	key := weak.Make(program)
	if stored, ok := programCaches.Load(key); ok {
		if entry, ok := stored.(*programCache); ok {
			return entry
		}
	}

	entry := &programCache{values: make(map[any]*cachedValue)}
	if stored, loaded := programCaches.LoadOrStore(key, entry); loaded {
		if existing, ok := stored.(*programCache); ok {
			return existing
		}
		return entry
	}

	// The cleanup takes the weak key rather than the Program, so registering
	// it does not make the Program permanently reachable.
	runtime.AddCleanup(program, func(dead weak.Pointer[compiler.Program]) {
		programCaches.Delete(dead)
	}, key)
	return entry
}

// CachedByProgram returns the value stored for one Program under key, calling
// build on the first request for it. build runs at most once per key even
// when rules on different files run concurrently, and only its key is
// untyped: what comes back is whatever T the caller asked for.
//
// A rule keys by whatever configuration changes the value it stores, so two
// rule configurations that disagree about the shape of that value get their
// own entries. Use an unexported key type so that keys from different
// packages cannot collide.
func CachedByProgram[T any](program *compiler.Program, key any, build func() T) T {
	if program == nil {
		return build()
	}

	cache := cacheFor(program)
	cache.mu.Lock()
	entry, ok := cache.values[key]
	if !ok {
		entry = &cachedValue{}
		cache.values[key] = entry
	}
	cache.mu.Unlock()

	// Built outside the cache lock so that one long build does not hold up
	// every other key; the once still admits exactly one builder per key.
	entry.once.Do(func() { entry.value = build() })
	value, _ := entry.value.(T)
	return value
}
