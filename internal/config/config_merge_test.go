package config

import (
	"encoding/json"
	"testing"
)

func TestGetConfigForFile_ExplicitRulesOnly(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": "error",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil merged config")
		return
	}

	// Only explicitly listed rules should be present
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger rule to be present")
	}
	if len(merged.Rules) != 1 {
		t.Errorf("Expected exactly 1 rule, got %d", len(merged.Rules))
	}
}

func TestGetConfigForFile_WithoutNormalize_PluginDoesNotAutoEnable(t *testing.T) {
	RegisterAllRules()

	// Without normalizeJSONConfig, plugins should not auto-enable rules
	config := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil merged config")
		return
	}

	// No rules should be enabled (JS config behavior)
	if len(merged.Rules) != 0 {
		t.Errorf("Expected 0 rules without normalization, got %d", len(merged.Rules))
	}
}

func TestGetConfigForFile_GlobalIgnores(t *testing.T) {
	config := RslintConfig{
		{
			Ignores: []string{"dist/**"},
		},
		{
			Rules: Rules{"no-debugger": "error"},
		},
	}

	// File in dist should be ignored
	merged := config.GetConfigForFile("dist/bundle.js", "")
	if merged != nil {
		t.Error("Expected nil for globally ignored file")
	}

	// File not in dist should not be ignored
	merged = config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil for non-ignored file")
		return
	}
}

func TestGetConfigForFile_EntryIgnores_NoMatch(t *testing.T) {
	config := RslintConfig{
		{
			Files:   []string{"**/*.ts"},
			Ignores: []string{"**/*.test.ts"},
			Rules:   Rules{"no-debugger": "error"},
		},
	}

	// Test file is ignored by entry-level ignores and no other entry matches
	// Should return nil (file should not be linted)
	merged := config.GetConfigForFile("src/app.test.ts", "")
	if merged != nil {
		t.Error("Expected nil for file ignored by all entries")
	}
}

func TestGetConfigForFile_EntryIgnores_OtherEntryMatches(t *testing.T) {
	config := RslintConfig{
		{
			Files:   []string{"**/*.ts"},
			Ignores: []string{"**/*.test.ts"},
			Rules:   Rules{"no-debugger": "error"},
		},
		{
			Files: []string{"**/*.test.ts"},
			Rules: Rules{"no-console": "warn"},
		},
	}

	// Test file is ignored by first entry but matched by second
	merged := config.GetConfigForFile("src/app.test.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config (matched by second entry)")
		return
	}
	if _, ok := merged.Rules["no-debugger"]; ok {
		t.Error("Expected no-debugger to not be present (from ignored entry)")
	}
	if _, ok := merged.Rules["no-console"]; !ok {
		t.Error("Expected no-console from second entry")
	}
}

func TestGetConfigForFile_FilesMatching(t *testing.T) {
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-debugger": "error"},
		},
	}

	// TS file should match
	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger for matching .ts file")
	}

	// JS file should not match — no entry matches, return nil
	merged = config.GetConfigForFile("src/app.js", "")
	if merged != nil {
		t.Error("Expected nil for non-matching file with no other entries")
	}
}

