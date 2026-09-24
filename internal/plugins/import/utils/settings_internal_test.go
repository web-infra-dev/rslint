package utils

import (
	"slices"
	"testing"
)

// TestModuleSettingsKeySeparatesListElements locks in that the cache key keeps
// list boundaries: two ignore lists whose elements merely concatenate alike
// must not share one compiled-settings cache entry.
func TestModuleSettingsKeySeparatesListElements(t *testing.T) {
	t.Parallel()

	joined := moduleSettingsKey(map[string]interface{}{"import/ignore": []interface{}{"foo bar"}})
	split := moduleSettingsKey(map[string]interface{}{"import/ignore": []interface{}{"foo", "bar"}})
	if joined == split {
		t.Fatalf("moduleSettingsKey collided on %q", joined)
	}

	typed := moduleSettingsKey(map[string]interface{}{"import/ignore": []string{"foo", "bar"}})
	if typed != split {
		t.Fatalf("moduleSettingsKey separated equal settings: %q vs %q", typed, split)
	}
}

func TestCompiledModuleSettingsCacheKeyIncludesInternalRegex(t *testing.T) {
	t.Parallel()

	left := map[string]interface{}{"import/internal-regex": "^@left/"}
	right := map[string]interface{}{"import/internal-regex": "^@right/"}
	if moduleSettingsKey(left) != moduleSettingsKey(right) {
		t.Fatal("module-index key should ignore classification-only settings")
	}
	if compiledModuleSettingsCacheKey(left) == compiledModuleSettingsCacheKey(right) {
		t.Fatal("compiled settings cache reused a different internal regexp")
	}
}

func TestCompiledModuleSettingsCacheKeyIncludesCoreModules(t *testing.T) {
	t.Parallel()

	left := map[string]interface{}{"import/core-modules": []interface{}{"left"}}
	right := map[string]interface{}{"import/core-modules": []interface{}{"right"}}
	if moduleSettingsKey(left) != moduleSettingsKey(right) {
		t.Fatal("module-index key should ignore classification-only settings")
	}
	if compiledModuleSettingsCacheKey(left) == compiledModuleSettingsCacheKey(right) {
		t.Fatal("compiled settings cache reused different core modules")
	}
}

func TestModuleSettingsKeySeparatesDefaultAndExplicitEmptyFolders(t *testing.T) {
	t.Parallel()

	defaultKey := moduleSettingsKey(nil)
	emptyStringsKey := moduleSettingsKey(map[string]interface{}{
		"import/external-module-folders": []string{},
	})
	emptyInterfacesKey := moduleSettingsKey(map[string]interface{}{
		"import/external-module-folders": []interface{}{},
	})
	if defaultKey == emptyStringsKey {
		t.Fatal("module settings cache merged the default folders with an explicit empty list")
	}
	if emptyStringsKey != emptyInterfacesKey {
		t.Fatal("equivalent typed empty folder lists produced different cache keys")
	}
}

func TestFileExtensions(t *testing.T) {
	for _, tc := range []struct {
		settings map[string]any
		want     []string
	}{
		{nil, []string{".js", ".mjs", ".cjs"}},
		{map[string]any{"import/extensions": []any{}}, nil},
		{map[string]any{"import/extensions": []any{".js", ".ts", ".js"}}, []string{".js", ".ts"}},
		{map[string]any{"import/extensions": []any{".tsx"}, "import/parsers": map[string]any{"typescript": []any{".ts", ".tsx"}}}, []string{".tsx", ".ts"}},
	} {
		if got := FileExtensions(tc.settings); !slices.Equal(got, tc.want) {
			t.Fatalf("FileExtensions(%v) = %v, want %v", tc.settings, got, tc.want)
		}
	}
	extensions := make([]string, 1, 2)
	extensions[0] = ".js"
	settings := map[string]any{"import/extensions": extensions, "import/parsers": map[string]any{"typescript": []string{".ts"}}}
	FileExtensions(settings)
	if extensions[:2][1] != "" {
		t.Fatal("extension settings were mutated")
	}
}
