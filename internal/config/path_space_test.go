package config

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
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

func TestResolveConfigFilePathSpace(t *testing.T) {
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
