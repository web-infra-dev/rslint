package program

import (
	"runtime"
	"sync"
	"weak"

	"github.com/microsoft/typescript-go/shim/compiler"
)

// derivedCache belongs to one immutable source generation. Compiler-backed
// Programs that wrap the same ts-go Program share a weakly owned cache;
// parser-backed Programs own their cache directly. Cached values must not hold
// a strong reference to Program, because that would keep a weak compiler key
// alive through its own value.
type derivedCache struct {
	mu     sync.Mutex
	values map[any]*derivedValue
}

type derivedValue struct {
	once  sync.Once
	value any
}

var compilerCaches sync.Map // weak.Pointer[compiler.Program] -> *derivedCache

func cacheForCompiler(raw *compiler.Program) *derivedCache {
	key := weak.Make(raw)
	if stored, ok := compilerCaches.Load(key); ok {
		if cache, ok := stored.(*derivedCache); ok {
			return cache
		}
	}

	cache := &derivedCache{values: make(map[any]*derivedValue)}
	if stored, loaded := compilerCaches.LoadOrStore(key, cache); loaded {
		if existing, ok := stored.(*derivedCache); ok {
			return existing
		}
		return cache
	}
	runtime.AddCleanup(raw, func(dead weak.Pointer[compiler.Program]) {
		compilerCaches.Delete(dead)
	}, key)
	return cache
}

// Cached returns immutable data derived from p, building it once per key even
// when rules on different files request it concurrently. Keys should use an
// unexported named type owned by the consumer package.
//
// A compiler-backed cache is weakly owned by the underlying ts-go generation,
// so build must not return a value that strongly references p. Source files,
// symbols, module edges, and indexes derived from them are safe.
func Cached[T any](p *Program, key any, build func() T) T {
	if !p.IsValid() || p.cache == nil {
		return build()
	}
	p.cache.mu.Lock()
	if p.cache.values == nil {
		p.cache.values = make(map[any]*derivedValue)
	}
	entry := p.cache.values[key]
	if entry == nil {
		entry = &derivedValue{}
		p.cache.values[key] = entry
	}
	p.cache.mu.Unlock()

	entry.once.Do(func() { entry.value = build() })
	value, _ := entry.value.(T)
	return value
}
