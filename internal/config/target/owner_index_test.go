package target

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rules"
)

type configPathSpaceFS struct {
	vfs.FS
	realPaths     map[string]string
	caseSensitive bool
}

type rejectRealpathAfterFreezeFS struct {
	*configPathSpaceFS
	reject bool
}

func (fs *rejectRealpathAfterFreezeFS) Realpath(filePath string) string {
	if fs.reject {
		panic("owner lookup consulted the live filesystem")
	}
	return fs.configPathSpaceFS.Realpath(filePath)
}

type retargetingConfigDirectoryFS struct {
	vfs.FS
	configPath  string
	configCalls int
}

func (fs *retargetingConfigDirectoryFS) UseCaseSensitiveFileNames() bool { return true }
func (fs *retargetingConfigDirectoryFS) Realpath(filePath string) string {
	if filePath != fs.configPath {
		return filePath
	}
	fs.configCalls++
	if fs.configCalls == 1 {
		return "/owner-a"
	}
	return "/owner-b"
}

func (fs *configPathSpaceFS) UseCaseSensitiveFileNames() bool { return fs.caseSensitive }
func (fs *configPathSpaceFS) Realpath(filePath string) string {
	if realPath := fs.realPaths[filePath]; realPath != "" {
		return realPath
	}
	return filePath
}

func resolveConfigOwner(filePath string, configMap map[string]rslintconfig.RslintConfig) (string, rslintconfig.RslintConfig) {
	return resolveConfigOwnerWithFS(filePath, configMap, nil)
}

func resolveConfigOwnerWithFS(
	filePath string,
	configMap map[string]rslintconfig.RslintConfig,
	fsys vfs.FS,
) (string, rslintconfig.RslintConfig) {
	target := FreezeFileIdentity(filePath, fsys)
	owner, ok := NewOwnerIndex(configMap, fsys).Resolve(target)
	if !ok {
		return "", nil
	}
	return owner, configMap[owner]
}

func TestOwnerIndexUsesVerifiedNativeCaseAliasBeforeFileRealpath(t *testing.T) {
	fs := &configPathSpaceFS{
		caseSensitive: false,
		realPaths: map[string]string{
			"/repo/Project":         "/repo/Project",
			"/repo/project":         "/repo/Project",
			"/repo/project/link.ts": "/repo/shared.ts",
		},
	}
	configMap := map[string]rslintconfig.RslintConfig{
		"/repo/Project": {{Rules: rslintconfig.Rules{"rule": "error"}}},
	}
	dir, cfg := resolveConfigOwnerWithFS("/repo/project/link.ts", configMap, fs)
	if dir != "/repo/Project" || cfg == nil {
		t.Fatalf("native case alias did not retain config owner: dir=%q cfg=%v", dir, cfg)
	}
}

func TestOwnerIndexResolveUsesOnlyFrozenIdentity(t *testing.T) {
	fs := &rejectRealpathAfterFreezeFS{configPathSpaceFS: &configPathSpaceFS{
		caseSensitive: false,
		realPaths: map[string]string{
			"/repo/Project": "/repo/Project",
		},
	}}
	index := NewOwnerIndex(
		map[string]rslintconfig.RslintConfig{"/repo/Project": nil},
		fs,
	)
	fs.reject = true

	owner, ok := index.Resolve(rslintconfig.PathIdentity{
		Path:                "/repo/project/src/index.ts",
		CanonicalPath:       "/repo/Project/src/index.ts",
		CanonicalParentPath: "/repo/Project/src",
	})
	if !ok || owner != "/repo/Project" {
		t.Fatalf("frozen native-case owner = %q, %v", owner, ok)
	}
}

