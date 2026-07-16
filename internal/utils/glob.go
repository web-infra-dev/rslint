package utils

import (
	"sync"

	"github.com/bmatcuk/doublestar/v4"
)

var globValidityCache sync.Map // map[string]bool

// MatchGlob reports whether path matches the doublestar glob pattern,
// matching doublestar.Match's boolean result. Pattern validity is checked
// once per distinct pattern and cached; invalid patterns short-circuit to
// false, and valid ones are matched with MatchUnvalidated to skip
// doublestar's per-call re-validation.
func MatchGlob(pattern, path string) bool {
	valid, ok := globValidityCache.Load(pattern)
	if !ok {
		valid = doublestar.ValidatePattern(pattern)
		globValidityCache.Store(pattern, valid)
	}
	if !valid.(bool) {
		return false
	}
	return doublestar.MatchUnvalidated(pattern, path)
}
