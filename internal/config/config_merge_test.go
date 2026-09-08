package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

func projectPolicyForTargetForTest(entries RslintConfig, owner string, fsys vfs.FS, target PathIdentity) (ProjectPolicy, error) {
	resolved := NewFileConfigResolverWithFS(entries, owner, fsys, baseRuleCatalog()).ResolveTarget(target)
	return ResolveProjectPolicy(resolved)
}

func TestProjectPolicyFlatConfigOverrides(t *testing.T) {
	owner := tspath.NormalizePath(t.TempDir())
	for _, test := range []struct {
		name, entries string
		want          ProjectPolicy
		error         string
	}{
		{name: "preset conflict", entries: `[{"projectService":true},{"project":"custom.json"}]`, error: "remove project"},
		{name: "empty array conflict", entries: `[{"projectService":true,"project":[]}]`, error: "remove project"},
		{name: "automatic conflict", entries: `[{"projectService":true,"project":true}]`, error: "remove project"},
		{name: "clear project false", entries: `[{"project":"custom.json"},{"projectService":true,"project":false}]`, want: ProjectPolicy{ProjectService: true}},
		{name: "clear project null", entries: `[{"project":"custom.json"},{"projectService":true,"project":null}]`, want: ProjectPolicy{ProjectService: true}},
		{name: "clear service false", entries: `[{"projectService":true},{"projectService":false,"project":"custom.json"}]`, want: ProjectPolicy{DefaultProjectDisabled: true}},
		{name: "clear service null", entries: `[{"projectService":true},{"projectService":null}]`, want: ProjectPolicy{DefaultProjectDisabled: true}},
		{name: "matched false reset", entries: `[{"project":"custom.json"},{"project":false}]`, want: ProjectPolicy{ProjectDisabled: true}},
		{name: "matched null reset", entries: `[{"project":"custom.json"},{"project":null}]`, want: ProjectPolicy{ProjectDisabled: true}},
		{name: "ordinary empty array", entries: `[{"project":"custom.json"},{"project":[]}]`},
		{name: "restore after reset", entries: `[{"project":false},{"project":"custom.json"}]`},
		{name: "automatic unsupported", entries: `[{"project":true}]`, error: "not supported"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var options []ParserOptions
			if err := json.Unmarshal([]byte(test.entries), &options); err != nil {
				t.Fatal(err)
			}
			var entries RslintConfig
			for index := range options {
				entries = append(entries, ConfigEntry{Files: []string{"**/*.ts"}, LanguageOptions: &LanguageOptions{ParserOptions: &options[index]}})
			}
			entries = append(entries, ConfigEntry{Files: []string{"other/**"}, LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{Project: ProjectPaths{"missing.json"}}}})
			encoded, err := json.Marshal(entries)
			if err != nil {
				t.Fatal(err)
			}
			var roundTrip RslintConfig
			if err := json.Unmarshal(encoded, &roundTrip); err != nil {
				t.Fatal(err)
			}
			for _, candidate := range []RslintConfig{entries, roundTrip} {
				policy, err := projectPolicyForTargetForTest(candidate, owner, nil, PathIdentity{Path: owner + "/src/file.ts"})
				if test.error != "" {
					if err == nil || !strings.Contains(err.Error(), test.error) {
						t.Fatalf("want %q, got %v", test.error, err)
					}
				} else if err != nil || policy != test.want {
					t.Fatalf("policy=%+v error=%v, want %+v", policy, err, test.want)
				}
			}
		})
	}
}

func TestHasProjectOptions(t *testing.T) {
	for _, test := range []struct {
		options string
		want    bool
	}{
		{`{}`, false}, {`{"project":"one.json"}`, false}, {`{"project":["one.json","two.json"]}`, false}, {`{"project":[]}`, false},
		{`{"projectService":true}`, true}, {`{"projectService":false}`, true}, {`{"projectService":null}`, true},
		{`{"project":false}`, true}, {`{"project":null}`, true}, {`{"project":true}`, true},
		{`{"tsconfigRootDir":null}`, true}, {`{"tsconfigRootDir":""}`, true},
	} {
		t.Run(test.options, func(t *testing.T) {
			var options ParserOptions
			if err := json.Unmarshal([]byte(test.options), &options); err != nil {
				t.Fatal(err)
			}
			entries := RslintConfig{{Rules: Rules{"no-debugger": "error"}}, {Files: []string{"unused.ts"}, LanguageOptions: &LanguageOptions{ParserOptions: &options}}}
			if got := HasProjectOptions(entries); got != test.want {
				t.Fatalf("HasProjectOptions=%v, want %v", got, test.want)
			}
		})
	}
}

