package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"gotest.tools/v3/assert"
)

func TestContainsGlobPattern(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{name: "plain file", path: "./tsconfig.json", expected: false},
		{name: "single wildcard", path: "./packages/*/tsconfig.json", expected: true},
		{name: "recursive wildcard", path: "./packages/**/tsconfig.json", expected: true},
		{name: "question wildcard", path: "./tsconfig?.json", expected: true},
		{name: "character class", path: "./tsconfig[0-9].json", expected: true},
		{name: "negated character class", path: "./tsconfig[!a].json", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, containsGlobPattern(tt.path), tt.expected)
		})
	}
}

func TestRelativeGlobPatternPreservesFilesystemRoots(t *testing.T) {
	tests := []struct {
		root     string
		pattern  string
		expected string
	}{
		{root: "/", pattern: "/*/tsconfig.json", expected: "*/tsconfig.json"},
		{root: "/repo", pattern: "/repo/packages/*/tsconfig.json", expected: "packages/*/tsconfig.json"},
		{root: "C:/", pattern: "C:/*/tsconfig.json", expected: "*/tsconfig.json"},
		{root: "//server/share/", pattern: "//server/share/*/tsconfig.json", expected: "*/tsconfig.json"},
	}
	for _, tt := range tests {
		assert.Equal(t, relativeGlobPattern(tt.root, tt.pattern), tt.expected)
	}
}

func TestNewConfigLoaderRequiresRuleCatalog(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil rule catalog to panic")
		}
	}()
	NewConfigLoader(osvfs.FS(), t.TempDir(), nil)
}

func TestLoadRslintConfig_RejectsEmptyFilesArray(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "rslint.jsonc")
	if err := os.WriteFile(configPath, []byte(`[
		{
			"files": [],
			"rules": { "no-console": "error" }
		}
	]`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	_, _, err := loader.LoadRslintConfig("rslint.jsonc")
	assert.ErrorContains(t, err, `key "files": expected value to be a non-empty array`)
}

func TestLoadRslintConfig_AllowsMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "rslint.jsonc")
	if err := os.WriteFile(configPath, []byte(`[
		{
			"rules": { "no-console": "error" }
		}
	]`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	cfg, _, err := loader.LoadRslintConfig("rslint.jsonc")
	assert.NilError(t, err)
	assert.Equal(t, len(cfg), 1)
	assert.Assert(t, cfg[0].Files == nil)
}

func TestLoadTsConfigsFromRslintConfig_GlobExpansion(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "packages/ui/tsconfig.json"))
	createTestFile(t, filepath.Join(tmpDir, "packages/utils/tsconfig.json"))
	createTestFile(t, filepath.Join(tmpDir, "apps/web/tsconfig.json"))

	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{
						"./packages/*/tsconfig.json",
						"./apps/*/tsconfig.json",
					},
				},
			},
		},
	}

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err)
	assert.DeepEqual(t, tsConfigs, []string{
		filepath.ToSlash(filepath.Join(tmpDir, "packages/ui/tsconfig.json")),
		filepath.ToSlash(filepath.Join(tmpDir, "packages/utils/tsconfig.json")),
		filepath.ToSlash(filepath.Join(tmpDir, "apps/web/tsconfig.json")),
	})
}

func TestLoadTsConfigsFromRslintConfig_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{"./packages/*/tsconfig.json"},
				},
			},
		},
	}

	_, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.ErrorContains(t, err, "glob pattern")
}

func TestLoadTsConfigsFromRslintConfig_MixedGlobAndNonGlob(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "tsconfig.json"))
	createTestFile(t, filepath.Join(tmpDir, "packages/ui/tsconfig.json"))

	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{
						"./tsconfig.json",
						"./packages/*/tsconfig.json",
					},
				},
			},
		},
	}

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err)
	assert.DeepEqual(t, tsConfigs, []string{
		filepath.ToSlash(filepath.Join(tmpDir, "tsconfig.json")),
		filepath.ToSlash(filepath.Join(tmpDir, "packages/ui/tsconfig.json")),
	})
}

func TestLoadTsConfigsFromRslintConfig_DeduplicatesMatches(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "packages/ui/tsconfig.json"))

	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{
						"./packages/ui/tsconfig.json",
						"./packages/*/tsconfig.json",
						"./packages/ui/*.json",
					},
				},
			},
		},
	}

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err)
	assert.DeepEqual(t, tsConfigs, []string{
		filepath.ToSlash(filepath.Join(tmpDir, "packages/ui/tsconfig.json")),
	})
}

