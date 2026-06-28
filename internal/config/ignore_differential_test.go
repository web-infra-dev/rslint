package config

import (
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
)

// Permanent differential guard for the IgnorePattern refactor.
//
// This refactor replaced the pre-refactor []string ignore pipeline
// (isFileIgnored / isFileLevelPattern / isDirPathBlocked, each re-deriving
// normalization and directory-role from the raw string per call) with the
// structured IgnorePattern pipeline (ParseIgnorePattern → Glob/Negated/Kind,
// consumed by isFileIgnored / isDirAbsolutelyBlocked). The refactor's contract
// is byte-for-byte identical LINT decisions; the `./*` regression slipped past
// one review round precisely because nothing pinned old==new.
//
// The functions below (mainX_*) are vendored verbatim from main @930924a7's
// internal/config/config.go. They are the OLD oracle. We assert that the OLD
// per-file lint decision (dir-block OR file-ignore) equals the NEW one over a
// corpus engineered to stress every divergence-prone edge:
//   - normalize-unstable raw patterns (`./`, `//`, `/.`, `/..`) whose suffix
//     class differs before vs after normalization (the `./*` class),
//   - the three suffix roles (X/**/* file-level cover, X/* single level,
//     X/** absolute) and their `!`-negated forms,
//   - glob metachars (`?`, `[abc]`, `{a,b}`, `**`) in both name and extension
//     position,
//   - sequential `!` override ordering, rooted vs unrooted.
//
// If a future change to ParseIgnorePattern / isDirAbsolutelyBlocked /
// isFileIgnored reintroduces a `./*`-class misclassification, the new decision
// diverges from this frozen oracle and the test fails. Keep mainX_* frozen — do
// NOT "fix" them to track the new code; that would defeat the guard.

// --- OLD oracle, verbatim from main @930924a7 (operates on []string) ---

func mainX_isFileIgnored(filePath string, ignorePatterns []string, cwd string) bool {
	if cwd == "" {
		return mainX_isFileIgnoredSimple(filePath, ignorePatterns)
	}
	normalizedPath := normalizePath(filePath, cwd)
	unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")
	ignored := false
	for _, pattern := range ignorePatterns {
		negated := false
		if strings.HasPrefix(pattern, "!") {
			negated = true
			pattern = pattern[1:]
		}
		normalizedPattern := normalizePattern(pattern)
		matched := matchGlob(normalizedPattern, normalizedPath)
		if !matched && unixPath != normalizedPath {
			matched = matchGlob(normalizedPattern, unixPath)
		}
		if matched {
			ignored = !negated
		}
	}
	return ignored
}

func mainX_isFileLevelPattern(pattern string) bool {
	return strings.HasSuffix(pattern, "/**/*") ||
		(strings.HasSuffix(pattern, "/*") && !strings.HasSuffix(pattern, "/**"))
}

func mainX_isDirBlockedByIgnores(filePath string, ignorePatterns []string, cwd string) bool {
	var dirPath string
	if cwd != "" {
		dirPath = normalizePath(tspath.GetDirectoryPath(filePath), cwd)
	} else {
		dirPath = tspath.GetDirectoryPath(filePath)
	}
	dirPath = strings.ReplaceAll(dirPath, "\\", "/")
	dirPath = strings.TrimSuffix(dirPath, "/")
	if dirPath == "" || dirPath == "." {
		return false
	}
	return mainX_isDirPathBlocked(dirPath, ignorePatterns)
}

func mainX_isDirPathBlocked(dirPath string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		if pattern == "" || strings.HasPrefix(pattern, "!") {
			continue
		}
		if mainX_isFileLevelPattern(pattern) {
			continue
		}
		normalizedPattern := normalizePattern(pattern)
		if matchGlob(normalizedPattern, dirPath) || matchGlob(normalizedPattern, dirPath+"/x") {
			return true
		}
		segments := strings.Split(dirPath, "/")
		for i := 1; i < len(segments); i++ {
			partial := strings.Join(segments[:i], "/")
			if matchGlob(normalizedPattern, partial) || matchGlob(normalizedPattern, partial+"/x") {
				return true
			}
		}
	}
	return false
}

func mainX_isFileIgnoredSimple(filePath string, ignorePatterns []string) bool {
	ignored := false
	for _, pattern := range ignorePatterns {
		negated := false
		if strings.HasPrefix(pattern, "!") {
			negated = true
			pattern = pattern[1:]
		}
		normalizedPattern := normalizePattern(pattern)
		if matched, err := doublestar.Match(normalizedPattern, filePath); err == nil && matched {
			ignored = !negated
		}
	}
	return ignored
}

