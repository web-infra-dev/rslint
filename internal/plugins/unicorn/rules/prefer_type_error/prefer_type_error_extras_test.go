package prefer_type_error_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_type_error"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferTypeErrorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_type_error.PreferTypeErrorRule,
		[]rule_tester.ValidTestCase{
			valid("function f(Error) { if (typeof value === 'string') { throw new Error(); } }"),
			valid("if (typeof value === 'string') throw new Error();"),
			{
				Code:            "if (typeof value === 'string') { throw new (Error as any)(); }",
				FileName:        "case.ts",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			valid("if (value instanceof ns.Error) { throw new Error(); }"),
		},
		[]rule_tester.InvalidTestCase{
			invalid("if (typeof value === 'string') {} else { throw new Error(); }"),
			invalid("if (value instanceof ns['Error']) { throw new Error(); }"),
			invalid("if (!!Array.isArray(value)) { throw new Error(); }"),
			{
				Code:            "function f(TypeError) { if (typeof value === 'string') { throw new Error(); } }",
				FileName:        "case.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Output:          []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "prefer-type-error",
					Message:     preferTypeErrorMessage,
					Line:        1,
					Column:      66,
					EndLine:     1,
					EndColumn:   71,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				}},
			},
		},
	)
}
