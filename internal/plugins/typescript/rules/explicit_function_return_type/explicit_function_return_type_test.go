package explicit_function_return_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExplicitFunctionReturnTypeRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		{Code: `function test(): void { return; }`},
		{
			Code: `
declare function fn(cb: () => void): void;
fn(() => {});
`,
			Options: []interface{}{map[string]interface{}{"allowExpressions": true}},
		},
		{
			Code: `
type Foo = () => string;
const foo: Foo = () => 'test';
`,
			Options: []interface{}{map[string]interface{}{"allowTypedFunctionExpressions": true}},
		},
		{
			Code:    `const higher = () => (): void => {};`,
			Options: []interface{}{map[string]interface{}{"allowHigherOrderFunctions": true}},
		},
		{
			Code:    `const func = (value: number) => ({ type: 'X', value }) as const;`,
			Options: []interface{}{map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true}},
		},
		{
			Code:    `const log = (message: string) => void console.log(message);`,
			Options: []interface{}{map[string]interface{}{"allowConciseArrowFunctionExpressionsStartingWithVoid": true}},
		},
		{
			Code:    `const log = (a: string) => a;`,
			Options: []interface{}{map[string]interface{}{"allowFunctionsWithoutTypeParameters": true}},
		},
		{
			Code: `
function test1() {
  return;
}
`,
			Options: []interface{}{map[string]interface{}{"allowedNames": []interface{}{"test1"}}},
		},
		{
			Code:    `const foo = (function () { return 1; })();`,
			Options: []interface{}{map[string]interface{}{"allowIIFEs": true}},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `function test() { return; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingReturnType",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `var arrowFn = () => 'test';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingReturnType",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 20,
				},
			},
		},
		{
			Code: "class Test {\n  method() {\n    return;\n  }\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingReturnType",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 9,
				},
			},
		},
		{
			Code:    `const log = <A,>(a: A) => a;`,
			Options: []interface{}{map[string]interface{}{"allowFunctionsWithoutTypeParameters": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingReturnType",
					Line:      1,
					Column:    24,
					EndLine:   1,
					EndColumn: 26,
				},
			},
		},
		{
			Code:    `const log = (message: string) => void console.log(message);`,
			Options: []interface{}{map[string]interface{}{"allowConciseArrowFunctionExpressionsStartingWithVoid": false}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingReturnType",
					Line:      1,
					Column:    31,
					EndLine:   1,
					EndColumn: 33,
				},
			},
		},
		{
			Code:    `const func = (value: number) => ({ type: 'X', value }) as const;`,
			Options: []interface{}{map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": false}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingReturnType",
					Line:      1,
					Column:    30,
					EndLine:   1,
					EndColumn: 32,
				},
			},
		},
	})
}
