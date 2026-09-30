package valid_title

import (
	"testing"

	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func TestSemanticName(t *testing.T) {
	cfg := Config{LegacyAliases: map[string]string{"fit": "it", "xdescribe": "describe"}}
	cases := []struct {
		parsed testFramework.ParsedCall
		want   string
	}{
		{testFramework.ParsedCall{Name: "test", Kind: testFramework.FnKindTest}, "test"},
		{testFramework.ParsedCall{Name: "fit", Kind: testFramework.FnKindTest}, "it"},
		{testFramework.ParsedCall{Name: "xdescribe", Kind: testFramework.FnKindDescribe}, "describe"},
		// Playwright keeps the root API name on a describe registration.
		{testFramework.ParsedCall{Name: "test", Kind: testFramework.FnKindDescribe}, "describe"},
		// Without an alias table the name is taken as is.
		{testFramework.ParsedCall{Name: "xit", Kind: testFramework.FnKindTest}, "xit"},
	}
	for _, c := range cases {
		if got := cfg.semanticName(&c.parsed); got != c.want {
			t.Errorf("semanticName(%s, kind %v) = %q, want %q", c.parsed.Name, c.parsed.Kind, got, c.want)
		}
	}
}

func TestDuplicatePrefixReplacement(t *testing.T) {
	cases := []struct {
		raw, name, want string
		ok              bool
	}{
		{`"test foo"`, "test", `"foo"`, true},
		{`'Test foo'`, "test", `'foo'`, true},
		{"`TEST foo`", "test", "`foo`", true},
		// Nothing follows the name, so there is no prefix to remove.
		{`"test"`, "test", "", false},
		// Escaped spelling is reported but never edited.
		{`"te\u0073t foo"`, "test", "", false},
		{`"test\u0020foo"`, "test", "", false},
		{`"foo"`, "", "", false},
	}
	for _, c := range cases {
		got, ok := duplicatePrefixReplacement(c.raw, c.name)
		if got != c.want || ok != c.ok {
			t.Errorf("duplicatePrefixReplacement(%s, %q) = (%q, %v), want (%q, %v)", c.raw, c.name, got, ok, c.want, c.ok)
		}
	}
}

func TestAccidentalSpaceReplacement(t *testing.T) {
	cases := map[string]string{
		`"  foo  "`: `"foo"`,
		"` foo`":    "`foo`",
		`'foo '`:    `'foo'`,
		`"foo"`:     `"foo"`,
		`"\u0020f"`: `"\u0020f"`,
	}
	for raw, want := range cases {
		if got := accidentalSpaceReplacement(raw); got != want {
			t.Errorf("accidentalSpaceReplacement(%s) = %s, want %s", raw, got, want)
		}
	}
}

func TestParseCompiledOptions(t *testing.T) {
	co := parseCompiledOptions([]any{map[string]any{
		"mustMatch":    map[string]any{"describe": "^a", "test": []any{"^b", "custom"}},
		"mustNotMatch": "^c",
	}})
	if len(co.invalidPatterns) != 0 {
		t.Fatalf("unexpected invalid patterns: %v", co.invalidPatterns)
	}
	if co.mustMatch.describe.re == nil || co.mustMatch.it.re != nil {
		t.Error("an object form must only fill the named groups")
	}
	if co.mustMatch.test.customText != "custom" {
		t.Errorf("custom text = %q", co.mustMatch.test.customText)
	}
	if co.mustNotMatch.describe.re == nil || co.mustNotMatch.test.re == nil || co.mustNotMatch.it.re == nil {
		t.Error("a string form must fill every group")
	}

	bad := parseCompiledOptions([]any{map[string]any{"mustMatch": map[string]any{"it": "("}}})
	if len(bad.invalidPatterns) != 1 || bad.invalidPatterns[0].optionPath != "mustMatch.it" {
		t.Errorf("invalid patterns = %v", bad.invalidPatterns)
	}
}

func TestMatcherForHasNoFallbackGroup(t *testing.T) {
	ms := parseCompiledOptions([]any{map[string]any{"mustMatch": "^a"}}).mustMatch
	if matcherFor("it", ms).re == nil {
		t.Error("it must resolve its group")
	}
	if matcherFor("bench", ms).re != nil {
		t.Error("an unknown API must not inherit the it group")
	}
}
