package lsp

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/config"
)

// tsConfigPathsForURI used to return an empty (non-nil)
// slice when a config matched but its tsconfig resolution turned up no
// paths. Callers that test `paths != nil` to decide whether type-aware
// rules can run (runLintWithSession in this file) would incorrectly
// treat that case as "type info present" and disable type-aware rules
// for every file under that config. The fix collapses empty-non-nil to
// nil so the nil-check is the single truth.
func TestTsConfigPathsForURI_EmptySliceNormalizedToNil(t *testing.T) {
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}

	// A config that exists in jsConfigs but resolves to NO tsconfig paths.
	// Without the empty-to-nil normalization, byConfig stores `[]string{}` here; with the fix,
	// the public accessor returns nil.
	s.jsConfigs = map[string]config.RslintConfig{
		"file:///project": {
			{
				LanguageOptions: &config.LanguageOptions{
					// No ParserOptions.Project → after rebuild, the
					// byConfig map either stores nil or [] depending
					// on the rebuild path. Test the accessor contract.
				},
			},
		},
	}
	// Manually pin the empty-non-nil case so we explicitly exercise the
	// failure mode that motivated the fix. rebuildTsConfigPaths might
	// produce nil naturally; we want to verify the GUARD on the way out.
	s.tsConfigPathsByConfig = map[string][]string{
		"file:///project": {}, // empty but non-nil — the danger
	}

	got := s.tsConfigPathsForURI("file:///project/src/index.ts")
	if got != nil {
		t.Errorf("expected nil (empty-non-nil collapsed), got %#v (len=%d)", got, len(got))
	}
}

func TestTsConfigPathsForURI_ServerLevelEmptyAlsoNil(t *testing.T) {
	// Same invariant for the no-jsConfigs branch (server-level
	// tsConfigPaths fallback). Empty server-level slice must also
	// normalize to nil so callers don't have to test both.
	s := newTestServer()
	s.fs = &mockFS{files: map[string]bool{}}
	s.jsConfigs = nil
	s.tsConfigPaths = []string{} // empty but non-nil

	got := s.tsConfigPathsForURI("file:///anywhere/x.ts")
	if got != nil {
		t.Errorf("expected nil (server-level empty collapsed), got %#v", got)
	}
}

// isTypeScriptFile was missing .mts/.cts/.mjs/.cjs.
// The CLI's createProgramsForConfig (programs.go) accepts all four
// modern Node module extensions; the LSP rejecting them silently caused
// "lint works from CLI but not in vscode" reports. Cover every extension
// the CLI accepts, plus a couple of must-not-match negatives.
func TestIsTypeScriptFile_CoversAllModernExtensions(t *testing.T) {
	cases := []struct {
		path string
		want bool
		why  string
	}{
		// Core extensions — covered before the fix and after.
		{"file:///x/a.ts", true, "core ts"},
		{"file:///x/a.tsx", true, "core tsx"},
		{"file:///x/a.js", true, "core js"},
		{"file:///x/a.jsx", true, "core jsx"},
		// Modern Node module variants — added by the fix.
		{"file:///x/a.mts", true, "ESM TS"},
		{"file:///x/a.cts", true, "CJS TS"},
		{"file:///x/a.mjs", true, "ESM JS"},
		{"file:///x/a.cjs", true, "CJS JS"},
		// Case-insensitive — must still match.
		{"file:///x/a.MTS", true, "upper-case .MTS"},
		// Must NOT match.
		{"file:///x/a.json", false, "json — not a source file"},
		{"file:///x/a.md", false, "markdown — not a source file"},
		{"file:///x/a.txt", false, "plain text"},
		{"file:///x/a.tsbuildinfo", false, "build-info file"},
		{"file:///x/a", false, "no extension"},
		{"file:///x/.ts", true, "dotfile that happens to be .ts (consistent: extension only)"},
	}
	for _, c := range cases {
		got := isTypeScriptFile(c.path)
		if got != c.want {
			t.Errorf("isTypeScriptFile(%q) = %v, want %v (%s)", c.path, got, c.want, c.why)
		}
	}
}

// Belt-suspenders: the file ends in the right extension but is wrapped
// in a query string (uncommon on file:// but cheap to verify) — current
// behavior is to NOT match because suffix check is byte-strict. Pin
// that explicitly so a future suffix-tolerant change is a conscious
// choice, not an accident.
func TestIsTypeScriptFile_QueryStringDefeatsSuffix(t *testing.T) {
	if isTypeScriptFile("file:///x/a.ts?foo=1") {
		t.Error("query-string in URI should currently defeat the strict suffix match")
	}
	// Verify we didn't accidentally allow `.ts.something`.
	if isTypeScriptFile("file:///x/a.ts.swp") {
		t.Error(".ts.swp must not match — only the final extension counts")
	}
	// Sanity: confirm trailing dot doesn't match.
	if isTypeScriptFile("file:///x/a.ts.") {
		t.Error("trailing dot must not match")
	}
	// Sanity: case folding works for capital extensions.
	if !isTypeScriptFile("file:///x/A.TSX") {
		t.Error("upper-case .TSX should match")
	}
	_ = strings.ToLower // import use
}
