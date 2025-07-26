package array_type

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestArrayTypeRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ArrayTypeRule, []rule_tester.ValidTestCase{
		{Code: `
		let a: Array<string> = [];
	`}}, []rule_tester.InvalidTestCase{
		// {
		// 	Code: `
		// type Foo = Array<string>;
		// `,
		// 	Errors: []rule_tester.InvalidTestCaseError{
		// 		{
		// 			MessageId: "await",
		// 			Line:      1,
		// 			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
		// 				{
		// 					MessageId: "errorStringGeneric",
		// 					Output:    " 0;",
		// 				},
		// 			},
		// 		},
		// 	},
		// },
	})
}
