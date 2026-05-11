package config

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// ──────────────────────────────────────────────────────────────────────
// EslintPluginEntry / FullRuleNames
// ──────────────────────────────────────────────────────────────────────

func TestEslintPluginEntry_FullRuleNames(t *testing.T) {
	e := EslintPluginEntry{
		Prefix:    "uc",
		RuleNames: []string{"no-null", "prefer-array-some"},
	}
	got := e.FullRuleNames()
	want := []string{"uc/no-null", "uc/prefer-array-some"}
	if len(got) != len(want) {
		t.Fatalf("FullRuleNames len: got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("FullRuleNames[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}

// ──────────────────────────────────────────────────────────────────────
// RegisterEslintPluginRules — placeholder registration
// ──────────────────────────────────────────────────────────────────────

func TestRegisterEslintPluginRules_PlaceholderRegistration(t *testing.T) {
	// Unique prefix isolates this test from any prior global registry state.
	prefix := "spike1__"
	entries := []EslintPluginEntry{
		{
			Prefix:    prefix,
			RuleNames: []string{"alpha", "beta"},
		},
	}
	RegisterEslintPluginRules(entries)

	for _, ruleName := range []string{prefix + "/alpha", prefix + "/beta"} {
		r, ok := GlobalRuleRegistry.GetRule(ruleName)
		if !ok {
			t.Fatalf("rule %q not registered", ruleName)
		}
		if !r.IsEslintPluginRule {
			t.Errorf("rule %q: IsEslintPluginRule=false, want true", ruleName)
		}
		if r.RequiresTypeInfo {
			t.Errorf("rule %q: RequiresTypeInfo=true, want false", ruleName)
		}
		// Run is a no-op placeholder; calling it must not panic and must
		// return nil listeners.
		if listeners := r.Run(rule.RuleContext{}, nil); listeners != nil {
			t.Errorf("rule %q: placeholder Run should return nil, got %v", ruleName, listeners)
		}
	}
}

func TestRegisterEslintPluginRules_SkipsEmptyPrefix(t *testing.T) {
	// Empty prefix → log + skip; no panic, no rule registered.
	entries := []EslintPluginEntry{
		{Prefix: "", RuleNames: []string{"alpha"}},
	}
	RegisterEslintPluginRules(entries) // must not panic
	if _, ok := GlobalRuleRegistry.GetRule("/alpha"); ok {
		t.Errorf("entry with empty prefix should not register a rule")
	}
}

// TestRegisterEslintPluginRules_PlaceholderDedupe verifies that registering
// the SAME placeholder twice (which happens whenever multiple ConfigEntry
// instances declare the same plugin under the same prefix) is a silent
// no-op for the second call — and crucially, that it doesn't get
// mis-classified as native shadowing (which used to produce a misleading
// "native implementation exists" stderr line in monorepo multi-config
// configurations).
func TestRegisterEslintPluginRules_PlaceholderDedupe(t *testing.T) {
	prefix := "dedupe_test__"
	ruleName := "rule-x"
	full := prefix + "/" + ruleName

	entries := []EslintPluginEntry{
		{Prefix: prefix, RuleNames: []string{ruleName}},
	}
	RegisterEslintPluginRules(entries)
	got1, ok1 := GlobalRuleRegistry.GetRule(full)
	if !ok1 {
		t.Fatalf("first register should have created a placeholder")
	}
	if !got1.IsEslintPluginRule {
		t.Fatalf("first register: IsEslintPluginRule=%v, want true", got1.IsEslintPluginRule)
	}

	// Capture stderr and re-register the same entry — the function should
	// neither panic nor write to stderr (no "native takes precedence" line
	// because the existing entry is itself a placeholder).
	r, w, _ := os.Pipe()
	origStderr := os.Stderr
	os.Stderr = w
	RegisterEslintPluginRules(entries)
	_ = w.Close()
	os.Stderr = origStderr
	captured, _ := io.ReadAll(r)
	if len(captured) != 0 {
		t.Errorf("re-registering a placeholder must be silent; got stderr: %q", string(captured))
	}

	// And the original placeholder is still there with its IsEslintPluginRule flag intact.
	got2, ok2 := GlobalRuleRegistry.GetRule(full)
	if !ok2 || !got2.IsEslintPluginRule {
		t.Errorf("after dedupe, placeholder must remain registered with IsEslintPluginRule=true")
	}
}

// TestRegisterEslintPluginRules_NativeShadowingDoesWarn covers the OTHER
// branch of the same conditional: when a NATIVE rule already owns the
// fully-qualified name, the plugin's same-named rule must be skipped AND
// the user must see a clear stderr warning so the shadowing isn't silent.
func TestRegisterEslintPluginRules_NativeShadowingDoesWarn(t *testing.T) {
	prefix := "shadow_test__"
	ruleName := "native-rule"
	full := prefix + "/" + ruleName

	// Pre-register a fake "native" rule at this fully-qualified name —
	// IsEslintPluginRule=false to mark it as native.
	GlobalRuleRegistry.Register(full, rule.Rule{
		Name:               full,
		RequiresTypeInfo:   false,
		IsEslintPluginRule: false,
		Run: func(_ rule.RuleContext, _ any) rule.RuleListeners {
			return nil
		},
	})

	r, w, _ := os.Pipe()
	origStderr := os.Stderr
	os.Stderr = w
	RegisterEslintPluginRules([]EslintPluginEntry{
		{Prefix: prefix, RuleNames: []string{ruleName}},
	})
	_ = w.Close()
	os.Stderr = origStderr
	captured, _ := io.ReadAll(r)
	captStr := string(captured)
	if !strings.Contains(captStr, full) {
		t.Errorf("warning must name the fully-qualified rule %q; got %q", full, captStr)
	}
	if !strings.Contains(captStr, "native") {
		t.Errorf("warning must mention native precedence; got %q", captStr)
	}

	// And the native rule is still in the registry, untouched.
	got, _ := GlobalRuleRegistry.GetRule(full)
	if got.IsEslintPluginRule {
		t.Errorf("native rule must NOT be overwritten by placeholder; IsEslintPluginRule=%v", got.IsEslintPluginRule)
	}
}

// ──────────────────────────────────────────────────────────────────────
// GetConfigForFile — EslintPlugins flow into MergedConfig as union
// ──────────────────────────────────────────────────────────────────────

func TestGetConfigForFile_EslintPluginsAppended(t *testing.T) {
	// Go side does NOT coalesce by prefix anymore — that's the runner's
	// (cli.ts → engine.ts) job. Each ConfigEntry's EslintPlugins is appended
	// into the merged config; same prefix appearing twice is preserved so
	// per-config dispatch can route each occurrence to its own plugin
	// instance.
	cfg := RslintConfig{
		{
			Files: []string{"**/*.ts"},
			EslintPlugins: []EslintPluginEntry{
				{Prefix: "uc", RuleNames: []string{"no-null"}},
			},
			Rules: Rules{"uc/no-null": "error"},
		},
		{
			Files: []string{"**/*.ts"},
			EslintPlugins: []EslintPluginEntry{
				{Prefix: "uc", RuleNames: []string{"prefer-array-some"}},
			},
		},
	}
	merged := cfg.GetConfigForFile("/p/test.ts", "/p")
	if merged == nil {
		t.Fatalf("expected non-nil merged config")
	}
	if len(merged.EslintPlugins) != 2 {
		t.Fatalf("expected 2 entries (Go appends, does not coalesce), got %d", len(merged.EslintPlugins))
	}
	if merged.EslintPlugins[0].Prefix != "uc" || merged.EslintPlugins[1].Prefix != "uc" {
		t.Errorf("both entries should preserve prefix=uc, got %v / %v",
			merged.EslintPlugins[0].Prefix, merged.EslintPlugins[1].Prefix)
	}
	// Each entry must carry ITS OWN ruleNames — a buggy "append but also
	// secretly union" path would put ['no-null','prefer-array-some'] on
	// both. Order is the order ConfigEntries appeared above.
	if len(merged.EslintPlugins[0].RuleNames) != 1 ||
		merged.EslintPlugins[0].RuleNames[0] != "no-null" {
		t.Errorf("entry[0].RuleNames: got %v, want [no-null]",
			merged.EslintPlugins[0].RuleNames)
	}
	if len(merged.EslintPlugins[1].RuleNames) != 1 ||
		merged.EslintPlugins[1].RuleNames[0] != "prefer-array-some" {
		t.Errorf("entry[1].RuleNames: got %v, want [prefer-array-some]",
			merged.EslintPlugins[1].RuleNames)
	}
}

func TestIsGlobalIgnoreEntry_NotTriggeredByEslintPlugins(t *testing.T) {
	// An entry that has only `ignores` and `eslintPlugins` is NOT a global
	// ignore entry — eslintPlugins counts as a meaningful field.
	entry := ConfigEntry{
		Ignores: []string{"node_modules/**"},
		EslintPlugins: []EslintPluginEntry{
			{Prefix: "uc"},
		},
	}
	if isGlobalIgnoreEntry(entry) {
		t.Errorf("entry with eslintPlugins must not be classified as global-ignore-only")
	}
}
