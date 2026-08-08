package config

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils"
)

// These helpers freeze the pre-reuse path preparation. The lower-level glob
// and ignore evaluators are intentionally shared so this oracle isolates the
// behavior changed by fileMatchPath.
func preReusePositiveFilePatternMatched(filePath string, pattern string, cwd string) bool {
	normalizedPath := filePath
	if cwd != "" {
		normalizedPath = normalizePath(filePath, cwd)
	}
	normalizedPattern := normalizePattern(pattern)
	if utils.MatchGlob(normalizedPattern, normalizedPath) {
		return true
	}
	if normalizedPath != filePath && utils.MatchGlob(normalizedPattern, filePath) {
		return true
	}
	unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")
	return unixPath != normalizedPath && utils.MatchGlob(normalizedPattern, unixPath)
}

func preReuseSingleFilePatternMatched(filePath string, pattern string, cwd string) bool {
	negated := false
	for strings.HasPrefix(pattern, "!") {
		negated = !negated
		pattern = strings.TrimPrefix(pattern, "!")
	}
	matched := preReusePositiveFilePatternMatched(filePath, pattern, cwd)
	if negated {
		return !matched
	}
	return matched
}

func preReuseFileIgnored(filePath string, patterns []IgnorePattern, cwd string) bool {
	if cwd == "" {
		return isFileIgnoredNormalized(filePath, strings.ReplaceAll(filePath, "\\", "/"), patterns)
	}
	normalizedPath := normalizePath(filePath, cwd)
	unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")
	if pathEscapesCwd(unixPath) && hasCaseInsensitivePattern(patterns) {
		normalizedPath = normalizePathWithCaseSensitivity(filePath, cwd, false)
		unixPath = strings.ReplaceAll(normalizedPath, "\\", "/")
	}
	return isFileIgnoredNormalized(normalizedPath, unixPath, patterns)
}

func TestFileMatchPathMatchesPreReuseBehavior(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		cwd      string
	}{
		{name: "posix within cwd", filePath: "/repo/src/app.ts", cwd: "/repo"},
		{name: "posix outside cwd", filePath: "/other/src/app.ts", cwd: "/repo"},
		{name: "relative normalized", filePath: "./src/../src/app.ts", cwd: "/repo"},
		{name: "empty cwd windows separators", filePath: `src\windows\app.ts`, cwd: ""},
		{name: "windows drive case", filePath: "C:/Users/Project/src/App.ts", cwd: "c:/users/project"},
		{name: "windows backslashes", filePath: `C:\Users\Project\src\App.ts`, cwd: `c:\users\project`},
		{name: "unc share case", filePath: "//SERVER/Share/Repo/src/App.ts", cwd: "//server/share/repo"},
		{name: "empty path", filePath: "", cwd: ""},
	}
	selectorPatterns := []string{
		"src/**/*.ts",
		"**/app.ts",
		"C:/Users/Project/src/*.ts",
		`src\windows\*.ts`,
		"!!src/**",
		"!!!**/*.test.ts",
		"",
	}
	ignoreSets := [][]IgnorePattern{
		nil,
		ParseIgnorePatterns([]string{"**/*.ts"}),
		ParseIgnorePatterns([]string{"src/**", "!src/app.ts"}),
	}
	caseInsensitive := ParseIgnorePatterns([]string{"SRC/GENERATED/**", "src/app.ts"})
	for index := range caseInsensitive {
		caseInsensitive[index].CaseInsensitive = true
	}
	ignoreSets = append(ignoreSets, caseInsensitive)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matchPath := newFileMatchPath(test.filePath, test.cwd)
			for _, pattern := range selectorPatterns {
				want := preReuseSingleFilePatternMatched(test.filePath, pattern, test.cwd)
				if got := matchPath.matchesSingle(pattern); got != want {
					t.Fatalf("selector %q: got %v, want %v", pattern, got, want)
				}
			}
			for _, patterns := range ignoreSets {
				want := preReuseFileIgnored(test.filePath, patterns, test.cwd)
				if got := matchPath.isIgnored(patterns); got != want {
					t.Fatalf("ignores %#v: got %v, want %v", patterns, got, want)
				}
			}
		})
	}
}