func TestGetConfigForFile_RulesShallowMerge(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": "error",
				"no-console":  "error",
			},
		},
		{
			Rules: Rules{
				"no-debugger":   "warn",
				"for-direction": "error",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	// no-debugger should be overridden to "warn"
	if merged.Rules["no-debugger"].Level != "warn" {
		t.Errorf("Expected no-debugger to be 'warn', got %q", merged.Rules["no-debugger"].Level)
	}
	// no-console should remain
	if merged.Rules["no-console"].Level != "error" {
		t.Errorf("Expected no-console to be 'error', got %q", merged.Rules["no-console"].Level)
	}
	// for-direction should be added
	if merged.Rules["for-direction"].Level != "error" {
		t.Errorf("Expected for-direction to be 'error', got %q", merged.Rules["for-direction"].Level)
	}
}

func TestGetConfigForFile_SettingsShallowMerge(t *testing.T) {
	config := RslintConfig{
		{
			Settings: Settings{
				"importResolver": "node",
				"react":          "17",
			},
		},
		{
			Settings: Settings{
				"react": "18",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	if merged.Settings["importResolver"] != "node" {
		t.Errorf("Expected importResolver to be 'node', got %v", merged.Settings["importResolver"])
	}
	if merged.Settings["react"] != "18" {
		t.Errorf("Expected react to be '18' (overridden), got %v", merged.Settings["react"])
	}
}

func TestMergeLanguageOptions(t *testing.T) {
	t.Run("nil override returns base", func(t *testing.T) {
		base := &LanguageOptions{
			ParserOptions: &ParserOptions{
				ProjectService: BoolPtr(true),
			},
		}
		result := mergeLanguageOptions(base, nil)
		if result != base {
			t.Error("Expected base to be returned when override is nil")
		}
	})

	t.Run("nil base returns override", func(t *testing.T) {
		override := &LanguageOptions{
			ParserOptions: &ParserOptions{
				ProjectService: BoolPtr(true),
			},
		}
		result := mergeLanguageOptions(nil, override)
		if result != override {
			t.Error("Expected override to be returned when base is nil")
		}
	})

	t.Run("deep merge parserOptions", func(t *testing.T) {
		base := &LanguageOptions{
			ParserOptions: &ParserOptions{
				ProjectService: BoolPtr(true),
			},
		}
		override := &LanguageOptions{
			ParserOptions: &ParserOptions{
				ProjectService: BoolPtr(false),
				Project:        ProjectPaths{"./tsconfig.json"},
			},
		}
		result := mergeLanguageOptions(base, override)

		if result.ParserOptions.ProjectService == nil || *result.ParserOptions.ProjectService != false {
			t.Error("Expected ProjectService to be overridden to false")
		}
		if len(result.ParserOptions.Project) != 1 || result.ParserOptions.Project[0] != "./tsconfig.json" {
			t.Error("Expected Project to be set from override")
		}
	})

	t.Run("nil ProjectService in override preserves base", func(t *testing.T) {
		base := &LanguageOptions{
			ParserOptions: &ParserOptions{
				ProjectService: BoolPtr(true),
			},
		}
		override := &LanguageOptions{
			ParserOptions: &ParserOptions{
				Project: ProjectPaths{"./tsconfig.json"},
			},
		}
		result := mergeLanguageOptions(base, override)

		if result.ParserOptions.ProjectService == nil || *result.ParserOptions.ProjectService != true {
			t.Error("Expected ProjectService to be preserved from base")
		}
	})

	// Opaque-compat merge pinning. Verifies that arbitrary ESLint
	// flat-config fields stashed in `Compat` deep-merge the same way
	// the pre-refactor typed Globals/EcmaVersion/etc. did — without
	// Go needing to know individual field names.

	t.Run("opaque compat: globals deep-merge across entries", func(t *testing.T) {
		base := &LanguageOptions{
			Compat: map[string]any{
				"globals": map[string]any{"Foo": "readonly", "Bar": "readonly"},
			},
		}
		override := &LanguageOptions{
			Compat: map[string]any{
				"globals": map[string]any{"Bar": "writable", "Baz": "readonly"},
			},
		}
		result := mergeLanguageOptions(base, override)
		g := result.Compat["globals"].(map[string]any)
		// Base-only key survives.
		if g["Foo"] != "readonly" {
			t.Errorf("expected Foo readonly survives, got %v", g["Foo"])
		}
		// Conflict: override wins.
		if g["Bar"] != "writable" {
			t.Errorf("expected Bar overridden to writable, got %v", g["Bar"])
		}
		// Override-only key included.
		if g["Baz"] != "readonly" {
			t.Errorf("expected Baz readonly, got %v", g["Baz"])
		}
	})

	t.Run("opaque compat: parserOptions deep-merge under parserOptions key", func(t *testing.T) {
		// Even though parserOptions.{Project,ProjectService} are typed
		// fields, OTHER parserOptions fields (ecmaVersion, sourceType,
		// ecmaFeatures, ...) flow through ParserOptions.Compat. Their
		// merge happens via the same deepMergeMap.
		base := &LanguageOptions{
			ParserOptions: &ParserOptions{
				Compat: map[string]any{
					"ecmaVersion":  2020,
					"sourceType":   "module",
					"ecmaFeatures": map[string]any{"jsx": true},
				},
			},
		}
		override := &LanguageOptions{
			ParserOptions: &ParserOptions{
				Compat: map[string]any{
					"ecmaVersion":  2024,
					"ecmaFeatures": map[string]any{"globalReturn": true},
				},
			},
		}
		result := mergeLanguageOptions(base, override)
		po := result.ParserOptions
		// Scalar override wins.
		if po.Compat["ecmaVersion"] != 2024 {
			t.Errorf("expected ecmaVersion=2024, got %v", po.Compat["ecmaVersion"])
		}
		// Base-only key survives (override didn't declare).
		if po.Compat["sourceType"] != "module" {
			t.Errorf("expected sourceType=module, got %v", po.Compat["sourceType"])
		}
		// Nested map deep-merge.
		feats := po.Compat["ecmaFeatures"].(map[string]any)
		if feats["jsx"] != true {
			t.Errorf("expected ecmaFeatures.jsx=true survives, got %v", feats["jsx"])
		}
		if feats["globalReturn"] != true {
			t.Errorf("expected ecmaFeatures.globalReturn=true, got %v", feats["globalReturn"])
		}
	})

	t.Run("opaque compat: unknown future field flows through unchanged", func(t *testing.T) {
		// The whole point of the refactor: a brand-new ESLint flat-config
		// field that Go knows NOTHING about still merges correctly.
		base := &LanguageOptions{
			Compat: map[string]any{"newField": map[string]any{"a": 1, "b": 2}},
		}
		override := &LanguageOptions{
			Compat: map[string]any{"newField": map[string]any{"b": 99, "c": 3}},
		}
		result := mergeLanguageOptions(base, override)
		nf := result.Compat["newField"].(map[string]any)
		if nf["a"] != 1 || nf["b"] != 99 || nf["c"] != 3 {
			t.Errorf("unexpected merged shape: %v", nf)
		}
	})
}

// JSON round-trip pinning. The typed+opaque split MUST be transparent
// to JSON in both directions — what the user wrote == what comes out of
// MarshalJSON (modulo key order). And `parserOptions.project` /
// `projectService` lift into typed fields while everything else lands in
// Compat.
func TestLanguageOptionsJSONRoundTrip(t *testing.T) {
	t.Run("unmarshal splits typed from compat", func(t *testing.T) {
		input := []byte(`{
			"parserOptions": {
				"project": ["./tsconfig.json"],
				"projectService": true,
				"ecmaVersion": 2024,
				"sourceType": "module",
				"ecmaFeatures": {"jsx": true}
			},
			"globals": {"Foo": "readonly"},
			"futureField": "anything"
		}`)
		var lo LanguageOptions
		if err := json.Unmarshal(input, &lo); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		// Typed native fields.
		if lo.ParserOptions == nil || len(lo.ParserOptions.Project) != 1 ||
			lo.ParserOptions.Project[0] != "./tsconfig.json" {
			t.Errorf("project not extracted: %+v", lo.ParserOptions)
		}
		if lo.ParserOptions.ProjectService == nil || !*lo.ParserOptions.ProjectService {
			t.Error("projectService not extracted")
		}
		// Opaque parserOptions compat.
		if lo.ParserOptions.Compat["ecmaVersion"] != float64(2024) {
			t.Errorf("ecmaVersion stayed in Compat: %v", lo.ParserOptions.Compat["ecmaVersion"])
		}
		if lo.ParserOptions.Compat["sourceType"] != "module" {
			t.Errorf("sourceType stayed in Compat: %v", lo.ParserOptions.Compat["sourceType"])
		}
		// Top-level opaque.
		globals := lo.Compat["globals"].(map[string]any)
		if globals["Foo"] != "readonly" {
			t.Errorf("globals.Foo not captured: %v", globals)
		}
		if lo.Compat["futureField"] != "anything" {
			t.Errorf("futureField dropped: %v", lo.Compat["futureField"])
		}
	})

	t.Run("ToCompatWire excludes native fields", func(t *testing.T) {
		ps := true
		lo := &LanguageOptions{
			ParserOptions: &ParserOptions{
				ProjectService: &ps,
				Project:        ProjectPaths{"./t.json"},
				Compat:         map[string]any{"ecmaVersion": 2024},
			},
			Compat: map[string]any{"globals": map[string]any{"X": "readonly"}},
		}
		wire := lo.ToCompatWire()
		// Native fields MUST NOT be on the wire — the runner has no
		// business seeing them.
		po := wire["parserOptions"].(map[string]any)
		if _, found := po["project"]; found {
			t.Error("project leaked onto wire")
		}
		if _, found := po["projectService"]; found {
			t.Error("projectService leaked onto wire")
		}
		// Compat fields present.
		if po["ecmaVersion"] != 2024 {
			t.Errorf("ecmaVersion missing from wire: %v", po)
		}
		globals := wire["globals"].(map[string]any)
		if globals["X"] != "readonly" {
			t.Errorf("globals.X missing from wire: %v", globals)
		}
	})

	t.Run("ToCompatWire returns nil when nothing compat-relevant", func(t *testing.T) {
		ps := true
		lo := &LanguageOptions{
			// Native-only: project + projectService. No compat.
			ParserOptions: &ParserOptions{
				ProjectService: &ps,
				Project:        ProjectPaths{"./t.json"},
			},
		}
		if wire := lo.ToCompatWire(); wire != nil {
			t.Errorf("expected nil wire for native-only config, got %v", wire)
		}
	})
}

// Direct deepMergeMap unit tests — independent of LanguageOptions
// wrappers so the merge algorithm is testable in isolation.
func TestDeepMergeMap(t *testing.T) {
	t.Run("nil base returns clone of override", func(t *testing.T) {
		out := deepMergeMap(nil, map[string]any{"a": 1})
		if out["a"] != 1 {
			t.Errorf("expected a=1, got %v", out["a"])
		}
		out["a"] = 999
		// out is a fresh allocation, not aliasing override.
	})

	t.Run("nil override returns clone of base", func(t *testing.T) {
		base := map[string]any{"a": 1}
		out := deepMergeMap(base, nil)
		if out["a"] != 1 {
			t.Errorf("expected a=1, got %v", out["a"])
		}
		// Ensure freshly allocated (cloneAnyMap doesn't alias).
		out["a"] = 999
		if base["a"] == 999 {
			t.Error("base was mutated — cloneAnyMap must allocate")
		}
	})

	t.Run("scalar conflict: later wins", func(t *testing.T) {
		out := deepMergeMap(
			map[string]any{"a": 1, "b": "old"},
			map[string]any{"b": "new", "c": true},
		)
		if out["a"] != 1 {
			t.Errorf("a not preserved: %v", out["a"])
		}
		if out["b"] != "new" {
			t.Errorf("b not overridden: %v", out["b"])
		}
		if out["c"] != true {
			t.Errorf("c not added: %v", out["c"])
		}
	})

	t.Run("object values merge recursively", func(t *testing.T) {
		out := deepMergeMap(
			map[string]any{"x": map[string]any{"a": 1, "b": 2}},
			map[string]any{"x": map[string]any{"b": 99, "c": 3}},
		)
		x := out["x"].(map[string]any)
		if x["a"] != 1 || x["b"] != 99 || x["c"] != 3 {
			t.Errorf("nested merge wrong: %v", x)
		}
	})

	t.Run("type mismatch: later wins (replace, no merge)", func(t *testing.T) {
		// base has map, override has scalar. Override wins, not merged.
		out := deepMergeMap(
			map[string]any{"x": map[string]any{"a": 1}},
			map[string]any{"x": "replaced"},
		)
		if out["x"] != "replaced" {
			t.Errorf("type mismatch should replace, got %v", out["x"])
		}
	})

	t.Run("does not mutate inputs", func(t *testing.T) {
		base := map[string]any{"x": map[string]any{"a": 1}}
		override := map[string]any{"x": map[string]any{"b": 2}}
		deepMergeMap(base, override)
		baseX := base["x"].(map[string]any)
		if _, found := baseX["b"]; found {
			t.Error("base['x'] was mutated to include override's 'b'")
		}
	})
}

func TestIsGlobalIgnoreEntry(t *testing.T) {
	tests := []struct {
		name     string
		entry    ConfigEntry
		expected bool
	}{
		{
			name:     "only ignores",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}},
			expected: true,
		},
		{
			name:     "ignores with rules",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Rules: Rules{"no-debugger": "error"}},
			expected: false,
		},
		{
			name:     "ignores with files",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Files: []string{"**/*.ts"}},
			expected: false,
		},
		{
			name:     "ignores with plugins",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Plugins: []string{"@typescript-eslint"}},
			expected: false,
		},
		{
			name:     "ignores with languageOptions",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, LanguageOptions: &LanguageOptions{}},
			expected: false,
		},
		{
			name:     "ignores with settings",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Settings: Settings{"key": "val"}},
			expected: false,
		},
		{
			name:     "empty entry",
			entry:    ConfigEntry{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGlobalIgnoreEntry(tt.entry)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetConfigForFile_ArrayRuleConfig(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"array-type": []interface{}{"warn", map[string]interface{}{"default": "array-simple"}},
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	rc := merged.Rules["array-type"]
	if rc == nil {
		t.Fatal("Expected array-type rule to be present")
		return
	}
	if rc.Level != "warn" {
		t.Errorf("Expected level 'warn', got %q", rc.Level)
	}
	// rc.Options is the ESLint-aligned positional options array; the lone
	// object option lives at [0] (#5 — config.go no longer unwraps it).
	optsArr, _ := rc.Options.([]interface{})
	if len(optsArr) != 1 {
		t.Fatalf("Expected 1 positional option, got %v", rc.Options)
	}
	optsMap, _ := optsArr[0].(map[string]interface{})
	if optsMap == nil || optsMap["default"] != "array-simple" {
		t.Error("Expected options[0] to contain default: array-simple")
	}
}

func TestGetConfigForFile_SingleArrayOption_NotUnwrapped(t *testing.T) {
	// #5: a single ARRAY-valued option must stay wrapped in the positional
	// options array (ESLint's context.options shape). Pre-fix, config.go
	// unwrapped the lone element, so `["error", ["asc","desc"]]` collapsed
	// to options ["asc","desc"] (two scalar options) instead of
	// [["asc","desc"]] (one array option) — corrupting the compat layer's
	// context.options.
	config := RslintConfig{
		{
			Rules: Rules{
				"sort-keys": []interface{}{"error", []interface{}{"asc", "desc"}},
			},
		},
	}
	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
	}
	rc := merged.Rules["sort-keys"]
	if rc == nil {
		t.Fatal("Expected sort-keys rule present")
	}
	optsArr, _ := rc.Options.([]interface{})
	if len(optsArr) != 1 {
		t.Fatalf("Expected ONE positional option (the array), got %v", rc.Options)
	}
	inner, ok := optsArr[0].([]interface{})
	if !ok || len(inner) != 2 || inner[0] != "asc" || inner[1] != "desc" {
		t.Errorf("Expected options[0] == [asc desc], got %v", optsArr[0])
	}
	// Native rules still receive the historic unwrapped shape (the lone
	// option, not wrapped) via rule_registry's nativeRuleOptions.
	got, ok := nativeRuleOptions(rc.Options).([]interface{})
	if !ok || len(got) != 2 || got[0] != "asc" || got[1] != "desc" {
		t.Errorf("nativeRuleOptions should unwrap to [asc desc], got %v", nativeRuleOptions(rc.Options))
	}
}

func TestGetConfigForFile_RuleOff(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": "error",
			},
		},
		{
			Rules: Rules{
				"no-debugger": "off",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	rc := merged.Rules["no-debugger"]
	if rc == nil {
		t.Fatal("Expected no-debugger rule config to be present")
		return
	}
	if rc.IsEnabled() {
		t.Error("Expected no-debugger to be disabled after being turned off")
	}
}

func TestGetConfigForFile_MultipleEntries_LanguageOptionsMerge(t *testing.T) {
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					ProjectService: BoolPtr(true),
				},
			},
			Rules: Rules{"no-debugger": "error"},
		},
		{
			Files: []string{"**/*.ts"},
			LanguageOptions: &LanguageOptions{
				ParserOptions: &ParserOptions{
					ProjectService: BoolPtr(false),
					Project:        ProjectPaths{"./tsconfig.json"},
				},
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	if merged.LanguageOptions == nil || merged.LanguageOptions.ParserOptions == nil {
		t.Fatal("Expected languageOptions with parserOptions")
	}
	if merged.LanguageOptions.ParserOptions.ProjectService == nil || *merged.LanguageOptions.ParserOptions.ProjectService != false {
		t.Error("Expected projectService to be overridden to false")
	}
	if len(merged.LanguageOptions.ParserOptions.Project) != 1 {
		t.Error("Expected project to be set")
	}
}

func TestGetConfigForFile_ArrayRuleOff(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": []interface{}{"off"},
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	rc := merged.Rules["no-debugger"]
	if rc == nil {
		t.Fatal("Expected no-debugger rule config to be present")
		return
	}
	if rc.IsEnabled() {
		t.Error("Expected no-debugger to be disabled via [\"off\"] array syntax")
	}
}

func TestGetConfigForFile_EntryIgnores_NoFiles(t *testing.T) {
	// Entry with ignores but no files — applies to all files except ignored ones
	config := RslintConfig{
		{
			Ignores: []string{"**/*.test.ts"},
			Rules:   Rules{"no-debugger": "error"},
		},
	}

	// Non-ignored file should match
	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil for non-ignored file")
		return
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger for non-ignored file")
	}

	// Ignored file — no entry matches, return nil
	merged = config.GetConfigForFile("src/app.test.ts", "")
	if merged != nil {
		t.Error("Expected nil for ignored file with no other matching entry")
	}
}