func TestProjectPolicyPreservesOwnerDeclarations(t *testing.T) {
	owner := tspath.NormalizePath(t.TempDir())
	for _, name := range []string{"first.json", "second.json", "tsconfig.json"} {
		createTestFile(t, tspath.ResolvePath(owner, name))
	}
	for _, test := range []struct {
		name, options string
		matched       bool
		want          []string
	}{
		{name: "ordinary project order", options: `{}`, want: []string{"first.json", "second.json"}},
		{name: "empty array keeps earlier declarations", options: `{"project":[]}`, matched: true, want: []string{"first.json", "second.json"}},
		{name: "service false keeps declarations", options: `{"projectService":false}`, matched: true, want: []string{"first.json", "second.json"}},
		{name: "service null keeps declarations", options: `{"projectService":null}`, matched: true, want: []string{"first.json", "second.json"}},
		{name: "unmatched service neutral", options: `{"projectService":true}`, want: []string{"first.json", "second.json"}},
		{name: "unmatched root neutral", options: `{"tsconfigRootDir":"."}`, want: []string{"first.json", "second.json"}},
		{name: "unmatched reset neutral", options: `{"project":false}`, want: []string{"first.json", "second.json"}},
		{name: "matched false reset keeps raw declarations", options: `{"project":false}`, matched: true, want: []string{"first.json", "second.json"}},
		{name: "matched null reset keeps raw declarations", options: `{"project":null}`, matched: true, want: []string{"first.json", "second.json"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var options ParserOptions
			if err := json.Unmarshal([]byte(test.options), &options); err != nil {
				t.Fatal(err)
			}
			selector := "unused.ts"
			if test.matched {
				selector = "target.ts"
			}
			entries := RslintConfig{
				{LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{Project: ProjectPaths{"first.json"}}}},
				{Files: []string{"unused.ts"}, LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{Project: ProjectPaths{"second.json"}}}},
				{Files: []string{selector}, LanguageOptions: &LanguageOptions{ParserOptions: &options}},
			}
			policy, err := projectPolicyForTargetForTest(entries, owner, osvfs.FS(), PathIdentity{Path: owner + "/target.ts"})
			if err != nil {
				t.Fatal(err)
			}
			paths, err := ResolveTsConfigPathsWithPolicy(entries, owner, osvfs.FS(), policy)
			var want []string
			for _, name := range test.want {
				want = append(want, tspath.ResolvePath(owner, name))
			}
			if err != nil || !reflect.DeepEqual(paths, want) {
				t.Fatalf("paths=%v error=%v, want %v", paths, err, want)
			}
		})
	}
}

func TestProjectPolicyDefaultFallbackAndEmptyDeclarations(t *testing.T) {
	owner := tspath.NormalizePath(t.TempDir())
	createTestFile(t, owner+"/tsconfig.json")
	for _, declaration := range []ProjectPaths{nil, {}} {
		for _, test := range []struct {
			options              string
			matched, wantDefault bool
		}{
			{`{"projectService":true}`, false, true}, {`{"projectService":false}`, false, true},
			{`{"project":false}`, false, true}, {`{"project":null}`, false, true},
			{`{"projectService":false}`, true, false}, {`{"projectService":null}`, true, false},
			{`{"project":false}`, true, true}, {`{"project":null}`, true, true},
		} {
			t.Run(fmt.Sprintf("empty=%t/match=%t/%s", declaration != nil, test.matched, test.options), func(t *testing.T) {
				var options ParserOptions
				if err := json.Unmarshal([]byte(test.options), &options); err != nil {
					t.Fatal(err)
				}
				selector := "unused.ts"
				if test.matched {
					selector = "target.ts"
				}
				entries := RslintConfig{
					{LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{Project: declaration}}},
					{Files: []string{selector}, LanguageOptions: &LanguageOptions{ParserOptions: &options}},
				}
				policy, err := projectPolicyForTargetForTest(entries, owner, osvfs.FS(), PathIdentity{Path: owner + "/target.ts"})
				if err != nil {
					t.Fatal(err)
				}
				paths, err := ResolveTsConfigPathsWithPolicy(entries, owner, osvfs.FS(), policy)
				var want []string
				if test.wantDefault && declaration == nil {
					want = []string{owner + "/tsconfig.json"}
				}
				if err != nil || !reflect.DeepEqual(paths, want) {
					t.Fatalf("paths=%v error=%v, want %v", paths, err, want)
				}
			})
		}
	}
}

