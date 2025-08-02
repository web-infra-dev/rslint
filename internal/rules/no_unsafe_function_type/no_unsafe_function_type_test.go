package no_unsafe_function_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoUnsafeFunctionType(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnsafeFunctionTypeRule, []rule_tester.ValidTestCase{
		{
			Code: `let value: () => void;`,
		},
		{
			Code: `let value: <T>(t: T) => T;`,
		},
		{
			Code: `
// create a scope since it's illegal to declare a duplicate identifier
// 'Function' in the global script scope.
{
  type Function = () => void;
  let value: Function;
}`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `let value: Function;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedFunctionType",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: Function[];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedFunctionType",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: Function | number;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedFunctionType",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `
class Weird implements Function {
  // ...
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedFunctionType",
					Line:      2,
					Column:    24,
				},
			},
		},
		{
			Code: `
interface Weird extends Function {
  // ...
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedFunctionType",
					Line:      2,
					Column:    25,
				},
			},
		},
	})
}