func TestGetConfigForFile_EmptyConfig(t *testing.T) {
	config := RslintConfig{}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged != nil {
		t.Error("Expected nil for empty config (no entries)")
	}
}

func TestGetConfigForFile_MultipleEntries_DifferentFilesPatterns(t *testing.T) {
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-debugger": "error"},
		},
		{
			Files: []string{"**/*.js"},
			Rules: Rules{"no-console": "warn"},
		},
	}

	// .ts file: only entry1 matches
	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil for .ts file")
		return
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger from entry1")
	}
	if _, ok := merged.Rules["no-console"]; ok {
		t.Error("Expected no-console to not be present (entry2 doesn't match .ts)")
	}

	// .js file: only entry2 matches
	merged = config.GetConfigForFile("src/app.js", "")
	if merged == nil {
		t.Fatal("Expected non-nil for .js file")
		return
	}
	if _, ok := merged.Rules["no-console"]; !ok {
		t.Error("Expected no-console from entry2")
	}
	if _, ok := merged.Rules["no-debugger"]; ok {
		t.Error("Expected no-debugger to not be present (entry1 doesn't match .js)")
	}

	// .vue file: no entry matches → nil
	merged = config.GetConfigForFile("src/app.vue", "")
	if merged != nil {
		t.Error("Expected nil for .vue file (no entry matches)")
	}
}