func TestProjectPolicyPreservesLiteralBaseForProjectPatterns(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	owner, first, second, boundary := root+"/owner", root+"/first[1]", root+"/second[2]", root+"/boundary[3]"
	for _, test := range []struct {
		name, suffix string
		root         string
	}{
		{name: "authored bases"},
		{name: "root override", suffix: fmt.Sprintf(`,{"languageOptions":{"parserOptions":{"tsconfigRootDir":%q}}}`, boundary), root: boundary},
		{name: "null reset", suffix: fmt.Sprintf(`,{"languageOptions":{"parserOptions":{"tsconfigRootDir":%q}}},{"languageOptions":{"parserOptions":{"tsconfigRootDir":null}}}`, boundary)},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, dir := range []string{first, second, boundary} {
				for _, file := range []string{"first.json", "second.json"} {
					createTestFile(t, tspath.ResolvePath(dir, file))
				}
			}
			entries := ConfigWithAuthoredPathBase(RslintConfig{{LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{Project: ProjectPaths{"./first*.json"}}}}}, first)
			entries = append(entries, ConfigWithAuthoredPathBase(RslintConfig{{LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{Project: ProjectPaths{"./second.json"}}}}}, second)...)
			var override RslintConfig
			if err := json.Unmarshal([]byte(`[{"languageOptions":{"parserOptions":{"projectService":false}}}`+test.suffix+`]`), &override); err != nil {
				t.Fatal(err)
			}
			entries = append(entries, override...)
			policy, err := projectPolicyForTargetForTest(entries, owner, osvfs.FS(), PathIdentity{Path: owner + "/target.ts"})
			if err != nil || policy.TsconfigRootDir != test.root {
				t.Fatalf("policy=%+v error=%v", policy, err)
			}
			want := []string{first + "/first.json", second + "/second.json"}
			if test.root != "" {
				want = []string{boundary + "/first.json", boundary + "/second.json"}
			}
			paths, err := ResolveTsConfigPathsWithPolicy(entries, owner, osvfs.FS(), policy)
			if err != nil || !reflect.DeepEqual(paths, want) {
				t.Fatalf("paths=%v error=%v, want %v", paths, err, want)
			}
		})
	}
}

func TestProjectPolicyRootValidationAfterMerge(t *testing.T) {
	owner := tspath.NormalizePath(t.TempDir())
	previous := owner + "/previous-boundary"
	explicit := owner + "/boundary"
	for _, root := range []struct {
		name  string
		value string
	}{
		{"relative", "."},
		{"empty", ""},
	} {
		for _, service := range []bool{true, false} {
			for _, test := range []struct {
				name        string
				files       string
				suffix      string
				wantRoot    string
				wantError   bool
				wantIgnored bool
			}{
				{name: "matched", files: "src/**", wantError: true},
				{name: "unmatched", files: "elsewhere/**", wantRoot: previous},
				{name: "overridden", files: "src/**", suffix: fmt.Sprintf(`,{"languageOptions":{"parserOptions":{"tsconfigRootDir":%q}}}`, explicit), wantRoot: explicit},
				{name: "null reset", files: "src/**", suffix: `,{"languageOptions":{"parserOptions":{"tsconfigRootDir":null}}}`, wantRoot: ""},
				{name: "global ignore", files: "src/**", suffix: `,{"ignores":["src/**"]}`, wantIgnored: true},
			} {
				name := fmt.Sprintf("root=%s/service=%t/%s", root.name, service, test.name)
				t.Run(name, func(t *testing.T) {
					input := fmt.Sprintf(`[{
						"files":["**/*.ts"],
						"languageOptions":{"parserOptions":{"projectService":%t,"tsconfigRootDir":%q}}
					},{"files":[%q],"languageOptions":{"parserOptions":{"tsconfigRootDir":%q}}}%s]`, service, previous, test.files, root.value, test.suffix)
					var original RslintConfig
					if err := json.Unmarshal([]byte(input), &original); err != nil {
						t.Fatalf("root validation ran before matching and merging: %v", err)
					}
					encoded, err := json.Marshal(original)
					if err != nil {
						t.Fatal(err)
					}
					var roundTrip RslintConfig
					if err := json.Unmarshal(encoded, &roundTrip); err != nil {
						t.Fatal(err)
					}
					for _, config := range []struct {
						name    string
						entries RslintConfig
					}{
						{"original", original},
						{"JSON round trip", roundTrip},
					} {
						t.Run(config.name, func(t *testing.T) {
							options := config.entries[1].LanguageOptions.ParserOptions
							if options.TsconfigRootDir == nil || *options.TsconfigRootDir != root.value {
								t.Fatalf("explicit %q root became omitted or null: %+v", root.value, options)
							}
							policy, err := projectPolicyForTargetForTest(config.entries, owner, nil, PathIdentity{Path: owner + "/src/file.ts"})
							if test.wantError {
								if err == nil || !strings.Contains(err.Error(), "absolute path") {
									t.Fatalf("expected final root validation even when service=%t, got %+v, %v", service, policy, err)
								}
								return
							}
							want := ProjectPolicy{ProjectService: service, TsconfigRootDir: test.wantRoot, DefaultProjectDisabled: !service}
							if test.wantIgnored {
								want = ProjectPolicy{}
							}
							if err != nil || !reflect.DeepEqual(policy, want) {
								t.Fatalf("policy = %+v, error = %v; want %+v", policy, err, want)
							}
						})
					}
				})
			}
		}
	}
}

