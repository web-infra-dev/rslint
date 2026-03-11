package config

import (
	"os"
	"path/filepath"
	"testing"

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, containsGlobPattern(tt.path), tt.expected)
		})
	}
}

func TestLoadTsConfigsFromRslintConfig_GlobExpansion(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFile(t, filepath.Join(tmpDir, "packages/ui/tsconfig.json"))
	createTestFile(t, filepath.Join(tmpDir, "packages/utils/tsconfig.json"))
	createTestFile(t, filepath.Join(tmpDir, "apps/web/tsconfig.json"))

	loader := NewConfigLoader(osvfs.FS(), tmpDir)
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
	loader := NewConfigLoader(osvfs.FS(), tmpDir)
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

	loader := NewConfigLoader(osvfs.FS(), tmpDir)
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

	loader := NewConfigLoader(osvfs.FS(), tmpDir)
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

	loader := NewConfigLoader(utils.NewOverlayVFS(osvfs.FS(), virtualFiles), tmpDir)
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
	loader := NewConfigLoader(osvfs.FS(), tmpDir)
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

	loader := NewConfigLoader(osvfs.FS(), tmpDir)
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

func createTestFile(t *testing.T, path string) {
	t.Helper()

	assert.NilError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NilError(t, os.WriteFile(path, []byte(`{}`), 0o644))
}