func TestGetConfigForFile_MultipleEntries_PartialMatch(t *testing.T) {
	// entry1: only TS files; entry2: only Vue files; entry3: all files (no files pattern)
	config := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-debugger": "error"},
		},
		{
			Files: []string{"**/*.vue"},
			Rules: Rules{"no-console": "warn"},
		},
		{
			// No files → applies to all
			Rules: Rules{"for-direction": "error"},
		},
	}

	// .ts file: matches entry1 + entry3
	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil for .ts file")
		return
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger from entry1")
	}
	if _, ok := merged.Rules["for-direction"]; !ok {
		t.Error("Expected for-direction from entry3")
	}
	if _, ok := merged.Rules["no-console"]; ok {
		t.Error("Expected no-console to not be present (entry2 doesn't match .ts)")
	}

	// .vue file: matches entry2 + entry3
	merged = config.GetConfigForFile("src/app.vue", "")
	if merged == nil {
		t.Fatal("Expected non-nil for .vue file")
		return
	}
	if _, ok := merged.Rules["no-console"]; !ok {
		t.Error("Expected no-console from entry2")
	}
	if _, ok := merged.Rules["for-direction"]; !ok {
		t.Error("Expected for-direction from entry3")
	}
	if _, ok := merged.Rules["no-debugger"]; ok {
		t.Error("Expected no-debugger to not be present (entry1 doesn't match .vue)")
	}
}

