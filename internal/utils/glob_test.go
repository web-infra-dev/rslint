package utils

import "testing"

func TestMatchGlob(t *testing.T) {
	type c struct {
		name    string
		pattern string
		path    string
		want    bool
	}
	cases := []c{
		// ---- Plain wildcards ----
		{"star match", "*.go", "main.go", true},
		{"star nomatch extension", "*.go", "main.ts", false},
		{"star nomatch separator", "*.go", "src/main.go", false},
		{"question one char", "fil?.txt", "file.txt", true},
		{"question nomatch length", "fil?.txt", "file.txt.txt", false},

		// ---- Globstar (recursive) ----
		{"globstar deep match", "src/**/*.go", "src/a/b/main.go", true},
		{"globstar zero dirs match", "src/**/*.go", "src/main.go", true},
		{"globstar nomatch other root", "src/**/*.go", "lib/main.go", false},

		// ---- Character class ----
		{"class match", "[abc].txt", "a.txt", true},
		{"class nomatch", "[abc].txt", "d.txt", false},
		{"class range match", "[a-c].txt", "b.txt", true},
		{"class range nomatch", "[a-c].txt", "d.txt", false},
		{"negated class match", "[!a].txt", "b.txt", true},
		{"negated class nomatch", "[!a].txt", "a.txt", false},

		// ---- Brace expansion ----
		{"brace match a", "{dist,build}/**", "dist/index.js", true},
		{"brace match b", "{dist,build}/**", "build/index.js", true},
		{"brace nomatch", "{dist,build}/**", "src/index.js", false},

		// ---- Empty inputs ----
		{"empty path nomatch", "*.go", "", false},
		{"empty pattern nomatch", "", "main.go", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MatchGlob(tc.pattern, tc.path)
			if got != tc.want {
				t.Errorf("MatchGlob(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
			}
		})
	}
}

// TestMatchGlob_MalformedPatternFalsePositives documents cases where
// MatchGlob (backed by doublestar.MatchUnvalidated, which skips pattern
// validation) diverges from doublestar.Match on malformed patterns —
// an unbalanced "[" inside a brace alternative can make an otherwise
// non-matching path match. doublestar.Match rejects these patterns
// with ErrBadPattern instead. Skipped until we decide whether/how to
// re-add validation; see PR #1280 for the tradeoff this reverted.
func TestMatchGlob_MalformedPatternFalsePositives(t *testing.T) {
	t.Skip("known false positives on malformed patterns; fix tracked separately")

	type c struct {
		name    string
		pattern string
		path    string
		want    bool
	}
	cases := []c{
		{"unbalanced bracket in brace alt", "{dist[,build}/**", "build/index.js", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MatchGlob(tc.pattern, tc.path)
			if got != tc.want {
				t.Errorf("MatchGlob(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
			}
		})
	}
}
