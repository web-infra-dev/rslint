package regexp

import (
	"strings"
	"sync"
)

// Bound process-wide retention for long-lived API and LSP hosts. Large or
// expanding patterns still compile normally, but are not retained in the cache.
const (
	patternCacheCapacity  = 256
	maxCachedSourceBytes  = 4 << 10
	maxCachedRewriteBytes = 64 << 10
)

type patternKey struct {
	source string
	flags  string
}

type patternResult struct {
	pattern *compiledPattern
	err     error
}

type patternCache struct {
	mu      sync.RWMutex
	entries map[patternKey]patternResult
	order   [patternCacheCapacity]patternKey
	next    int
}

var patterns patternCache

func (c *patternCache) compile(key patternKey) (*compiledPattern, error) {
	if len(key.source)+len(key.flags) > maxCachedSourceBytes {
		return compilePattern(key)
	}
	c.mu.RLock()
	result, ok := c.entries[key]
	c.mu.RUnlock()
	if ok {
		return result.pattern, result.err
	}

	// A source may be a small slice of an entire linted file. Retain only the
	// pattern, including any substrings captured by parsing or error messages.
	key = patternKey{source: strings.Clone(key.source), flags: strings.Clone(key.flags)}
	pattern, err := compilePattern(key)
	if pattern != nil && len(pattern.rewritten) > maxCachedRewriteBytes {
		return pattern, err
	}

	// Expensive compilation stays outside the lock. Concurrent cold misses may
	// compile independently; publish one result and discard redundant work.
	c.mu.Lock()
	defer c.mu.Unlock()
	if result, ok = c.entries[key]; ok {
		return result.pattern, result.err
	}
	if c.entries == nil {
		c.entries = make(map[patternKey]patternResult, patternCacheCapacity)
	}
	if len(c.entries) == patternCacheCapacity {
		delete(c.entries, c.order[c.next])
	}
	c.entries[key] = patternResult{pattern: pattern, err: err}
	c.order[c.next] = key
	c.next = (c.next + 1) % patternCacheCapacity
	return pattern, err
}
