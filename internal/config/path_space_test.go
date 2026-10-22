package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rules"
)

type pathSpaceTestFS struct {
	vfs.FS
	realPaths     map[string]string
	caseSensitive bool
}

func (fs *pathSpaceTestFS) UseCaseSensitiveFileNames() bool { return fs.caseSensitive }
func (fs *pathSpaceTestFS) Realpath(filePath string) string {
	if realPath := fs.realPaths[filePath]; realPath != "" {
		return realPath
	}
	return filePath
}

func (fs *pathSpaceTestFS) Stat(path string) vfs.FileInfo {
	if fs.FS != nil {
		return fs.FS.Stat(path)
	}
	return nil
}

func TestFreezeFileIdentityAncestorFallback(t *testing.T) {
	for _, test := range []struct {
		name, file, canonical, parent string
		realPaths                     map[string]string
	}{
		{"one missing directory", "/alias/src/new.ts", "/physical/pkg/src/new.ts", "/physical/pkg/src", map[string]string{"/alias": "/physical/pkg"}},
		{"multiple missing directories", "/alias/src/deep/new.ts", "/physical/pkg/src/deep/new.ts", "/physical/pkg/src/deep", map[string]string{"/alias": "/physical/pkg"}},
		{"file symlink", "/alias/link.ts", "/other/source.ts", "/physical/pkg", map[string]string{"/alias": "/physical/pkg", "/physical/pkg/link.ts": "/other/source.ts"}},
		{"drive root", "C:/missing/new.ts", "C:/missing/new.ts", "C:/missing", map[string]string{"C:": "/must-not-resolve"}},
		{"UNC share root", "//server/share/missing/new.ts", "//server/share/missing/new.ts", "//server/share/missing", map[string]string{"//server/": "/must-not-resolve"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			identity := FreezeFileIdentity(test.file, &pathSpaceTestFS{realPaths: test.realPaths})
			if identity.Path != test.file || identity.CanonicalPath != test.canonical || identity.CanonicalParentPath != test.parent {
				t.Fatalf("identity = %+v, want file %s and parent %s", identity, test.canonical, test.parent)
			}
			if path := GitPath(identity); path != tspath.CombinePaths(test.parent, tspath.GetBaseFileName(test.file)) {
				t.Fatalf("Git coordinate followed the leaf target: %s", path)
			}
		})
	}
	t.Run("existing directory stops ancestor lookup", func(t *testing.T) {
		directory := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
		file := tspath.ResolvePath(directory, "new.ts")
		fsys := &pathSpaceTestFS{FS: osvfs.FS(), realPaths: map[string]string{tspath.GetDirectoryPath(directory): "/must-not-resolve"}}
		if identity := FreezeFileIdentity(file, fsys); identity.CanonicalParentPath != directory || identity.CanonicalPath != file {
			t.Fatalf("existing directory was replaced by an ancestor: %+v", identity)
		}
	})
}

type pathIdentityHookFS struct {
	vfs.FS
	realpath func(string) string
}

func (fsys *pathIdentityHookFS) Realpath(path string) string { return fsys.realpath(path) }

