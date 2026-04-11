package explicit_function_return_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBodyless(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		// declare function — no body, skip (ESLint treats as TSDeclareFunction)
		{Code: `declare function foo(): void;`},
		{Code: `declare function foo();`},
		// Abstract methods — no body, skip (ESLint treats as TSAbstractMethodDefinition)
		{Code: `
abstract class Foo {
  abstract method(): void;
}
		`},
		{Code: `
abstract class Foo {
  abstract method();
}
		`},
		// Overload signatures — no body, skip. Only the implementation body matters.
		{Code: `
function foo(x: number): void;
function foo(x: string): void;
function foo(x: any): void {}
		`},
		{Code: `
class Foo {
  method(x: number): void;
  method(x: string): void;
  method(x: any): void {}
}
		`},
		// declare class method
		{Code: `
declare class Foo {
  method(): void;
}
		`},
		{Code: `
declare class Foo {
  method();
}
		`},
	}, []rule_tester.InvalidTestCase{
		// Overload: the implementation body still needs a return type
		{
			Code: `
function foo(x: number): void;
function foo(x: any) { return x; }
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 1}},
		},
	})
}

func TestGetFunctionHeadLocEdgeCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// async function — head starts at `async`, not `function`
		{
			Code:   `async function foo() {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 1}},
		},
		// export default function — head starts at `function`
		{
			Code:   `export default function() {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 16}},
		},
		// export async function — head starts at `async`
		{
			Code:   `export async function foo() {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 8}},
		},
		// export default async function — head starts at `async`
		{
			Code:   `export default async function foo() {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 16}},
		},
		// class static async method — head includes all modifiers
		{
			Code: `
class A {
  static async method() {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
		// class private method — head includes `private`
		{
			Code: `
class A {
  private method() {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
		// class property with function expression — head includes property name
		{
			Code: `
class A {
  static foo = function() {};
}
			`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
		// arrow in object property — head includes property key
		{
			Code: `
const x = {
  foo: () => {},
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
	})
}