func TestLoadTsConfigsFromRslintConfig_GlobExpansionWithOverlayVFS(t *testing.T) {
	tmpDir := t.TempDir()
	virtualFiles := map[string]string{
		filepath.ToSlash(filepath.Join(tmpDir, "packages/ui/tsconfig.json")):    `{}`,
		filepath.ToSlash(filepath.Join(tmpDir, "packages/utils/tsconfig.json")): `{}`,
	}

	loader := NewConfigLoader(utils.NewOverlayVFS(osvfs.FS(), virtualFiles), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{"./packages/*/tsconfig.json"},
				},
			},
		},
	}

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err)
	assert.DeepEqual(t, tsConfigs, []string{
		filepath.ToSlash(filepath.Join(tmpDir, "packages/ui/tsconfig.json")),
		filepath.ToSlash(filepath.Join(tmpDir, "packages/utils/tsconfig.json")),
	})
}

func TestLoadTsConfigsFromRslintConfig_NonExistentNonGlobFile(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{"./missing.json"},
				},
			},
		},
	}

	_, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.ErrorContains(t, err, "doesn't exist")
}

func TestLoadTsConfigsFromRslintConfig_DoubleStarPattern(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "packages/ui/tsconfig.json"))
	createTestFile(t, filepath.Join(tmpDir, "packages/ui/subpackage/tsconfig.json"))
	createTestFile(t, filepath.Join(tmpDir, "packages/stores/tsconfig.json"))

	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{"./packages/**/tsconfig.json"},
				},
			},
		},
	}

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err)
	assert.DeepEqual(t, tsConfigs, []string{
		filepath.ToSlash(filepath.Join(tmpDir, "packages/stores/tsconfig.json")),
		filepath.ToSlash(filepath.Join(tmpDir, "packages/ui/subpackage/tsconfig.json")),
		filepath.ToSlash(filepath.Join(tmpDir, "packages/ui/tsconfig.json")),
	})
}

func TestLoadTsConfigsFromRslintConfig_SingleStarDoesNotMatchNested(t *testing.T) {
	tmpDir := t.TempDir()
	// Direct child — should match
	createTestFile(t, filepath.Join(tmpDir, "packages/ui/tsconfig.json"))
	// Nested deeper — should NOT match with single *
	createTestFile(t, filepath.Join(tmpDir, "packages/ui/node_modules/foo/tsconfig.json"))
	createTestFile(t, filepath.Join(tmpDir, "packages/ui/src/tsconfig.json"))

	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{"./packages/*/tsconfig.json"},
				},
			},
		},
	}

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err)
	assert.DeepEqual(t, tsConfigs, []string{
		filepath.ToSlash(filepath.Join(tmpDir, "packages/ui/tsconfig.json")),
	})
}

func TestLoadTsConfigsFromRslintConfig_NonExistentSearchRoot(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{"./nonexistent/*/tsconfig.json"},
				},
			},
		},
	}

	_, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.ErrorContains(t, err, "glob pattern")
}

func TestLoadTsConfigsFromRslintConfig_DoubleStarWithSymlinkCycle(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "packages/ui/tsconfig.json"))
	// Create symlink cycle: packages/ui/loop -> packages
	assert.NilError(t, os.Symlink(
		filepath.Join(tmpDir, "packages"),
		filepath.Join(tmpDir, "packages/ui/loop"),
	))

	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{"./packages/**/tsconfig.json"},
				},
			},
		},
	}

	// Should complete without hanging, finding only the real tsconfig
	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err)
	assert.Assert(t, len(tsConfigs) >= 1, "should find at least the real tsconfig.json")

	found := false
	for _, c := range tsConfigs {
		if filepath.Base(filepath.Dir(c)) == "ui" {
			found = true
		}
	}
	assert.Assert(t, found, "should find packages/ui/tsconfig.json")
}

