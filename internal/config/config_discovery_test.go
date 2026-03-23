package config

import (
	"testing"
)

func TestFindNearestConfig_DirectMatch(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project": {{Rules: Rules{"no-console": "error"}}},
	}

	dir, cfg := FindNearestConfig("/project/src/a.ts", configMap)
	if dir != "/project" {
		t.Errorf("Expected dir /project, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
}

func TestFindNearestConfig_Subdirectory(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project": {{Rules: Rules{"rule-a": "error"}}},
	}

	dir, cfg := FindNearestConfig("/project/src/deep/nested/file.ts", configMap)
	if dir != "/project" {
		t.Errorf("Expected dir /project, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
}

func TestFindNearestConfig_NearestWins(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project":              {{Rules: Rules{"root-rule": "error"}}},
		"/project/packages/foo": {{Rules: Rules{"foo-rule": "error"}}},
	}

	dir, cfg := FindNearestConfig("/project/packages/foo/src/a.ts", configMap)
	if dir != "/project/packages/foo" {
		t.Errorf("Expected nearest config dir /project/packages/foo, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
	if _, ok := cfg[0].Rules["foo-rule"]; !ok {
		t.Error("Expected foo-rule in config")
	}
}

func TestFindNearestConfig_NoMatch(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project": {{Rules: Rules{"rule-a": "error"}}},
	}

	dir, cfg := FindNearestConfig("/other/file.ts", configMap)
	if dir != "" {
		t.Errorf("Expected empty dir, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config for file outside all config dirs")
	}
}

func TestFindNearestConfig_EmptyMap(t *testing.T) {
	configMap := map[string]RslintConfig{}

	dir, cfg := FindNearestConfig("/project/a.ts", configMap)
	if dir != "" {
		t.Errorf("Expected empty dir, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config for empty map")
	}
}

func TestFindNearestConfig_FileInConfigDir(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project": {{Rules: Rules{"rule-a": "error"}}},
	}

	// File directly in config dir (not in a subdirectory)
	dir, cfg := FindNearestConfig("/project/a.ts", configMap)
	if dir != "/project" {
		t.Errorf("Expected dir /project, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
}

func TestFindNearestConfig_MultipleConfigsSameDepth(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project/packages/foo": {{Rules: Rules{"foo-rule": "error"}}},
		"/project/packages/bar": {{Rules: Rules{"bar-rule": "error"}}},
	}

	// File in foo should get foo's config
	dir, cfg := FindNearestConfig("/project/packages/foo/src/a.ts", configMap)
	if dir != "/project/packages/foo" {
		t.Errorf("Expected /project/packages/foo, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
	if _, ok := cfg[0].Rules["foo-rule"]; !ok {
		t.Error("Expected foo-rule")
	}

	// File in bar should get bar's config
	dir, cfg = FindNearestConfig("/project/packages/bar/src/b.ts", configMap)
	if dir != "/project/packages/bar" {
		t.Errorf("Expected /project/packages/bar, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
	if _, ok := cfg[0].Rules["bar-rule"]; !ok {
		t.Error("Expected bar-rule")
	}
}

func TestFindNearestConfig_SimilarPrefixNoFalseMatch(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project/src": {{Rules: Rules{"rule-a": "error"}}},
	}

	// /project/src-other should NOT match /project/src
	dir, cfg := FindNearestConfig("/project/src-other/a.ts", configMap)
	if dir != "" {
		t.Errorf("Expected no match for src-other, got dir %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config for src-other")
	}
}

func TestFindNearestConfig_NestedConfigDirs(t *testing.T) {
	// /project/src and /project/src/components both have configs.
	// File in components should pick the deeper config.
	configMap := map[string]RslintConfig{
		"/project/src":            {{Rules: Rules{"src-rule": "error"}}},
		"/project/src/components": {{Rules: Rules{"components-rule": "error"}}},
	}

	dir, cfg := FindNearestConfig("/project/src/components/Button.tsx", configMap)
	if dir != "/project/src/components" {
		t.Errorf("Expected /project/src/components, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
	if _, ok := cfg[0].Rules["components-rule"]; !ok {
		t.Error("Expected components-rule")
	}

	// File in src (not components) should pick src config
	dir, cfg = FindNearestConfig("/project/src/utils.ts", configMap)
	if dir != "/project/src" {
		t.Errorf("Expected /project/src, got %s", dir)
	}
	if _, ok := cfg[0].Rules["src-rule"]; !ok {
		t.Error("Expected src-rule")
	}
}

func TestFindNearestConfig_RootConfig(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/": {{Rules: Rules{"root-rule": "error"}}},
	}

	dir, cfg := FindNearestConfig("/any/deep/path/file.ts", configMap)
	if dir != "/" {
		t.Errorf("Expected /, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
}

func TestFindNearestConfig_FileAboveAllConfigs(t *testing.T) {
	// Config only in a subdirectory; file in parent should not match
	configMap := map[string]RslintConfig{
		"/project/packages/foo": {{Rules: Rules{"foo-rule": "error"}}},
	}

	dir, cfg := FindNearestConfig("/project/root-file.ts", configMap)
	if dir != "" {
		t.Errorf("Expected no match, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil for file above config dir")
	}
}

func TestFindNearestConfig_TrailingSlashInKey(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project/": {{Rules: Rules{"rule-a": "error"}}},
	}

	dir, cfg := FindNearestConfig("/project/src/a.ts", configMap)
	if dir != "/project/" {
		t.Errorf("Expected /project/, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config with trailing slash key, got nil")
	}
}

func TestFindNearestConfig_EmptyStringKey(t *testing.T) {
	configMap := map[string]RslintConfig{
		"": {{Rules: Rules{"rule-a": "error"}}},
	}

	// Empty string key should not match anything
	dir, cfg := FindNearestConfig("/project/a.ts", configMap)
	if dir != "" || cfg != nil {
		t.Errorf("Expected no match for empty key, got dir=%q cfg=%v", dir, cfg)
	}
}

func TestFindNearestConfig_NilMap(t *testing.T) {
	// nil map should not panic and should return no match
	dir, cfg := FindNearestConfig("/project/a.ts", nil)
	if dir != "" {
		t.Errorf("Expected empty dir for nil map, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config for nil map")
	}
}

func TestFindNearestConfig_FilePathEqualsConfigDir(t *testing.T) {
	// filePath == configDir: StartsWithDirectory returns false,
	// so the file should NOT match (it's not "inside" the directory)
	configMap := map[string]RslintConfig{
		"/project/src": {{Rules: Rules{"rule-a": "error"}}},
	}

	dir, cfg := FindNearestConfig("/project/src", configMap)
	if dir != "" {
		t.Errorf("Expected no match when filePath == configDir, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config when filePath == configDir")
	}
}

func TestFindNearestConfig_SingleConfigFallback(t *testing.T) {
	// Single config that matches — should always work for files under it
	configMap := map[string]RslintConfig{
		"/monorepo": {{Rules: Rules{"rule-a": "error"}}},
	}

	// File deep under config dir
	dir, cfg := FindNearestConfig("/monorepo/packages/foo/src/deep/file.ts", configMap)
	if dir != "/monorepo" {
		t.Errorf("Expected /monorepo, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}

	// File NOT under config dir
	_, cfg = FindNearestConfig("/other-repo/file.ts", configMap)
	if cfg != nil {
		t.Error("Expected nil for file outside config dir")
	}
}
