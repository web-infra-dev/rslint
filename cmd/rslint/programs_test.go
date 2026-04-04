package main

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestFilterNonTypeAwareRules(t *testing.T) {
	rules := []linter.ConfiguredRule{
		{Name: "syntax-rule", RequiresTypeInfo: false},
		{Name: "type-rule", RequiresTypeInfo: true},
		{Name: "another-syntax", RequiresTypeInfo: false},
	}

	filtered := linter.FilterNonTypeAwareRules(rules)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(filtered))
	}
	if filtered[0].Name != "syntax-rule" {
		t.Errorf("expected syntax-rule, got %s", filtered[0].Name)
	}
	if filtered[1].Name != "another-syntax" {
		t.Errorf("expected another-syntax, got %s", filtered[1].Name)
	}
}

func TestFilterNonTypeAwareRules_AllTypeAware(t *testing.T) {
	rules := []linter.ConfiguredRule{
		{Name: "type-rule-1", RequiresTypeInfo: true},
		{Name: "type-rule-2", RequiresTypeInfo: true},
	}

	filtered := linter.FilterNonTypeAwareRules(rules)

	if len(filtered) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(filtered))
	}
}

func TestFilterNonTypeAwareRules_NoneTypeAware(t *testing.T) {
	rules := []linter.ConfiguredRule{
		{Name: "rule-a", RequiresTypeInfo: false},
		{Name: "rule-b", RequiresTypeInfo: false},
	}

	filtered := linter.FilterNonTypeAwareRules(rules)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(filtered))
	}
}

func TestFilterNonTypeAwareRules_Empty(t *testing.T) {
	filtered := linter.FilterNonTypeAwareRules(nil)
	if len(filtered) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(filtered))
	}
}

// Verify RequiresTypeInfo propagates through CreateRule.
func TestRequiresTypeInfo_Propagation(t *testing.T) {
	r := rule.CreateRule(rule.Rule{
		Name:             "test-rule",
		RequiresTypeInfo: true,
		Run:              func(ctx rule.RuleContext, options any) rule.RuleListeners { return nil },
	})

	if r.Name != "@typescript-eslint/test-rule" {
		t.Errorf("unexpected name: %s", r.Name)
	}
	if !r.RequiresTypeInfo {
		t.Error("RequiresTypeInfo should be true after CreateRule")
	}
}
