package config

import (
	"context"
	"encoding/json"
	"sort"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

func TestBasePathScopesSelectorsCascadeAndLocalIgnores(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	app := tspath.CombinePaths(root, "packages/app")
	config := RslintConfig{
		{Files: []string{"**/*.js"}, Rules: Rules{"root": "error"}},
		{BasePath: "packages/app", Files: []string{"src/**/*.js"}, Rules: Rules{"scoped-files": "error"}},
		{BasePath: "packages/app", Rules: Rules{"scoped-cascade": "error"}},
		{BasePath: "packages/app", Files: []string{"**/*.js"}, Ignores: []string{"generated/**"}, Rules: Rules{"scoped-local": "error"}},
	}

	tests := []struct {
		name  string
		path  string
		rules []string
	}{
		{
			name:  "inside scoped files",
			path:  tspath.CombinePaths(app, "src/a.js"),
			rules: []string{"root", "scoped-files", "scoped-cascade", "scoped-local"},
		},
		{
			name:  "local ignore only suppresses its entry",
			path:  tspath.CombinePaths(app, "generated/a.js"),
			rules: []string{"root", "scoped-cascade"},
		},
		{
			name:  "outside scoped base",
			path:  tspath.CombinePaths(root, "packages/other/src/a.js"),
			rules: []string{"root"},
		},
	}

	resolver := NewFileConfigResolver(config, root, false, nil)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			direct := config.GetConfigForFile(test.path, root)
			cached := resolver.ConfigForFile(test.path)
			assertRuleSet(t, direct, test.rules)
			assertRuleSet(t, cached, test.rules)
		})
	}

	absolute := RslintConfig{{
		BasePath: app,
		Files:    []string{"src/**/*.js"},
		Rules:    Rules{"absolute": "error"},
	}}
	assertRuleSet(t, absolute.GetConfigForFile(tspath.CombinePaths(app, "src/a.js"), root), []string{"absolute"})
	assertRuleSet(t, absolute.GetConfigForFile(tspath.CombinePaths(root, "packages/other/src/a.js"), root), []string{})
}

func TestRelativeBasePathScopesEntriesWhenCWDIsEmpty(t *testing.T) {
	config := RslintConfig{
		{Files: []string{"**/*.js"}, Rules: Rules{"root": "error"}},
		{BasePath: "scope", Rules: Rules{"scoped": "error"}},
	}

	assertRuleSet(t, config.GetConfigForFile("scope/a.js", ""), []string{"root", "scoped"})
	assertRuleSet(t, config.GetConfigForFile("other.js", ""), []string{"root"})
}

