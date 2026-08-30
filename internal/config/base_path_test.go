package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

func TestBasePathScopesExistingFilesAndIgnoresMatchers(t *testing.T) {
	basePath := "pkg"
	config := ConfigWithResolvedBasePaths(RslintConfig{{
		BasePath: &basePath,
		Files:    []string{"**/*.ts"},
		Ignores:  []string{"**/*.test.ts"},
		Rules:    Rules{"no-debugger": "error"},
	}}, "/repo")

	for _, test := range []struct {
		name string
		path string
		want bool
	}{
		{name: "inside", path: "/repo/pkg/src/app.ts", want: true},
		{name: "local ignore", path: "/repo/pkg/src/app.test.ts", want: false},
		{name: "outside", path: "/repo/other/app.ts", want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, config.GetConfigForFile(test.path, "/repo") != nil, test.want)
		})
	}

	escape := ConfigWithResolvedBasePaths(RslintConfig{{
		BasePath: &basePath,
		Files:    []string{"../outside.ts"},
		Rules:    Rules{"no-debugger": "error"},
	}}, "/repo")
	assert.Assert(t, escape.GetConfigForFile("/repo/outside.ts", "/repo") == nil)

	wholeEntry := ConfigWithResolvedBasePaths(RslintConfig{{
		BasePath: &basePath,
		Rules:    Rules{"no-debugger": "error"},
	}}, "/repo")
	assert.Assert(t, wholeEntry.GetConfigForFile("/repo/pkg/app.ts", "/repo") != nil)
	assert.Assert(t, wholeEntry.GetConfigForFile("/repo/app.ts", "/repo") == nil)

	baseItself := ConfigWithResolvedBasePaths(RslintConfig{{
		BasePath: &basePath,
		Files:    []string{"**"},
		Rules:    Rules{"no-debugger": "error"},
	}}, "/repo")
	assert.Assert(t, baseItself.GetConfigForFile("/repo/pkg", "/repo") != nil)
}

func TestBasePathDotReusesExistingMatchers(t *testing.T) {
	basePath := "."
	plain := RslintConfig{
		{Ignores: []string{"dist/**"}},
		{Files: []string{"**/*.ts"}, Ignores: []string{"**/*.test.ts"}, Rules: Rules{"no-debugger": "error"}},
	}
	scoped := ConfigWithResolvedBasePaths(RslintConfig{
		{BasePath: &basePath, Ignores: []string{"dist/**"}},
		{BasePath: &basePath, Files: []string{"**/*.ts"}, Ignores: []string{"**/*.test.ts"}, Rules: Rules{"no-debugger": "error"}},
	}, "/repo")

	for _, path := range []string{
		"/repo/src/app.ts",
		"/repo/src/app.test.ts",
		"/repo/dist/app.ts",
		"/repo/src/app.js",
	} {
		assert.Equal(
			t,
			scoped.GetConfigForFile(path, "/repo") != nil,
			plain.GetConfigForFile(path, "/repo") != nil,
			path,
		)
	}
}

func TestBasePathScopesFilesystemFreeWindowsCoordinates(t *testing.T) {
	for _, test := range []struct {
		name     string
		root     string
		spelling string
	}{
		{name: "drive", root: "C:/Repo", spelling: "c:/repo"},
		{name: "UNC", root: "//SERVER/share/Repo", spelling: "//server/SHARE/repo"},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, basePath := range []string{".", "pkg"} {
				t.Run(basePath, func(t *testing.T) {
					relativeTarget := "src/app.ts"
					if basePath != "." {
						relativeTarget = basePath + "/" + relativeTarget
					}

					files := ConfigWithResolvedBasePaths(RslintConfig{{
						BasePath: &basePath,
						Files:    []string{"**/*.ts"},
						Rules:    Rules{"no-debugger": "error"},
					}}, test.root)
					assert.Assert(t, files.GetConfigForFile(test.spelling+"/"+relativeTarget, test.root) != nil)

					globalIgnore := ConfigWithResolvedBasePaths(RslintConfig{
						{BasePath: &basePath, Ignores: []string{"**/*.ts"}},
						{Rules: Rules{}},
					}, test.root)
					assert.Assert(t, globalIgnore.IsFileIgnored(test.spelling+"/"+relativeTarget, test.root))
				})
			}
		})
	}
}

