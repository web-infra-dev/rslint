package max_params

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMaxParamsUpstream migrates the full valid/invalid suite from upstream eslint/tests/lib/rules/max-params.js 1:1.
// Position assertions cover line/column for every invalid case. rslint-specific lock-in cases live in the max_params_extras_test.go file.
func TestMaxParamsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxParamsRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint core valid ----
			{Code: "function test(d, e, f) {}"},
			{Code: "var test = function(a, b, c) {};", Options: option(3)},
			{Code: "var test = (a, b, c) => {};", Options: option(3)},
			{Code: "var test = function test(a, b, c) {};", Options: option(3)},

			// ---- ESLint core valid: object property options ----
			{Code: "var test = function(a, b, c) {};", Options: option(map[string]interface{}{"max": 3})},

			// ---- TypeScript valid ----
			{Code: "function foo() {}"},
			{Code: "const foo = function () {};"},
			{Code: "const foo = () => {};"},
			{Code: "function foo(a) {}"},
			{Code: `
class Foo {
  constructor(a) {}
}
			`},
			{Code: `
class Foo {
  method(this: void, a, b, c) {}
}
			`},
			{Code: `
class Foo {
  method(this: Foo, a, b) {}
}
			`},
			{Code: "function foo(a, b, c, d) {}", Options: option(map[string]interface{}{"max": 4})},
			{Code: "function foo(a, b, c, d) {}", Options: option(map[string]interface{}{"maximum": 4})},
			{Code: `
class Foo {
  method(this: void) {}
}
			`, Options: option(map[string]interface{}{"max": 0})},
			{Code: `
class Foo {
  method(this: void, a) {}
}
			`, Options: option(map[string]interface{}{"max": 1})},
			{Code: `
class Foo {
  method(this: void, a) {}
}
			`, Options: option(map[string]interface{}{"countVoidThis": true, "max": 2})},
			{Code: "function testD(this: void, a) {}", Options: option(map[string]interface{}{"max": 1})},
			{Code: "function testD(this: void, a) {}", Options: option(map[string]interface{}{"countVoidThis": true, "max": 2})},
			{Code: "const testE = function (this: void, a) {}", Options: option(map[string]interface{}{"max": 1})},
			{Code: "const testE = function (this: void, a) {}", Options: option(map[string]interface{}{"countVoidThis": true, "max": 2})},
			{Code: `
declare function makeDate(m: number, d: number, y: number): Date;
			`, Options: option(map[string]interface{}{"max": 3})},
			{Code: `
type sum = (a: number, b: number) => number;
			`, Options: option(map[string]interface{}{"max": 2})},
			{Code: "function foo(this: unknown[], a, b, c) {}", Options: option(map[string]interface{}{"max": 3, "countThis": "never"})},
			{Code: "function foo(this: void, a, b, c) {}", Options: option(map[string]interface{}{"max": 3, "countThis": "except-void"})},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ESLint core invalid ----
			{
				Code:    "function test(a, b, c) {}",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'test'", 3, 2), Line: 1, Column: 1, EndLine: 1, EndColumn: 14}},
			},
			{
				Code:   "function test(a, b, c, d) {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'test'", 4, 3), Line: 1, Column: 1, EndLine: 1, EndColumn: 14}},
			},
			{
				Code:    "var test = function(a, b, c, d) {};",
				Options: option(3),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 4, 3), Line: 1, Column: 12, EndLine: 1, EndColumn: 20}},
			},
			{
				Code:    "var test = (a, b, c, d) => {};",
				Options: option(3),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Arrow function", 4, 3), Line: 1, Column: 25, EndLine: 1, EndColumn: 27}},
			},
			{
				Code:    "(function(a, b, c, d) {});",
				Options: option(3),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 4, 3), Line: 1, Column: 2, EndLine: 1, EndColumn: 10}},
			},
			{
				Code:    "var test = function test(a, b, c) {};",
				Options: option(1),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'test'", 3, 1), Line: 1, Column: 12, EndLine: 1, EndColumn: 25}},
			},

			// ---- ESLint core invalid: object property options ----
			{
				Code:    "function test(a, b, c) {}",
				Options: option(map[string]interface{}{"max": 2}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'test'", 3, 2), Line: 1, Column: 1, EndLine: 1, EndColumn: 14}},
			},
			{
				Code:    "function test(a, b, c, d) {}",
				Options: option(map[string]interface{}{}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'test'", 4, 3), Line: 1, Column: 1, EndLine: 1, EndColumn: 14}},
			},
			{
				Code:    "function test(a) {}",
				Options: option(map[string]interface{}{"max": 0}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'test'", 1, 0), Line: 1, Column: 1, EndLine: 1, EndColumn: 14}},
			},
			{
				Code: `function test(a, b, c) {
              // Just to make it longer
            }`,
				Options: option(map[string]interface{}{"max": 2}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1, EndLine: 1, EndColumn: 14}},
			},

			// ---- TypeScript invalid ----
			{
				Code:   "function foo(a, b, c, d) {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1}},
			},
			{
				Code:   "const foo = function (a, b, c, d) {};",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 13}},
			},
			{
				Code:   "const foo = (a, b, c, d) => {};",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 26}},
			},
			{
				Code:    "const foo = a => {};",
				Options: option(map[string]interface{}{"max": 0}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 15}},
			},
			{
				Code: `
class Foo {
  method(this: void, a, b, c, d) {}
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 3}},
			},
			{
				Code: `
class Foo {
  method(this: void, a) {}
}
				`,
				Options: option(map[string]interface{}{"countVoidThis": true, "max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 3}},
			},
			{
				Code: `
class Foo {
  method(this: void, a) {}
}
				`,
				Options: option(map[string]interface{}{"countThis": "always", "max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 3}},
			},
			{
				Code:    "function testD(this: void, a) {}",
				Options: option(map[string]interface{}{"countVoidThis": true, "max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1}},
			},
			{
				Code:    "function testD(this: void, a) {}",
				Options: option(map[string]interface{}{"countThis": "always", "max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1}},
			},
			{
				Code:    "const testE = function (this: void, a) {}",
				Options: option(map[string]interface{}{"countThis": "always", "max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 15}},
			},
			{
				Code:    "function testFunction(test: void, a: number) {}",
				Options: option(map[string]interface{}{"countThis": "except-void", "max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1}},
			},
			{
				Code:    "const testE = function (this: void, a) {}",
				Options: option(map[string]interface{}{"countVoidThis": true, "max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 15}},
			},
			{
				Code:    "function testFunction(test: void, a: number) {}",
				Options: option(map[string]interface{}{"countVoidThis": false, "max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1}},
			},
			{
				Code: `
class Foo {
  method(this: Foo, a, b, c) {}
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 3, Column: 3}},
			},
			{
				Code: `
declare function makeDate(m: number, d: number, y: number): Date;
				`,
				Options: option(map[string]interface{}{"max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 2, Column: 1}},
			},
			{
				Code: `
type sum = (a: number, b: number) => number;
				`,
				Options: option(map[string]interface{}{"max": 1}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 2, Column: 12}},
			},
			{
				Code:    "function foo(this: unknown[], a, b, c) {}",
				Options: option(map[string]interface{}{"max": 3, "countThis": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1}},
			},
			{
				Code:    "function foo(this: unknown[], a, b, c) {}",
				Options: option(map[string]interface{}{"max": 3, "countThis": "except-void"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 1}},
			},
		},
	)
}

func option(v any) []interface{} {
	return []interface{}{v}
}

func exceedMessage(name string, count, maxAllowed int) string {
	return fmt.Sprintf("%s has too many parameters (%d). Maximum allowed is %d.", name, count, maxAllowed)
}