func TestLoadTsConfigsFromRslintConfig_QuestionMarkPattern(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "packages/a/tsconfig.json"))
	createTestFile(t, filepath.Join(tmpDir, "packages/b/tsconfig.json"))
	// "ab" is two chars — should NOT match single ?
	createTestFile(t, filepath.Join(tmpDir, "packages/ab/tsconfig.json"))

	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{"./packages/?/tsconfig.json"},
				},
			},
		},
	}

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err)
	assert.DeepEqual(t, tsConfigs, []string{
		filepath.ToSlash(filepath.Join(tmpDir, "packages/a/tsconfig.json")),
		filepath.ToSlash(filepath.Join(tmpDir, "packages/b/tsconfig.json")),
	})
}

func TestLoadTsConfigsFromRslintConfig_CharacterClassPattern(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "tsconfig1.json"))
	createTestFile(t, filepath.Join(tmpDir, "tsconfig2.json"))
	// "a" is not in [0-9] — should NOT match
	createTestFile(t, filepath.Join(tmpDir, "tsconfiga.json"))

	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: ProjectPaths{"./tsconfig[0-9].json"},
				},
			},
		},
	}

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err)
	assert.DeepEqual(t, tsConfigs, []string{
		filepath.ToSlash(filepath.Join(tmpDir, "tsconfig1.json")),
		filepath.ToSlash(filepath.Join(tmpDir, "tsconfig2.json")),
	})
}

func TestLoadTsConfigsFromRslintConfig_NegatedCharacterClass(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "packages/a/tsconfig.json"))
	createTestFile(t, filepath.Join(tmpDir, "packages/b/tsconfig.json"))
	createTestFile(t, filepath.Join(tmpDir, "packages/c/tsconfig.json"))

	loader := NewConfigLoader(osvfs.FS(), tmpDir, nativeRuleCatalog())
	rslintConfig := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					// [!a] matches any single char except "a"
					Project: ProjectPaths{"./packages/[!a]/tsconfig.json"},
				},
			},
		},
	}

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err)
	assert.DeepEqual(t, tsConfigs, []string{
		filepath.ToSlash(filepath.Join(tmpDir, "packages/b/tsconfig.json")),
		filepath.ToSlash(filepath.Join(tmpDir, "packages/c/tsconfig.json")),
	})
}

func TestGlobSearchRoot(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		fallback string
		expected string
	}{
		{name: "no glob", pattern: "/a/b/c", fallback: "/fallback", expected: "/a/b/c"},
		{name: "star after slash", pattern: "/a/b/*/c.json", fallback: "/fallback", expected: "/a/b"},
		{name: "doublestar after slash", pattern: "/a/b/**/c.json", fallback: "/fallback", expected: "/a/b"},
		{name: "star at start", pattern: "*/c.json", fallback: "/fallback", expected: "/fallback"},
		{name: "star mid segment", pattern: "/a/b/ts*.json", fallback: "/fallback", expected: "/a/b"},
		{name: "question mark", pattern: "/a/b/?.json", fallback: "/fallback", expected: "/a/b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, globSearchRoot(tt.pattern, tt.fallback), tt.expected)
		})
	}
}

func createTestFile(t *testing.T, path string) {
	t.Helper()

	assert.NilError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NilError(t, os.WriteFile(path, []byte(`{}`), 0o644))
}

func TestResolveTsConfigPaths_ReturnsConfiguredPaths(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "packages/foo/tsconfig.json"))

	cfg := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: []string{"./packages/foo/tsconfig.json"},
				},
			},
		},
	}

	paths, err := ResolveTsConfigPaths(cfg, tmpDir, osvfs.FS())
	assert.NilError(t, err)
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}
	expected := filepath.ToSlash(filepath.Join(tmpDir, "packages/foo/tsconfig.json"))
	if paths[0] != expected {
		t.Errorf("expected path %q, got %q", expected, paths[0])
	}
}

func TestResolveTsConfigPaths_FallbackToDefaultTsConfig(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "tsconfig.json"))

	// No parserOptions.project → auto-detect tsconfig.json
	cfg := RslintConfig{{}}

	paths, err := ResolveTsConfigPaths(cfg, tmpDir, osvfs.FS())
	assert.NilError(t, err)
	if len(paths) != 1 {
		t.Fatalf("expected 1 path (auto-detected), got %d", len(paths))
	}
	expected := filepath.ToSlash(filepath.Join(tmpDir, "tsconfig.json"))
	if paths[0] != expected {
		t.Errorf("expected auto-detected path %q, got %q", expected, paths[0])
	}
}

