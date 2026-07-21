package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

func TestCLIAndAPIIgnoreConformance(t *testing.T) {
	tests := []struct {
		name          string
		configDir     string
		relative      string
		globalIgnores []string
		entryIgnores  []string
		gitignores    map[string]string
		symlinkDir    bool
		targetIgnore  string
		wantLinted    bool
	}{
		{
			name:          "global config ignore suppresses explicit target",
			relative:      "global.ts",
			globalIgnores: []string{"global.ts"},
		},
		{
			name:         "entry ignore keeps target but removes rules",
			relative:     "entry.ts",
			entryIgnores: []string{"entry.ts"},
			wantLinted:   true,
		},
		{
			name:       "root gitignore suppresses explicit target",
			relative:   "ignored.ts",
			gitignores: map[string]string{".gitignore": "ignored.ts\n"},
		},
		{
			name:          "config negation restores gitignored explicit target",
			relative:      "dist/important.ts",
			globalIgnores: []string{"!dist/important.ts"},
			gitignores:    map[string]string{".gitignore": "dist/\n"},
			wantLinted:    true,
		},
		{
			name:     "nested negation restores explicit target",
			relative: "nested/keep.ts",
			gitignores: map[string]string{
				".gitignore":        "nested/*.ts\n",
				"nested/.gitignore": "!keep.ts\n",
			},
			wantLinted: true,
		},
		{
			name:     "ignored parent blocks nested source",
			relative: "blocked/keep.ts",
			gitignores: map[string]string{
				".gitignore":         "blocked/\n",
				"blocked/.gitignore": "!keep.ts\n",
			},
		},
		{
			name:      "parent ignore does not affect nested config",
			configDir: "packages/app",
			relative:  "ignored/keep.ts",
			gitignores: map[string]string{
				".gitignore":                      "/packages/app/ignored/\n",
				"packages/app/ignored/.gitignore": "!keep.ts\n",
			},
			wantLinted: true,
		},
		{
			name:     "pruned nested source does not override root negation",
			relative: "dist/types/private.ts",
			gitignores: map[string]string{
				".gitignore":            "dist/\n!dist/types/\n",
				"dist/types/.gitignore": "private.ts\n",
			},
			wantLinted: true,
		},
		{
			name:       "directory symlink remains lintable without ignore",
			relative:   "link/source.ts",
			symlinkDir: true,
			wantLinted: true,
		},
		{
			name:       "directory symlink obeys lexical root gitignore",
			relative:   "link/source.ts",
			gitignores: map[string]string{".gitignore": "link/source.ts\n"},
			symlinkDir: true,
		},
		{
			name:         "directory symlink skips target gitignore source",
			relative:     "link/source.ts",
			symlinkDir:   true,
			targetIgnore: "source.ts\n",
			wantLinted:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspace := t.TempDir()
			configDir := workspace
			if test.configDir != "" {
				configDir = filepath.Join(workspace, test.configDir)
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if test.symlinkDir {
				targetDir := t.TempDir()
				if test.targetIgnore != "" {
					if err := os.WriteFile(filepath.Join(targetDir, ".gitignore"), []byte(test.targetIgnore), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Symlink(targetDir, filepath.Join(configDir, "link")); err != nil {
					t.Skipf("directory symlink unavailable: %v", err)
				}
			}
			for relative, content := range test.gitignores {
				filePath := filepath.Join(workspace, relative)
				if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			target := filepath.Join(configDir, test.relative)
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, []byte("debugger; const value = ;\n"), 0o644); err != nil {
				t.Fatal(err)
			}

			entries := make([]map[string]any, 0, 2)
			if test.globalIgnores != nil {
				entries = append(entries, map[string]any{"ignores": test.globalIgnores})
			}
			entry := map[string]any{
				"files": []string{"**/*.ts"},
				"rules": map[string]string{"no-debugger": "error"},
			}
			if test.entryIgnores != nil {
				entry["ignores"] = test.entryIgnores
			}
			entries = append(entries, entry)
			configJSON, err := json.Marshal(entries)
			if err != nil {
				t.Fatal(err)
			}
			configPath := filepath.Join(configDir, "rslint.json")
			if err := os.WriteFile(configPath, configJSON, 0o644); err != nil {
				t.Fatal(err)
			}

			code, stdout, stderr := runLintPipelineForTest(t, configDir, lintArgs{
				Config:         configPath,
				AllowFiles:     []string{tspath.NormalizePath(target)},
				Format:         "default",
				NoColor:        true,
				SingleThreaded: true,
			})
			cliLinted := strings.Contains(stdout, "TypeScript(TS")
			if cliLinted != test.wantLinted {
				t.Fatalf("CLI linted=%v, want %v: code=%d stdout=%q stderr=%q", cliLinted, test.wantLinted, code, stdout, stderr)
			}
			if test.wantLinted && code == 0 {
				t.Fatalf("CLI syntax diagnostic must fail: stdout=%q stderr=%q", stdout, stderr)
			}
			if !test.wantLinted && code != 0 {
				t.Fatalf("CLI ignored target must exit cleanly: stdout=%q stderr=%q", stdout, stderr)
			}

			response, err := (&IPCHandler{}).HandleLint(api.LintRequest{
				Config:           configJSON,
				ConfigDirectory:  configDir,
				WorkingDirectory: configDir,
				Files:            []string{target},
			})
			if err != nil {
				t.Fatalf("API lint: %v", err)
			}
			apiLinted := response.FileCount == 1 && len(response.Diagnostics) > 0 &&
				strings.HasPrefix(response.Diagnostics[0].RuleName, "TypeScript(TS")
			if apiLinted != test.wantLinted {
				t.Fatalf("API linted=%v, want %v: response=%+v", apiLinted, test.wantLinted, response)
			}
		})
	}
}

func TestCLIMultiConfigGitignoreIsolation(t *testing.T) {
	workspace := t.TempDir()
	firstDir := filepath.Join(workspace, "packages", "first")
	secondDir := filepath.Join(workspace, "packages", "second")
	firstTarget := filepath.Join(firstDir, "source.ts")
	secondTarget := filepath.Join(secondDir, "source.ts")
	for _, target := range []string{firstTarget, secondTarget} {
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(firstDir, ".gitignore"), []byte("source.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	entry := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		Rules: rslintconfig.Rules{"no-debugger": "error"},
	}}
	catalog := &discovery.ConfigCatalog{
		Configs: map[string]rslintconfig.RslintConfig{
			tspath.NormalizePath(firstDir): rslintconfig.ConfigWithGitignore(
				entry,
				tspath.NormalizePath(firstDir),
				osvfs.FS(),
				[]string{tspath.NormalizePath(firstTarget)},
			),
			tspath.NormalizePath(secondDir): rslintconfig.ConfigWithGitignore(
				entry,
				tspath.NormalizePath(secondDir),
				osvfs.FS(),
				[]string{tspath.NormalizePath(secondTarget)},
			),
		},
		Scopes: map[string]rslintconfig.LintDiscoveryScope{
			tspath.NormalizePath(firstDir):  {Files: []string{tspath.NormalizePath(firstTarget)}, ExplicitOnly: true},
			tspath.NormalizePath(secondDir): {Files: []string{tspath.NormalizePath(secondTarget)}, ExplicitOnly: true},
		},
	}

	code, stdout, stderr := runLintPipelineForTest(t, workspace, lintArgs{
		ConfigCatalog:  catalog,
		Format:         "jsonline",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 1 {
		t.Fatalf("expected the non-ignored config target to fail lint: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if strings.Contains(stdout, "packages/first/source.ts") {
		t.Fatalf("first config's gitignored target was linted: stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stdout, "packages/second/source.ts") {
		t.Fatalf("second config lost its independently lintable target: stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestCLIExplicitOnlyConfigDoesNotBlockParentGitignore(t *testing.T) {
	workspace := t.TempDir()
	ignoredDir := filepath.Join(workspace, "ignored")
	explicitTarget := filepath.Join(ignoredDir, "explicit.js")
	automaticTarget := filepath.Join(ignoredDir, "automatic.js")
	if err := os.MkdirAll(ignoredDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ignoredDir, ".gitignore"), []byte("automatic.js\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(explicitTarget, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(automaticTarget, []byte("console.log('automatic');\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	catalog := &discovery.ConfigCatalog{
		Configs: map[string]rslintconfig.RslintConfig{
			tspath.NormalizePath(workspace): rslintconfig.ConfigWithGitignore(
				rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"no-console": "error"}}},
				tspath.NormalizePath(workspace),
				osvfs.FS(),
				nil,
			),
			tspath.NormalizePath(ignoredDir): rslintconfig.ConfigWithGitignore(
				rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
				tspath.NormalizePath(ignoredDir),
				osvfs.FS(),
				[]string{tspath.NormalizePath(explicitTarget)},
			),
		},
		Scopes: map[string]rslintconfig.LintDiscoveryScope{
			tspath.NormalizePath(ignoredDir): {
				Files:        []string{tspath.NormalizePath(explicitTarget)},
				ExplicitOnly: true,
			},
		},
	}
	code, stdout, stderr := runLintPipelineForTest(t, workspace, lintArgs{
		ConfigCatalog:  catalog,
		AllowFiles:     []string{tspath.NormalizePath(explicitTarget)},
		AllowDirs:      []string{tspath.NormalizePath(ignoredDir)},
		Format:         "jsonline",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 1 {
		t.Fatalf("explicit target should fail lint: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if strings.Contains(stdout, "automatic.js") {
		t.Fatalf("parent-owned target escaped nested .gitignore: stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stdout, "explicit.js") || !strings.Contains(stdout, "no-debugger") {
		t.Fatalf("explicit-only target lost its config: stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestCLIMultiConfigGitignoreOwnershipBoundaries(t *testing.T) {
	workspace := t.TempDir()
	childDir := filepath.Join(workspace, "packages", "app")
	parentOwnedTarget := filepath.Join(childDir, "parent-owned.ts")
	parentIgnoredTarget := filepath.Join(childDir, "parent-ignored.ts")
	childOwnedTarget := filepath.Join(childDir, "child-owned.ts")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{parentOwnedTarget, parentIgnoredTarget, childOwnedTarget} {
		if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(workspace, ".gitignore"), []byte("child-owned.ts\nparent-ignored.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(childDir, ".gitignore"), []byte("parent-owned.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	entry := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		Rules: rslintconfig.Rules{"no-debugger": "error"},
	}}
	catalog := &discovery.ConfigCatalog{
		Configs: map[string]rslintconfig.RslintConfig{
			tspath.NormalizePath(workspace): rslintconfig.ConfigWithGitignoreWithBoundaries(
				entry,
				tspath.NormalizePath(workspace),
				osvfs.FS(),
				[]string{
					tspath.NormalizePath(parentOwnedTarget),
					tspath.NormalizePath(parentIgnoredTarget),
				},
				[]string{tspath.NormalizePath(childDir)},
			),
			tspath.NormalizePath(childDir): rslintconfig.ConfigWithGitignore(
				entry,
				tspath.NormalizePath(childDir),
				osvfs.FS(),
				[]string{tspath.NormalizePath(childOwnedTarget)},
			),
		},
		Scopes: map[string]rslintconfig.LintDiscoveryScope{
			tspath.NormalizePath(workspace): {
				Files: []string{
					tspath.NormalizePath(parentOwnedTarget),
					tspath.NormalizePath(parentIgnoredTarget),
				},
				ExplicitOnly: true,
			},
			tspath.NormalizePath(childDir): {
				Files:        []string{tspath.NormalizePath(childOwnedTarget)},
				ExplicitOnly: true,
			},
		},
	}
	for _, test := range []struct {
		name           string
		singleThreaded bool
	}{
		{name: "single threaded", singleThreaded: true},
		{name: "concurrent"},
	} {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := runLintPipelineForTest(t, workspace, lintArgs{
				ConfigCatalog:  catalog,
				Format:         "jsonline",
				NoColor:        true,
				SingleThreaded: test.singleThreaded,
			})
			if code != 1 {
				t.Fatalf("expected both boundary targets to fail lint: code=%d stdout=%q stderr=%q", code, stdout, stderr)
			}
			for _, targetName := range []string{"parent-owned.ts", "child-owned.ts"} {
				if !strings.Contains(stdout, targetName) {
					t.Fatalf("%s was polluted by another config's .gitignore: stdout=%q stderr=%q", targetName, stdout, stderr)
				}
			}
			if strings.Contains(stdout, "parent-ignored.ts") {
				t.Fatalf("parent-owned .gitignore did not apply before child boundary: stdout=%q stderr=%q", stdout, stderr)
			}
		})
	}
}
