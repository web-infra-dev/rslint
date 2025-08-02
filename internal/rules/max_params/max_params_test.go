package max_params

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestMaxParamsRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		{Code: "function foo() {}"},
		{Code: "const foo = function () {};"},
		{Code: "const foo = () => {};"},
		{Code: "function foo(a) {}"},
		{Code: `
class Foo {
  constructor(a) {}
}`},
		{Code: `
class Foo {
  method(this: void, a, b, c) {}
}`},
		{Code: `
class Foo {
  method(this: Foo, a, b) {}
}`},
		{Code: "function foo(a, b, c, d) {}", Options: map[string]interface{}{"max": 4}},
		{Code: "function foo(a, b, c, d) {}", Options: map[string]interface{}{"maximum": 4}},
		{Code: `
class Foo {
  method(this: void) {}
}`, Options: map[string]interface{}{"max": 0}},
		{Code: `
class Foo {
  method(this: void, a) {}
}`, Options: map[string]interface{}{"max": 1}},
		{Code: `
class Foo {
  method(this: void, a) {}
}`, Options: map[string]interface{}{"countVoidThis": true, "max": 2}},
		{Code: `declare function makeDate(m: number, d: number, y: number): Date;`, Options: map[string]interface{}{"max": 3}},
		{Code: `type sum = (a: number, b: number) => number;`, Options: map[string]interface{}{"max": 2}},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		{
			Code: "function foo(a, b, c, d) {}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1}},
		},
		{
			Code: "const foo = function (a, b, c, d) {};",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 13}},
		},
		{
			Code: "const foo = (a, b, c, d) => {};",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 13}},
		},
		{
			Code: "const foo = a => {};",
			Options: map[string]interface{}{"max": 0},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 13}},
		},
		{
			Code: `
class Foo {
  method(this: void, a, b, c, d) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 3}},
		},
		{
			Code: `
class Foo {
  method(this: void, a) {}
}`,
			Options: map[string]interface{}{"countVoidThis": true, "max": 1},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 3}},
		},
		{
			Code: `
class Foo {
  method(this: Foo, a, b, c) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 3}},
		},
		{
			Code: `declare function makeDate(m: number, d: number, y: number): Date;`,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1}},
		},
		{
			Code: `type sum = (a: number, b: number) => number;`,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 12}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MaxParamsRule, validTestCases, invalidTestCases)
}