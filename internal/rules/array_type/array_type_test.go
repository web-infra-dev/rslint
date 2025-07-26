package array_type

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestArrayTypeRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ArrayTypeRule, []rule_tester.ValidTestCase{
		{Code: `
		// let a: string[] = [];
	`}}, []rule_tester.InvalidTestCase{
		{
			Code:   `let a: Array<string> = [];`,
			Output: []string{`let a: string[] = [];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorStringArray",
					Line:      1,
				},
			},
		},
	})
}
