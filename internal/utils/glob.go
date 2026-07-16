package utils

import (
	"sync"

	"github.com/bmatcuk/doublestar/v4"
)

var (
	globValidityCacheMu sync.RWMutex
	globValidityCache   = make(map[string]bool)
)

func isValidGlob(pattern string) bool {
	globValidityCacheMu.RLock()
	valid, ok := globValidityCache[pattern]
	globValidityCacheMu.RUnlock()
	if ok {
		return valid
	}

	valid = doublestar.ValidatePattern(pattern)
	globValidityCacheMu.Lock()
	globValidityCache[pattern] = valid
	globValidityCacheMu.Unlock()
	return valid
}

// MatchGlob reports whether path matches the doublestar glob pattern,
// matching doublestar.Match's boolean result. Pattern validity is checked
// once per distinct pattern and cached; invalid patterns short-circuit to
// false, and valid ones are matched with MatchUnvalidated to skip
// doublestar's per-call re-validation.
func MatchGlob(pattern, path string) bool {
	if !isValidGlob(pattern) {
		return false
	}
	return doublestar.MatchUnvalidated(pattern, path)
}