func TestProjectPolicyRootPathNormalization(t *testing.T) {
	owner := tspath.NormalizePath(t.TempDir())
	root := tspath.NormalizePath(filepath.VolumeName(owner) + string(filepath.Separator))
	for _, test := range []struct {
		name, value, want string
		invalid           bool
	}{
		{name: "absolute", value: owner, want: owner},
		{name: "trailing separator", value: owner + "/", want: owner},
		{name: "dot segments", value: owner + "/nested/.././", want: owner},
		{name: "filesystem root", value: root, want: root},
		{name: "relative", value: "relative", invalid: true},
		{name: "empty", value: "", invalid: true},
		{name: "drive path", value: "C:/repo/", want: "C:/repo", invalid: runtime.GOOS != "windows"},
		{name: "backslash root", value: `\repo`, invalid: true},
		{name: "network root", value: `\\server\share\repo`, invalid: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			entries := RslintConfig{{LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{TsconfigRootDir: &test.value}}}}
			policy, err := projectPolicyForTargetForTest(entries, owner, nil, PathIdentity{Path: owner + "/file.ts"})
			if test.invalid {
				if err == nil || !strings.Contains(err.Error(), "absolute path") {
					t.Fatalf("expected invalid host path, got %+v, %v", policy, err)
				}
				return
			}
			if err != nil || policy.TsconfigRootDir != test.want {
				t.Fatalf("root=%q error=%v, want %q", policy.TsconfigRootDir, err, test.want)
			}
		})
	}
}

func TestParserOptionsRejectUnsupportedServiceOptions(t *testing.T) {
	for _, input := range []string{
		`{"projectService":{}}`,
		`{"projectService":{"allowDefaultProject":["*.js"]}}`,
		`{"tsconfigRootDir":42}`,
	} {
		var options ParserOptions
		if err := json.Unmarshal([]byte(input), &options); err == nil {
			t.Fatalf("expected unsupported options to fail: %s", input)
		}
	}
}

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