func TestGetConfigForFile_ThreeEntries_CascadingOverride(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-debugger": "error",
				"no-console":  "error",
			},
		},
		{
			// Override no-debugger to warn, add for-direction
			Rules: Rules{
				"no-debugger":   "warn",
				"for-direction": "error",
			},
		},
		{
			// Turn off for-direction
			Rules: Rules{
				"for-direction": "off",
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	// no-debugger: entry1 "error" → entry2 "warn" → final "warn"
	if merged.Rules["no-debugger"].Level != "warn" {
		t.Errorf("Expected no-debugger 'warn', got %q", merged.Rules["no-debugger"].Level)
	}
	// no-console: entry1 "error", never overridden → final "error"
	if merged.Rules["no-console"].Level != "error" {
		t.Errorf("Expected no-console 'error', got %q", merged.Rules["no-console"].Level)
	}
	// for-direction: entry2 "error" → entry3 "off" → final "off"
	if merged.Rules["for-direction"].IsEnabled() {
		t.Error("Expected for-direction to be disabled (turned off in entry3)")
	}
}

func TestGetConfigForFile_MultipleEntries_ArrayRuleOverridesString(t *testing.T) {
	config := RslintConfig{
		{
			Rules: Rules{
				"no-console": "error",
			},
		},
		{
			// Later entry overrides string config with array config
			Rules: Rules{
				"no-console": []interface{}{"warn", map[string]interface{}{"allow": []interface{}{"error", "warn"}}},
			},
		},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil config")
		return
	}

	rc := merged.Rules["no-console"]
	if rc == nil {
		t.Fatal("Expected no-console in merged rules")
		return
	}
	if rc.Level != "warn" {
		t.Errorf("Expected level 'warn' from array override, got %q", rc.Level)
	}
	// rc.Options is the ESLint-aligned positional options array; the lone
	// object option lives at [0] (#5 — config.go no longer unwraps it).
	optsArr2, _ := rc.Options.([]interface{})
	if len(optsArr2) != 1 {
		t.Fatalf("Expected 1 positional option from array config, got %v", rc.Options)
	}
	optsMap2, _ := optsArr2[0].(map[string]interface{})
	if optsMap2 == nil {
		t.Fatal("Expected options[0] from array config")
		return
	}
	allow, ok := optsMap2["allow"].([]interface{})
	if !ok || len(allow) != 2 {
		t.Error("Expected allow option with 2 items")
	}
}

func TestGetConfigForFile_GlobalIgnore_PlusEntryIgnores(t *testing.T) {
	config := RslintConfig{
		{
			// Global ignore for dist
			Ignores: []string{"dist/**"},
		},
		{
			// Entry with its own ignores for test files
			Ignores: []string{"**/*.test.ts"},
			Rules:   Rules{"no-debugger": "error"},
		},
	}

	// File in dist: global ignore → nil
	merged := config.GetConfigForFile("dist/bundle.js", "")
	if merged != nil {
		t.Error("Expected nil for dist file (global ignore)")
	}

	// Test file: entry-level ignore, no other entry matches → nil
	merged = config.GetConfigForFile("src/app.test.ts", "")
	if merged != nil {
		t.Error("Expected nil for test file (entry-level ignore, no other match)")
	}

	// Normal file: not ignored anywhere, entry2 matches
	merged = config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("Expected non-nil for normal file")
		return
	}
	if _, ok := merged.Rules["no-debugger"]; !ok {
		t.Error("Expected no-debugger from entry2")
	}
}

