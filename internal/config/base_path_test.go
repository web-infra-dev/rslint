package config

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestResolveRelativeBasePath(t *testing.T) {
	root := filepath.FromSlash("/repo")
	if runtime.GOOS == "windows" {
		root = `C:\repo`
	}

	got, err := resolveRelativeBasePath("packages/foo", root, true)
	if err != nil {
		t.Fatalf("resolveRelativeBasePath: %v", err)
	}
	if got != "packages/foo" {
		t.Fatalf("relative basePath: got %q, want packages/foo", got)
	}

	got, err = resolveRelativeBasePath(".", root, true)
	if err != nil {
		t.Fatalf("resolveRelativeBasePath: %v", err)
	}
	if got != "" {
		t.Fatalf("dot basePath: got %q, want empty", got)
	}

	got, err = resolveRelativeBasePath("packages/foo/../bar", root, true)
	if err != nil {
		t.Fatalf("resolveRelativeBasePath: %v", err)
	}
	if got != "packages/bar" {
		t.Fatalf("normalized basePath: got %q, want packages/bar", got)
	}
}

func TestResolveRelativeBasePath_EscapesGlobMeta(t *testing.T) {
	root := filepath.FromSlash("/repo")
	if runtime.GOOS == "windows" {
		root = `C:\repo`
	}

	got, err := resolveRelativeBasePath("packages/[locale]", root, true)
	if err != nil {
		t.Fatalf("resolveRelativeBasePath: %v", err)
	}
	// [locale] must be treated as a literal directory, not a character class.
	// The leading '[' is escaped as a class ([[]); the trailing ']' is already
	// literal outside a class and needs no escaping.
	if got != "packages/[[]locale]" {
		t.Fatalf("escaped basePath: got %q, want packages/[[]locale]", got)
	}

	got, err = resolveRelativeBasePath("packages/foo", root, true)
	if err != nil {
		t.Fatalf("resolveRelativeBasePath: %v", err)
	}
	if got != "packages/foo" {
		t.Fatalf("plain basePath must be unchanged, got %q", got)
	}
}

func TestResolveRelativeBasePath_RejectsOutsideRoot(t *testing.T) {
	root := filepath.FromSlash("/repo")
	if runtime.GOOS == "windows" {
		root = `C:\repo`
	}

	for _, base := range []string{"../shared", "../../outside"} {
		if _, err := resolveRelativeBasePath(base, root, true); err == nil {
			t.Fatalf("expected %q to be rejected as outside the match root", base)
		}
	}
}

func TestResolveRelativeBasePath_CaseInsensitiveRebase(t *testing.T) {
	root := filepath.FromSlash("/repo")
	caseDiffers := "/Repo/packages/foo"
	if runtime.GOOS == "windows" {
		root = `C:\repo`
		caseDiffers = `C:\Repo\packages\foo`
	}
	// An absolute basePath whose casing differs from the match root must still
	// resolve on a case-insensitive filesystem (useCaseSensitive=false models a
	// case-insensitive vfs), instead of manufacturing ../ patterns.
	got, err := resolveRelativeBasePath(caseDiffers, root, false)
	if err != nil {
		t.Fatalf("case-insensitive resolveRelativeBasePath: %v", err)
	}
	if got == "" || strings.HasPrefix(got, "..") {
		t.Fatalf("expected a match-root-relative path, got %q", got)
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

	got, err := ResolveBasePaths(config, "/repo", nil)
	if err != nil {
		t.Fatalf("ResolveBasePaths: %v", err)
	}
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
	got, err := ResolveBasePaths(config, "/repo", nil)
	if err != nil {
		t.Fatalf("ResolveBasePaths: %v", err)
	}
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
	got, err := ResolveBasePaths(config, "/repo", nil)
	if err != nil {
		t.Fatalf("ResolveBasePaths: %v", err)
	}
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
	got, err := ResolveBasePaths(config, "/repo", nil)
	if err != nil {
		t.Fatalf("ResolveBasePaths: %v", err)
	}
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
	config, err := ResolveBasePaths(RslintConfig{
		{
			BasePath: "packages/foo",
			Files:    []string{"src/**/*.ts"},
			Rules:    Rules{"no-console": "error"},
		},
	}, "/repo", nil)
	if err != nil {
		t.Fatalf("ResolveBasePaths: %v", err)
	}

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

func TestUnmarshalJSON_RejectsEmptyBasePath(t *testing.T) {
	var config RslintConfig
	raw := []byte(`[{"basePath":"","rules":{"no-console":"error"}}]`)
	err := config.UnmarshalJSON(raw)
	if err == nil {
		t.Fatal("expected empty basePath to be rejected")
	}
	if got := err.Error(); got != `config entry at index 0: key "basePath": expected value to be a non-empty string` {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestUnmarshalJSON_RejectsNullBasePath(t *testing.T) {
	var config RslintConfig
	raw := []byte(`[{"basePath":null,"rules":{"no-console":"error"}}]`)
	err := config.UnmarshalJSON(raw)
	if err == nil {
		t.Fatal("expected null basePath to be rejected")
	}
	if got := err.Error(); got != `config entry at index 0: key "basePath": expected value to be a non-empty string` {
		t.Fatalf("unexpected error: %q", got)
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

func TestResolveBasePaths_RejectsOutsideRoot(t *testing.T) {
	config := RslintConfig{
		{
			BasePath: "../shared",
			Files:    []string{"src/**"},
			Rules:    Rules{"no-console": "error"},
		},
	}
	_, err := ResolveBasePaths(config, "/repo", nil)
	if err == nil {
		t.Fatal("expected outside-root basePath to be rejected")
	}
	if !strings.Contains(err.Error(), "config entry at index 0") {
		t.Fatalf("expected entry index in error, got %q", err)
	}
	if !strings.Contains(err.Error(), "outside the config match root") {
		t.Fatalf("expected outside-root message in error, got %q", err)
	}
}

func TestResolveBasePaths_GlobEscapesBasePathPrefix(t *testing.T) {
	config := RslintConfig{
		{
			BasePath: "packages/[locale]",
			Files:    []string{"src/**/*.ts"},
			Rules:    Rules{"no-console": "error"},
		},
	}
	got, err := ResolveBasePaths(config, "/repo", nil)
	if err != nil {
		t.Fatalf("ResolveBasePaths: %v", err)
	}
	want := []string{"packages/[[]locale]/src/**/*.ts"}
	if !reflect.DeepEqual(got[0].Files, want) {
		t.Fatalf("Files = %#v, want %#v", got[0].Files, want)
	}
}