func TestGetConfigForFile_PluginDeclarationDoesNotAutoEnableRules(t *testing.T) {

	// A plugin declaration gates namespaced rules but does not enable rules.
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

	// No rules should be enabled without an explicit rule or preset.
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

func TestGetConfigForFile_SeverityOnlyRuleOverridePreservesOptions(t *testing.T) {
	for _, override := range []any{"warn", 1, []interface{}{float64(1)}} {
		t.Run(fmt.Sprintf("%T_%v", override, override), func(t *testing.T) {
			config := RslintConfig{
				{Rules: Rules{"example": []interface{}{"error", map[string]interface{}{"mode": "strict"}, "tail"}}},
				{Rules: Rules{"example": override}},
			}

			merged := config.GetConfigForFile("src/app.ts", "")
			if merged == nil || merged.Rules["example"] == nil {
				t.Fatal("expected merged example rule")
				return
			}
			ruleConfig := merged.Rules["example"]
			if ruleConfig.Level != "warn" {
				t.Fatalf("severity = %q, want warn", ruleConfig.Level)
			}
			if len(ruleConfig.Options) != 2 || ruleConfig.Options[1] != "tail" {
				t.Fatalf("severity-only override lost prior options: %#v", ruleConfig.Options)
			}
		})
	}
}

func TestGetConfigForFile_NumericRuleSeverities(t *testing.T) {
	config := RslintConfig{{Rules: Rules{
		"off":   0,
		"warn":  float64(1),
		"error": uint8(2),
	}}}

	merged := config.GetConfigForFile("src/app.ts", "")
	if merged == nil {
		t.Fatal("expected merged config")
		return
	}
	for name, want := range map[string]string{"off": "off", "warn": "warn", "error": "error"} {
		if got := merged.Rules[name]; got == nil || got.Level != want {
			t.Errorf("rule %q = %#v, want level %q", name, got, want)
		}
	}
}

func TestGetConfigForFile_NumericRuleOverrideWithOptionsReplacesOptions(t *testing.T) {
	config := RslintConfig{
		{Rules: Rules{"example": []interface{}{"error", "old", map[string]any{"keep": false}}}},
		{Rules: Rules{"example": []interface{}{1, "new", true}}},
	}

	merged := config.GetConfigForFile("src/app.ts", "")
	ruleConfig := merged.Rules["example"]
	if ruleConfig.Level != "warn" {
		t.Fatalf("severity = %q, want warn", ruleConfig.Level)
	}
	if !reflect.DeepEqual(ruleConfig.Options, []interface{}{"new", true}) {
		t.Fatalf("explicit positional options were not replaced: %#v", ruleConfig.Options)
	}
}

func TestGetConfigForFile_SettingsDeepMerge(t *testing.T) {
	config := RslintConfig{
		{
			Settings: Settings{
				"importResolver": "node",
				"react": map[string]any{
					"version": "17",
					"pragma":  "h",
				},
				"extensions": []any{".js"},
			},
		},
		{
			Settings: Settings{
				"react":      map[string]any{"version": "18"},
				"extensions": []any{".ts"},
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
	react, ok := merged.Settings["react"].(map[string]any)
	if !ok || react["version"] != "18" || react["pragma"] != "h" {
		t.Fatalf("nested settings were not deeply merged: %#v", merged.Settings["react"])
	}
	extensions, ok := merged.Settings["extensions"].([]any)
	if !ok || len(extensions) != 1 || extensions[0] != ".ts" {
		t.Fatalf("settings arrays should be replaced, got %#v", merged.Settings["extensions"])
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

	t.Run("raw parser options are recursively merged", func(t *testing.T) {
		base := &LanguageOptions{Raw: map[string]any{
			"parserOptions": map[string]any{
				"alpha": map[string]any{"one": float64(1), "two": float64(2)},
			},
		}}
		override := &LanguageOptions{Raw: map[string]any{
			"parserOptions": map[string]any{
				"alpha": map[string]any{"one": float64(9)},
			},
		}}

		result := mergeLanguageOptions(base, override)
		parserOptions := result.Raw["parserOptions"].(map[string]any)
		alpha := parserOptions["alpha"].(map[string]any)
		if alpha["one"] != float64(9) || alpha["two"] != float64(2) {
			t.Fatalf("raw parserOptions were not deeply merged: %#v", parserOptions)
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
			name:     "ignores with name",
			entry:    ConfigEntry{Name: "global ignores", Ignores: []string{"dist/**"}},
			expected: true,
		},
		{
			name:     "ignores with rules",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Rules: Rules{"no-debugger": "error"}},
			expected: false,
		},
		{
			name:     "ignores with empty rules",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Rules: Rules{}},
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
			name:     "ignores with empty plugins",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Plugins: []string{}},
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
			name:     "ignores with empty settings",
			entry:    ConfigEntry{Ignores: []string{"dist/**"}, Settings: Settings{}},
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
	if len(rc.Options) != 1 {
		t.Fatalf("Expected 1 option, got %d", len(rc.Options))
	}
	optsMap, _ := rc.Options[0].(map[string]interface{})
	if optsMap == nil || optsMap["default"] != "array-simple" {
		t.Error("Expected options to contain default: array-simple")
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
		return
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

func TestConfigDecode_PreservesNonGlobalIgnoreObjectShape(t *testing.T) {
	for _, raw := range []string{
		`[{"ignores":["dist/**"],"rules":null}]`,
		`[{"ignores":["dist/**"],"processor":"example"}]`,
		`[{"ignores":["dist/**"],"plugins":[]}]`,
		`[{"ignores":["dist/**"],"settings":{}}]`,
	} {
		var cfg RslintConfig
		if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
			t.Fatalf("decode %s: %v", raw, err)
		}
		if len(cfg) != 1 || isGlobalIgnoreEntry(cfg[0]) {
			t.Fatalf("entry with another authored key must not become a global ignore: %#v", cfg)
		}
		if ignores := extractConfigIgnores(cfg); len(ignores) != 0 {
			t.Fatalf("entry-level ignore leaked into global ignores: %#v", ignores)
		}

		encoded, err := json.Marshal(cfg)
		if err != nil {
			t.Fatalf("re-encode %s: %v", raw, err)
		}
		var roundTripped RslintConfig
		if err := json.Unmarshal(encoded, &roundTripped); err != nil {
			t.Fatalf("round-trip %s via %s: %v", raw, encoded, err)
		}
		if len(roundTripped) != 1 || isGlobalIgnoreEntry(roundTripped[0]) {
			t.Fatalf("round-trip changed entry-local ignore into global ignore: raw=%s encoded=%s decoded=%#v", raw, encoded, roundTripped)
		}
	}
}

func TestConfigJSONRoundTripPreservesGlobalIgnoreObjectShape(t *testing.T) {
	original := RslintConfig{{Ignores: []string{"coverage/**"}}}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal global ignore: %v", err)
	}
	if strings.Contains(string(encoded), `"rules"`) {
		t.Fatalf("marshal invented a non-global rules key: %s", encoded)
	}
	var decoded RslintConfig
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal global ignore: %v", err)
	}
	if len(decoded) != 1 || !isGlobalIgnoreEntry(decoded[0]) {
		t.Fatalf("round-trip changed global ignore shape: %#v", decoded)
	}
}

func TestConfigJSONRoundTripPreservesNonNilEmptyRulesShape(t *testing.T) {
	original := RslintConfig{{
		Ignores: []string{"coverage/**"},
		Rules:   Rules{},
	}}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal empty rules entry: %v", err)
	}
	if !strings.Contains(string(encoded), `"rules":{}`) {
		t.Fatalf("marshal dropped an authored empty rules key: %s", encoded)
	}
	var decoded RslintConfig
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal empty rules entry: %v", err)
	}
	if len(decoded) != 1 || isGlobalIgnoreEntry(decoded[0]) {
		t.Fatalf("round-trip changed entry-local ignore into global ignore: %#v", decoded)
	}
}

func TestGetConfigForFile_UnscopedEntriesStayWithinConfigSelectorUnion(t *testing.T) {
	config := RslintConfig{
		{Rules: Rules{"base-rule": "error"}},
		{Files: []string{"**/*.JS"}, Rules: Rules{"uppercase-rule": "error"}},
	}

	defaultFile := config.GetConfigForFile("src/app.ts", "")
	if defaultFile == nil || defaultFile.Rules["base-rule"] == nil {
		t.Fatalf("default baseline file should receive the unscoped entry: %+v", defaultFile)
	}

	explicitlySelected := config.GetConfigForFile("src/app.JS", "")
	if explicitlySelected == nil || explicitlySelected.Rules["base-rule"] == nil || explicitlySelected.Rules["uppercase-rule"] == nil {
		t.Fatalf("explicit selector should make unscoped entries cascade: %+v", explicitlySelected)
	}

	outsideSelector := RslintConfig{{Rules: Rules{"base-rule": "error"}}}.GetConfigForFile("src/app.JS", "")
	if outsideSelector != nil {
		t.Fatalf("unscoped entry must not configure a path outside the implicit selector: %+v", outsideSelector)
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
	if len(rc.Options) != 1 {
		t.Fatalf("Expected 1 option, got %d", len(rc.Options))
	}
	optsMap2, _ := rc.Options[0].(map[string]interface{})
	if optsMap2 == nil {
		t.Fatal("Expected options from array config")
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

func TestGetConfigForFile_FilesAndGroupsUseOrOutsideAndInside(t *testing.T) {
	cfg := RslintConfig{{
		Files: []string{"special.ts"},
		FilePatternGroups: [][]string{
			{"src/**", "**/*.js", "!**/*.test.js"},
		},
		Rules: Rules{"no-console": "error"},
	}}

	for _, filePath := range []string{"/repo/special.ts", "/repo/src/app.js"} {
		merged := cfg.GetConfigForFile(filePath, "/repo")
		if merged == nil || merged.Rules["no-console"] == nil {
			t.Fatalf("expected %q to match a files selector", filePath)
		}
	}
	for _, filePath := range []string{"/repo/src/app.test.js", "/repo/other/app.js", "/repo/src/app.ts"} {
		if merged := cfg.GetConfigForFile(filePath, "/repo"); merged != nil {
			t.Fatalf("expected %q to fail the complete AND group, got %+v", filePath, merged)
		}
	}
}
