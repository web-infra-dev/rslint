package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

func TestContainsGlobPattern(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "no glob characters",
			path:     "./tsconfig.json",
			expected: false,
		},
		{
			name:     "asterisk wildcard",
			path:     "./packages/*/tsconfig.json",
			expected: true,
		},
		{
			name:     "double asterisk",
			path:     "./packages/**/tsconfig.json",
			expected: true,
		},
		{
			name:     "question mark wildcard",
			path:     "./tsconfig?.json",
			expected: true,
		},
		{
			name:     "character class",
			path:     "./tsconfig[0-9].json",
			expected: true,
		},
		{
			name:     "multiple glob patterns",
			path:     "./*/test/*.json",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsGlobPattern(tt.path)
			assert.Equal(t, tt.expected, result, "containsGlobPattern(%q) should be %v", tt.path, tt.expected)
		})
	}
}

func TestLoadTsConfigsFromRslintConfig_GlobExpansion(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test directory structure
	testDirs := []string{
		"packages/ui",
		"packages/stores",
		"packages/utils",
		"apps/web",
		"apps/mobile",
	}

	for _, dir := range testDirs {
		dirPath := filepath.Join(tmpDir, dir)
		err := os.MkdirAll(dirPath, 0755)
		assert.NilError(t, err, "Failed to create directory %s", dirPath)

		// Create a tsconfig.json in each directory
		tsconfigPath := filepath.Join(dirPath, "tsconfig.json")
		err = os.WriteFile(tsconfigPath, []byte(`{"compilerOptions": {}}`), 0644)
		assert.NilError(t, err, "Failed to create tsconfig.json in %s", dirPath)
	}

	// Create rslint config in the tmp directory
	rslintConfig := RslintConfig{
		ConfigEntry{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					ProjectService: false,
					Project: ProjectPaths{
						"./packages/*/tsconfig.json",
						"./apps/*/tsconfig.json",
					},
				},
			},
		},
	}

	// Create loader with OS filesystem
	loader := NewConfigLoader(osvfs.FS(), tmpDir)

	// Test glob expansion
	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err, "LoadTsConfigsFromRslintConfig should not return error")

	// Verify we got all 5 tsconfig files
	assert.Equal(t, 5, len(tsConfigs), "Should find 5 tsconfig files")

	// Verify all expected paths are present
	expectedPaths := []string{
		filepath.Join(tmpDir, "packages/ui/tsconfig.json"),
		filepath.Join(tmpDir, "packages/stores/tsconfig.json"),
		filepath.Join(tmpDir, "packages/utils/tsconfig.json"),
		filepath.Join(tmpDir, "apps/web/tsconfig.json"),
		filepath.Join(tmpDir, "apps/mobile/tsconfig.json"),
	}

	for _, expectedPath := range expectedPaths {
		found := false
		for _, actualPath := range tsConfigs {
			if actualPath == expectedPath {
				found = true
				break
			}
		}
		assert.Assert(t, found, "Expected path %s not found in results", expectedPath)
	}
}

func TestLoadTsConfigsFromRslintConfig_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()

	// Create rslint config with glob that matches nothing
	rslintConfig := RslintConfig{
		ConfigEntry{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					ProjectService: false,
					Project: ProjectPaths{
						"./nonexistent/*/tsconfig.json",
					},
				},
			},
		},
	}

	loader := NewConfigLoader(osvfs.FS(), tmpDir)

	// Test that no matches returns an error
	_, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.ErrorContains(t, err, "glob pattern", "Should return error when glob matches no files")
}

func TestLoadTsConfigsFromRslintConfig_MixedGlobAndNonGlob(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	os.MkdirAll(filepath.Join(tmpDir, "packages/ui"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "packages/ui/tsconfig.json"), []byte(`{}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(`{}`), 0644)

	// Create config with both glob and non-glob paths
	rslintConfig := RslintConfig{
		ConfigEntry{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					ProjectService: false,
					Project: ProjectPaths{
						"./tsconfig.json",           // non-glob
						"./packages/*/tsconfig.json", // glob
					},
				},
			},
		},
	}

	loader := NewConfigLoader(osvfs.FS(), tmpDir)

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err, "Should handle mixed glob and non-glob paths")
	assert.Equal(t, 2, len(tsConfigs), "Should find 2 tsconfig files")
}

func TestLoadTsConfigsFromRslintConfig_Deduplication(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	os.MkdirAll(filepath.Join(tmpDir, "packages/ui"), 0755)
	tsconfigPath := filepath.Join(tmpDir, "packages/ui/tsconfig.json")
	os.WriteFile(tsconfigPath, []byte(`{}`), 0644)

	// Create config that would match the same file multiple times
	rslintConfig := RslintConfig{
		ConfigEntry{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					ProjectService: false,
					Project: ProjectPaths{
						"./packages/ui/tsconfig.json",  // explicit path
						"./packages/*/tsconfig.json",    // glob that also matches it
						"./packages/ui/*.json",          // another glob
					},
				},
			},
		},
	}

	loader := NewConfigLoader(osvfs.FS(), tmpDir)

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err, "Should handle duplicate paths")
	assert.Equal(t, 1, len(tsConfigs), "Should deduplicate to 1 unique file")
	assert.Equal(t, tsconfigPath, tsConfigs[0], "Should have correct deduplicated path")
}

func TestLoadTsConfigsFromRslintConfig_NonExistentNonGlobFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with non-glob path that doesn't exist
	rslintConfig := RslintConfig{
		ConfigEntry{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					ProjectService: false,
					Project: ProjectPaths{
						"./nonexistent.json",
					},
				},
			},
		},
	}

	loader := NewConfigLoader(osvfs.FS(), tmpDir)

	// Should return error for non-existent non-glob file
	_, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.ErrorContains(t, err, "doesn't exist", "Should return error for non-existent file")
}

func TestLoadTsConfigsFromRslintConfig_DoubleStarPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	testPaths := []string{
		"packages/ui/tsconfig.json",
		"packages/ui/subpackage/tsconfig.json",
		"packages/stores/tsconfig.json",
	}

	for _, testPath := range testPaths {
		fullPath := filepath.Join(tmpDir, testPath)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte(`{}`), 0644)
	}

	// Create config with ** pattern
	rslintConfig := RslintConfig{
		ConfigEntry{
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					ProjectService: false,
					Project: ProjectPaths{
						"./packages/**/tsconfig.json", // Should match nested files
					},
				},
			},
		},
	}

	loader := NewConfigLoader(osvfs.FS(), tmpDir)

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, tmpDir)
	assert.NilError(t, err, "Should handle ** patterns")
	assert.Equal(t, 3, len(tsConfigs), "Should find all nested tsconfig files")
}