// TestGetConfigForFile_CwdAffectsMatching verifies that the cwd parameter
// controls how files/ignores patterns are matched against absolute file paths.
// This is critical for monorepo sub-package configs where the config directory
// differs from the process cwd.
func TestGetConfigForFile_CwdAffectsMatching(t *testing.T) {
	config := RslintConfig{
		{
			Files: []string{"src/**/*.ts"},
			Rules: Rules{"no-console": "error"},
		},
	}

	// Absolute path: /monorepo/packages/foo/src/index.ts
	absPath := "/monorepo/packages/foo/src/index.ts"

	// With cwd = config's own directory (/monorepo/packages/foo),
	// relative path = src/index.ts → matches src/**/*.ts ✓
	merged := config.GetConfigForFile(absPath, "/monorepo/packages/foo")
	if merged == nil {
		t.Fatal("Expected match when cwd is the config directory")
		return
	}
	if merged.Rules["no-console"] == nil {
		t.Error("Expected no-console rule to be enabled")
	}

	// With cwd = monorepo root (/monorepo),
	// relative path = packages/foo/src/index.ts → does NOT match src/**/*.ts ✗
	merged = config.GetConfigForFile(absPath, "/monorepo")
	if merged != nil {
		t.Error("Expected no match when cwd is the monorepo root (wrong base for pattern)")
	}
}