func TestFreezeFileIdentityAnchorsLeafInPhysicalParent(t *testing.T) {
	for _, cached := range []bool{false, true} {
		for _, suffix := range []string{"", "missing/deep"} {
			for _, leafSymlink := range []bool{false, true} {
				t.Run(fmt.Sprintf("cached=%t/suffix=%s/leafSymlink=%t", cached, suffix, leafSymlink), func(t *testing.T) {
					root := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
					first, second := tspath.ResolvePath(root, "first"), tspath.ResolvePath(root, "second")
					for _, directory := range []string{first, second} {
						if err := os.Mkdir(directory, 0o755); err != nil {
							t.Fatal(err)
						}
						if err := os.Mkdir(tspath.ResolvePath(directory, ".git"), 0o755); err != nil {
							t.Fatal(err)
						}
					}
					if err := os.WriteFile(tspath.ResolvePath(first, ".gitignore"), []byte("new.ts\n"), 0o644); err != nil {
						t.Fatal(err)
					}
					alias := tspath.ResolvePath(root, "alias")
					if err := os.Symlink(first, alias); err != nil {
						t.Skipf("directory symlink unavailable: %v", err)
					}
					parent := tspath.ResolvePath(first, suffix)
					physicalFile := tspath.ResolvePath(parent, "new.ts")
					wantCanonical := physicalFile
					if leafSymlink {
						if err := os.MkdirAll(parent, 0o755); err != nil {
							t.Fatal(err)
						}
						wantCanonical = tspath.ResolvePath(second, "source.ts")
						if err := os.WriteFile(wantCanonical, []byte("debugger;"), 0o644); err != nil {
							t.Fatal(err)
						}
						if err := os.Symlink(wantCanonical, physicalFile); err != nil {
							t.Skipf("file symlink unavailable: %v", err)
						}
					}
					file := tspath.ResolvePath(alias, suffix, "new.ts")
					moved := false
					hooks := &pathIdentityHookFS{FS: osvfs.FS()}
					hooks.realpath = func(path string) string {
						if path == file {
							t.Fatal("leaf resolution revisited the moving directory alias")
						}
						if path == physicalFile && !moved {
							moved = true
							if err := os.Remove(alias); err != nil {
								t.Fatal(err)
							}
							if err := os.Symlink(second, alias); err != nil {
								t.Fatal(err)
							}
						}
						return hooks.FS.Realpath(path)
					}
					var fsys vfs.FS = hooks
					if cached {
						fsys = cachedvfs.From(hooks)
					}
					identity := FreezeFileIdentity(file, fsys)
					if !moved || identity.Path != file || identity.CanonicalPath != wantCanonical || identity.CanonicalParentPath != parent {
						t.Fatalf("mixed directory generations after alias moved: %+v", identity)
					}
					policy := CollectGitPolicy(nil, root, fsys, []PathIdentity{identity}, nil)
					if !policy.ignored(GitPath(identity)) {
						t.Fatal("frozen target lost its containing repository's Git policy")
					}
				})
			}
		}
	}
}

