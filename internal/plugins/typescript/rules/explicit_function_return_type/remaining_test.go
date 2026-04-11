package explicit_function_return_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRemainingEdgeCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Setter in object literal — always valid (no return type needed)
		// ============================================================
		{Code: `const obj = { set foo(v) {} };`},
		{Code: `const obj = { set foo(v: string) {} };`},

		// ============================================================
		// Generator with return type
		// ============================================================
		{Code: `class Foo { *gen(): Generator { yield 1; } }`},

		// ============================================================
		// Namespace / module export function
		// ============================================================
		{Code: `
namespace N {
  export function foo(): void {}
}
		`},

		// ============================================================
		// allowExpressions contexts that should be valid
		// ============================================================
		// Ternary
		{
			Code:    `true ? () => {} : () => {};`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		// Logical
		{
			Code:    `false || (() => {});`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		// Array element
		{
			Code:    `[() => {}, function() {}];`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		// Parenthesized expression
		{
			Code:    `(function() {});`,
			Options: map[string]interface{}{"allowExpressions": true},
		},

		// ============================================================
		// allowTypedFunctionExpressions with parenthesized type assertions
		// ============================================================
		{
			Code:    `const x = ((() => {}) as Foo);`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},

		// ============================================================
		// allowHigherOrderFunctions: method returning function
		// ============================================================
		{
			Code: `
class Foo {
  method() {
    return function inner(): void {};
  }
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},

		// ============================================================
		// allowIIFEs: function expression called via .call / .apply / .bind
		// These are NOT IIFEs — the function is an argument, not callee
		// ============================================================
		// (Not testing .call/.apply because they would be method calls,
		//  the function would be accessed via property, not wrapped in CallExpression directly)

		// ============================================================
		// Optional chaining on call expression argument
		// ============================================================
		{
			Code: `
declare const foo: undefined | ((cb: () => void) => void);
foo?.(() => {});
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},

		// ============================================================
		// allowFunctionsWithoutTypeParameters: method / getter
		// ============================================================
		{
			Code: `
class Foo {
  method(x: number) { return x; }
}
			`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
		},
		{
			Code: `
const obj = {
  get foo() { return 1; },
};
			`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
		},
		// Generic method still needs return type
		{
			Code: `
class Foo {
  method<T>(x: T): T { return x; }
}
			`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
		},

		// ============================================================
		// ancestorHasReturnType through VariableDeclaration with type
		// ============================================================
		// Expression body: outer bodyless arrow → VariableDeclaration has type
		{
			Code: `
type Fn = () => () => void;
const foo: Fn = () => () => {};
			`,
		},
		// Block body: inner arrow in return → walk up finds VariableDeclaration type
		{
			Code: `
type Fn = () => () => void;
const foo: Fn = () => {
  return () => {};
};
			`,
		},
		// Block body inner arrow with block body — ancestor walk still finds type
		{
			Code: `
type Fn = () => () => void;
const foo: Fn = () => {
  return () => {
    console.log('hi');
  };
};
			`,
		},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Generator method without return type
		// ============================================================
		{
			Code:   `class Foo { *gen() { yield 1; } }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 13}},
		},

		// ============================================================
		// Namespace export function without return type
		// ============================================================
		{
			Code: `
namespace N {
  export function foo() {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 10}},
		},

		// ============================================================
		// allowExpressions: declarations in various contexts still invalid
		// ============================================================
		// export default arrow
		{
			Code:    `export default () => {};`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 19}},
		},
		// Variable assignment
		{
			Code:    `const foo = () => {};`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 16}},
		},
		// Class property assignment
		{
			Code: `
class Foo {
  foo = () => {};
}
			`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},

		// ============================================================
		// allowFunctionsWithoutTypeParameters: generic method needs type
		// ============================================================
		{
			Code: `
class Foo {
  method<T>(x: T) { return x; }
}
			`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},

		// ============================================================
		// allowTypedFunctionExpressions: false — parens don't help
		// ============================================================
		{
			Code:    `const x = ((() => {})) as Foo;`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 16}},
		},

		// ============================================================
		// ancestorHasReturnType fails: no typed ancestor
		// ============================================================
		// Untyped variable — inner arrow has no ancestor with return type
		{
			Code: `
const foo = () => {
  return () => {
    console.log('hi');
  };
};
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": false},
			Errors: []rule_tester.InvalidTestCaseError{
				// Inner arrow reported first (exit order)
				{MessageId: "missingReturnType", Line: 3, Column: 13},
				{MessageId: "missingReturnType", Line: 2, Column: 16},
			},
		},
	})
}
