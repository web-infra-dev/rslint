package config

import (
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

func TestValidateRuleOptionsAcceptsValidConfig(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		"with-schema": []any{"error", map[string]any{"allow": []any{"warn"}}},
		"no-options":  "error",
		"unmigrated":  []any{"warn", map[string]any{"whatever": true}},
	})
	if errs := ValidateRuleOptions(config, registry); len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateRuleOptionsReportsEveryFailure(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		// `allow` must be an array of strings.
		"with-schema": []any{"error", map[string]any{"allow": "warn"}},
		// A no-options rule given an option.
		"no-options": []any{"error", map[string]any{"unexpected": true}},
	})
	errs := ValidateRuleOptions(config, registry)
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

func TestValidateRuleOptionsSkipsDisabledUnknownAndUnmigrated(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		// Disabled: options would be invalid, but the rule never runs.
		"with-schema": []any{"off", map[string]any{"allow": "warn"}},
		// Unknown rule names are not this step's concern (planned separately).
		"no-such-rule": []any{"error", map[string]any{"allow": "warn"}},
		// No schema declared yet: passes through unvalidated.
		"unmigrated": []any{"error", map[string]any{"whatever": true}},
	})
	if errs := ValidateRuleOptions(config, registry); len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateRuleOptionsSurfacesSchemaCompileError(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{"broken-schema": "error"})
	errs := ValidateRuleOptions(config, registry)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].RuleName != "broken-schema" {
		t.Errorf("expected the compile failure to name the rule, got %q", errs[0].RuleName)
	}
}

func TestValidateRuleOptionsValidatesEachEntryAndDeduplicates(t *testing.T) {
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
	errs := ValidateRuleOptions(config, registry)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors (duplicate collapsed, distinct kept), got %d: %v", len(errs), errs)
	}
	for _, err := range errs {
		if err.RuleName != "with-schema" {
			t.Errorf("expected all errors to name with-schema, got %q", err.RuleName)
		}
	}
}

func TestValidateRuleOptionsAcceptsSeverityOnlyAndEmptyOptions(t *testing.T) {
	registry := newOptionsTestRegistry()
	config := configWithRules(Rules{
		// Bare severity string: options are nil → validated as [].
		"with-schema": "error",
		// Array form with severity only.
		"no-options": []any{"warn"},
	})
	if errs := ValidateRuleOptions(config, registry); len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}
