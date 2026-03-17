package prefer_rest_params

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferRestParamsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferRestParamsRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `arguments;`},
			{Code: `function foo(arguments: any) { arguments; }`},
			{Code: `function foo() { var arguments: any; arguments; }`},
			{Code: `var foo = () => arguments;`},
			{Code: `function foo(...args: any[]) { args; }`},
			{Code: `function foo() { arguments.length; }`},
			{Code: `function foo() { arguments.callee; }`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `function foo() { arguments; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 18},
				},
			},
			{
				Code: `function foo() { arguments[0]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 18},
				},
			},
			{
				Code: `function foo() { arguments[1]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 18},
				},
			},
		},
	)
}