func TestOwnerIndexFreezesOwnerAndAuthoredBaseTogether(t *testing.T) {
	const configDirectory = "/config-link"
	fs := &retargetingConfigDirectoryFS{
		FS:         &configPathSpaceFS{caseSensitive: true},
		configPath: configDirectory,
	}
	configMap := map[string]rslintconfig.RslintConfig{
		configDirectory: {{
			Files: []string{"src/*.ts"},
			Rules: rslintconfig.Rules{"selected": "error"},
		}},
	}
	resolver := NewOwnerIndex(configMap, fs)
	target := File{
		PathIdentity: rslintconfig.PathIdentity{
			Path:                "/outside-link/src/index.ts",
			CanonicalPath:       "/owner-a/src/index.ts",
			CanonicalParentPath: "/owner-a/src",
		},
	}

	configDir, ok := resolver.Resolve(target.Identity())
	entries := configMap[configDir]
	fileResolver, err := rslintconfig.NewFileConfigResolverWithPathSpaces(
		entries,
		configDir,
		fs,
		resolver.PathSpaces(),
		rules.All(),
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	resolved := fileResolver.ResolveTarget(target.Identity())
	if !ok || configDir != configDirectory || resolved.MergedConfig == nil {
		t.Fatalf("frozen owner/config resolution = %q, %+v, %v", configDir, resolved, ok)
	}
	if _, selected := resolved.MergedConfig.Rules["selected"]; !selected {
		t.Fatalf("frozen authored base lost selected rules: %+v", resolved.MergedConfig.Rules)
	}
	if fs.configCalls != 1 {
		t.Fatalf("config directory Realpath calls = %d, want one catalog observation", fs.configCalls)
	}
}

func TestOwnerIndexPrefersLexicalHierarchy(t *testing.T) {
	fs := &configPathSpaceFS{
		caseSensitive: true,
		realPaths: map[string]string{
			"/alias/pkg": "/real/pkg",
		},
	}
	configMap := map[string]rslintconfig.RslintConfig{
		"/real":      {{Rules: rslintconfig.Rules{"root": "error"}}},
		"/alias/pkg": {{Rules: rslintconfig.Rules{"package": "error"}}},
	}
	resolver := NewOwnerIndex(configMap, fs)

	dir, ok := resolver.Resolve(FreezeFileIdentity("/real/pkg/src/a.ts", fs))
	if dir != "/real" || !ok {
		t.Fatalf("physical config replaced lexical owner: dir=%q ok=%v", dir, ok)
	}

	index := newConfigDirectoryIndex(configMap, fs)
	children := index.childConfigDirs("/real")
	if len(children) != 0 {
		t.Fatalf("physical hierarchy created lexical child boundaries: %v", children)
	}
}

func TestOwnerIndexChildOwnerDirectories(t *testing.T) {
	resolver := NewOwnerIndex(map[string]rslintconfig.RslintConfig{
		"/repo":                  nil,
		"/repo/packages/app":     nil,
		"/repo/packages/lib":     nil,
		"/repo/packages/app/e2e": nil,
	}, nil)

	children := resolver.ChildOwnerDirectories("/repo")
	if len(children) != 2 || children[0] != "/repo/packages/app" || children[1] != "/repo/packages/lib" {
		t.Fatalf("root child boundaries = %v", children)
	}
	if deep := resolver.ChildOwnerDirectories("/repo/packages/app"); len(deep) != 1 || deep[0] != "/repo/packages/app/e2e" {
		t.Fatalf("nested child boundaries = %v", deep)
	}

	children[0] = "/mutated"
	if fresh := resolver.ChildOwnerDirectories("/repo"); fresh[0] != "/repo/packages/app" {
		t.Fatalf("ChildOwnerDirectories exposed index state: %v", fresh)
	}
}

func TestOwnerIndexForAutomaticTargetsExcludesExplicitOnlyBoundary(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/repo":          nil,
		"/repo/ignored":  nil,
		"/repo/packages": nil,
	}
	scopes := map[string]OwnerScope{
		"/repo/ignored": {ExplicitOnly: true},
	}

	automatic := newOwnerIndexForAutomaticTargets(configMap, scopes, nil)
	dir, _ := automatic.Resolve(FreezeFileIdentity("/repo/ignored/automatic.ts", nil))
	if dir != "/repo" {
		t.Fatalf("explicit-only config claimed automatic target: %q", dir)
	}
	children := automatic.ChildOwnerDirectories("/repo")
	if len(children) != 1 || children[0] != "/repo/packages" {
		t.Fatalf("automatic child boundaries = %v", children)
	}

	complete := NewOwnerIndex(configMap, nil)
	dir, _ = complete.Resolve(FreezeFileIdentity("/repo/ignored/explicit.ts", nil))
	if dir != "/repo/ignored" {
		t.Fatalf("complete resolver lost literal owner: %q", dir)
	}
}

