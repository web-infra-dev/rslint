package utils

import (
	"github.com/bmatcuk/doublestar/v4"
)

// MatchGlob reports whether path matches the doublestar glob pattern.
func MatchGlob(pattern, path string) bool {
	return doublestar.MatchUnvalidated(pattern, path)
}
