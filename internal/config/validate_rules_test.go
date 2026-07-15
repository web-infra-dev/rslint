package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// newOptionsTestRegistry builds a registry with a representative mix of
// schema situations: a rule with a real options schema, a no-options rule on
// the shared EmptyArraySchema, a not-yet-migrated rule without a schema, and
// a rule whose schema JSON is broken (an authoring bug).
func newOptionsTestRegistry() *RuleRegistry {
	registry := NewRuleRegistry()
	noopRun := func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	}
	registry.Register("with-schema", rule.Rule{
		Name: "with-schema",
		Schema: rule.NewSchema([]byte(`{
			"type": "array",
			"items": [
				{
					"type": "object",
					"properties": {
						"allow": { "type": "array", "items": { "type": "string" } }
					},
					"additionalProperties": false
				}
			],
			"minItems": 0,
			"maxItems": 1
		}`)),
		Run: noopRun,
	})
	registry.Register("no-options", rule.Rule{
		Name:   "no-options",
		Schema: rule.EmptyArraySchema,
		Run:    noopRun,
	})
	registry.Register("unmigrated", rule.Rule{
		Name: "unmigrated",
		Run:  noopRun,
	})
	registry.Register("broken-schema", rule.Rule{
		Name:   "broken-schema",
		Schema: rule.NewSchema([]byte(`not json`)),
		Run:    noopRun,
	})
	return registry
}

func configWithRules(rules Rules) RslintConfig {
	return RslintConfig{ConfigEntry{Rules: rules}}
}

func TestValidateRulesAcceptsValidConfig(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		"with-schema": []any{"error", map[string]any{"allow": []any{"warn"}}},
		"no-options":  "error",
		"unmigrated":  []any{"warn", map[string]any{"whatever": true}},
	})
	if errs := ValidateRules(config, registry, nil); len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateRulesReportsEveryOptionsFailure(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		// `allow` must be an array of strings.
		"with-schema": []any{"error", map[string]any{"allow": "warn"}},
		// A no-options rule given an option.
		"no-options": []any{"error", map[string]any{"unexpected": true}},
	})
	errs := ValidateRules(config, registry, nil)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
	// Deterministic rule-name order.
	if errs[0].RuleName != "no-options" || errs[1].RuleName != "with-schema" {
		t.Errorf("expected errors sorted by rule name, got %q, %q", errs[0].RuleName, errs[1].RuleName)
	}
	if !strings.Contains(errs[1].Error(), `"with-schema"`) {
		t.Errorf("expected the message to name the rule, got: %v", errs[1])
	}
}

func TestValidateRulesReportsUnknownRuleNames(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		"no-such-rule":             "error",
		"import/no-such-rule":      "error",
		"some-plugin/no-such-rule": "error",
	})
	errs := ValidateRules(config, registry, nil)
	if len(errs) != 3 {
		t.Fatalf("expected 3 unknown-rule errors, got %d: %v", len(errs), errs)
	}
	want := []string{
		`unknown rule "import/no-such-rule": plugin "import" has no such rule`,
		`unknown rule "no-such-rule"`,
		`unknown rule "some-plugin/no-such-rule": plugin "some-plugin" is not registered`,
	}
	for i, message := range want {
		if errs[i].Error() != message {
			t.Errorf("unexpected message at %d:\n got: %s\nwant: %s", i, errs[i].Error(), message)
		}
	}
}

func TestValidateRulesReportsUnknownNamesAlongsideOptionsFailures(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		"no-such-rule": "error",
		"with-schema":  []any{"error", map[string]any{"allow": "warn"}},
	})
	errs := ValidateRules(config, registry, nil)
	if len(errs) != 2 {
		t.Fatalf("expected 1 unknown-name + 1 options error, got %d: %v", len(errs), errs)
	}
	if errs[0].RuleName != "no-such-rule" || !strings.Contains(errs[0].Error(), "unknown rule") {
		t.Errorf("expected the unknown-name error first, got: %v", errs[0])
	}
	if errs[1].RuleName != "with-schema" || !strings.Contains(errs[1].Error(), "invalid options") {
		t.Errorf("expected the options error second, got: %v", errs[1])
	}
}

func TestValidateRulesResolvesMountedPluginRules(t *testing.T) {
	registry := newOptionsTestRegistry()
	mounted := []EslintPluginEntry{
		{Prefix: "import-x", RuleNames: []string{"no-unresolved"}},
	}
	config := configWithRules(Rules{
		"import-x/no-unresolved": "error", // mounted → known
		"import-x/no-such-rule":  "error", // typo within a mounted plugin
	})
	errs := ValidateRules(config, registry, mounted)
	if len(errs) != 1 {
		t.Fatalf("expected 1 unknown-rule error, got %d: %v", len(errs), errs)
	}
	want := `unknown rule "import-x/no-such-rule": plugin "import-x" has no such rule`
	if errs[0].Error() != want {
		t.Errorf("unexpected message:\n got: %s\nwant: %s", errs[0].Error(), want)
	}
}

