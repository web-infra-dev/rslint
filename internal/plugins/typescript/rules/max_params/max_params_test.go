package max_params

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMaxParamsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MaxParamsRule, []rule_tester.ValidTestCase{
		{Code: `function foo() {}`},
		{Code: `const foo = function () {};`},
		{Code: `const foo = () => {};`},
		{Code: `function foo(a) {}`},
		{
			Code: `
class Foo {
  constructor(a) {}
}
			`,
		},
		{
			Code: `
class Foo {
  method(this: void, a, b, c) {}
}
			`,
		},
		{
			Code: `
class Foo {
  method(this: Foo, a, b) {}
}
			`,
		},
		{
			Code:    `function foo(a, b, c, d) {}`,
			Options: []interface{}{map[string]interface{}{"max": 4}},
		},
		{
			Code:    `function foo(a, b, c, d) {}`,
			Options: []interface{}{map[string]interface{}{"maximum": 4}},
		},
		{
			Code: `
class Foo {
  method(this: void) {}
}
			`,
			Options: []interface{}{map[string]interface{}{"max": 0}},
		},
		{
			Code: `
class Foo {
  method(this: void, a) {}
}
			`,
			Options: []interface{}{map[string]interface{}{"max": 1}},
		},
		{
			Code: `
class Foo {
  method(this: void, a) {}
}
			`,
			Options: []interface{}{map[string]interface{}{"countVoidThis": true, "max": 2}},
		},
		{
			Code: `
declare function makeDate(m: number, d: number, y: number): Date;
			`,
			Options: []interface{}{map[string]interface{}{"max": 3}},
		},
		{
			Code: `
type sum = (a: number, b: number) => number;
			`,
			Options: []interface{}{map[string]interface{}{"max": 2}},
		},
		{
			Code: `
interface Foo {
  method(a: number, b: number, c: number, d: number): void;
}
			`,
		},
		{
			Code: `
type CallSig = {
  (a: number, b: number, c: number, d: number): void;
};
			`,
		},
		{
			Code: `
type Ctor = new (a: number, b: number, c: number, d: number) => Foo;
			`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `function foo(a, b, c, d) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
			},
		},
		{
			Code: `const foo = function (a, b, c, d) {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exceed", Line: 1, Column: 13, EndLine: 1, EndColumn: 37},
			},
		},
		{
			Code: `const foo = (a, b, c, d) => {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exceed", Line: 1, Column: 13, EndLine: 1, EndColumn: 31},
			},
		},
		{
			Code:    `const foo = a => {};`,
			Options: []interface{}{map[string]interface{}{"max": 0}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exceed", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
			},
		},
		{
			Code: `
class Foo {
  method(this: void, a, b, c, d) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exceed", Line: 3, Column: 3, EndLine: 3, EndColumn: 36},
			},
		},
		{
			Code: `
class Foo {
  method(this: void, a) {}
}
			`,
			Options: []interface{}{map[string]interface{}{"countVoidThis": true, "max": 1}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exceed", Line: 3, Column: 3, EndLine: 3, EndColumn: 27},
			},
		},
		{
			Code: `
class Foo {
  method(this: Foo, a, b, c) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exceed", Line: 3, Column: 3, EndLine: 3, EndColumn: 32},
			},
		},
		{
			Code: `
declare function makeDate(m: number, d: number, y: number): Date;
			`,
			Options: []interface{}{map[string]interface{}{"max": 1}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exceed", Line: 2, Column: 1, EndLine: 2, EndColumn: 66},
			},
		},
		{
			Code: `
type sum = (a: number, b: number) => number;
			`,
			Options: []interface{}{map[string]interface{}{"max": 1}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exceed", Line: 2, Column: 12, EndLine: 2, EndColumn: 44},
			},
		},
	})
}