// oldDecision is the OLD linter's per-file ignore decision: a file is excluded
// iff its directory is absolutely blocked OR it is file-ignored. Mirrors the
// pre-refactor GetConfigForFile/IsFileIgnored composition.
func oldDecision(filePath string, patterns []string, cwd string) bool {
	return mainX_isDirBlockedByIgnores(filePath, patterns, cwd) ||
		mainX_isFileIgnored(filePath, patterns, cwd)
}

// newDecision is the SAME composition on the structured pipeline.
func newDecision(filePath string, patterns []string, cwd string) bool {
	parsed := ParseIgnorePatterns(patterns)
	return isDirBlockedByIgnores(filePath, parsed, cwd) ||
		isFileIgnored(filePath, parsed, cwd)
}

// differentialCorpus pairs each ignore-pattern set with probe paths chosen so
// the dir-block vs file-ignore split, the suffix classes, and the normalize
// edges all matter. Every (patternSet × path) must yield old==new.
var differentialCorpus = []struct {
	name     string
	patterns []string
	paths    []string
}{
	// The regression itself: `./*` normalizes to bare `*`. Old isFileLevelPattern
	// sees raw `./*` ends `/*` → file-level → NOT a dir block. New classifies raw
	// `./*` → dirNone likewise. A nested file must stay lintable under both.
	{"dotstar regression", []string{"./*"}, []string{
		"src/app/main.ts", "src/main.ts", "top.ts", "a/b/c/d.ts",
	}},
	{"dotstar double-slash", []string{".//*"}, []string{"src/app/main.ts", "top.ts"}},
	{"dotstar negated", []string{"**/*", "!./*"}, []string{"src/main.ts", "top.ts"}},

	// Suffix roles + negation. file-level cover allows `!` re-include; absolute
	// block does not; single-level covers one depth only.
	{"file-level cover", []string{"target/**/*"}, []string{
		"target/a.ts", "target/sub/b.ts", "target.ts", "src/a.ts",
	}},
	{"file-level cover negated child", []string{"target/**/*", "!target/keep/**/*"}, []string{
		"target/a.ts", "target/keep/k.ts", "target/keep/deep/d.ts", "target/other/o.ts",
	}},
	{"absolute block", []string{"build/**"}, []string{
		"build/a.ts", "build/sub/b.ts", "build", "buildx/a.ts", "src/a.ts",
	}},
	{"absolute block defeats negation", []string{"build/**", "!build/keep.ts"}, []string{
		"build/keep.ts", "build/other.ts",
	}},
	{"single level X/*", []string{"dist/*"}, []string{
		"dist/a.ts", "dist/sub/b.ts", "dist/sub/deep/c.ts",
	}},

	// Normalize-unstable mid-pattern segments.
	{"dotdot segment", []string{"src/../lib/**/*"}, []string{"lib/a.ts", "src/a.ts", "lib/sub/b.ts"}},
	{"dot segment", []string{"src/./gen/**/*"}, []string{"src/gen/a.ts", "src/other/b.ts"}},
	{"double-slash mid", []string{"a//b/**/*"}, []string{"a/b/c.ts", "a/b.ts", "a/x/c.ts"}},

	// Glob metachars across name and extension positions.
	{"question name", []string{"di?t/**"}, []string{"dist/a.ts", "dixt/a.ts", "distx/a.ts"}},
	{"bracket name", []string{"[bd]ist/**"}, []string{"dist/a.ts", "bist/a.ts", "list/a.ts"}},
	{"brace name cover", []string{"**/{target,dist}/**/*"}, []string{
		"target/a.ts", "dist/sub/b.ts", "other/c.ts",
	}},
	{"brace ext filter not dir-level", []string{"target/**/*.{ts,tsx}"}, []string{
		"target/a.ts", "target/sub/b.tsx", "target/c.js",
	}},
	{"ext filter not dir-level", []string{"logs/**/*.log"}, []string{
		"logs/a.log", "logs/sub/b.log", "logs/a.ts",
	}},
	{"doublestar bare", []string{"**/node_modules/**"}, []string{
		"node_modules/x/a.ts", "pkg/node_modules/y/b.ts", "src/a.ts",
	}},

	// Sequential override ordering (file-level so ! is meaningful at file level).
	{"sequential re-ignore", []string{"**/*", "!src/**/*", "src/test/**/*", "!src/test/keep.ts"}, []string{
		"README.md", "src/index.ts", "src/test/u.ts", "src/test/keep.ts",
	}},
	{"negation before positive", []string{"!build/test.js", "build/**"}, []string{
		"build/test.js", "build/other.js",
	}},

	// Bare names and rooted forms (dir-level; ancestor scan).
	{"bare name dir", []string{"target"}, []string{"target/a.ts", "target/sub/b.ts", "targetx/a.ts"}},
	{"wildcard mid dir", []string{"packages/*/dist/**"}, []string{
		"packages/app/dist/a.ts", "packages/app/src/b.ts", "packages/dist/c.ts",
	}},
}

