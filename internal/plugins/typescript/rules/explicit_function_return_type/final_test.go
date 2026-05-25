package explicit_function_return_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestFinalEdgeCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		// ============================================================
		// satisfies at top level is NOT a typed context (matches ESLint)
		// ============================================================
		// satisfies DOES NOT exempt the function — only `as` does
		// (The ESLint rule only treats AsExpression/TypeAssertionExpression as typed,
		//  not SatisfiesExpression)

		// ============================================================
		// Nested method shorthand in typed context
		// ============================================================
		{
			Code: `
const x: Foo = {
  a: {
    method() { return 1; },
  },
};
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// Nested arrow in new expression argument
		{
			Code: `
new Foo(() => {});
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// Deeply nested method in typed call expression
		{
			Code: `
declare function setup(opts: { hooks: { onInit: () => void } }): void;
setup({
  hooks: {
    onInit() {},
  },
});
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// module.exports = arrow — with allowExpressions, BinaryExpression parent is not excluded
		{
			Code:    `module.exports = () => {};`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		// Decorator on class method — method has return type, valid
		{
			Code: `
function decorator(target: any, key: string, desc: any): void {}
class A {
  @decorator
  method(): void {}
}
			`,
		},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// satisfies at top level — arrow still needs return type
		// ============================================================
		{
			Code:   `const fn = (() => {}) satisfies Foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 16}},
		},
		// new Foo({ callback: () => {} }) — arrow in property of object in new expression
		// isConstructorArgument is only checked at top level, not inside isPropertyOfObjectWithType
		{
			Code: `
new Foo({ callback: () => {} });
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 11}},
		},
		// Decorator on method — method without return type still needs one
		{
			Code: `
function decorator(target: any, key: string, desc: any): void {}
class A {
  @decorator
  method() {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType"}},
		},
		// allowExpressions: true + allowTypedFunctionExpressions: false
		// allowExpressions is gated behind allowTypedFunctionExpressions, so it has no effect
		{
			Code: `[() => {}];`,
			Options: map[string]interface{}{
				"allowExpressions":              true,
				"allowTypedFunctionExpressions": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType"}},
		},
	})
}
