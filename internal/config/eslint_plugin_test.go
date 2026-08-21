package config

import (
	"encoding/json"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func preserveGlobalRuleRegistryEntries(t *testing.T, names ...string) {
	t.Helper()

	type savedRule struct {
		value  rule.Rule
		exists bool
	}
	saved := make(map[string]savedRule, len(names))
	for _, name := range names {
		value, exists := GlobalRuleRegistry.GetRule(name)
		saved[name] = savedRule{value: value, exists: exists}
	}

	t.Cleanup(func() {
		for name, entry := range saved {
			if entry.exists {
				GlobalRuleRegistry.rules[name] = entry.value
			} else {
				delete(GlobalRuleRegistry.rules, name)
			}
		}
	})
}

func TestRegisterEslintPluginRules_RegistersPlaceholders(t *testing.T) {
	preserveGlobalRuleRegistryEntries(t, "testplugA/no-foo", "testplugA/no-bar")

	RegisterEslintPluginRules([]EslintPluginEntry{
		{Prefix: "testplugA", RuleNames: []string{"no-foo", "no-bar"}},
	})

	r, ok := GlobalRuleRegistry.GetRule("testplugA/no-foo")
	if !ok {
		t.Fatal("expected testplugA/no-foo to be registered as a placeholder")
	}
	if !r.IsEslintPluginRule {
		t.Error("expected IsEslintPluginRule=true for a plugin placeholder")
	}
	if r.RequiresTypeInfo {
		t.Error("expected RequiresTypeInfo=false for a plugin placeholder")
	}
	if _, ok := GlobalRuleRegistry.GetRule("testplugA/no-bar"); !ok {
		t.Error("expected testplugA/no-bar to be registered")
	}
}

func TestRegisterEslintPluginRules_NativeWins(t *testing.T) {
	preserveGlobalRuleRegistryEntries(t, "testplugB/native-rule")

	// Pre-register a native rule (IsEslintPluginRule=false), then try to
	// mount a plugin rule of the same fully-qualified name.
	GlobalRuleRegistry.Register("testplugB/native-rule", rule.Rule{
		Name:               "testplugB/native-rule",
		IsEslintPluginRule: false,
	})
	RegisterEslintPluginRules([]EslintPluginEntry{
		{Prefix: "testplugB", RuleNames: []string{"native-rule"}},
	})

	r, ok := GlobalRuleRegistry.GetRule("testplugB/native-rule")
	if !ok {
		t.Fatal("expected testplugB/native-rule present")
	}
	if r.IsEslintPluginRule {
		t.Error("native rule must win: IsEslintPluginRule should stay false")
	}
}

func TestLanguageOptions_RawCaptureAndMerge(t *testing.T) {
	// UnmarshalJSON must capture the FULL raw object (sourceType / globals /
	// ecmaFeatures the Go core doesn't model) so Go can forward them to the
	// plugin worker — not just the typed ParserOptions.
	var lo LanguageOptions
	if err := json.Unmarshal([]byte(`{"sourceType":"module","ecmaVersion":2020,"globals":{"foo":"readonly"},"parserOptions":{"project":["./tsconfig.json"],"ecmaFeatures":{"globalReturn":true}}}`), &lo); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if lo.Raw["sourceType"] != "module" {
		t.Errorf("Raw should capture sourceType, got %v", lo.Raw["sourceType"])
	}
	if _, ok := lo.Raw["globals"]; !ok {
		t.Error("Raw should capture globals")
	}
	if lo.ParserOptions == nil || len(lo.ParserOptions.Project) != 1 {
		t.Error("typed ParserOptions.Project should still parse")
	}

	encoded, err := json.Marshal(lo)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var roundTripped map[string]any
	if err := json.Unmarshal(encoded, &roundTripped); err != nil {
		t.Fatalf("decode marshaled language options: %v", err)
	}
	if roundTripped["sourceType"] != "module" {
		t.Errorf("marshal lost sourceType: %s", encoded)
	}
	if _, ok := roundTripped["globals"]; !ok {
		t.Errorf("marshal lost globals: %s", encoded)
	}
	if roundTripped["ecmaVersion"] != float64(2020) {
		t.Errorf("marshal lost ecmaVersion: %s", encoded)
	}
	parserOptions, _ := roundTripped["parserOptions"].(map[string]any)
	if project, _ := parserOptions["project"].([]any); len(project) != 1 || project[0] != "./tsconfig.json" {
		t.Errorf("marshal lost typed parserOptions.project: %s", encoded)
	}
	if _, ok := parserOptions["ecmaFeatures"]; !ok {
		t.Errorf("marshal lost unknown parserOptions keys: %s", encoded)
	}

	var ecmaVersion LanguageOptions
	if err := json.Unmarshal([]byte(`{"ecmaVersion":"latest"}`), &ecmaVersion); err != nil {
		t.Fatalf("unmarshal ecmaVersion: %v", err)
	}
	if ecmaVersion.Raw["ecmaVersion"] != "latest" {
		t.Errorf("Raw should capture ecmaVersion, got %v", ecmaVersion.Raw["ecmaVersion"])
	}

	// mergeLanguageOptions: override wins per key; base-only keys retained;
	// base map not mutated (per-file merge must not corrupt the shared base).
	base := &LanguageOptions{Raw: map[string]any{"sourceType": "script", "ecmaVersion": float64(2020)}}
	override := &LanguageOptions{Raw: map[string]any{"sourceType": "module"}}
	merged := mergeLanguageOptions(base, override)
	if merged.Raw["sourceType"] != "module" {
		t.Errorf("override should win for sourceType, got %v", merged.Raw["sourceType"])
	}
	if merged.Raw["ecmaVersion"] != float64(2020) {
		t.Errorf("base-only key should be retained, got %v", merged.Raw["ecmaVersion"])
	}
	if base.Raw["sourceType"] != "script" {
		t.Errorf("base map must not be mutated, got %v", base.Raw["sourceType"])
	}
}

func TestGetEnabledRules_PluginGateAndResolution(t *testing.T) {
	preserveGlobalRuleRegistryEntries(t, "testplugC/no-null")

	RegisterEslintPluginRules([]EslintPluginEntry{
		{Prefix: "testplugC", RuleNames: []string{"no-null"}},
	})
	cwd := "/proj"

	// Plugin prefix declared (as normalizeConfig writes it into `plugins`)
	// → the rule resolves as a plugin rule.
	cfg := RslintConfig{
		{
			Plugins: []string{"testplugC"},
			Rules:   Rules{"testplugC/no-null": "error"},
		},
	}
	rules, _ := GlobalRuleRegistry.GetEnabledRules(cfg, "/proj/a.ts", cwd, true)
	if len(rules) != 1 {
		t.Fatalf("expected exactly 1 enabled rule, got %d", len(rules))
	}
	if rules[0].Name != "testplugC/no-null" {
		t.Errorf("expected testplugC/no-null, got %q", rules[0].Name)
	}
	if !rules[0].IsEslintPluginRule {
		t.Error("expected IsEslintPluginRule=true on the resolved rule")
	}
	if rules[0].Severity != rule.SeverityError {
		t.Errorf("expected SeverityError, got %v", rules[0].Severity)
	}

	// Prefix NOT declared in `plugins` → the plugin gate (enforcePlugins=true)
	// drops the rule entirely.
	cfgNoGate := RslintConfig{
		{
			Rules: Rules{"testplugC/no-null": "error"},
		},
	}
	rulesNoGate, _ := GlobalRuleRegistry.GetEnabledRules(cfgNoGate, "/proj/a.ts", cwd, true)
	if len(rulesNoGate) != 0 {
		t.Errorf("expected the gate to drop the rule when its prefix is not declared, got %d", len(rulesNoGate))
	}
}

// TestGetEnabledRules_SplitEntryNativeAndCommunity pins the documented combine
// workflow: native plugins in one (array-form) entry and community plugins in a
// separate entry. For a file matching both, GetEnabledRules must return BOTH
// rules, each carrying the correct IsEslintPluginRule routing flag (native runs
// in Go, community routes to the worker).
func TestGetEnabledRules_SplitEntryNativeAndCommunity(t *testing.T) {
	preserveGlobalRuleRegistryEntries(t, "testplugSplit/no-foo")

	RegisterAllRules()
	RegisterEslintPluginRules([]EslintPluginEntry{
		{Prefix: "testplugSplit", RuleNames: []string{"no-foo"}},
	})
	cfg := RslintConfig{
		{
			Plugins: []string{"@typescript-eslint"},
			Rules:   Rules{"@typescript-eslint/no-explicit-any": "error"},
		},
		{
			Plugins: []string{"testplugSplit"},
			Rules:   Rules{"testplugSplit/no-foo": "error"},
		},
	}
	rules, _ := GlobalRuleRegistry.GetEnabledRules(cfg, "/proj/a.ts", "/proj", true)

	routing := map[string]bool{} // rule name -> IsEslintPluginRule
	for _, r := range rules {
		routing[r.Name] = r.IsEslintPluginRule
	}

	native, hasNative := routing["@typescript-eslint/no-explicit-any"]
	if !hasNative {
		t.Fatalf("native rule @typescript-eslint/no-explicit-any missing from %v", routing)
	}
	if native {
		t.Error("native rule must have IsEslintPluginRule=false (runs in Go)")
	}

	community, hasCommunity := routing["testplugSplit/no-foo"]
	if !hasCommunity {
		t.Fatalf("community rule testplugSplit/no-foo missing from %v", routing)
	}
	if !community {
		t.Error("community rule must have IsEslintPluginRule=true (routes to the worker)")
	}
}

// Config resolution retains native and community rules independently of source
// capabilities. The linter plan owns type-aware eligibility.
func TestGetEnabledRulesKeepsNativeAndCommunityRules(t *testing.T) {
	preserveGlobalRuleRegistryEntries(t, "unicornGap/no-null")

	RegisterAllRules()
	RegisterEslintPluginRules([]EslintPluginEntry{
		{Prefix: "unicornGap", RuleNames: []string{"no-null"}},
	})
	cfg := RslintConfig{
		{
			Plugins:         []string{"@typescript-eslint"},
			LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{Project: []string{"./tsconfig.json"}}},
			Rules:           Rules{"@typescript-eslint/require-await": "error"},
		},
		{
			Plugins: []string{"unicornGap"},
			Rules:   Rules{"unicornGap/no-null": "error"},
		},
	}
	gap, _ := GlobalRuleRegistry.GetEnabledRules(cfg, "/proj/gap.ts", "/proj", true)
	gapRouting := map[string]bool{}
	for _, r := range gap {
		gapRouting[r.Name] = r.IsEslintPluginRule
	}
	if _, hasNative := gapRouting["@typescript-eslint/require-await"]; !hasNative {
		t.Error("config resolution must retain the type-aware native rule")
	}
	community, hasCommunity := gapRouting["unicornGap/no-null"]
	if !hasCommunity {
		t.Fatalf("community rule must survive on a gap file; got %v", gapRouting)
	}
	if !community {
		t.Error("the surviving community rule must keep IsEslintPluginRule=true (routes to the worker)")
	}

	covered, _ := GlobalRuleRegistry.GetEnabledRules(cfg, "/proj/covered.ts", "/proj", true)
	coveredNames := map[string]bool{}
	for _, r := range covered {
		coveredNames[r.Name] = true
	}
	if !coveredNames["@typescript-eslint/require-await"] {
		t.Error("type-aware native rule must be kept on a covered file")
	}
	if !coveredNames["unicornGap/no-null"] {
		t.Error("community rule must be present on a covered file too")
	}
}

