package dot_notation

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestDotNotationRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DotNotationRule, []rule_tester.ValidTestCase{
		// Valid cases
		{Code: "a.b;"},
		{Code: "a['12'];"},
		{Code: "a[b];"},
		{Code: "a[0];"},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases
		{
			Code: "a['b'];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useDot",
					Line:      1,
					Column:    3,
				},
			},
			Output: []string{"a.b;"},
		},
		{
			Code: "a['test'];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useDot",
					Line:      1,
					Column:    3,
				},
			},
			Output: []string{"a.test;"},
		},
	})
}
