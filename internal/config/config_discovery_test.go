package config

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
)

type configPathSpaceFS struct {
	vfs.FS
	realPaths     map[string]string
	caseSensitive bool
}

func (fs *configPathSpaceFS) UseCaseSensitiveFileNames() bool { return fs.caseSensitive }
func (fs *configPathSpaceFS) Realpath(filePath string) string {
	if realPath := fs.realPaths[filePath]; realPath != "" {
		return realPath
	}
	return filePath
}

func resolveConfigOwner(filePath string, configMap map[string]RslintConfig) (string, RslintConfig) {
	return NewConfigOwnerResolver(configMap, nil).Resolve(filePath)
}

func TestResolveConfigPathSpace(t *testing.T) {
	t.Run("symlink aliases use the physical config root", func(t *testing.T) {
		fs := &configPathSpaceFS{
			caseSensitive: true,
			realPaths: map[string]string{
				"/alias":          "/real",
				"/alias/src/a.ts": "/real/src/a.ts",
				"/real/src/a.ts":  "/real/src/a.ts",
			},
		}
		for _, filePath := range []string{"/alias/src/a.ts", "/real/src/a.ts"} {
			matchFile, matchDir := ResolveConfigPathSpace(filePath, "/alias", fs)
			if matchFile != "/real/src/a.ts" || matchDir != "/real" {
				t.Fatalf("ResolveConfigPathSpace(%q) = (%q, %q)", filePath, matchFile, matchDir)
			}
		}
	})

	t.Run("realpath-normalized casing retains one matching space", func(t *testing.T) {
		fs := &configPathSpaceFS{
			caseSensitive: false,
			realPaths: map[string]string{
				"C:/Repo":              "C:/Repo",
				"c:/repo/src/index.ts": "C:/Repo/src/index.ts",
			},
		}
		matchFile, matchDir := ResolveConfigPathSpace("c:/repo/src/index.ts", "C:/Repo", fs)
		if matchFile != "C:/Repo/src/index.ts" || matchDir != "C:/Repo" {
			t.Fatalf("case-insensitive path space = (%q, %q)", matchFile, matchDir)
		}
	})

	t.Run("distinct physical casing is not collapsed by a global case flag", func(t *testing.T) {
		fs := &configPathSpaceFS{
			caseSensitive: false,
			realPaths: map[string]string{
				"C:/Repo":              "C:/Repo",
				"c:/repo/src/index.ts": "c:/repo/src/index.ts",
			},
		}
		matchFile, matchDir := ResolveConfigPathSpace("c:/repo/src/index.ts", "C:/Repo", fs)
		if matchFile != "c:/repo/src/index.ts" || matchDir != "C:/Repo" {
			t.Fatalf("distinct physical roots were collapsed: (%q, %q)", matchFile, matchDir)
		}
	})

	t.Run("native case alias retains relative path when the file symlink escapes", func(t *testing.T) {
		fs := &configPathSpaceFS{
			caseSensitive: false,
			realPaths: map[string]string{
				"/repo/Project":         "/repo/Project",
				"/repo/project":         "/repo/Project",
				"/repo/project/link.ts": "/repo/shared.ts",
			},
		}
		matchFile, matchDir := ResolveConfigPathSpace("/repo/project/link.ts", "/repo/Project", fs)
		if matchFile != "/repo/Project/link.ts" || matchDir != "/repo/Project" {
			t.Fatalf("native alias path space = (%q, %q)", matchFile, matchDir)
		}
	})
}

func TestConfigOwnerResolverUsesVerifiedNativeCaseAliasBeforeFileRealpath(t *testing.T) {
	fs := &configPathSpaceFS{
		caseSensitive: false,
		realPaths: map[string]string{
			"/repo/Project":         "/repo/Project",
			"/repo/project":         "/repo/Project",
			"/repo/project/link.ts": "/repo/shared.ts",
		},
	}
	configMap := map[string]RslintConfig{
		"/repo/Project": {{Rules: Rules{"rule": "error"}}},
	}
	dir, cfg := NewConfigOwnerResolver(configMap, fs).Resolve("/repo/project/link.ts")
	if dir != "/repo/Project" || cfg == nil {
		t.Fatalf("native case alias did not retain config owner: dir=%q cfg=%v", dir, cfg)
	}
}