// TestDifferential_OldEqualsNew is the permanent guard: the NEW structured
// pipeline's per-file lint decision must equal the OLD []string pipeline's, for
// every (patternSet × path) in the corpus, under both cwd modes.
func TestDifferential_OldEqualsNew(t *testing.T) {
	const cwd = "/repo"
	for _, c := range differentialCorpus {
		for _, rel := range c.paths {
			// cwd mode: absolute file resolved relative to cwd.
			absPath := "/repo/" + rel
			oldA := oldDecision(absPath, c.patterns, cwd)
			newA := newDecision(absPath, c.patterns, cwd)
			if oldA != newA {
				t.Errorf("[cwd] %s: decision diverged for %q under %v: old=%v new=%v",
					c.name, rel, c.patterns, oldA, newA)
			}
			// cwd="" simple mode: raw relative path, doublestar.Match only.
			oldS := oldDecision(rel, c.patterns, "")
			newS := newDecision(rel, c.patterns, "")
			if oldS != newS {
				t.Errorf("[simple] %s: decision diverged for %q under %v: old=%v new=%v",
					c.name, rel, c.patterns, oldS, newS)
			}
		}
	}
}

// TestDifferential_DirBlockComponentMatches isolates the directory-block
// component (the part the `./*` regression corrupted): the NEW
// isDirAbsolutelyBlocked over parsed patterns must agree with the OLD
// isDirPathBlocked over raw strings, directly on directory paths (no file
// wrapper). This pins the Kind classification independently of file-level
// matching, so a misroute can't be masked by isFileIgnored happening to agree.
func TestDifferential_DirBlockComponentMatches(t *testing.T) {
	dirCorpus := []struct {
		patterns []string
		dirs     []string
	}{
		{[]string{"./*"}, []string{"src", "src/app", "a/b/c"}},
		{[]string{"target/**"}, []string{"target", "target/sub", "targetx"}},
		{[]string{"target/**/*"}, []string{"target", "target/sub"}},
		{[]string{"dist/*"}, []string{"dist", "dist/sub"}},
		{[]string{"**/build"}, []string{"build", "a/build", "a/b/build", "buildx"}},
		{[]string{"packages/*/dist/**"}, []string{"packages/app/dist", "packages/app/src", "packages/dist"}},
		{[]string{"[bd]ist/**"}, []string{"dist", "bist", "list"}},
		{[]string{"di?t/**"}, []string{"dist", "dixt", "distx"}},
		{[]string{"**/{target,dist}/**/*"}, []string{"target", "dist/sub", "other"}},
		{[]string{"src/../lib/**"}, []string{"lib", "lib/sub", "src"}},
		{[]string{"a//b/**"}, []string{"a/b", "a/b/c", "a/x"}},
		{[]string{"foo"}, []string{"foo", "foo/bar", "foobar"}},
		// Empty-normalizing patterns (`foo/..` → ""). Pre-refactor isDirPathBlocked
		// keeps them absolute, and its empty normalized glob matches the empty
		// LEADING segment of an ABSOLUTE dir path via the ancestor-segment scan —
		// so it blocks "/repo" but not relative "src". Two review agents wrongly
		// declared this unobservable (the empty glob looks harmless until the
		// absolute-path ancestor loop). The empty guard keys on the raw body, not
		// the normalized Glob, precisely to keep this byte-equivalent. Pin both
		// directions (absolute dir blocked, relative dir not).
		{[]string{"foo/.."}, []string{"/repo", "/a/b", "src", "src/app"}},
		{[]string{"./a/.."}, []string{"/repo", "src"}},
		{[]string{"."}, []string{"/repo", "src"}},
	}
	for _, c := range dirCorpus {
		parsed := ParseIgnorePatterns(c.patterns)
		for _, d := range c.dirs {
			oldB := mainX_isDirPathBlocked(d, c.patterns)
			newB := isDirAbsolutelyBlocked(d, parsed)
			if oldB != newB {
				t.Errorf("dir-block diverged for dir %q under %v: old=%v new=%v",
					d, c.patterns, oldB, newB)
			}
		}
	}
}