func TestResolveConfigFilePathSpace(t *testing.T) {
	t.Run("drive spelling preserves lexical selectors through a directory symlink", func(t *testing.T) {
		fs := &pathSpaceTestFS{
			caseSensitive: false,
			realPaths: map[string]string{
				"C:/Repo":                           "C:/Repo",
				"c:/Repo/workspace/linked-src":      "C:/Physical/src",
				"c:/Repo/workspace/linked-src/a.ts": "C:/Physical/src/a.ts",
			},
		}
		const file = "c:/Repo/workspace/linked-src/a.ts"
		entries := RslintConfig{{
			Files: []string{"workspace/linked-src/*.ts"},
			Rules: Rules{"rule": "error"},
		}}
		resolved := NewFileConfigResolverWithFS(entries, "C:/Repo", fs, rules.All()).ResolveTarget(PathIdentity{
			Path: file, CanonicalPath: "C:/Physical/src/a.ts", CanonicalParentPath: "C:/Physical/src",
		})
		if resolved.MergedConfig == nil || resolved.MergedConfig.Rules["rule"] == nil {
			t.Fatalf("drive spelling lost the lexical selector: %#v", resolved)
		}
		entries[0].Ignores = append(literalFileIgnoresForTest(8), "workspace/linked-src/a.ts")
		ignored := NewFileConfigResolverWithFS(entries, "C:/Repo", fs, rules.All()).ResolveTarget(PathIdentity{
			Path: file, CanonicalPath: "C:/Physical/src/a.ts", CanonicalParentPath: "C:/Physical/src",
		})
		if ignored.MergedConfig != nil {
			t.Fatalf("indexed ignore lost the lexical target path: %#v", ignored)
		}
	})
	t.Run("symlink aliases use the physical config root", func(t *testing.T) {
		fs := &pathSpaceTestFS{
			caseSensitive: true,
			realPaths: map[string]string{
				"/alias":          "/real",
				"/alias/src/a.ts": "/real/src/a.ts",
				"/real/src/a.ts":  "/real/src/a.ts",
			},
		}
		for _, filePath := range []string{"/alias/src/a.ts", "/real/src/a.ts"} {
			matchFile, matchDir := ResolveConfigFilePathSpace(filePath, "/alias", fs)
			if matchFile != "/real/src/a.ts" || matchDir != "/real" {
				t.Fatalf("ResolveConfigFilePathSpace(%q) = (%q, %q)", filePath, matchFile, matchDir)
			}
		}
	})

	t.Run("directory alias preserves a leaf symlink's containing selector", func(t *testing.T) {
		fs := &pathSpaceTestFS{caseSensitive: true}
		identity := PathIdentity{
			Path: "/link/leaf.ts", CanonicalPath: "/outside/source.ts", CanonicalParentPath: "/real/sub",
		}
		entries := RslintConfig{{Files: []string{"sub/leaf.ts"}, Rules: Rules{"rule": "error"}}}
		resolved := NewFileConfigResolverWithFS(entries, "/real", fs, rules.All()).ResolveTarget(identity)
		if resolved.MergedConfig == nil || resolved.MergedConfig.Rules["rule"] == nil {
			t.Fatalf("containing selector lost to leaf destination: %#v", resolved)
		}
	})

	t.Run("realpath-normalized casing retains one matching space", func(t *testing.T) {
		fs := &pathSpaceTestFS{
			caseSensitive: false,
			realPaths: map[string]string{
				"C:/Repo":              "C:/Repo",
				"c:/repo/src/index.ts": "C:/Repo/src/index.ts",
			},
		}
		matchFile, matchDir := ResolveConfigFilePathSpace("c:/repo/src/index.ts", "C:/Repo", fs)
		if matchFile != "C:/Repo/src/index.ts" || matchDir != "C:/Repo" {
			t.Fatalf("case-insensitive path space = (%q, %q)", matchFile, matchDir)
		}
	})

	t.Run("distinct physical casing is not collapsed by a global case flag", func(t *testing.T) {
		fs := &pathSpaceTestFS{
			caseSensitive: false,
			realPaths: map[string]string{
				"C:/Repo":              "C:/Repo",
				"c:/repo/src/index.ts": "c:/repo/src/index.ts",
			},
		}
		matchFile, matchDir := ResolveConfigFilePathSpace("c:/repo/src/index.ts", "C:/Repo", fs)
		if matchFile != "c:/repo/src/index.ts" || matchDir != "C:/Repo" {
			t.Fatalf("distinct physical roots were collapsed: (%q, %q)", matchFile, matchDir)
		}
	})

	t.Run("native case alias retains relative path when the file symlink escapes", func(t *testing.T) {
		fs := &pathSpaceTestFS{
			caseSensitive: false,
			realPaths: map[string]string{
				"/repo/Project":         "/repo/Project",
				"/repo/project":         "/repo/Project",
				"/repo/project/link.ts": "/repo/shared.ts",
			},
		}
		matchFile, matchDir := ResolveConfigFilePathSpace("/repo/project/link.ts", "/repo/Project", fs)
		if matchFile != "/repo/Project/link.ts" || matchDir != "/repo/Project" {
			t.Fatalf("native alias path space = (%q, %q)", matchFile, matchDir)
		}
	})

	t.Run("external target canonical identity does not replace its lexical selector path", func(t *testing.T) {
		fs := &pathSpaceTestFS{
			caseSensitive: true,
			realPaths: map[string]string{
				"/repo/config":            "/physical/config",
				"/repo/workspace/link.js": "/elsewhere/source.js",
			},
		}
		matchFile, matchDir := ResolveConfigFilePathSpaceWithCanonical(
			"/repo/workspace/link.js",
			"/elsewhere/source.js",
			"/repo/config",
			fs,
		)
		if matchFile != "/repo/workspace/link.js" || matchDir != "/repo/config" {
			t.Fatalf("external lexical path space = (%q, %q)", matchFile, matchDir)
		}

		matchFile, matchDir = ResolveConfigFilePathSpaceWithCanonical(
			"/repo/workspace/inside-link.js",
			"/physical/config/source.js",
			"/repo/config",
			fs,
		)
		if matchFile != "/repo/workspace/inside-link.js" || matchDir != "/repo/config" {
			t.Fatalf("external link into config tree changed selector path = (%q, %q)", matchFile, matchDir)
		}
	})

	t.Run("shared alias ancestor keeps sibling relative selectors", func(t *testing.T) {
		fs := &pathSpaceTestFS{
			caseSensitive: true,
			realPaths: map[string]string{
				"/alias":                     "/real",
				"/alias/config":              "/real/config",
				"/real/workspace":            "/real/workspace",
				"/real/workspace/visible.ts": "/real/workspace/visible.ts",
			},
		}
		matchFile, matchDir := ResolveConfigFilePathSpace(
			"/real/workspace/visible.ts",
			"/alias/config",
			fs,
		)
		if matchFile != "/real/workspace/visible.ts" || matchDir != "/real/config" {
			t.Fatalf("shared alias path space = (%q, %q)", matchFile, matchDir)
		}

		entries := RslintConfig{{
			Files: []string{"../workspace/*.ts"},
			Rules: Rules{"rule": "error"},
		}}
		if merged := NewFileConfigResolverWithFS(entries, "/alias/config", fs, rules.All()).
			ConfigForFile("/real/workspace/visible.ts"); merged == nil || merged.Rules["rule"] == nil {
			t.Fatalf("shared alias selector did not match: %#v", merged)
		}
	})

	t.Run("shared alias ancestor preserves a file symlink basename", func(t *testing.T) {
		fs := &pathSpaceTestFS{
			caseSensitive: true,
			realPaths: map[string]string{
				"/alias":                  "/real",
				"/alias/config":           "/real/config",
				"/real/workspace":         "/real/workspace",
				"/real/workspace/link.ts": "/elsewhere/source.ts",
				"/elsewhere/source.ts":    "/elsewhere/source.ts",
			},
		}
		matchFile, matchDir := ResolveConfigFilePathSpace(
			"/real/workspace/link.ts",
			"/alias/config",
			fs,
		)
		if matchFile != "/real/workspace/link.ts" || matchDir != "/real/config" {
			t.Fatalf("shared alias file-symlink path space = (%q, %q)", matchFile, matchDir)
		}

		entries := RslintConfig{{
			Files: []string{"../workspace/link.ts"},
			Rules: Rules{"rule": "error"},
		}}
		if merged := NewFileConfigResolverWithFS(entries, "/alias/config", fs, rules.All()).
			ConfigForFile("/real/workspace/link.ts"); merged == nil || merged.Rules["rule"] == nil {
			t.Fatalf("shared alias file-symlink selector did not match: %#v", merged)
		}
	})
}

