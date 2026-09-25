package config

import (
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
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

func TestFileIgnoreMatcherMatchesSequentialEvaluation(t *testing.T) {
	random := rand.New(rand.NewSource(0x51a7))
	literals := []string{
		"src/app.ts", "src/other.ts", "src/app.test.ts", "src/nested/app.ts",
		"", "src/", "src/中文.ts", "src/UPPER.ts", "src/name!.ts", "src/../root.ts",
	}
	otherPatterns := []string{
		"!src/app.ts", "!src/**/*.ts", "src/**", "src/*.{js,ts}",
		"src/[ab]*.ts", "src/?.ts", "src/[", "src/{", "src/\ufffd.ts", "src/\xff.ts",
	}
	targets := []struct{ path, cwd string }{
		{"/repo/src/app.ts", "/repo"},
		{"/repo/src/other.ts", "/repo"},
		{"/repo/src/app.test.ts", "/repo"},
		{"/repo/src/nested/app.ts", "/repo"},
		{"/repo/src/中文.ts", "/repo"},
		{"/repo/src/UPPER.ts", "/repo"},
		{"/repo/src/upper.ts", "/repo"},
		{"/repo/src/\xff.ts", "/repo"},
		{"/repo/root.ts", "/repo"},
		{"/other/src/app.ts", "/repo"},
		{"", ""},
		{`src\app.ts`, ""},
		{`C:\Repo\src\app.ts`, "C:/Repo"},
		{"//server/share/src/app.ts", "//server/share"},
	}
	for sample := range 300 {
		var raw []string
		for range 3 {
			for range 8 + random.Intn(8) {
				raw = append(raw, literals[random.Intn(len(literals))])
			}
			for range 1 + random.Intn(4) {
				raw = append(raw, otherPatterns[random.Intn(len(otherPatterns))])
			}
		}
		patterns := ParseIgnorePatterns(raw)
		matcher := newFileIgnoreMatcher(patterns)
		if len(matcher.steps) == 0 {
			t.Fatal("the differential corpus must exercise indexed literals")
		}
		for _, target := range targets {
			legacy := newFileMatchPath(target.path, target.cwd)
			prepared := newFileMatchPath(target.path, target.cwd)
			if got, want := matcher.isIgnored(&prepared), legacy.isIgnored(patterns); got != want {
				t.Fatalf("sample %d target %+v: got %v, want %v; patterns %q", sample, target, got, want, raw)
			}
		}
	}

	for _, unsupported := range []IgnorePattern{
		{Glob: "src/app.ts", GitPattern: true},
		{Glob: "src/app.ts", CaseInsensitive: true},
		{Glob: "src/app.ts", MatchDirectory: "/repo"},
		{Glob: "src/app.ts", PhysicalMatchDirectory: "/repo"},
		{Glob: "src/app.ts", LexicalMatchDirectory: "/repo"},
	} {
		patterns := append(ParseIgnorePatterns(literals), unsupported)
		if matcher := newFileIgnoreMatcher(patterns); matcher.steps != nil {
			t.Fatalf("special path semantics must use the original matcher: %+v", unsupported)
		}
	}
}

func literalFileIgnoresForTest(count int) []string {
	patterns := make([]string, count)
	for index := range patterns {
		patterns[index] = fmt.Sprintf("src/generated/file%d.ts", index)
	}
	return patterns
}

func TestFileIgnoreMatcherOrderedOverrides(t *testing.T) {
	literals := literalFileIgnoresForTest(8)
	target := literals[0]
	for _, test := range []struct {
		name     string
		patterns []string
		path     string
		cwd      string
		want     bool
	}{
		{"literal hit", literals, target, "", true},
		{"literal miss", literals, "src/other.ts", "", false},
		{"literal is not a directory block", literals, target + "/child.ts", "", false},
		{"trailing separator", literals, target + "/", "", false},
		{"case remains significant", literals, strings.ToUpper(target), "", false},
		{"unix fallback", literals, strings.ReplaceAll(target, "/", `\`), "", true},
		{"relative normalization", literals, "/repo/src/../" + target, "/repo", true},
		{"later negation", slices.Concat(literals, []string{"!" + target}), target, "", false},
		{"earlier negation", slices.Concat([]string{"!" + target}, literals), target, "", true},
		{"later literal segment", slices.Concat(literals, []string{"!**/*.ts"}, literals), target, "", true},
		{"earlier glob survives literal miss", slices.Concat([]string{"src/**"}, literals), "src/other.ts", "", true},
		{"later negated glob", slices.Concat(literals, []string{"!src/**"}), target, "", false},
		{"later positive glob", slices.Concat(literals, []string{"!" + target, "src/**"}), target, "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			matcher := newFileIgnoreMatcher(ParseIgnorePatterns(test.patterns))
			path := newFileMatchPath(test.path, test.cwd)
			if got := matcher.isIgnored(&path); got != test.want {
				t.Fatalf("isIgnored(%q) = %v, want %v", test.path, got, test.want)
			}
		})
	}
}

func TestFileIgnoreMatcherFallback(t *testing.T) {
	for _, count := range []int{0, 1, 7, 8} {
		matcher := newFileIgnoreMatcher(ParseIgnorePatterns(literalFileIgnoresForTest(count)))
		if indexed := len(matcher.steps) > 0; indexed != (count == 8) {
			t.Fatalf("%d literals: indexed = %v", count, indexed)
		}
		path := newFileMatchPath("/repo/src/generated/file0.ts", "/repo")
		if got := matcher.isIgnored(&path); got != (count > 0) {
			t.Fatalf("%d literals: ignored = %v", count, got)
		}
		if count == 0 && path.ready {
			t.Fatal("empty ignores must not normalize the file path")
		}
	}
	// These patterns cannot use byte equality, even when every pattern in the
	// list is the same. The glob engine decodes invalid bytes as RuneError.
	for _, test := range []struct{ glob, path string }{
		{`src/\*.ts`, "src/*.ts"},
		{"src/\ufffd.ts", "src/\xff.ts"},
		{"src/\xfe.ts", "src/\xff.ts"},
	} {
		patterns := make([]IgnorePattern, 8)
		for index := range patterns {
			patterns[index] = IgnorePattern{Glob: test.glob}
		}
		matcher := newFileIgnoreMatcher(patterns)
		path := newFileMatchPath(test.path, "")
		if len(matcher.steps) != 0 || !matcher.isIgnored(&path) {
			t.Fatalf("pattern %q must keep glob semantics for %q", test.glob, test.path)
		}
	}
}

func FuzzFileIgnoreMatcherMatchesSequentialEvaluation(f *testing.F) {
	f.Add("!src/generated/file0.ts", "src/generated/file0.ts", "")
	f.Add("src/**\n!src/generated/file0.ts", "/repo/src/generated/file0.ts", "/repo")
	f.Add("src/\ufffd.ts", "src/\xff.ts", "")
	f.Add("!**/*.ts\nsrc/*.ts", `src\generated\file0.ts`, "")
	f.Add("src/[a-z].ts\nsrc/{a,b}.ts", "C:/Repo/src/a.ts", "C:/Repo")
	f.Fuzz(func(t *testing.T, encodedPatterns, filePath, cwd string) {
		if len(encodedPatterns) > 512 || len(filePath) > 256 || len(cwd) > 256 {
			t.Skip()
		}
		// Surround arbitrary ordered patterns with literal runs so mutations
		// exercise transitions into and out of indexed segments as well.
		raw := slices.Concat(literalFileIgnoresForTest(8), strings.Split(encodedPatterns, "\n"))
		for index := range 8 {
			raw = append(raw, fmt.Sprintf("other/file%d.ts", index))
		}
		patterns := ParseIgnorePatterns(raw)
		matcher := newFileIgnoreMatcher(patterns)
		path := newFileMatchPath(filePath, cwd)
		legacy := newFileMatchPath(filePath, cwd)
		if got, want := matcher.isIgnored(&path), legacy.isIgnored(patterns); got != want {
			t.Fatalf("file=%q cwd=%q patterns=%q: got %v, want %v", filePath, cwd, raw, got, want)
		}
	})
}

func BenchmarkFileIgnoreMatcher(b *testing.B) {
	for _, count := range []int{0, 4, 8, 512} {
		matcher := newFileIgnoreMatcher(ParseIgnorePatterns(literalFileIgnoresForTest(count)))
		for _, target := range []struct{ name, path string }{
			{"miss_shared_prefix", "/repo/src/generated/not-ignored.ts"},
			{"miss_other_prefix", "/repo/packages/example/main.ts"},
			{"hit", "/repo/src/generated/file0.ts"},
		} {
			b.Run(fmt.Sprintf("patterns=%d/%s", count, target.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					path := newFileMatchPath(target.path, "/repo")
					matcher.isIgnored(&path)
				}
			})
		}
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

func TestConfigShapeResolutionMatchesLegacyAlgorithm(t *testing.T) {
	random := rand.New(rand.NewSource(0x5a17_2026))
	paths := []string{
		"/repo/src/app.ts",
		"/repo/src/app.test.ts",
		"/repo/src/special.ts",
		"/repo/src/generated/app.ts",
		"/repo/test/app.js",
		"/repo/component.vue",
		"/repo/component.unsupported",
	}

	for sample := range 100 {
		entryCount := 1 + random.Intn(80)
		config := make(RslintConfig, entryCount)
		for index := range config {
			entry := ConfigEntry{Name: fmt.Sprintf("sample-%d-entry-%d", sample, index)}
			switch random.Intn(6) {
			case 1:
				entry.Files = []string{"src/**/*.ts"}
			case 2:
				entry.Files = []string{"**/*.vue"}
			case 3:
				entry.Files = []string{"src/special.ts"}
			case 4:
				entry.FilePatternGroups = [][]string{{"src/**", "**/*.ts", "!**/*.test.ts"}}
			case 5:
				entry.FilePatternGroups = [][]string{{}}
			}
			if random.Intn(4) == 0 {
				ignores := slices.Concat(literalFileIgnoresForTest(16), []string{
					"src/app.ts", "**/*.test.ts", "src/generated/**", "!src/generated/keep.ts",
				})
				entry.Ignores = ignores[:len(ignores)-random.Intn(4)]
			}
			if random.Intn(3) != 0 {
				severities := []any{"off", "warn", "error", []any{"error", map[string]any{"allow": []any{"warn"}}}}
				entry.Rules = Rules{"no-console": severities[random.Intn(len(severities))]}
			}
			if random.Intn(4) == 0 {
				entry.Settings = Settings{
					"shared": map[string]any{
						"sample": sample,
						"entry":  index,
					},
				}
			}
			if random.Intn(6) == 0 {
				entry.LanguageOptions = &LanguageOptions{Raw: map[string]any{
					"globals": map[string]any{"testGlobal": "readonly"},
				}}
			}
			if random.Intn(8) == 0 {
				entry.Plugins = []string{"@typescript-eslint"}
			}
			config[index] = entry
		}

		resolver := NewFileConfigResolver(config, "/repo", baseRuleCatalog())
		for _, filePath := range paths {
			want := legacyMergedConfigForTest(config, filePath, "/repo", extractConfigIgnores(config))
			if got := config.GetConfigForFile(filePath, "/repo"); !reflect.DeepEqual(got, want) {
				t.Fatalf("sample %d path %s direct mismatch:\nnew:    %#v\nlegacy: %#v", sample, filePath, got, want)
			}
			if got := resolver.ConfigForFile(filePath); !reflect.DeepEqual(got, want) {
				t.Fatalf("sample %d path %s resolver mismatch:\nnew:    %#v\nlegacy: %#v", sample, filePath, got, want)
			}
		}
	}
}

func TestFileConfigResolverMatchesDirectResolutionAcrossShapes(t *testing.T) {
	config := RslintConfig{
		{Ignores: []string{"generated/**"}},
		{
			Rules: Rules{
				"no-console": []any{"warn", map[string]any{"allow": []any{"warn"}}},
			},
			Settings: Settings{"shared": map[string]any{"base": true}},
		},
		{
			Files: []string{"src/**/*.ts"},
			Rules: Rules{
				"no-debugger": "error",
			},
			Settings: Settings{"shared": map[string]any{"typed": true}},
			LanguageOptions: &LanguageOptions{Raw: map[string]any{
				"globals": map[string]any{"typedGlobal": "readonly"},
			}},
		},
		{
			Files: []string{"src/special.ts"},
			Rules: Rules{
				"no-console": "error",
			},
		},
		{
			Files:   []string{"src/**/*.ts"},
			Ignores: append(literalFileIgnoresForTest(8), "**/*.test.ts"),
			Rules: Rules{
				"eqeqeq": "warn",
			},
		},
		{
			FilePatternGroups: [][]string{{"src/**", "**/*.ts"}},
			Rules: Rules{
				"no-alert": "warn",
			},
		},
		{
			Files: []string{"**/*.vue"},
			Rules: Rules{
				"no-debugger": "warn",
			},
		},
	}

	resolver := NewFileConfigResolver(config, "/repo", baseRuleCatalog())
	paths := []string{
		"/repo/src/a.ts",
		"/repo/src/b.ts",
		"/repo/src/special.ts",
		"/repo/src/a.test.ts",
		"/repo/src/a.js",
		"/repo/src/generated/file0.ts",
		"/repo/src/generated/file0.ts/child.ts",
		"/repo/components/view.vue",
		"/repo/generated/skip.ts",
		"/repo/outside.unsupported",
	}
	for _, filePath := range paths {
		directMerged := config.GetConfigForFile(filePath, "/repo")
		directRules := ConfiguredRules(baseRuleCatalog(), directMerged)
		resolvedRules, resolvedMerged := resolver.EnabledRulesForFile(filePath)
		if !reflect.DeepEqual(resolvedMerged, directMerged) {
			t.Fatalf("%s merged config mismatch:\nresolver: %#v\ndirect:   %#v", filePath, resolvedMerged, directMerged)
		}
		if got, want := configuredRuleViews(resolvedRules), configuredRuleViews(directRules); !reflect.DeepEqual(got, want) {
			t.Fatalf("%s enabled rules mismatch:\nresolver: %#v\ndirect:   %#v", filePath, got, want)
		}
	}

	first := resolver.planForFile("/repo/src/a.ts")
	if second := resolver.planForFile("/repo/src/b.ts"); first == nil || second != first {
		t.Fatal("files with the same exact config shape did not share one effective plan")
	}
	if special := resolver.planForFile("/repo/src/special.ts"); special == nil || special == first {
		t.Fatal("files with different config shapes shared an effective plan")
	}
}

func TestFileConfigResolverDistinguishesMatchedZeroRulesAndMisses(t *testing.T) {
	config := RslintConfig{{Settings: Settings{}}}
	resolver := NewFileConfigResolver(config, "/repo", baseRuleCatalog())

	rules, merged := resolver.EnabledRulesForFile("/repo/src/file.ts")
	if merged == nil || len(rules) != 0 {
		t.Fatalf("default-selected zero-rule file resolved to rules=%v merged=%#v", rules, merged)
	}
	if rules, merged = resolver.EnabledRulesForFile("/repo/src/file.vue"); rules != nil || merged != nil {
		t.Fatalf("unsupported unselected file resolved to rules=%v merged=%#v", rules, merged)
	}
}

func TestFileConfigResolverShapeKeyPreservesEntriesBeyond64(t *testing.T) {
	config := make(RslintConfig, 72)
	for index := range config {
		config[index] = ConfigEntry{Name: fmt.Sprintf("base-%d", index)}
	}
	config[63] = ConfigEntry{
		Name:  "low-special",
		Files: []string{"src/low.ts"},
		Rules: Rules{"no-alert": "warn"},
	}
	config[64] = ConfigEntry{
		Name:  "tail-special",
		Files: []string{"src/tail.ts"},
		Rules: Rules{"no-debugger": "error"},
	}
	config[71] = ConfigEntry{
		Name:  "tail-vue",
		Files: []string{"**/*.vue"},
		Rules: Rules{"no-console": "warn"},
	}

	resolver := NewFileConfigResolver(config, "/repo", baseRuleCatalog())
	paths := []string{
		"/repo/src/ordinary.ts",
		"/repo/src/low.ts",
		"/repo/src/tail.ts",
		"/repo/component.vue",
		"/repo/component.unsupported",
	}
	plans := make(map[string]*effectiveConfigPlan, len(paths))
	for _, filePath := range paths {
		resolvedRules, resolvedMerged := resolver.EnabledRulesForFile(filePath)
		directMerged := config.GetConfigForFile(filePath, "/repo")
		directRules := ConfiguredRules(baseRuleCatalog(), directMerged)
		if !reflect.DeepEqual(resolvedMerged, directMerged) ||
			!reflect.DeepEqual(configuredRuleViews(resolvedRules), configuredRuleViews(directRules)) {
			t.Fatalf("resolution mismatch for %s", filePath)
		}
		plans[filePath] = resolver.planForFile(filePath)
	}
	if plans["/repo/src/ordinary.ts"] == plans["/repo/src/low.ts"] ||
		plans["/repo/src/ordinary.ts"] == plans["/repo/src/tail.ts"] ||
		plans["/repo/src/ordinary.ts"] == plans["/repo/component.vue"] {
		t.Fatal("entry bits at or beyond the 64-entry boundary collided")
	}
	if plans["/repo/component.unsupported"] != nil {
		t.Fatal("unsupported file with only unscoped matches should remain unselected")
	}
}

func TestFileConfigResolverPreservesWindowsPathMatching(t *testing.T) {
	config := RslintConfig{{
		Files:   []string{"src/**/*.ts"},
		Ignores: []string{"src/generated/**"},
		Rules:   Rules{"no-console": "error"},
	}}
	tests := []struct {
		filePath string
		cwd      string
	}{
		{"C:/Users/project/src/index.ts", "C:/Users/project"},
		{"C:/Users/project/src/index.ts", `C:\Users\project`},
		{"C:/Users/project/src/generated/index.ts", `C:\Users\project`},
		{"C:/repo/packages/foo/src/index.ts", "C:/repo"},
	}
	for _, test := range tests {
		resolver := NewFileConfigResolver(config, test.cwd, baseRuleCatalog())
		got := resolver.ConfigForFile(test.filePath)
		want := config.GetConfigForFile(test.filePath, test.cwd)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("filePath=%q cwd=%q: resolver=%#v direct=%#v", test.filePath, test.cwd, got, want)
		}
	}
}

func TestFileConfigResolverConcurrentShapePublication(t *testing.T) {
	config := RslintConfig{{
		Ignores: literalFileIgnoresForTest(16),
		Files:   []string{"src/**/*.ts"},
		Rules: Rules{
			"no-console":  "warn",
			"no-debugger": "error",
		},
	}}
	resolver := NewFileConfigResolver(config, "/repo", baseRuleCatalog())
	plans := make(chan *effectiveConfigPlan, 128)
	var waitGroup sync.WaitGroup
	for index := range 128 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			plans <- resolver.planForFile(fmt.Sprintf("/repo/src/file-%d.ts", index))
		}()
	}
	waitGroup.Wait()
	close(plans)

	var first *effectiveConfigPlan
	for plan := range plans {
		if plan == nil {
			t.Fatal("matching file resolved to a nil plan")
		}
		if first == nil {
			first = plan
			continue
		}
		if plan != first {
			t.Fatal("concurrent files with one shape published multiple effective plans")
		}
	}
}

func TestFileConfigResolverConcurrentDirectoryBlockCache(t *testing.T) {
	config := RslintConfig{
		{Ignores: []string{"dist/**"}},
		{Files: []string{"**/*.ts"}, Settings: Settings{}},
	}
	resolver := NewFileConfigResolver(config, "/repo", baseRuleCatalog())

	var waitGroup sync.WaitGroup
	results := make(chan *MergedConfig, 128)
	for index := range 128 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			results <- resolver.ConfigForFile(fmt.Sprintf("/repo/src/pkg/file-%d.ts", index))
		}()
	}
	waitGroup.Wait()
	close(results)
	for merged := range results {
		if merged == nil {
			t.Fatal("non-ignored file resolved to nil config")
		}
	}

	if merged := resolver.ConfigForFile("/repo/dist/pkg/blocked.ts"); merged != nil {
		t.Fatal("directory-blocked file resolved to a config")
	}
}

type configuredRuleView struct {
	name               string
	settings           map[string]interface{}
	globals            map[string]utils.GlobalAccess
	severity           int
	requiresTypeInfo   bool
	isEslintPluginRule bool
	options            []any
	hasRun             bool
}

func configuredRuleViews(rules []rule.ConfiguredRule) []configuredRuleView {
	views := make([]configuredRuleView, len(rules))
	for index, configuredRule := range rules {
		views[index] = configuredRuleView{
			name:               configuredRule.Name,
			settings:           configuredRule.Environment.Settings,
			globals:            configuredRule.Environment.Globals,
			severity:           int(configuredRule.Severity),
			requiresTypeInfo:   configuredRule.RequiresTypeInfo,
			isEslintPluginRule: configuredRule.IsEslintPluginRule,
			options:            configuredRule.Options,
			hasRun:             configuredRule.Run != nil,
		}
	}
	return views
}

// legacyMergedConfigForTest is the pre-interning selection and merge algorithm.
// Keeping this independent test-only oracle attacks semantic drift introduced
// when the production path combines selection and entry matching into one pass.
func legacyMergedConfigForTest(
	config RslintConfig,
	filePath string,
	cwd string,
	globalIgnorePatterns []IgnorePattern,
) *MergedConfig {
	merged := &MergedConfig{
		Rules:   make(map[string]*RuleConfig),
		Plugins: make(map[string]struct{}),
	}
	if len(globalIgnorePatterns) > 0 &&
		(isDirBlockedByIgnores(filePath, globalIgnorePatterns, cwd) ||
			isFileIgnored(filePath, globalIgnorePatterns, cwd)) {
		return nil
	}
	if !isFileSelectedByConfig(config, filePath, cwd) {
		return nil
	}

	entryMatched := false
	for _, entry := range config {
		if isGlobalIgnoreEntry(entry) {
			continue
		}
		if hasFileSelectors(entry) && !isFileMatchedByConfigEntry(filePath, entry, cwd) {
			continue
		}
		if isFileIgnored(filePath, ParseIgnorePatterns(entry.Ignores), cwd) {
			continue
		}
		entryMatched = true

		for ruleName, ruleValue := range entry.Rules {
			next, hasOptions, err := parseRuleConfigValue(ruleValue)
			if err != nil {
				continue
			}
			if previous := merged.Rules[ruleName]; !hasOptions && previous != nil {
				next.Options = append([]interface{}(nil), previous.Options...)
			}
			merged.Rules[ruleName] = next
		}
		for _, plugin := range entry.Plugins {
			merged.Plugins[NormalizePluginName(plugin)] = struct{}{}
		}
		if entry.Settings != nil {
			merged.Settings = Settings(deepMergeConfigObjects(
				map[string]any(merged.Settings),
				map[string]any(entry.Settings),
			))
		}
		merged.LanguageOptions = mergeLanguageOptions(merged.LanguageOptions, entry.LanguageOptions)
	}
	if !entryMatched {
		return nil
	}
	return merged
}
