package config

import (
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"

	"github.com/web-infra-dev/rslint/internal/linter"
)

func TestConfigShapeResolutionMatchesLegacyAlgorithm(t *testing.T) {
	RegisterAllRules()
	random := rand.New(rand.NewSource(0x5a17_2026))
	paths := []string{
		"/repo/src/app.ts",
		"/repo/src/app.test.ts",
		"/repo/src/special.ts",
		"/repo/src/generated/app.ts",
		"/repo/test/app.js",
		"/repo/component.vue",
		"/repo/component.unsupported",
	}

	for sample := range 100 {
		entryCount := 1 + random.Intn(80)
		config := make(RslintConfig, entryCount)
		for index := range config {
			entry := ConfigEntry{Name: fmt.Sprintf("sample-%d-entry-%d", sample, index)}
			switch random.Intn(6) {
			case 1:
				entry.Files = []string{"src/**/*.ts"}
			case 2:
				entry.Files = []string{"**/*.vue"}
			case 3:
				entry.Files = []string{"src/special.ts"}
			case 4:
				entry.FilePatternGroups = [][]string{{"src/**", "**/*.ts", "!**/*.test.ts"}}
			case 5:
				entry.FilePatternGroups = [][]string{{}}
			}
			if random.Intn(4) == 0 {
				ignores := []string{"**/*.test.ts", "src/generated/**", "!src/generated/keep.ts"}
				entry.Ignores = ignores[:1+random.Intn(len(ignores))]
			}
			if random.Intn(3) != 0 {
				severities := []any{"off", "warn", "error", []any{"error", map[string]any{"allow": []any{"warn"}}}}
				entry.Rules = Rules{"no-console": severities[random.Intn(len(severities))]}
			}
			if random.Intn(4) == 0 {
				entry.Settings = Settings{
					"shared": map[string]any{
						"sample": sample,
						"entry":  index,
					},
				}
			}
			if random.Intn(6) == 0 {
				entry.LanguageOptions = &LanguageOptions{Raw: map[string]any{
					"globals": map[string]any{"testGlobal": "readonly"},
				}}
			}
			if random.Intn(8) == 0 {
				entry.Plugins = []string{"@typescript-eslint"}
			}
			config[index] = entry
		}

		resolver := NewFileConfigResolver(config, "/repo", false)
		for _, filePath := range paths {
			want := legacyMergedConfigForTest(config, filePath, "/repo", extractConfigIgnores(config))
			if got := config.GetConfigForFile(filePath, "/repo"); !reflect.DeepEqual(got, want) {
				t.Fatalf("sample %d path %s direct mismatch:\nnew:    %#v\nlegacy: %#v", sample, filePath, got, want)
			}
			if got := resolver.ConfigForFile(filePath); !reflect.DeepEqual(got, want) {
				t.Fatalf("sample %d path %s resolver mismatch:\nnew:    %#v\nlegacy: %#v", sample, filePath, got, want)
			}
		}
	}
}

func TestFileConfigResolverMatchesDirectResolutionAcrossShapes(t *testing.T) {
	RegisterAllRules()
	config := RslintConfig{
		{Ignores: []string{"generated/**"}},
		{
			Rules: Rules{
				"no-console": []any{"warn", map[string]any{"allow": []any{"warn"}}},
			},
			Settings: Settings{"shared": map[string]any{"base": true}},
		},
		{
			Files: []string{"src/**/*.ts"},
			Rules: Rules{
				"no-debugger": "error",
			},
			Settings: Settings{"shared": map[string]any{"typed": true}},
			LanguageOptions: &LanguageOptions{Raw: map[string]any{
				"globals": map[string]any{"typedGlobal": "readonly"},
			}},
		},
		{
			Files: []string{"src/special.ts"},
			Rules: Rules{
				"no-console": "error",
			},
		},
		{
			Files:   []string{"src/**/*.ts"},
			Ignores: []string{"**/*.test.ts"},
			Rules: Rules{
				"eqeqeq": "warn",
			},
		},
		{
			FilePatternGroups: [][]string{{"src/**", "**/*.ts"}},
			Rules: Rules{
				"no-alert": "warn",
			},
		},
		{
			Files: []string{"**/*.vue"},
			Rules: Rules{
				"no-debugger": "warn",
			},
		},
	}

	resolver := NewFileConfigResolver(config, "/repo", false)
	paths := []string{
		"/repo/src/a.ts",
		"/repo/src/b.ts",
		"/repo/src/special.ts",
		"/repo/src/a.test.ts",
		"/repo/src/a.js",
		"/repo/components/view.vue",
		"/repo/generated/skip.ts",
		"/repo/outside.unsupported",
	}
	for _, filePath := range paths {
		directMerged := config.GetConfigForFile(filePath, "/repo")
		directRules := GlobalRuleRegistry.GetEnabledRulesForMergedConfig(directMerged, false)
		resolvedRules, resolvedMerged := resolver.EnabledRulesForFile(filePath)
		if !reflect.DeepEqual(resolvedMerged, directMerged) {
			t.Fatalf("%s merged config mismatch:\nresolver: %#v\ndirect:   %#v", filePath, resolvedMerged, directMerged)
		}
		if got, want := configuredRuleViews(resolvedRules), configuredRuleViews(directRules); !reflect.DeepEqual(got, want) {
			t.Fatalf("%s enabled rules mismatch:\nresolver: %#v\ndirect:   %#v", filePath, got, want)
		}
	}

	first := resolver.planForFile("/repo/src/a.ts")
	if second := resolver.planForFile("/repo/src/b.ts"); first == nil || second != first {
		t.Fatal("files with the same exact config shape did not share one effective plan")
	}
	if special := resolver.planForFile("/repo/src/special.ts"); special == nil || special == first {
		t.Fatal("files with different config shapes shared an effective plan")
	}
}