func TestOwnerIndexUsesPhysicalFallbackWithoutLexicalOwner(t *testing.T) {
	fs := &configPathSpaceFS{
		caseSensitive: true,
		realPaths: map[string]string{
			"/alias/pkg":         "/real/pkg",
			"/outside/link/a.ts": "/real/pkg/src/a.ts",
		},
	}
	configMap := map[string]rslintconfig.RslintConfig{
		"/alias/pkg": {{Rules: rslintconfig.Rules{"package": "error"}}},
	}

	dir, cfg := resolveConfigOwnerWithFS("/outside/link/a.ts", configMap, fs)
	if dir != "/alias/pkg" || cfg == nil {
		t.Fatalf("physical fallback did not resolve aliased config: dir=%q cfg=%v", dir, cfg)
	}
}

func TestOwnerIndexKeepsLexicalOwnerAcrossUnrelatedRealpathTree(t *testing.T) {
	fs := &configPathSpaceFS{
		caseSensitive: true,
		realPaths: map[string]string{
			"/lex/link/a.ts": "/physical/a.ts",
		},
	}
	configMap := map[string]rslintconfig.RslintConfig{
		"/lex":      {{Rules: rslintconfig.Rules{"lexical": "error"}}},
		"/physical": {{Rules: rslintconfig.Rules{"physical": "error"}}},
	}

	dir, cfg := resolveConfigOwnerWithFS("/lex/link/a.ts", configMap, fs)
	if dir != "/lex" || cfg == nil {
		t.Fatalf("unrelated physical tree replaced lexical owner: dir=%q cfg=%v", dir, cfg)
	}
}

func TestOwnerIndex_DirectMatch(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/project": {{Rules: rslintconfig.Rules{"no-console": "error"}}},
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

func TestOwnerIndex_Subdirectory(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/project": {{Rules: rslintconfig.Rules{"rule-a": "error"}}},
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

func TestOwnerIndex_NearestWins(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/project":              {{Rules: rslintconfig.Rules{"root-rule": "error"}}},
		"/project/packages/foo": {{Rules: rslintconfig.Rules{"foo-rule": "error"}}},
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

func TestOwnerIndex_NoMatch(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/project": {{Rules: rslintconfig.Rules{"rule-a": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/other/file.ts", configMap)
	if dir != "" {
		t.Errorf("Expected empty dir, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config for file outside all config dirs")
	}
}

func TestOwnerIndex_EmptyMap(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{}

	dir, cfg := resolveConfigOwner("/project/a.ts", configMap)
	if dir != "" {
		t.Errorf("Expected empty dir, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config for empty map")
	}
}

func TestOwnerIndexSnapshotsDirectorySet(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/project": {{Rules: rslintconfig.Rules{"root-rule": "error"}}},
	}
	resolver := NewOwnerIndex(configMap, nil)

	delete(configMap, "/project")
	configMap["/other"] = rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"other-rule": "error"}}}

	dir, ok := resolver.Resolve(FreezeFileIdentity("/project/src/a.ts", nil))
	if dir != "/project" || !ok {
		t.Fatalf("resolver changed after caller map mutation: dir=%q ok=%v", dir, ok)
	}
	if dir, ok := resolver.Resolve(FreezeFileIdentity("/other/a.ts", nil)); dir != "" || ok {
		t.Fatalf("resolver observed a directory added after construction: dir=%q ok=%v", dir, ok)
	}
}

