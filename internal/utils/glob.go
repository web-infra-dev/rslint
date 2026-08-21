package utils

import (
	//nolint:depguard // Neither dialect the guard names is the one read here.
	"github.com/bmatcuk/doublestar/v4"
)

// MatchGlob reports whether path matches the doublestar glob pattern.
//
// `no-restricted-imports`, the only rule that calls this, hands its patterns to
// the `ignore` package, which reads a pattern the way a .gitignore file is
// read. That is a third dialect, so neither doublestar nor minimatch 3 is what
// upstream answers with and the guard's advice does not apply.
func MatchGlob(pattern, path string) bool {
	return doublestar.MatchUnvalidated(pattern, path)
}