func TestFileConfigResolverDistinguishesMatchedZeroRulesAndMisses(t *testing.T) {
	config := RslintConfig{{Settings: Settings{}}}
	resolver := NewFileConfigResolver(config, "/repo", false)

	rules, merged := resolver.EnabledRulesForFile("/repo/src/file.ts")
	if merged == nil || len(rules) != 0 {
		t.Fatalf("default-selected zero-rule file resolved to rules=%v merged=%#v", rules, merged)
	}
	if rules, merged = resolver.EnabledRulesForFile("/repo/src/file.vue"); rules != nil || merged != nil {
		t.Fatalf("unsupported unselected file resolved to rules=%v merged=%#v", rules, merged)
	}
}

func TestFileConfigResolverShapeKeyPreservesEntriesBeyond64(t *testing.T) {
	RegisterAllRules()
	config := make(RslintConfig, 72)
	for index := range config {
		config[index] = ConfigEntry{Name: fmt.Sprintf("base-%d", index)}
	}
	config[63] = ConfigEntry{
		Name:  "low-special",
		Files: []string{"src/low.ts"},
		Rules: Rules{"no-alert": "warn"},
	}
	config[64] = ConfigEntry{
		Name:  "tail-special",
		Files: []string{"src/tail.ts"},
		Rules: Rules{"no-debugger": "error"},
	}
	config[71] = ConfigEntry{
		Name:  "tail-vue",
		Files: []string{"**/*.vue"},
		Rules: Rules{"no-console": "warn"},
	}

	resolver := NewFileConfigResolver(config, "/repo", false)
	paths := []string{
		"/repo/src/ordinary.ts",
		"/repo/src/low.ts",
		"/repo/src/tail.ts",
		"/repo/component.vue",
		"/repo/component.unsupported",
	}
	plans := make(map[string]*effectiveConfigPlan, len(paths))
	for _, filePath := range paths {
		resolvedRules, resolvedMerged := resolver.EnabledRulesForFile(filePath)
		directMerged := config.GetConfigForFile(filePath, "/repo")
		directRules := GlobalRuleRegistry.GetEnabledRulesForMergedConfig(directMerged, false)
		if !reflect.DeepEqual(resolvedMerged, directMerged) ||
			!reflect.DeepEqual(configuredRuleViews(resolvedRules), configuredRuleViews(directRules)) {
			t.Fatalf("resolution mismatch for %s", filePath)
		}
		plans[filePath] = resolver.planForFile(filePath)
	}
	if plans["/repo/src/ordinary.ts"] == plans["/repo/src/low.ts"] ||
		plans["/repo/src/ordinary.ts"] == plans["/repo/src/tail.ts"] ||
		plans["/repo/src/ordinary.ts"] == plans["/repo/component.vue"] {
		t.Fatal("entry bits at or beyond the 64-entry boundary collided")
	}
	if plans["/repo/component.unsupported"] != nil {
		t.Fatal("unsupported file with only unscoped matches should remain unselected")
	}
}