func TestBasePathDoesNotChangeFilesystemFreeMatcherCoordinates(t *testing.T) {
	basePath := "."
	unrelatedBasePath := "D:/other"
	for _, test := range []struct {
		name   string
		root   string
		target string
	}{
		{name: "drive", root: "C:/Repo", target: "c:/repo/src/app.ts"},
		{name: "UNC", root: "//SERVER/share/Repo", target: "//server/SHARE/repo/src/app.ts"},
	} {
		t.Run(test.name, func(t *testing.T) {
			plainFiles := RslintConfig{{
				Files: []string{"src/*.ts"},
				Rules: Rules{"no-debugger": "error"},
			}}
			dotFiles := ConfigWithResolvedBasePaths(RslintConfig{{
				BasePath: &basePath,
				Files:    []string{"src/*.ts"},
				Rules:    Rules{"no-debugger": "error"},
			}}, test.root)
			withUnrelatedBasePath := append(append(RslintConfig(nil), plainFiles...), ConfigEntry{
				BasePath: &unrelatedBasePath,
				Files:    []string{"never"},
			})
			wantFiles := plainFiles.GetConfigForFile(test.target, test.root) != nil
			assert.Equal(t, dotFiles.GetConfigForFile(test.target, test.root) != nil, wantFiles)
			assert.Equal(t, withUnrelatedBasePath.GetConfigForFile(test.target, test.root) != nil, wantFiles)

			plainIgnore := RslintConfig{
				{Ignores: []string{"src/*.ts"}},
				{Rules: Rules{}},
			}
			dotIgnore := ConfigWithResolvedBasePaths(RslintConfig{
				{BasePath: &basePath, Ignores: []string{"src/*.ts"}},
				{Rules: Rules{}},
			}, test.root)
			withUnrelatedBasePath = append(append(RslintConfig(nil), plainIgnore...), ConfigEntry{
				BasePath: &unrelatedBasePath,
				Files:    []string{"never"},
			})
			wantIgnored := plainIgnore.IsFileIgnored(test.target, test.root)
			assert.Equal(t, dotIgnore.IsFileIgnored(test.target, test.root), wantIgnored)
			assert.Equal(t, withUnrelatedBasePath.IsFileIgnored(test.target, test.root), wantIgnored)
		})
	}
}

func TestBasePathIsLiteralNotGlob(t *testing.T) {
	basePath := "pkg*"
	config := ConfigWithResolvedBasePaths(RslintConfig{{
		BasePath: &basePath,
		Files:    []string{"**/*.ts"},
		Rules:    Rules{"no-debugger": "error"},
	}}, "/repo")

	assert.Assert(t, config.GetConfigForFile("/repo/pkg*/app.ts", "/repo") != nil)
	assert.Assert(t, config.GetConfigForFile("/repo/pkg1/app.ts", "/repo") == nil)
}

func TestBasePathScopesGlobalIgnores(t *testing.T) {
	basePath := "pkg"
	config := ConfigWithResolvedBasePaths(RslintConfig{
		{BasePath: &basePath, Ignores: []string{"blocked/**"}},
		{Files: []string{"**/*.js"}, Rules: Rules{"no-debugger": "error"}},
	}, "/repo")

	assert.Assert(t, config.GetConfigForFile("/repo/pkg/blocked/app.js", "/repo") == nil)
	assert.Assert(t, config.GetConfigForFile("/repo/blocked/app.js", "/repo") != nil)
	assert.Assert(t, config.GetConfigForFile("/repo/pkg/app.js", "/repo") != nil)
	assert.Assert(t, !newConfigTargetResolver(config, "/repo", nil).canPruneDirectory("/repo/pkg", ""))
}

func TestBasePathGlobalIgnoreKeepsConfigArrayRootReachable(t *testing.T) {
	basePath := ".."
	config := ConfigWithResolvedBasePaths(RslintConfig{
		{BasePath: &basePath, Ignores: []string{"*"}},
		{Files: []string{"**/*.js"}, Rules: Rules{"no-debugger": "error"}},
	}, "/repo")
	resolver := newConfigTargetResolver(config, "/repo", nil)

	assert.Assert(t, config.GetConfigForFile("/repo/visible.js", "/repo") != nil)
	assert.Assert(t, !resolver.canPruneDirectory("/repo", ""))
}

