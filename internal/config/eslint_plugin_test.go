package config

import (
	"encoding/json"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestRegisterEslintPluginRules_RegistersPlaceholders(t *testing.T) {
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
	if err := json.Unmarshal([]byte(`{"sourceType":"module","globals":{"foo":"readonly"},"parserOptions":{"project":["./tsconfig.json"]}}`), &lo); err != nil {
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