func TestResolveTsConfigPaths_ExplicitEmptyProjectDisablesFallback(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "tsconfig.json"))
	cfg := RslintConfig{{
		LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{
			Project: ProjectPaths{},
		}},
	}}

	paths, err := ResolveTsConfigPaths(cfg, tmpDir, osvfs.FS())
	assert.NilError(t, err)
	if paths != nil {
		t.Fatalf("explicit project: [] selected an implicit tsconfig: %v", paths)
	}
}

func TestResolveTsConfigPaths_NilFS(t *testing.T) {
	paths, err := ResolveTsConfigPaths(RslintConfig{{}}, "/any", nil)
	assert.NilError(t, err)
	if paths != nil {
		t.Errorf("expected nil paths for nil FS, got %v", paths)
	}
}

func TestResolveTsConfigPaths_ErrorOnNonExistentTsConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := RslintConfig{
		{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					Project: []string{"./nonexistent/tsconfig.json"},
				},
			},
		},
	}

	_, err := ResolveTsConfigPaths(cfg, tmpDir, osvfs.FS())
	if err == nil {
		t.Error("expected error for non-existent tsconfig")
	}
}

func TestAuthoredPathBaseAppliesToFilesIgnoresAndProjects(t *testing.T) {
	root := t.TempDir()
	configDirectory := filepath.Join(root, "config")
	authoredDirectory := filepath.Join(root, "workspace")
	target := filepath.Join(authoredDirectory, "src", "index.ts")
	ignored := filepath.Join(authoredDirectory, "src", "ignored.ts")
	ownerProject := filepath.Join(configDirectory, "tsconfig.json")
	authoredProject := filepath.Join(authoredDirectory, "tsconfig.json")
	for _, path := range []string{target, ignored, ownerProject, authoredProject} {
		createTestFile(t, path)
	}

	base := RslintConfig{{
		LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{
			Project: ProjectPaths{"./tsconfig.json"},
		}},
	}}
	override := RslintConfig{
		{Ignores: []string{"src/ignored.ts"}},
		{
			Files: []string{"src/**/*.ts"},
			LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{
				Project: ProjectPaths{"./tsconfig.json"},
			}},
			Rules: Rules{"no-debugger": "error"},
		},
	}
	composed := append(base, ConfigWithAuthoredPathBase(override, authoredDirectory)...)
	if override[0].authoredPathBase != nil || override[1].authoredPathBase != nil {
		t.Fatal("ConfigWithAuthoredPathBase mutated its input")
	}
	composed, optionErrors := ValidateRuleOptions(composed, nativeRuleCatalog())
	if len(optionErrors) != 0 {
		t.Fatalf("ValidateRuleOptions returned errors: %v", optionErrors)
	}

	merged := composed.GetConfigForFile(target, configDirectory)
	if merged == nil || merged.Rules["no-debugger"] == nil {
		t.Fatalf("authored files base did not select target: %+v", merged)
	}
	if !composed.IsFileIgnored(ignored, configDirectory) {
		t.Fatal("authored ignores base did not exclude target")
	}

	projects, err := ResolveTsConfigPaths(composed, configDirectory, osvfs.FS())
	assert.NilError(t, err)
	wantProjects := []string{
		tspath.NormalizePath(ownerProject),
		tspath.NormalizePath(authoredProject),
	}
	assert.DeepEqual(t, projects, wantProjects)
}

func TestResolveTsConfigPathsDefaultRemainsAtOwningConfigBase(t *testing.T) {
	root := t.TempDir()
	configDirectory := filepath.Join(root, "config")
	authoredDirectory := filepath.Join(root, "workspace")
	ownerProject := filepath.Join(configDirectory, "tsconfig.json")
	for _, path := range []string{
		ownerProject,
		filepath.Join(authoredDirectory, "tsconfig.json"),
	} {
		createTestFile(t, path)
	}

	config := ConfigWithAuthoredPathBase(RslintConfig{{
		Rules: Rules{"no-debugger": "error"},
	}}, authoredDirectory)
	projects, err := ResolveTsConfigPaths(config, configDirectory, osvfs.FS())
	assert.NilError(t, err)
	assert.DeepEqual(t, projects, []string{tspath.NormalizePath(ownerProject)})
}