func TestFileConfigResolverPreservesWindowsPathMatching(t *testing.T) {
	config := RslintConfig{{
		Files:   []string{"src/**/*.ts"},
		Ignores: []string{"src/generated/**"},
		Rules:   Rules{"no-console": "error"},
	}}
	tests := []struct {
		filePath string
		cwd      string
	}{
		{"C:/Users/project/src/index.ts", "C:/Users/project"},
		{"C:/Users/project/src/index.ts", `C:\Users\project`},
		{"C:/Users/project/src/generated/index.ts", `C:\Users\project`},
		{"C:/repo/packages/foo/src/index.ts", "C:/repo"},
	}
	for _, test := range tests {
		resolver := NewFileConfigResolver(config, test.cwd, false)
		got := resolver.ConfigForFile(test.filePath)
		want := config.GetConfigForFile(test.filePath, test.cwd)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("filePath=%q cwd=%q: resolver=%#v direct=%#v", test.filePath, test.cwd, got, want)
		}
	}
}

func TestFileConfigResolverConcurrentShapePublication(t *testing.T) {
	RegisterAllRules()
	config := RslintConfig{{
		Files: []string{"src/**/*.ts"},
		Rules: Rules{
			"no-console":  "warn",
			"no-debugger": "error",
		},
	}}
	resolver := NewFileConfigResolver(config, "/repo", false)
	plans := make(chan *effectiveConfigPlan, 128)
	var waitGroup sync.WaitGroup
	for index := range 128 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			plans <- resolver.planForFile(fmt.Sprintf("/repo/src/file-%d.ts", index))
		}()
	}
	waitGroup.Wait()
	close(plans)

	var first *effectiveConfigPlan
	for plan := range plans {
		if plan == nil {
			t.Fatal("matching file resolved to a nil plan")
		}
		if first == nil {
			first = plan
			continue
		}
		if plan != first {
			t.Fatal("concurrent files with one shape published multiple effective plans")
		}
	}
}

type configuredRuleView struct {
	name               string
	settings           map[string]interface{}
	globals            map[string]bool
	severity           int
	requiresTypeInfo   bool
	isEslintPluginRule bool
	options            []any
	hasRun             bool
}

func configuredRuleViews(rules []linter.ConfiguredRule) []configuredRuleView {
	views := make([]configuredRuleView, len(rules))
	for index, configuredRule := range rules {
		views[index] = configuredRuleView{
			name:               configuredRule.Name,
			settings:           configuredRule.Settings,
			globals:            configuredRule.Globals,
			severity:           int(configuredRule.Severity),
			requiresTypeInfo:   configuredRule.RequiresTypeInfo,
			isEslintPluginRule: configuredRule.IsEslintPluginRule,
			options:            configuredRule.Options,
			hasRun:             configuredRule.Run != nil,
		}
	}
	return views
}

// legacyMergedConfigForTest is the pre-interning selection and merge algorithm.
// Keeping this independent test-only oracle attacks semantic drift introduced
// when the production path combines selection and entry matching into one pass.
func legacyMergedConfigForTest(
	config RslintConfig,
	filePath string,
	cwd string,
	globalIgnorePatterns []IgnorePattern,
) *MergedConfig {
	merged := &MergedConfig{
		Rules:   make(map[string]*RuleConfig),
		Plugins: make(map[string]struct{}),
	}
	if len(globalIgnorePatterns) > 0 &&
		(isDirBlockedByIgnores(filePath, globalIgnorePatterns, cwd) ||
			isFileIgnored(filePath, globalIgnorePatterns, cwd)) {
		return nil
	}
	if !isFileSelectedByConfig(config, filePath, cwd) {
		return nil
	}

	entryMatched := false
	for _, entry := range config {
		if isGlobalIgnoreEntry(entry) {
			continue
		}
		if hasFileSelectors(entry) && !isFileMatchedByConfigEntry(filePath, entry, cwd) {
			continue
		}
		if isFileIgnored(filePath, ParseIgnorePatterns(entry.Ignores), cwd) {
			continue
		}
		entryMatched = true

		for ruleName, ruleValue := range entry.Rules {
			next, hasOptions, err := parseRuleConfigValue(ruleValue)
			if err != nil {
				continue
			}
			if previous := merged.Rules[ruleName]; !hasOptions && previous != nil {
				next.Options = append([]interface{}(nil), previous.Options...)
			}
			merged.Rules[ruleName] = next
		}
		for _, plugin := range entry.Plugins {
			merged.Plugins[NormalizePluginName(plugin)] = struct{}{}
		}
		if entry.Settings != nil {
			merged.Settings = Settings(deepMergeConfigObjects(
				map[string]any(merged.Settings),
				map[string]any(entry.Settings),
			))
		}
		merged.LanguageOptions = mergeLanguageOptions(merged.LanguageOptions, entry.LanguageOptions)
	}
	if !entryMatched {
		return nil
	}
	return merged
}