// TestGetConfigForFile_CwdIgnoresMatching verifies cwd affects ignores resolution.
func TestGetConfigForFile_CwdIgnoresMatching(t *testing.T) {
	config := RslintConfig{
		{
			Ignores: []string{"dist/**"},
		},
		{
			Files: []string{"**/*.ts"},
			Rules: Rules{"no-console": "error"},
		},
	}

	absPath := "/project/dist/bundle.ts"

	// With cwd = /project, relative path = dist/bundle.ts → matches dist/** → globally ignored
	merged := config.GetConfigForFile(absPath, "/project")
	if merged != nil {
		t.Error("Expected file to be ignored when cwd matches config directory")
	}

	// With wrong cwd = /other, relative path won't start with dist/ → NOT ignored
	merged = config.GetConfigForFile(absPath, "/other")
	if merged == nil {
		t.Fatal("Expected file to NOT be ignored with wrong cwd")
		return
	}
}

// TestGetConfigForFile_WindowsPaths verifies cwd matching works with Windows-style paths.
// uriToPath produces forward-slash paths (C:/Users/...) and os.Getwd may produce
// backslash paths (C:\Users\...). Both must compute correct relative paths.
func TestGetConfigForFile_WindowsPaths(t *testing.T) {
	cfg := RslintConfig{
		{
			Files: []string{"src/**/*.ts"},
			Rules: Rules{"no-console": "error"},
		},
	}

	tests := []struct {
		name     string
		filePath string
		cwd      string
		wantHit  bool
	}{
		{
			name:     "forward-slash cwd (from uriToPath)",
			filePath: "C:/Users/project/src/index.ts",
			cwd:      "C:/Users/project",
			wantHit:  true,
		},
		{
			name:     "backslash cwd (from os.Getwd on Windows)",
			filePath: "C:/Users/project/src/index.ts",
			cwd:      "C:\\Users\\project",
			wantHit:  true,
		},
		{
			name:     "monorepo sub-package Windows cwd",
			filePath: "C:/repo/packages/foo/src/index.ts",
			cwd:      "C:/repo/packages/foo",
			wantHit:  true,
		},
		{
			name:     "wrong cwd on Windows — should not match",
			filePath: "C:/repo/packages/foo/src/index.ts",
			cwd:      "C:/repo",
			wantHit:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := cfg.GetConfigForFile(tt.filePath, tt.cwd)
			if tt.wantHit && merged == nil {
				t.Errorf("expected match for filePath=%q cwd=%q", tt.filePath, tt.cwd)
			}
			if !tt.wantHit && merged != nil {
				t.Errorf("expected no match for filePath=%q cwd=%q", tt.filePath, tt.cwd)
			}
		})
	}
}