func TestOwnerIndex_FileInConfigDir(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/project": {{Rules: rslintconfig.Rules{"rule-a": "error"}}},
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

func TestOwnerIndex_MultipleConfigsSameDepth(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/project/packages/foo": {{Rules: rslintconfig.Rules{"foo-rule": "error"}}},
		"/project/packages/bar": {{Rules: rslintconfig.Rules{"bar-rule": "error"}}},
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

func TestOwnerIndex_SimilarPrefixNoFalseMatch(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/project/src": {{Rules: rslintconfig.Rules{"rule-a": "error"}}},
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

func TestOwnerIndex_UsesExactCasing(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"C:/Repo":              {{Rules: rslintconfig.Rules{"root": "error"}}},
		"C:/Repo/Packages/App": {{Rules: rslintconfig.Rules{"app": "error"}}},
	}

	dir, cfg := resolveConfigOwner("c:/repo/packages/app/src/a.ts", configMap)
	if dir != "" || cfg != nil {
		t.Fatalf("exact lookup should not match different casing, got dir=%q cfg=%v", dir, cfg)
	}
}

func TestOwnerIndex_NestedConfigDirs(t *testing.T) {
	// /project/src and /project/src/components both have configs.
	// File in components should pick the deeper config.
	configMap := map[string]rslintconfig.RslintConfig{
		"/project/src":            {{Rules: rslintconfig.Rules{"src-rule": "error"}}},
		"/project/src/components": {{Rules: rslintconfig.Rules{"components-rule": "error"}}},
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

func TestOwnerIndex_RootConfig(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/": {{Rules: rslintconfig.Rules{"root-rule": "error"}}},
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

func TestOwnerIndex_FileAboveAllConfigs(t *testing.T) {
	// Config only in a subdirectory; file in parent should not match
	configMap := map[string]rslintconfig.RslintConfig{
		"/project/packages/foo": {{Rules: rslintconfig.Rules{"foo-rule": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/project/root-file.ts", configMap)
	if dir != "" {
		t.Errorf("Expected no match, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil for file above config dir")
	}
}

func TestOwnerIndex_TrailingSlashInKey(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/project/": {{Rules: rslintconfig.Rules{"rule-a": "error"}}},
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

func TestOwnerIndex_EmptyStringKey(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"": {{Rules: rslintconfig.Rules{"rule-a": "error"}}},
	}

	// Empty string key should not match anything
	dir, cfg := resolveConfigOwner("/project/a.ts", configMap)
	if dir != "" || cfg != nil {
		t.Errorf("Expected no match for empty key, got dir=%q cfg=%v", dir, cfg)
	}
}

func TestOwnerIndex_NilMap(t *testing.T) {
	// nil map should not panic and should return no match
	dir, cfg := resolveConfigOwner("/project/a.ts", nil)
	if dir != "" {
		t.Errorf("Expected empty dir for nil map, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config for nil map")
	}
}

func TestOwnerIndex_FilePathEqualsConfigDir(t *testing.T) {
	// filePath == configDir: StartsWithDirectory returns false,
	// so the file should NOT match (it's not "inside" the directory)
	configMap := map[string]rslintconfig.RslintConfig{
		"/project/src": {{Rules: rslintconfig.Rules{"rule-a": "error"}}},
	}

	dir, cfg := resolveConfigOwner("/project/src", configMap)
	if dir != "" {
		t.Errorf("Expected no match when filePath == configDir, got %s", dir)
	}
	if cfg != nil {
		t.Error("Expected nil config when filePath == configDir")
	}
}

func TestOwnerIndex_SingleConfigFallback(t *testing.T) {
	// Single config that matches — should always work for files under it
	configMap := map[string]rslintconfig.RslintConfig{
		"/monorepo": {{Rules: rslintconfig.Rules{"rule-a": "error"}}},
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
