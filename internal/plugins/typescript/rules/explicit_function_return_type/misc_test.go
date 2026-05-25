package explicit_function_return_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMiscEdgeCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		// Private class members with return types
		{Code: `
class A {
  #method(): void {}
  #arrow = (): void => {};
  private fn = (): void => {};
}
		`},
		// Override method with return type
		{Code: `
class A { method(): void {} }
class B extends A {
  override method(): void {}
}
		`},
		// Parenless arrow in class property with type annotation
		{Code: `
class A {
  fn: (x: number) => number = x => x + 1;
}
		`},
		// Abstract getter with return type (no body)
		{Code: `
abstract class A {
  abstract get foo(): number;
}
		`},
		// Object method with async
		{Code: `
const obj = {
  async foo(): Promise<void> {},
};
		`},
		// Object method with generator
		{Code: `
const obj = {
  *gen(): Generator { yield 1; },
};
		`},
		// Nested IIFE with typed result
		{
			Code:    `const x = (() => (() => 'foo')())();`,
			Options: map[string]interface{}{"allowIIFEs": true},
		},
		// allowExpressions with object shorthand method
		{
			Code: `
const obj = {
  foo() {},
  get bar() { return 1; },
  async baz() {},
};
			`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
	}, []rule_tester.InvalidTestCase{
		// Private members without return types
		{
			Code: `
class A {
  #method() {}
  #arrow = () => {};
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 3},
				{MessageId: "missingReturnType", Line: 4, Column: 3},
			},
		},
		// Override method without return type
		{
			Code: `
class A { method(): void {} }
class B extends A {
  override method() {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 4, Column: 3},
			},
		},
		// Object async method without return type
		{
			Code: `
const obj = {
  async foo() {},
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 3},
			},
		},
		// allowedNames does NOT match private identifiers (matches ESLint behavior)
		{
			Code: `
class A {
  #method() {}
}
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"#method"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
		// Parenless arrow in untyped class property
		{
			Code: `
class A {
  fn = x => x + 1;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
		// allowExpressions does NOT apply to class methods/getters
		{
			Code: `
class A {
  foo() {}
  get bar() { return 1; }
}
			`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 3},
				{MessageId: "missingReturnType", Line: 4, Column: 3},
			},
		},
	})
}
