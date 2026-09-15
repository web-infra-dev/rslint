package no_restricted_matchers

import (
	"reflect"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestNewRuleDoesNotPrepareWithoutRestrictions(t *testing.T) {
	prepared := false
	r := NewRule(Config{
		Name: "test/no-restricted-matchers",
		Prepare: func(rule.RuleContext) Runtime {
			prepared = true
			return Runtime{}
		},
	})

	if listeners := r.Run(rule.RuleContext{}, nil); len(listeners) != 0 {
		t.Fatalf("Run() returned %d listeners, want none", len(listeners))
	}
	if prepared {
		t.Fatal("Run() called Prepare without configured restrictions")
	}
}

func TestNewRuleSchema(t *testing.T) {
	r := NewRule(Config{Name: "test/no-restricted-matchers"})
	valid := [][]any{
		{},
		{map[string]any{}},
		{map[string]any{"toBe": nil, "not.toEqual": "Use another matcher"}},
	}
	for _, options := range valid {
		if err := r.Schema.Validate(options); err != nil {
			t.Errorf("Schema.Validate(%#v) returned %v", options, err)
		}
	}

	invalid := [][]any{
		{"toBe"},
		{map[string]any{"toBe": true}},
		{map[string]any{}, map[string]any{}},
	}
	for _, options := range invalid {
		if err := r.Schema.Validate(options); err == nil {
			t.Errorf("Schema.Validate(%#v) unexpectedly succeeded", options)
		}
	}
}

func TestParseOptionsUsesDeterministicSpecificityOrder(t *testing.T) {
	got := parseOptions([]any{map[string]any{
		"resolves":          nil,
		"resolves.not":      "specific",
		"rejects.not":       nil,
		"rejects.toBeFalsy": nil,
	}})

	want := []string{"rejects.not", "rejects.toBeFalsy", "resolves.not", "resolves"}
	chains := make([]string, len(got))
	for i, restriction := range got {
		chains[i] = restriction.chain
	}
	if !reflect.DeepEqual(chains, want) {
		t.Fatalf("parseOptions() chains = %v, want %v", chains, want)
	}
	if got[2].message != "specific" {
		t.Fatalf("parseOptions() message = %q, want specific", got[2].message)
	}
}

func TestParseOptionsRejectsEmptySegments(t *testing.T) {
	got := parseOptions([]any{map[string]any{
		"":       nil,
		"not.":   nil,
		".not":   nil,
		"not..x": nil,
		"not":    nil,
	}})
	if len(got) != 1 || got[0].chain != "not" {
		t.Fatalf("parseOptions() = %#v, want only not", got)
	}
}

func TestIsChainRestricted(t *testing.T) {
	modifiers := map[string]bool{"not": true, "resolves": true, "rejects": true}
	tests := []struct {
		chain       string
		restriction string
		want        bool
	}{
		// Locks in upstream isChainRestricted() arm 1: one modifier is a prefix restriction.
		{chain: "not.toBe", restriction: "not", want: true},
		{chain: "resolves.not.toBe", restriction: "resolves", want: true},
		{chain: "rejects.toThrow", restriction: "rejects", want: true},
		// Locks in upstream isChainRestricted() arm 2: a chain ending in not is a prefix restriction.
		{chain: "resolves.not.toBe", restriction: "resolves.not", want: true},
		// Locks in upstream isChainRestricted() arm 3: ordinary chains match exactly.
		{chain: "to.be.a", restriction: "to.be.a", want: true},
		{chain: "to.be.a.and.contain", restriction: "to.be.a", want: false},
		{chain: "not.toBe", restriction: "not.toBeUndefined", want: false},
	}
	for _, test := range tests {
		restriction := restrictedMatcher{chain: test.restriction, parts: strings.Split(test.restriction, ".")}
		if got := isChainRestricted(test.chain, restriction, modifiers); got != test.want {
			t.Errorf("isChainRestricted(%q, %q) = %v, want %v", test.chain, test.restriction, got, test.want)
		}
	}
}