// TestGetConfigForFile_SplitEntry_ProjectFromNativeEntrySurvives pins that in
// the split-entry combine workflow, parserOptions.project declared only on the
// native (array-form) entry survives the merge — so the type-aware native rule
// still gets type info even though the community (object-form) entry carries no
// languageOptions.
func TestGetConfigForFile_SplitEntry_ProjectFromNativeEntrySurvives(t *testing.T) {
	cfg := RslintConfig{
		{
			Plugins:         []string{"@typescript-eslint"},
			LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{Project: []string{"./tsconfig.json"}}},
			Rules:           Rules{"@typescript-eslint/require-await": "error"},
		},
		{
			Plugins: []string{"unicornProj"},
			Rules:   Rules{"unicornProj/no-null": "error"},
		},
	}
	merged := cfg.GetConfigForFile("/proj/a.ts", "/proj")
	if merged == nil {
		t.Fatal("merged config should not be nil")
		return
	}
	if merged.LanguageOptions == nil || merged.LanguageOptions.ParserOptions == nil {
		t.Fatal("merged languageOptions/parserOptions must survive from the native entry")
		return
	}
	if len(merged.LanguageOptions.ParserOptions.Project) != 1 ||
		merged.LanguageOptions.ParserOptions.Project[0] != "./tsconfig.json" {
		t.Errorf("project must survive the merge, got %v", merged.LanguageOptions.ParserOptions.Project)
	}
}

// TestPluginMergedMaps pins the three branches the plugin dispatch relies on:
// nil merged → both nil; nil LanguageOptions → languageOptions nil but settings
// forwarded; both present → both forwarded.
func TestPluginMergedMaps(t *testing.T) {
	if lo, s := PluginMergedMaps(nil); lo != nil || s != nil {
		t.Errorf("nil merged -> (nil,nil), got (%v,%v)", lo, s)
	}

	if lo, s := PluginMergedMaps(&MergedConfig{Settings: Settings{"k": 1}}); lo != nil || s["k"] != 1 {
		t.Errorf("nil LanguageOptions -> (nil, settings), got (%v,%v)", lo, s)
	}

	raw := map[string]any{"sourceType": "module"}
	lo, s := PluginMergedMaps(&MergedConfig{
		Settings:        Settings{"foo": "bar"},
		LanguageOptions: &LanguageOptions{Raw: raw},
	})
	if lo["sourceType"] != "module" {
		t.Errorf("languageOptions = %v, want the raw map", lo)
	}
	if s["foo"] != "bar" {
		t.Errorf("settings = %v, want forwarded", s)
	}
}
