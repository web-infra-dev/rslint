package config

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestResolveRelativeBasePath(t *testing.T) {
	root := filepath.FromSlash("/repo")
	if runtime.GOOS == "windows" {
		root = `C:\repo`
	}

	got := resolveRelativeBasePath("packages/foo", root)
	if got != "packages/foo" {
		t.Fatalf("relative basePath: got %q, want packages/foo", got)
	}

	got = resolveRelativeBasePath(".", root)
	if got != "" {
		t.Fatalf("dot basePath: got %q, want empty", got)
	}

	got = resolveRelativeBasePath("packages/foo/../bar", root)
	if got != "packages/bar" {
		t.Fatalf("normalized basePath: got %q, want packages/bar", got)
	}
}

func TestRebasePattern(t *testing.T) {
	tests := []struct {
		pattern string
		base    string
		want    string
	}{
		{"src/**/*.ts", "packages/foo", "packages/foo/src/**/*.ts"},
		{"!fixtures/**", "packages/foo", "!packages/foo/fixtures/**"},
		{"./src/**", "packages/foo", "packages/foo/src/**"},
		{"!!src/**", "packages/foo", "packages/foo/src/**"},
		{"/abs/**", "packages/foo", "/abs/**"},
		{"src/**", "", "src/**"},
		{"**/*.ts", "packages/foo", "packages/foo/**/*.ts"},
		{"tsconfig.json", "packages/foo", "packages/foo/tsconfig.json"},
	}
	for _, tt := range tests {
		got := rebasePattern(tt.pattern, tt.base)
		if got != tt.want {
			t.Errorf("rebasePattern(%q, %q) = %q, want %q", tt.pattern, tt.base, got, tt.want)
		}
	}
}

func TestResolveBasePaths_FilesIgnoresProject(t *testing.T) {
	config := RslintConfig{
		{
			BasePath: "packages/foo",
			Files:    []string{"src/**/*.ts", "**/*.tsx"},
			FilePatternGroups: [][]string{
				{"**/*.js", "!**/*.test.js"},
			},
			Ignores: []string{"fixtures/**", "!fixtures/keep.ts"},
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: []string{"./tsconfig.json"},
				},
			},
			Rules: Rules{"no-console": "error"},
		},
	}

	got := ResolveBasePaths(config, "/repo")
	entry := got[0]
	if entry.BasePath != "" {
		t.Fatalf("BasePath should be cleared, got %q", entry.BasePath)
	}
	if !reflect.DeepEqual(entry.Files, []string{"packages/foo/src/**/*.ts", "packages/foo/**/*.tsx"}) {
		t.Fatalf("Files = %#v", entry.Files)
	}
	if !reflect.DeepEqual(entry.FilePatternGroups, [][]string{{"packages/foo/**/*.js", "!packages/foo/**/*.test.js"}}) {
		t.Fatalf("FilePatternGroups = %#v", entry.FilePatternGroups)
	}
	if !reflect.DeepEqual(entry.Ignores, []string{"packages/foo/fixtures/**", "!packages/foo/fixtures/keep.ts"}) {
		t.Fatalf("Ignores = %#v", entry.Ignores)
	}
	if !reflect.DeepEqual([]string(entry.LanguageOptions.ParserOptions.Project), []string{"packages/foo/tsconfig.json"}) {
		t.Fatalf("Project = %#v", entry.LanguageOptions.ParserOptions.Project)
	}
}

func TestResolveBasePaths_GlobalIgnoreWithBasePath(t *testing.T) {
	config := RslintConfig{
		{
			BasePath: "packages/foo",
			Ignores:   []string{"fixtures/**"},
		},
	}
	got := ResolveBasePaths(config, "/repo")
	entry := got[0]
	if entry.BasePath != "" {
		t.Fatalf("BasePath should be cleared")
	}
	if !reflect.DeepEqual(entry.Ignores, []string{"packages/foo/fixtures/**"}) {
		t.Fatalf("Ignores = %#v", entry.Ignores)
	}
	if !isGlobalIgnoreEntry(entry) {
		t.Fatal("expected global ignore entry after desugar")
	}
}

func TestResolveBasePaths_BareBasePathInjectsCatchAll(t *testing.T) {
	config := RslintConfig{
		{
			BasePath: "packages/foo",
			Rules:    Rules{"no-console": "error"},
		},
	}
	got := ResolveBasePaths(config, "/repo")
	entry := got[0]
	if !reflect.DeepEqual(entry.Files, []string{"packages/foo/**"}) {
		t.Fatalf("Files = %#v, want [packages/foo/**]", entry.Files)
	}
	if entry.BasePath != "" {
		t.Fatalf("BasePath should be cleared")
	}
}

func TestResolveBasePaths_NoBasePathUnchanged(t *testing.T) {
	config := RslintConfig{
		{
			Files:  []string{"src/**/*.ts"},
			Ignores: []string{"dist/**"},
			Rules:  Rules{"no-console": "error"},
		},
	}
	got := ResolveBasePaths(config, "/repo")
	if !reflect.DeepEqual(got[0].Files, []string{"src/**/*.ts"}) {
		t.Fatalf("Files mutated: %#v", got[0].Files)
	}
	if !reflect.DeepEqual(got[0].Ignores, []string{"dist/**"}) {
		t.Fatalf("Ignores mutated: %#v", got[0].Ignores)
	}
}

func TestIsGlobalIgnoreEntry_WithBasePathMeta(t *testing.T) {
	// basePath is meta: basePath + ignores is still a global ignore entry.
	entry := ConfigEntry{
		BasePath: "packages/foo",
		Ignores:   []string{"fixtures/**"},
	}
	if !isGlobalIgnoreEntry(entry) {
		t.Fatal("basePath + ignores should still be a global ignore entry")
	}
}

func TestGetConfigForFile_WithDesugaredBasePath(t *testing.T) {
	config := ResolveBasePaths(RslintConfig{
		{
			BasePath: "packages/foo",
			Files:    []string{"src/**/*.ts"},
			Rules:    Rules{"no-console": "error"},
		},
	}, "/repo")

	// Match using relative paths (cwd empty → raw path matching), same style as
	// the rest of the config unit tests.
	matched := config.GetConfigForFile("packages/foo/src/a.ts", "")
	if matched == nil {
		t.Fatal("expected match under basePath")
	}
	if matched.Rules["no-console"] == nil || matched.Rules["no-console"].Level != "error" {
		t.Fatalf("rules = %#v", matched.Rules)
	}

	outside := config.GetConfigForFile("packages/bar/src/a.ts", "")
	if outside != nil {
		t.Fatal("expected no match outside basePath")
	}
}

func TestUnmarshalJSON_BasePathIsMetaForGlobalIgnore(t *testing.T) {
	var config RslintConfig
	raw := []byte(`[{"basePath":"packages/foo","ignores":["fixtures/**"]}]`)
	if err := config.UnmarshalJSON(raw); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if len(config) != 1 {
		t.Fatalf("len = %d", len(config))
	}
	if config[0].BasePath != "packages/foo" {
		t.Fatalf("BasePath = %q", config[0].BasePath)
	}
	// Must remain global: no settings marker injected for basePath meta.
	if config[0].Settings != nil {
		t.Fatalf("Settings should be nil for basePath+ignores, got %#v", config[0].Settings)
	}
	if !isGlobalIgnoreEntry(config[0]) {
		t.Fatal("expected global ignore before desugar")
	}
}
