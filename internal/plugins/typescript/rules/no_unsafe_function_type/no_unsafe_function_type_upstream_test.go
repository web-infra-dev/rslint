package no_unsafe_function_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnsafeFunctionTypeUpstream migrates every `valid` / `invalid` case
// from
// https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/no-unsafe-function-type.test.ts
// verbatim. Additional augmentation cases live in
// no_unsafe_function_type_extras_test.go.
func TestNoUnsafeFunctionTypeUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnsafeFunctionTypeRule, []rule_tester.ValidTestCase{
		{Code: `let value: () => void;`},
		{Code: `let value: <T>(t: T) => T;`},
		{Code: `
      // create a scope since it's illegal to declare a duplicate identifier
      // 'Function' in the global script scope.
      {
        type Function = () => void;
        let value: Function;
      }
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `let value: Function;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 12},
			},
		},
		{
			Code: `let value: Function[];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 12},
			},
		},
		{
			Code: `let value: Function | number;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 12},
			},
		},
		{
			Code: `
        class Weird implements Function {
          // ...
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 2, Column: 32},
			},
		},
		{
			Code: `
        interface Weird extends Function {
          // ...
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 2, Column: 33},
			},
		},
	})
}