func TestConfigOwnerResolverPrefersLexicalHierarchy(t *testing.T) {
	fs := &configPathSpaceFS{
		caseSensitive: true,
		realPaths: map[string]string{
			"/alias/pkg": "/real/pkg",
		},
	}
	configMap := map[string]RslintConfig{
		"/real":      {{Rules: Rules{"root": "error"}}},
		"/alias/pkg": {{Rules: Rules{"package": "error"}}},
	}
	resolver := NewConfigOwnerResolver(configMap, fs)

	dir, cfg := resolver.Resolve("/real/pkg/src/a.ts")
	if dir != "/real" || cfg == nil {
		t.Fatalf("physical config replaced lexical owner: dir=%q cfg=%v", dir, cfg)
	}

	index := newConfigDirectoryIndex(configMap, fs)
	children := index.childConfigDirs("/real")
	if len(children) != 0 {
		t.Fatalf("physical hierarchy created lexical child boundaries: %v", children)
	}
}

func TestConfigOwnerResolverChildConfigDirs(t *testing.T) {
	resolver := NewConfigOwnerResolver(map[string]RslintConfig{
		"/repo":                  nil,
		"/repo/packages/app":     nil,
		"/repo/packages/lib":     nil,
		"/repo/packages/app/e2e": nil,
	}, nil)

	children := resolver.ChildConfigDirs("/repo")
	if len(children) != 2 || children[0] != "/repo/packages/app" || children[1] != "/repo/packages/lib" {
		t.Fatalf("root child boundaries = %v", children)
	}
	if deep := resolver.ChildConfigDirs("/repo/packages/app"); len(deep) != 1 || deep[0] != "/repo/packages/app/e2e" {
		t.Fatalf("nested child boundaries = %v", deep)
	}

	children[0] = "/mutated"
	if fresh := resolver.ChildConfigDirs("/repo"); fresh[0] != "/repo/packages/app" {
		t.Fatalf("ChildConfigDirs exposed resolver state: %v", fresh)
	}
}

func TestConfigOwnerResolverUsesPhysicalFallbackWithoutLexicalOwner(t *testing.T) {
	fs := &configPathSpaceFS{
		caseSensitive: true,
		realPaths: map[string]string{
			"/alias/pkg":         "/real/pkg",
			"/outside/link/a.ts": "/real/pkg/src/a.ts",
		},
	}
	configMap := map[string]RslintConfig{
		"/alias/pkg": {{Rules: Rules{"package": "error"}}},
	}

	dir, cfg := NewConfigOwnerResolver(configMap, fs).Resolve("/outside/link/a.ts")
	if dir != "/alias/pkg" || cfg == nil {
		t.Fatalf("physical fallback did not resolve aliased config: dir=%q cfg=%v", dir, cfg)
	}
}

func TestConfigOwnerResolverKeepsLexicalOwnerAcrossUnrelatedRealpathTree(t *testing.T) {
	fs := &configPathSpaceFS{
		caseSensitive: true,
		realPaths: map[string]string{
			"/lex/link/a.ts": "/physical/a.ts",
		},
	}
	configMap := map[string]RslintConfig{
		"/lex":      {{Rules: Rules{"lexical": "error"}}},
		"/physical": {{Rules: Rules{"physical": "error"}}},
	}

	dir, cfg := NewConfigOwnerResolver(configMap, fs).Resolve("/lex/link/a.ts")
	if dir != "/lex" || cfg == nil {
		t.Fatalf("unrelated physical tree replaced lexical owner: dir=%q cfg=%v", dir, cfg)
	}
}

func TestConfigOwnerResolver_DirectMatch(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project": {{Rules: Rules{"no-console": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/project/src/a.ts", configMap)
	if dir != "/project" {
		t.Errorf("Expected dir /project, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
		return
	}
}