func TestFileMatchPathPreservesLazyAndNestedSelectorSemantics(t *testing.T) {
	matchPath := newFileMatchPath("/repo/src/app.ts", "/repo")
	if matchPath.isIgnored(nil) || matchPath.matchesAny(nil) {
		t.Fatal("empty matcher unexpectedly matched")
	}
	if !matchPath.matchesConfigEntry(ConfigEntry{FilePatternGroups: [][]string{{}}}) {
		t.Fatal("empty nested selector must remain vacuously true")
	}
	if matchPath.ready {
		t.Fatal("empty matchers must not normalize the file path")
	}

	entry := ConfigEntry{FilePatternGroups: [][]string{{"src/**", "**/*.ts", "!**/*.test.ts"}}}
	if !matchPath.matchesConfigEntry(entry) {
		t.Fatal("nested AND selector unexpectedly rejected matching file")
	}
	if !matchPath.ready {
		t.Fatal("non-empty selector did not prepare the file path")
	}
}

func TestFileMatchPathCaseInsensitiveFallbackDoesNotLeak(t *testing.T) {
	const filePath = "/Repo/src/file.ts"
	const cwd = "/repo"
	matchPath := newFileMatchPath(filePath, cwd)
	baseNormalized, baseUnix := matchPath.normalizedPaths()
	fallbackNormalized := normalizePathWithCaseSensitivity(filePath, cwd, false)
	if fallbackNormalized == baseNormalized {
		t.Fatalf("attack fixture did not exercise fallback: both normalized to %q", baseNormalized)
	}

	caseInsensitive := ParseIgnorePatterns([]string{"src/other.ts"})
	caseInsensitive[0].CaseInsensitive = true
	if matchPath.isIgnored(caseInsensitive) {
		t.Fatal("non-matching case-insensitive pattern unexpectedly ignored file")
	}
	if matchPath.normalized != baseNormalized || matchPath.unix != baseUnix {
		t.Fatalf("fallback polluted cached path: got (%q, %q), want (%q, %q)",
			matchPath.normalized, matchPath.unix, baseNormalized, baseUnix)
	}

	caseSensitive := []IgnorePattern{{Glob: baseNormalized}}
	if got, want := matchPath.isIgnored(caseSensitive), preReuseFileIgnored(filePath, caseSensitive, cwd); got != want || !got {
		t.Fatalf("case-sensitive matcher after fallback: got %v, want %v", got, want)
	}
}

func FuzzFileMatchPathMatchesPreReuse(f *testing.F) {
	f.Add("/repo/src/app.ts", "/repo", "!!src/**/*.ts", "src/**", false)
	f.Add(`C:\Users\Project\src\App.ts`, `c:\users\project`, "src/**/*.ts", "SRC/**", true)
	f.Add("//SERVER/Share/Repo/src/App.ts", "//server/share/repo", "**/*.ts", "src/**", true)
	f.Add(`src\windows\app.ts`, "", "src/**/*.ts", "src/**", false)

	f.Fuzz(func(t *testing.T, filePath string, cwd string, selector string, ignore string, caseInsensitive bool) {
		if len(filePath) > 256 || len(cwd) > 256 || len(selector) > 256 || len(ignore) > 256 {
			t.Skip()
		}
		matchPath := newFileMatchPath(filePath, cwd)
		if got, want := matchPath.matchesSingle(selector), preReuseSingleFilePatternMatched(filePath, selector, cwd); got != want {
			t.Fatalf("selector mismatch: file=%q cwd=%q selector=%q got=%v want=%v",
				filePath, cwd, selector, got, want)
		}
		patterns := ParseIgnorePatterns([]string{ignore})
		patterns[0].CaseInsensitive = caseInsensitive
		if got, want := matchPath.isIgnored(patterns), preReuseFileIgnored(filePath, patterns, cwd); got != want {
			t.Fatalf("ignore mismatch: file=%q cwd=%q ignore=%q insensitive=%v got=%v want=%v",
				filePath, cwd, ignore, caseInsensitive, got, want)
		}
	})
}