func TestRelativePathWithinConfigRootWindowsDrive(t *testing.T) {
	for _, test := range []struct {
		file, root, relative string
		within               bool
	}{
		{"c:/Repo/src/a.ts", "C:/Repo", "src/a.ts", true},
		{"C:/Repo", "c:/Repo", "", true},
		{"c:/Repo/src/a.ts", "C:/", "Repo/src/a.ts", true},
		{"d:/Repo/src/a.ts", "C:/Repo", "", false},
		{"c:/repo/src/a.ts", "C:/Repo", "", false},
		{"c:/Repo-other/a.ts", "C:/Repo", "", false},
	} {
		t.Run(test.file+" within "+test.root, func(t *testing.T) {
			relative, within := RelativePathWithinConfigRoot(test.file, test.root, true)
			if relative != test.relative || within != test.within {
				t.Fatalf("relative containment = (%q, %v), want (%q, %v)", relative, within, test.relative, test.within)
			}
			if within := isPathWithinNormalizedRoot(test.file, test.root); within != test.within {
				t.Fatalf("fast containment = %v, want %v", within, test.within)
			}
		})
	}
}

func TestResolveConfigDirectoryPathSpace(t *testing.T) {
	fs := &pathSpaceTestFS{
		caseSensitive: true,
		realPaths: map[string]string{
			"/repo/config-link": "/physical/config",
			"/repo/workspace":   "/repo/workspace",
		},
	}

	matchDirectory, matchConfigDirectory := ResolveConfigDirectoryPathSpace(
		"/repo/config-link",
		"/physical/config",
		fs,
	)
	if matchDirectory != "/physical/config" || matchConfigDirectory != "/physical/config" {
		t.Fatalf("directory alias into config tree = (%q, %q)", matchDirectory, matchConfigDirectory)
	}

	matchDirectory, matchConfigDirectory = ResolveConfigDirectoryPathSpace(
		"/repo/workspace",
		"/repo/config-link",
		fs,
	)
	if matchDirectory != "/repo/workspace" || matchConfigDirectory != "/repo/config-link" {
		t.Fatalf("external directory projection = (%q, %q)", matchDirectory, matchConfigDirectory)
	}
}

func TestResolveConfigFilePathSpacePreservesExternalRelativeDepth(t *testing.T) {
	fs := &pathSpaceTestFS{
		caseSensitive: true,
		realPaths: map[string]string{
			"/alias/config":        "/real/config",
			"/real/workspace/a.ts": "/real/workspace/a.ts",
			"/real/workspace":      "/real/workspace",
		},
	}
	matchFile, matchDirectory := ResolveConfigFilePathSpaceWithCanonical(
		"/real/workspace/a.ts",
		"/real/workspace/a.ts",
		"/alias/config",
		fs,
	)
	if matchFile != "/real/workspace/a.ts" || matchDirectory != "/alias/config" {
		t.Fatalf("external lexical pair = (%q, %q)", matchFile, matchDirectory)
	}
	config := RslintConfig{{
		Files: []string{"../../real/workspace/**"},
		Rules: Rules{"rule": "error"},
	}}
	if config.GetConfigForFile(matchFile, matchDirectory) == nil {
		t.Fatal("external selector lost its lexical ../ depth")
	}
}
