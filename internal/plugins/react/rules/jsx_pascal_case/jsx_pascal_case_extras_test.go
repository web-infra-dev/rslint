package jsx_pascal_case

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Configured patterns preserve minimatch negation and expansion semantics.
func TestJsxPascalCaseRuleMinimatchPatterns(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxPascalCaseRule, []rule_tester.ValidTestCase{
		{
			Code:    `<Foo_Active />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{"Foo_!(Deprecated|Ignored)"}},
		},
		{
			Code:    `<FOO2 />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{"FOO{1..3}"}},
		},
		{
			Code:    `<OTHER />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{"!Foo"}},
		},
		{
			Code:    `<Foo_Active />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{" Foo_!(Deprecated|Ignored) "}},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    `<Foo_Deprecated />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{"Foo_!(Deprecated|Ignored)"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usePascalCase"}},
		},
	})
}

func TestIgnoreCheckMinimatchOptions(t *testing.T) {
	cfg := parseOptions([]any{map[string]interface{}{"ignore": []interface{}{"a/**/c", "#literal"}}})
	if ignoreCheck(cfg.ignore, "a/c") {
		t.Error("ignore must disable globstar, as upstream's noglobstar option does")
	}
	if !ignoreCheck(cfg.ignore, "#literal") {
		t.Error("an exact ignore entry must match even when minimatch treats it as a comment")
	}
}