func TestBasePathNegationKeepsReachableSubtreeOpen(t *testing.T) {
	basePath := "foo/pkg"
	config := ConfigWithResolvedBasePaths(RslintConfig{
		{Ignores: []string{"foo/**/*"}},
		{BasePath: &basePath, Ignores: []string{"!keep.js"}},
		{Files: []string{"**/*.js"}, Rules: Rules{"no-debugger": "error"}},
	}, "/repo")
	resolver := newConfigTargetResolver(config, "/repo", nil)

	assert.Assert(t, config.GetConfigForFile("/repo/foo/pkg/keep.js", "/repo") != nil)
	assert.Assert(t, !resolver.canPruneDirectory("/repo/foo", ""))
	assert.Assert(t, !resolver.canPruneDirectory("/repo/foo/pkg", ""))
}

func TestBasePathNegationReopensPhysicalAliasSubtree(t *testing.T) {
	root := t.TempDir()
	physical := filepath.Join(root, "physical")
	firstAlias := filepath.Join(root, "alias-one")
	secondAlias := filepath.Join(root, "alias-two")
	if err := os.MkdirAll(filepath.Join(physical, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(physical, firstAlias); err != nil {
		t.Skipf("directory symlinks unavailable: %v", err)
	}
	if err := os.Symlink(physical, secondAlias); err != nil {
		t.Skipf("directory symlinks unavailable: %v", err)
	}
	target := filepath.Join(secondAlias, "pkg", "keep.js")
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	basePath := filepath.Join(firstAlias, "pkg")
	config := ConfigWithResolvedBasePaths(RslintConfig{
		{Ignores: []string{"alias-two/**/*"}},
		{BasePath: &basePath, Ignores: []string{"!keep.js"}},
		{Files: []string{"**/*.js"}, Rules: Rules{"no-debugger": "error"}},
	}, root)
	fsys := osvfs.FS()
	root = tspath.NormalizePath(root)
	physical = tspath.NormalizePath(fsys.Realpath(physical))
	target = tspath.NormalizePath(target)
	snapshot := NewPathSpaceSnapshot(map[string]RslintConfig{root: config}, fsys)
	matcher, err := NewTargetMatcherWithPathSpaces(config, root, fsys, snapshot)
	assert.NilError(t, err)

	decision := matcher.MatchFile(PathIdentity{
		Path:                target,
		CanonicalPath:       tspath.NormalizePath(fsys.Realpath(target)),
		CanonicalParentPath: tspath.NormalizePath(fsys.Realpath(tspath.GetDirectoryPath(target))),
	})
	assert.Assert(t, decision.Selected)
	assert.Assert(t, !decision.GloballyIgnored)
	assert.Assert(t, !matcher.CanPruneDirectory(DirectoryIdentity{
		LexicalPath:   tspath.NormalizePath(secondAlias),
		CanonicalPath: physical,
	}))
}

func TestConfigWithResolvedBasePathsUsesOneInternalPathOrigin(t *testing.T) {
	basePath := "pkg"
	config := ConfigWithAuthoredPathBase(RslintConfig{
		{Rules: Rules{"cwd-rule": "error"}},
		{BasePath: &basePath, Rules: Rules{"base-rule": "error"}},
	}, "/cwd")

	effective := ConfigWithResolvedBasePaths(config, "/configs")
	assert.Equal(t, configEntryBaseDirectory(effective[0], "/owner"), "/cwd")
	assert.Equal(t, configEntryBaseDirectory(effective[1], "/owner"), "/configs/pkg")
	assert.Equal(t, configEntryPathOrigin(effective[1], "/owner").configArrayBase, "/configs")
	assert.Assert(t, configEntryPathOrigin(effective[1], "/owner").basePathScoped)
	assert.Assert(t, &effective[0] != &config[0])
	again := ConfigWithResolvedBasePaths(effective, "/other")
	assert.Assert(t, &again[0] == &effective[0])

	snapshot := NewPathSpaceSnapshot(map[string]RslintConfig{"/owner": effective}, nil)
	_, configArrayBaseFrozen := snapshot.PhysicalDirectory("/configs")
	assert.Assert(t, configArrayBaseFrozen)
	_, err := NewTargetMatcherWithPathSpaces(effective, "/owner", nil, snapshot)
	assert.NilError(t, err)
}