func TestConfigOwnerResolver_Subdirectory(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project": {{Rules: Rules{"rule-a": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/project/src/deep/nested/file.ts", configMap)
	if dir != "/project" {
		t.Errorf("Expected dir /project, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
		return
	}
}

func TestConfigOwnerResolver_NearestWins(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project":              {{Rules: Rules{"root-rule": "error"}}},
		"/project/packages/foo": {{Rules: Rules{"foo-rule": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/project/packages/foo/src/a.ts", configMap)
	if dir != "/project/packages/foo" {
		t.Errorf("Expected nearest config dir /project/packages/foo, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
		return
	}
	if _, ok := cfg[0].Rules["foo-rule"]; !ok {
		t.Error("Expected foo-rule in config")
	}
}

func TestConfigOwnerResolver_NoMatch(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project": {{Rules: Rules{"rule-a": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/other/file.ts", configMap)
	if dir != "" {
		t.Errorf("Expected empty dir, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config for file outside all config dirs")
	}
}

func TestConfigOwnerResolver_EmptyMap(t *testing.T) {
	configMap := map[string]RslintConfig{}

	dir, cfg := resolveConfigOwner("/project/a.ts", configMap)
	if dir != "" {
		t.Errorf("Expected empty dir, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config for empty map")
	}
}

func TestConfigOwnerResolverSnapshotsDirectorySet(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project": {{Rules: Rules{"root-rule": "error"}}},
	}
	resolver := NewConfigOwnerResolver(configMap, nil)

	delete(configMap, "/project")
	configMap["/other"] = RslintConfig{{Rules: Rules{"other-rule": "error"}}}

	dir, cfg := resolver.Resolve("/project/src/a.ts")
	if dir != "/project" || cfg == nil {
		t.Fatalf("resolver changed after caller map mutation: dir=%q cfg=%v", dir, cfg)
	}
	if dir, cfg := resolver.Resolve("/other/a.ts"); dir != "" || cfg != nil {
		t.Fatalf("resolver observed a directory added after construction: dir=%q cfg=%v", dir, cfg)
	}
}

func TestConfigOwnerResolver_FileInConfigDir(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project": {{Rules: Rules{"rule-a": "error"}}},
	}

	// File directly in config dir (not in a subdirectory)
	dir, cfg := resolveConfigOwner("/project/a.ts", configMap)
	if dir != "/project" {
		t.Errorf("Expected dir /project, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
		return
	}
}

func TestConfigOwnerResolver_MultipleConfigsSameDepth(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project/packages/foo": {{Rules: Rules{"foo-rule": "error"}}},
		"/project/packages/bar": {{Rules: Rules{"bar-rule": "error"}}},
	}

	// File in foo should get foo's config
	dir, cfg := resolveConfigOwner("/project/packages/foo/src/a.ts", configMap)
	if dir != "/project/packages/foo" {
		t.Errorf("Expected /project/packages/foo, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
		return
	}
	if _, ok := cfg[0].Rules["foo-rule"]; !ok {
		t.Error("Expected foo-rule")
	}

	// File in bar should get bar's config
	dir, cfg = resolveConfigOwner("/project/packages/bar/src/b.ts", configMap)
	if dir != "/project/packages/bar" {
		t.Errorf("Expected /project/packages/bar, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
		return
	}
	if _, ok := cfg[0].Rules["bar-rule"]; !ok {
		t.Error("Expected bar-rule")
	}
}

func TestConfigOwnerResolver_SimilarPrefixNoFalseMatch(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project/src": {{Rules: Rules{"rule-a": "error"}}},
	}

	// /project/src-other should NOT match /project/src
	dir, cfg := resolveConfigOwner("/project/src-other/a.ts", configMap)
	if dir != "" {
		t.Errorf("Expected no match for src-other, got dir %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config for src-other")
	}
}

func TestConfigOwnerResolver_UsesExactCasing(t *testing.T) {
	configMap := map[string]RslintConfig{
		"C:/Repo":              {{Rules: Rules{"root": "error"}}},
		"C:/Repo/Packages/App": {{Rules: Rules{"app": "error"}}},
	}

	dir, cfg := resolveConfigOwner("c:/repo/packages/app/src/a.ts", configMap)
	if dir != "" || cfg != nil {
		t.Fatalf("exact lookup should not match different casing, got dir=%q cfg=%v", dir, cfg)
	}
}

func TestConfigOwnerResolver_NestedConfigDirs(t *testing.T) {
	// /project/src and /project/src/components both have configs.
	// File in components should pick the deeper config.
	configMap := map[string]RslintConfig{
		"/project/src":            {{Rules: Rules{"src-rule": "error"}}},
		"/project/src/components": {{Rules: Rules{"components-rule": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/project/src/components/Button.tsx", configMap)
	if dir != "/project/src/components" {
		t.Errorf("Expected /project/src/components, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
		return
	}
	if _, ok := cfg[0].Rules["components-rule"]; !ok {
		t.Error("Expected components-rule")
	}

	// File in src (not components) should pick src config
	dir, cfg = resolveConfigOwner("/project/src/utils.ts", configMap)
	if dir != "/project/src" {
		t.Errorf("Expected /project/src, got %s", dir)
	}
	if _, ok := cfg[0].Rules["src-rule"]; !ok {
		t.Error("Expected src-rule")
	}
}

func TestConfigOwnerResolver_RootConfig(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/": {{Rules: Rules{"root-rule": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/any/deep/path/file.ts", configMap)
	if dir != "/" {
		t.Errorf("Expected /, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
		return
	}
}

func TestConfigOwnerResolver_FileAboveAllConfigs(t *testing.T) {
	// Config only in a subdirectory; file in parent should not match
	configMap := map[string]RslintConfig{
		"/project/packages/foo": {{Rules: Rules{"foo-rule": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/project/root-file.ts", configMap)
	if dir != "" {
		t.Errorf("Expected no match, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil for file above config dir")
	}
}

func TestConfigOwnerResolver_TrailingSlashInKey(t *testing.T) {
	configMap := map[string]RslintConfig{
		"/project/": {{Rules: Rules{"rule-a": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/project/src/a.ts", configMap)
	if dir != "/project/" {
		t.Errorf("Expected /project/, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config with trailing slash key, got nil")
		return
	}
}

func TestConfigOwnerResolver_EmptyStringKey(t *testing.T) {
	configMap := map[string]RslintConfig{
		"": {{Rules: Rules{"rule-a": "error"}}},
	}

	// Empty string key should not match anything
	dir, cfg := resolveConfigOwner("/project/a.ts", configMap)
	if dir != "" || cfg != nil {
		t.Errorf("Expected no match for empty key, got dir=%q cfg=%v", dir, cfg)
	}
}

func TestConfigOwnerResolver_NilMap(t *testing.T) {
	// nil map should not panic and should return no match
	dir, cfg := resolveConfigOwner("/project/a.ts", nil)
	if dir != "" {
		t.Errorf("Expected empty dir for nil map, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config for nil map")
	}
}

func TestConfigOwnerResolver_FilePathEqualsConfigDir(t *testing.T) {
	// filePath == configDir: StartsWithDirectory returns false,
	// so the file should NOT match (it's not "inside" the directory)
	configMap := map[string]RslintConfig{
		"/project/src": {{Rules: Rules{"rule-a": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/project/src", configMap)
	if dir != "" {
		t.Errorf("Expected no match when filePath == configDir, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config when filePath == configDir")
	}
}

func TestConfigOwnerResolver_SingleConfigFallback(t *testing.T) {
	// Single config that matches — should always work for files under it
	configMap := map[string]RslintConfig{
		"/monorepo": {{Rules: Rules{"rule-a": "error"}}},
	}

	// File deep under config dir
	dir, cfg := resolveConfigOwner("/monorepo/packages/foo/src/deep/file.ts", configMap)
	if dir != "/monorepo" {
		t.Errorf("Expected /monorepo, got %s", dir)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
		return
	}

	// File NOT under config dir
	_, cfg = resolveConfigOwner("/other-repo/file.ts", configMap)
	if cfg != nil {
		t.Error("Expected nil for file outside config dir")
	}
}