func TestEntrySelectorsAndLocalIgnoresAtExactBasePath(t *testing.T) {
	const target = "/repo/target.js"
	tests := []struct {
		name        string
		entry       ConfigEntry
		wantApplied bool
	}{
		{name: "no files cascades", entry: ConfigEntry{}, wantApplied: true},
		{name: "empty and group", entry: ConfigEntry{FilePatternGroups: [][]string{{}}}, wantApplied: true},
		{name: "doublestar", entry: ConfigEntry{Files: []string{"**"}}, wantApplied: true},
		{name: "empty string", entry: ConfigEntry{Files: []string{""}}, wantApplied: true},
		{name: "star", entry: ConfigEntry{Files: []string{"*"}}},
		{name: "doublestar slash star", entry: ConfigEntry{Files: []string{"**/*"}}},
		{name: "local doublestar ignore", entry: ConfigEntry{Ignores: []string{"**"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.entry.BasePath = target
			test.entry.Rules = Rules{"scoped": "error"}
			config := RslintConfig{
				{Files: []string{"**/*.js"}, Rules: Rules{"selected": "error"}},
				test.entry,
			}
			merged := config.GetConfigForFile(target, "/repo")
			if merged == nil {
				t.Fatal("outer selector must keep the target selected")
			}
			_, applied := merged.Rules["scoped"]
			if applied != test.wantApplied {
				t.Fatalf("scoped rule applied = %v, want %v", applied, test.wantApplied)
			}
		})
	}
}

func TestBasePathScopesOrderedGlobalIgnores(t *testing.T) {
	root := "/repo"
	keep := "/repo/packages/app/keep.js"
	drop := "/repo/packages/app/drop.js"
	other := "/repo/packages/other/keep.js"

	positiveThenNegative := RslintConfig{
		{BasePath: "packages", Ignores: []string{"**/*.js"}},
		{BasePath: "packages/app", Ignores: []string{"!keep.js"}},
	}
	if positiveThenNegative.IsFileIgnored(keep, root) {
		t.Fatal("nested later negation should re-include app/keep.js")
	}
	if !positiveThenNegative.IsFileIgnored(drop, root) || !positiveThenNegative.IsFileIgnored(other, root) {
		t.Fatal("positive outer scope should still ignore non-reincluded files")
	}

	negativeThenPositive := RslintConfig{
		{BasePath: "packages/app", Ignores: []string{"!keep.js"}},
		{BasePath: "packages", Ignores: []string{"**/*.js"}},
	}
	if !negativeThenPositive.IsFileIgnored(keep, root) {
		t.Fatal("a negation before the positive match must not pre-include a file")
	}

	directoryBoundary := RslintConfig{
		{BasePath: "packages/app", Ignores: []string{"generated/**"}},
		{BasePath: "packages/app", Ignores: []string{"!generated/keep/**"}},
	}
	if !directoryBoundary.IsFileIgnored("/repo/packages/app/generated/keep/a.js", root) {
		t.Fatal("a descendant negation cannot reopen an ignored parent directory")
	}

	ignored, err := NewConfigEvaluator(positiveThenNegative, root, nil, nil).
		IsDirectoryIgnored(context.Background(), "/repo/packages/app")
	if err != nil {
		t.Fatal(err)
	}
	if ignored {
		t.Fatal("a scoped global ignore must never ignore its own basePath root")
	}
}

func TestBasePathGlobalIgnoreShapeRoundTrip(t *testing.T) {
	var config RslintConfig
	if err := json.Unmarshal([]byte(`[{"name":"generated","basePath":"packages/app","ignores":["generated/**"]}]`), &config); err != nil {
		t.Fatal(err)
	}
	if len(config) != 1 || !isGlobalIgnoreEntry(config[0]) {
		t.Fatal("basePath and name must remain metadata on a global-ignore entry")
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	var roundTripped RslintConfig
	if err := json.Unmarshal(encoded, &roundTripped); err != nil {
		t.Fatal(err)
	}
	if len(roundTripped) != 1 || roundTripped[0].BasePath != "packages/app" || !isGlobalIgnoreEntry(roundTripped[0]) {
		t.Fatalf("unexpected round trip: %#v", roundTripped)
	}
}

func TestDiscoverLintFilesHonorsEntryBasePath(t *testing.T) {
	root, paths := setupDiscoveryFixture(t, []string{
		"packages/app/src/keep.JS",
		"packages/app/src/drop.JS",
		"packages/app/generated/other.JS",
		"packages/other/src/keep.JS",
	})
	config := RslintConfig{
		{
			BasePath: "packages/app",
			Files:    []string{"src/**/*.JS"},
			Ignores:  []string{"src/drop.JS"},
			Rules:    Rules{"scoped": "error"},
		},
	}

	got := DiscoverLintFiles(config, root, osvfs.FS(), nil, nil, true)
	want := []string{paths["packages/app/src/keep.JS"]}
	sort.Strings(want)
	assert.DeepEqual(t, got, want)
}

func TestAbsoluteBasePathUsesAuthoredConfigAliasPathSpace(t *testing.T) {
	fsys := &configPathSpaceFS{
		caseSensitive: true,
		realPaths: map[string]string{
			"/alias":                      "/real",
			"/real/packages/app/src/a.JS": "/real/packages/app/src/a.JS",
		},
	}
	matchFile, matchRoot := ResolveConfigPathSpaceWithCanonical(
		"/real/packages/app/src/a.JS",
		"/real/packages/app/src/a.JS",
		"/alias",
		fsys,
	)
	if matchFile != "/alias/packages/app/src/a.JS" || matchRoot != "/alias" {
		t.Fatalf("match space = (%q, %q)", matchFile, matchRoot)
	}
	config := RslintConfig{{
		BasePath: "/alias/packages/app",
		Files:    []string{"src/**/*.JS"},
		Rules:    Rules{"absolute-alias": "error"},
	}}
	assertRuleSet(t, NewFileConfigResolver(config, matchRoot, false, nil).ConfigForFile(matchFile), []string{"absolute-alias"})
}

func assertRuleSet(t *testing.T, merged *MergedConfig, want []string) {
	t.Helper()
	if merged == nil {
		t.Fatalf("config is nil, want rules %v", want)
	}
	got := make([]string, 0, len(merged.Rules))
	for name := range merged.Rules {
		got = append(got, name)
	}
	sort.Strings(got)
	want = append([]string{}, want...)
	sort.Strings(want)
	assert.DeepEqual(t, got, want)
}
