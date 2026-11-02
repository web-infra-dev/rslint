package utils

import (
	"slices"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

type resolverSettingMarshaler struct{ calls *atomic.Int32 }

func (value resolverSettingMarshaler) MarshalJSON() ([]byte, error) {
	value.calls.Add(1)
	return []byte("null"), nil
}

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

func TestImportResolverPreparedConfiguration(t *testing.T) {
	fs := rslintUtils.NewOverlayVFS(osvfs.FS(), map[string]string{
		"/import-prepare/input.ts": "export {};", "/import-prepare/other/input.ts": "export {};",
		"/import-prepare/target.js": "export {};", "/import-prepare/target.custom": "export {};",
		"/import-prepare/only-typed.ts":   "export {};",
		"/import-prepare/other/target.js": "export {};",
		"/import-prepare/a/pkg/index.js":  "export {};", "/import-prepare/b/pkg/index.js": "export {};",
	})
	p, err := program.NewFromRoots(program.RootOptions{
		Host: rslintUtils.CreateCompilerHost("/import-prepare", fs), CompilerOptions: &core.CompilerOptions{},
		RootFileNames: []string{"/import-prepare/input.ts", "/import-prepare/other/input.ts"}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := (rule.RuleContext{SourceFile: p.GetSourceFile("/import-prepare/input.ts")}).WithProgram(p)
	other := (rule.RuleContext{SourceFile: p.GetSourceFile("/import-prepare/other/input.ts")}).WithProgram(p)
	source := modules.SourceFromSpecifier(nil)
	check := func(t *testing.T, r *ImportResolver, name, want string, found bool, wantError string) {
		t.Helper()
		path, ok, err := r.ResolveName(name, source)
		if path != want || ok != found || err != wantError {
			t.Fatalf("ResolveName(%q) = (%q, %v, %q), want (%q, %v, %q)", name, path, ok, err, want, found, wantError)
		}
	}
	t.Run("share configuration with file-specific requests", func(t *testing.T) {
		first, second := NewImportResolver(ctx), NewImportResolver(other)
		if first.importResolverConfig != second.importResolverConfig {
			t.Fatal("equal default settings rebuilt configuration")
		}
		check(t, first, "./target", "/import-prepare/target.js", true, "")
		check(t, second, "./target", "/import-prepare/other/target.js", true, "")
	})
	t.Run("unused resolver options do not enter cache identity", func(t *testing.T) {
		for _, name := range []string{"typescript", "node", "eslint-import-resolver-node", "unavailable"} {
			configured := ctx
			configured.Settings = map[string]any{"import/resolver": map[string]any{name: map[string]any{"project": "tsconfig.json"}}}
			first := NewImportResolver(configured)
			// An ignored value must not be marshaled or recursively traversed.
			configured.Settings = map[string]any{"import/resolver": map[string]any{name: map[string]any{"project": make(chan int)}}}
			second := NewImportResolver(configured)
			if first.importResolverConfig != second.importResolverConfig {
				t.Fatalf("%s cached an unused project setting", name)
			}
		}
	})
	t.Run("immutable environment reuses preparation across file passes", func(t *testing.T) {
		environment := &rule.RuleEnvironment{Settings: map[string]any{
			"import/resolver": map[string]any{"node": map[string]any{"paths": []any{"."}}},
		}}
		firstContext, secondContext := ctx, other
		firstContext.Settings, secondContext.Settings = environment.Settings, environment.Settings
		first := NewImportResolver(firstContext.WithFileCache(rule.NewFileCacheWithEnvironment("/import-prepare/a", environment)))
		second := NewImportResolver(secondContext.WithFileCache(rule.NewFileCacheWithEnvironment("/import-prepare/a", environment)))
		if first.importResolverConfig != second.importResolverConfig {
			t.Fatal("new file pass rebuilt an immutable configuration")
		}
		check(t, first, "pkg", "/import-prepare/a/pkg/index.js", true, "")
		check(t, second, "./target", "/import-prepare/other/target.js", true, "")
		changedCwd := NewImportResolver(secondContext.WithFileCache(rule.NewFileCacheWithEnvironment("/import-prepare/b", environment)))
		check(t, changedCwd, "pkg", "/import-prepare/b/pkg/index.js", true, "")
		changed := &rule.RuleEnvironment{Settings: map[string]any{
			"import/resolver": map[string]any{"node": map[string]any{"extensions": []any{}}},
		}}
		secondContext.Settings = changed.Settings
		check(t, NewImportResolver(secondContext.WithFileCache(rule.NewFileCacheWithEnvironment("/import-prepare/a", changed))), "./target", "", false, "")
	})
	t.Run("warm environment does not encode settings again", func(t *testing.T) {
		var calls atomic.Int32
		environment := &rule.RuleEnvironment{Settings: map[string]any{
			"import/core-modules": resolverSettingMarshaler{&calls},
		}}
		firstContext, secondContext := ctx, other
		firstContext.Settings, secondContext.Settings = environment.Settings, environment.Settings
		first := NewImportResolver(firstContext.WithFileCache(rule.NewFileCacheWithEnvironment("/import-prepare", environment)))
		second := NewImportResolver(secondContext.WithFileCache(rule.NewFileCacheWithEnvironment("/import-prepare", environment)))
		if calls.Load() != 1 || first.importResolverConfig != second.importResolverConfig {
			t.Fatalf("warm configuration was prepared again: encodings=%d", calls.Load())
		}
		check(t, second, "./target", "/import-prepare/other/target.js", true, "")
	})
	t.Run("snapshot settings and separate changed configuration", func(t *testing.T) {
		extensions, core := []string{".custom"}, []string{"virtual"}
		configured := ctx
		configured.Settings = map[string]any{"import/resolver": map[string]any{"node": map[string]any{"extensions": extensions}}, "import/core-modules": core}
		first := NewImportResolver(configured)
		extensions[0], core[0] = ".js", "other"
		second := NewImportResolver(configured)
		if first.importResolverConfig == second.importResolverConfig {
			t.Fatal("mutated settings reused old configuration")
		}
		check(t, first, "./target", "/import-prepare/target.custom", true, "")
		check(t, second, "./target", "/import-prepare/target.js", true, "")
		check(t, first, "virtual", "", true, "")
		check(t, second, "virtual", "", false, "")
	})
	t.Run("working directory belongs to configuration identity", func(t *testing.T) {
		configured := ctx
		configured.Settings = map[string]any{"import/resolver": map[string]any{"node": map[string]any{"paths": []any{"."}}}}
		first := NewImportResolver(configured.WithFileCache(rule.NewFileCacheWithProcessCurrentDirectory("/import-prepare/a")))
		second := NewImportResolver(configured.WithFileCache(rule.NewFileCacheWithProcessCurrentDirectory("/import-prepare/b")))
		check(t, first, "pkg", "/import-prepare/a/pkg/index.js", true, "")
		check(t, second, "pkg", "/import-prepare/b/pkg/index.js", true, "")
	})
	t.Run("retain typed settings interpretation", func(t *testing.T) {
		configured := ctx
		configured.Settings = map[string]any{"import/resolver": []any{"node"}}
		check(t, NewImportResolver(configured), "./target", "/import-prepare/target.js", true, "")
		configured.Settings["import/resolver"] = []string{"node"}
		check(t, NewImportResolver(configured), "./target", "", false, "invalid resolver config")
		configured.Settings["import/resolver"] = map[string]any{"node": map[string]any{"extensions": []string{}}}
		check(t, NewImportResolver(configured), "./target", "", false, "")
		configured.Settings["import/resolver"] = map[string]any{"node": map[string][]string{"extensions": {}}}
		check(t, NewImportResolver(configured), "./target", "/import-prepare/target.js", true, "")
	})
	t.Run("nil and empty extensions stay distinct", func(t *testing.T) {
		configured := ctx
		configured.Settings = map[string]any{"import/resolver": map[string]any{"node": map[string]any{"extensions": nil}}}
		check(t, NewImportResolver(configured), "./target", "/import-prepare/target.js", true, "")
		configured.Settings["import/resolver"] = map[string]any{"node": map[string]any{"extensions": []any{}}}
		check(t, NewImportResolver(configured), "./target", "", false, "")
	})
	t.Run("typed nil resolver differs from default", func(t *testing.T) {
		configured := ctx
		configured.Settings = map[string]any{"import/resolver": []any(nil)}
		check(t, NewImportResolver(configured), "./target", "", false, "")
		configured.Settings["import/resolver"] = map[string]any(nil)
		check(t, NewImportResolver(configured), "./target", "", false, "")
		configured.Settings["import/resolver"] = nil
		check(t, NewImportResolver(configured), "./target", "/import-prepare/target.js", true, "")
	})
	t.Run("typed nil legacy config retains node-only resolution", func(t *testing.T) {
		configured := ctx
		configured.Settings = map[string]any{"import/resolve": map[string]any(nil)}
		check(t, NewImportResolver(configured), "./only-typed", "", false, "")
		configured.Settings["import/resolve"] = nil
		check(t, NewImportResolver(configured), "./only-typed", "/import-prepare/only-typed.ts", true, "")
	})
	t.Run("JSON replacement characters do not merge core settings", func(t *testing.T) {
		configured := ctx
		configured.Settings = map[string]any{"import/core-modules": []string{"\xff"}}
		check(t, NewImportResolver(configured), "\xff", "", true, "")
		configured.Settings["import/core-modules"] = []string{"\xfe"}
		check(t, NewImportResolver(configured), "\xff", "", false, "")
	})
}
