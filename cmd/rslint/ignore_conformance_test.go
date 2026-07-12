package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	api "github.com/web-infra-dev/rslint/internal/api"
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
			name:      "ancestor ignore blocks nested source",
			configDir: "packages/app",
			relative:  "ignored/keep.ts",
			gitignores: map[string]string{
				".gitignore":                      "/packages/app/ignored/\n",
				"packages/app/ignored/.gitignore": "!keep.ts\n",
			},
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

	entry := map[string]any{
		"files": []string{"**/*.ts"},
		"rules": map[string]string{"no-debugger": "error"},
	}
	payload, err := json.Marshal(map[string]any{
		"configs": []any{
			map[string]any{
				"configDirectory": firstDir,
				"entries":         []any{entry},
				"targetFiles":     []string{firstTarget},
				"explicitOnly":    true,
			},
			map[string]any{
				"configDirectory": secondDir,
				"entries":         []any{entry},
				"targetFiles":     []string{secondTarget},
				"explicitOnly":    true,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	payloadPath := filepath.Join(workspace, "config-payload.json")
	if err := os.WriteFile(payloadPath, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	stdin, err := os.Open(payloadPath)
	if err != nil {
		t.Fatal(err)
	}
	originalStdin := os.Stdin
	os.Stdin = stdin
	t.Cleanup(func() {
		os.Stdin = originalStdin
		_ = stdin.Close()
	})

	code, stdout, stderr := runLintPipelineForTest(t, workspace, lintArgs{
		ConfigStdin:    true,
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
