package rule

import (
	"runtime"
	"sync"
	"weak"

	"github.com/microsoft/typescript-go/shim/compiler"
)

// programCaches holds state derived from a Program for as long as that Program
// lives. Anything that is a property of the Program rather than of one file —
// a module's exports, a rule's whole-project index — belongs here, so that the
// first file of a run that needs it pays for it and the rest reuse it.
//
// The key is a weak pointer, so an entry never keeps its Program alive, and a
// cleanup drops the entry once the Program is collected. Callers must keep
// that property: nothing stored here may hold a strong reference back to the
// Program. Source files do not, which is why they are safe to store.
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

// CachedByTypeScriptProgram returns the value stored for one ts-go Program under key, calling
// build on the first request for it. build runs at most once per key even
// when rules on different files run concurrently, and only its key is
// untyped: what comes back is whatever T the caller asked for.
//
// A caller keys by whatever configuration changes the value it stores, so two
// rule configurations that disagree about the shape of that value get their
// own entries. Use an unexported key type so that keys from different packages
// cannot collide.
//
// The value is shared with every later caller, and rules on different files
// run concurrently, so build must return something no caller will mutate.
func CachedByTypeScriptProgram[T any](program *compiler.Program, key any, build func() T) T {
	if program == nil {
		return build()
	}
	return cachedValueFor(cacheFor(program), key, build)
}

// CachedByProgram shares derived state across every rule context backed by the
// same rslint Program. Compiler-backed Programs retain their weak ts-go
// lifetime cache; standalone Programs use the cache owned by their run-scoped
// module graph and cannot leak beyond that lint run.
func CachedByProgram[T any](ctx RuleContext, key any, build func() T) T {
	if typeScriptProgram := ctx.TypeScriptProgram(); typeScriptProgram != nil {
		return CachedByTypeScriptProgram(typeScriptProgram, key, build)
	}
	if ctx.Modules == nil {
		return build()
	}
	return cachedValueFor(&ctx.Modules.cache, key, build)
}

func cachedValueFor[T any](cache *programCache, key any, build func() T) T {
	cache.mu.Lock()
	if cache.values == nil {
		cache.values = make(map[any]*cachedValue)
	}
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