func TestValidateRulesIgnoresStalePluginPlaceholders(t *testing.T) {
	// A long-lived process may still hold placeholder registrations from an
	// earlier config or request; whether a mounted rule resolves is decided
	// by the current config's own plugin entries, not the registry.
	registry := newOptionsTestRegistry()
	registry.Register("stale/rule", rule.Rule{
		Name:               "stale/rule",
		IsEslintPluginRule: true,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			return rule.RuleListeners{}
		},
	})
	config := configWithRules(Rules{"stale/rule": "error"})
	errs := ValidateRules(config, registry, nil)
	if len(errs) != 1 {
		t.Fatalf("expected the stale placeholder to be unknown, got %d: %v", len(errs), errs)
	}
	if errs[0].RuleName != "stale/rule" {
		t.Errorf("expected unknown rule %q, got %q", "stale/rule", errs[0].RuleName)
	}
}

func TestValidateRulesSkipsDisabledUnknownRules(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		// Disabled entries are exempt from both checks, matching ESLint.
		"no-such-rule":    "off",
		"another-bad-one": []any{"off"},
	})
	if errs := ValidateRules(config, registry, nil); len(errs) != 0 {
		t.Errorf("expected no errors for disabled rules, got: %v", errs)
	}
}

func TestValidateRulesDeduplicatesUnknownNamesAcrossEntries(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := RslintConfig{
		ConfigEntry{Rules: Rules{"no-such-rule": "error"}},
		ConfigEntry{Rules: Rules{"no-such-rule": "warn"}},
	}
	errs := ValidateRules(config, registry, nil)
	if len(errs) != 1 {
		t.Fatalf("expected 1 deduplicated error, got %d: %v", len(errs), errs)
	}
}

func TestValidateRuleOptionsSkipsUnknownNames(t *testing.T) {
	// The options-only variant (used by the LSP) must keep tolerating
	// unresolved names: a degraded plugin host commits a catalog without
	// plugin entries, and rejecting the config over them would be wrong.
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		"no-such-rule": []any{"error", map[string]any{"allow": "warn"}},
		"with-schema":  []any{"error", map[string]any{"allow": "warn"}},
	})
	errs := ValidateRuleOptions(config, registry)
	if len(errs) != 1 || errs[0].RuleName != "with-schema" {
		t.Fatalf("expected only the with-schema options error, got: %v", errs)
	}
}

func TestValidateRulesSkipsDisabledAndUnmigratedOptions(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		// Disabled: options would be invalid, but the rule never runs.
		"with-schema": []any{"off", map[string]any{"allow": "warn"}},
		// No schema declared yet: passes through unvalidated.
		"unmigrated": []any{"error", map[string]any{"whatever": true}},
	})
	if errs := ValidateRules(config, registry, nil); len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateRulesSurfacesSchemaCompileError(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{"broken-schema": "error"})
	errs := ValidateRules(config, registry, nil)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].RuleName != "broken-schema" {
		t.Errorf("expected the compile failure to name the rule, got %q", errs[0].RuleName)
	}
}

func TestValidateRulesValidatesEachEntryAndDeduplicates(t *testing.T) {
	registry := newOptionsTestRegistry()
	badOptions := []any{"error", map[string]any{"allow": "warn"}}
	config := RslintConfig{
		// The identical bad entry twice: reported once.
		ConfigEntry{Rules: Rules{"with-schema": badOptions}},
		ConfigEntry{Rules: Rules{"with-schema": badOptions}},
		// A different bad entry for the same rule: reported separately.
		ConfigEntry{Rules: Rules{"with-schema": []any{"error", map[string]any{"allow": 42}}}},
		// A valid entry for the same rule doesn't mask the bad ones.
		ConfigEntry{Rules: Rules{"with-schema": []any{"error", map[string]any{"allow": []any{"warn"}}}}},
	}
	errs := ValidateRules(config, registry, nil)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors (duplicate collapsed, distinct kept), got %d: %v", len(errs), errs)
	}
	for _, err := range errs {
		if err.RuleName != "with-schema" {
			t.Errorf("expected all errors to name with-schema, got %q", err.RuleName)
		}
	}
}

func TestValidateRulesFillsSchemaDefaultsIntoConfig(t *testing.T) {
	// The useDefaults contract: validation mutates the options in place, and
	// because the validated options slice aliases the raw config entry value,
	// the defaults are visible to the per-file config merge afterwards.
	registry := NewRuleRegistry()
	registry.Register("with-defaults", rule.Rule{
		Name: "with-defaults",
		Schema: rule.NewSchema([]byte(`{
			"type": "array",
			"items": [
				{
					"type": "object",
					"properties": {
						"allow": { "type": "array", "items": { "type": "string" }, "default": [] },
						"strict": { "type": "boolean", "default": true }
					},
					"additionalProperties": false
				}
			],
			"minItems": 0,
			"maxItems": 1
		}`)),
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			return rule.RuleListeners{}
		},
	})

	options := map[string]any{"strict": false}
	config := configWithRules(Rules{"with-defaults": []any{"error", options}})
	if errs := ValidateRules(config, registry, nil); len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}

	want := map[string]any{"allow": []any{}, "strict": false}
	if !reflect.DeepEqual(options, want) {
		t.Errorf("expected defaults filled into the config's own options, got %#v, want %#v", options, want)
	}
}

func TestValidateRulesAcceptsSeverityOnlyAndEmptyOptions(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		// Bare severity string: options are nil → validated as [].
		"with-schema": "error",
		// Array form with severity only.
		"no-options": []any{"warn"},
	})
	if errs := ValidateRules(config, registry, nil); len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}
