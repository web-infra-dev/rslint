package max_params

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMaxParamsExtras locks in branches and edge shapes that the upstream test suite doesn't exercise.
// Each case carries an inline comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future refactors can't silently regress them without breaking a named lock-in.
func TestMaxParamsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxParamsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: declaration/container forms ----
			{Code: "async function f(a, b, c) {}"},
			{Code: "function* f(a, b, c) {}"},
			{Code: "async function* f(a, b, c) {}"},
			{Code: "class C { static method(a, b, c) {} }"},
			{Code: "class C { #method(a, b, c) {} }"},
			{Code: "const C = class { method(a, b, c) {} };"},
			{Code: "const obj = { method(a, b, c) {} };"},
			{Code: "const obj = { get value() { return 1; } };"},
			{Code: "class C { field = (a, b, c) => {}; }"},
			{Code: "class C { static #field = (a, b, c) => {}; }"},
			{Code: "const obj = { field: function(a, b, c) {} };"},
			{Code: "const obj = { field: (a, b, c) => {} };"},

			// ---- Dimension 4: access/key forms ----
			{Code: "class C { 'method'(a, b, c) {} }"},
			{Code: "class C { 0(a, b, c) {} }"},
			{Code: "class C { ['method'](a, b, c) {} }"},
			{Code: "class C { #method(a, b, c) {} }"},

			// ---- Dimension 4: receiver/expression wrappers on inspected function nodes ----
			{Code: "(function(a, b, c) {});"},
			{Code: "((function(a, b, c) {}));"},
			{Code: "const f = (function(a, b, c) {}) as Function;"},
			{Code: "const f = (function(a, b, c) {}) satisfies Function;"},
			{Code: "const f = (function(a, b, c) {})!;"},
			// N/A: optional chaining applies to calls/member access, not function definitions.

			// ---- Dimension 4: graceful degradation ----
			{Code: "function f() {}"},
			{Code: "function f(...args) {}"},
			{Code: "function f({a, b}, [c]) {}"},
			{Code: "function f(a = 1, b?: number, c?: number) {}"},
			{Code: "abstract class C { abstract method(a, b, c): void; }"},
			{Code: "interface I { method(a: number, b: number, c: number): void; }"},
			{Code: "interface I { method(a: number, b: number, c: number, d: number): void; }"},
			{Code: "interface I { (a: number, b: number, c: number, d: number): void; }"},
			{Code: "interface I { new (a: number, b: number, c: number, d: number): object; }"},
			{Code: "type Shape = { method(a: number, b: number, c: number, d: number): void };"},
			{Code: "type Fn = <T, U>(a: T, b: U, c: string) => void;"},
			{Code: "type Ctor = new (a: number, b: number, c: number, d: number) => object;"},
			{Code: "declare namespace API { export function request(a: string, b: string, c: string): void; }"},
			// N/A: object spread/rest do not appear in function parameter lists as standalone members.

			// Locks in upstream parseOptions arm 1: numeric option sets max directly.
			{Code: "function f(a, b, c, d) {}", Options: option(4)},
			// Locks in upstream parseOptions arm 2: bare object shape from CLI uses max.
			{Code: "function f(a, b, c, d) {}", Options: map[string]interface{}{"max": 4}},
			// Locks in upstream parseOptions arm 2b: bare object shape from CLI uses maximum.
			{Code: "function f(a, b, c, d) {}", Options: map[string]interface{}{"maximum": 4}},
			// Locks in upstream parseOptions arm 3: maximum wins when truthy.
			{Code: "function f(a, b, c, d) {}", Options: option(map[string]interface{}{"maximum": 4, "max": 1})},
			// Locks in upstream parseOptions arm 4: maximum: 0 falls through to max.
			{Code: "function f(a, b, c, d) {}", Options: option(map[string]interface{}{"maximum": 0, "max": 4})},
			// Locks in upstream parseOptions arm 5: maximum: 0 with no max disables reports.
			{Code: "function f(a, b, c, d, e) {}", Options: option(map[string]interface{}{"maximum": 0})},
			// Locks in upstream countThis arm 1: "never" strips any this parameter.
			{Code: "function f(this: Foo, a, b, c) {}", Options: option(map[string]interface{}{"max": 3, "countThis": "never"})},
			// Locks in upstream countThis arm 2: "except-void" strips this: void.
			{Code: "function f(this: void, a, b, c) {}", Options: option(map[string]interface{}{"max": 3, "countThis": "except-void"})},
			// Locks in upstream countVoidThis arm 1: false maps to except-void.
			{Code: "function f(this: void, a, b, c) {}", Options: option(map[string]interface{}{"max": 3, "countVoidThis": false})},

			// ---- Real-user: eslint/eslint#20107 countThis never for typed this ----
			{Code: "function doSomething(this: MyType, param1: unknown, param2: unknown): unknown {}", Options: option(map[string]interface{}{"max": 2, "countThis": "never"})},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: declaration/container forms ----
			{
				Code:   "async function f(a, b, c, d) {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Async function 'f'", 4, 3), Line: 1, Column: 1}},
			},
			{
				Code:   "function* f(a, b, c, d) {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Generator function 'f'", 4, 3), Line: 1, Column: 1}},
			},
			{
				Code:   "export default function(a, b, c, d) {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 4, 3)}},
			},
			{
				Code:   "const f = async <T>(a: T, b: T, c: T, d: T) => {};",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Async arrow function", 4, 3)}},
			},
			{
				Code:   "class C { static #method(a, b, c, d) {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Static private method '#method'", 4, 3), Line: 1, Column: 11}},
			},
			{
				Code:   "class C { field = (a, b, c, d) => {}; }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 11}},
			},
			{
				Code:   "class C { static #field = (a, b, c, d) => {}; }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Static private method '#field'", 4, 3)}},
			},
			{
				Code:   "const obj = { field: function(a, b, c, d) {} };",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'field'", 4, 3)}},
			},
			{
				Code:   "const obj = { field: (a, b, c, d) => {} };",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'field'", 4, 3)}},
			},
			{
				Code:   "const obj = { async method(a, b, c, d) {} };",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Async method 'method'", 4, 3)}},
			},
			{
				Code:   "const obj = { *method(a, b, c, d) {} };",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Generator method 'method'", 4, 3)}},
			},
			{
				Code:    "const obj = { set value(v) {} };",
				Options: option(0),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Setter 'value'", 1, 0)}},
			},

			// ---- Dimension 4: access/key forms ----
			{
				Code:   "class C { 'method'(a, b, c, d) {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'method'", 4, 3), Line: 1, Column: 11}},
			},
			{
				Code:   "class C { 0(a, b, c, d) {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method '0'", 4, 3), Line: 1, Column: 11}},
			},
			{
				Code:   "class C { [name](a, b, c, d) {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method", 4, 3), Line: 1, Column: 11}},
			},

			// ---- Dimension 4: receiver/expression wrappers on inspected function nodes ----
			{
				Code:   "((function(a, b, c, d) {}));",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 4, 3), Line: 1, Column: 3}},
			},
			{
				Code:   "const f = (function(a, b, c, d) {}) as Function;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 12}},
			},
			{
				Code:   "const f = (function(a, b, c, d) {})!;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 12}},
			},

			// ---- Dimension 4: nesting/traversal boundaries ----
			{
				Code: "function outer(a, b, c, d) { function inner(e, f, g, h) {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: exceedMessage("Function 'outer'", 4, 3), Line: 1, Column: 1},
					{MessageId: "exceed", Message: exceedMessage("Function 'inner'", 4, 3), Line: 1, Column: 30},
				},
			},
			{
				Code: "class Outer { method(a, b, c, d) { const inner = (e, f, g, h) => {}; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: exceedMessage("Method 'method'", 4, 3)},
					{MessageId: "exceed", Message: exceedMessage("Arrow function", 4, 3)},
				},
			},
			{
				Code: "function outer(a, b, c) { class C { method(d, e, f, g) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: exceedMessage("Method 'method'", 4, 3)},
				},
			},
			{
				Code: "class C { static { function f(a, b, c, d) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: exceedMessage("Function 'f'", 4, 3)},
				},
			},

			// ---- Dimension 4: graceful degradation ----
			{
				Code:    "function f(...args) {}",
				Options: option(0),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'f'", 1, 0), Line: 1, Column: 1}},
			},
			{
				Code:    "abstract class C { abstract method(a, b, c, d): void; }",
				Options: option(3),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Line: 1, Column: 20}},
			},
			{
				Code:    "class C { set value(v) {} }",
				Options: option(0),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Setter 'value'", 1, 0)}},
			},
			{
				Code:    "class C { constructor(public a, private b, readonly c, d) {} }",
				Options: option(3),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Constructor", 4, 3)}},
			},
			{
				Code:    "type Fn = <T>(a: T, b: T, c: T, d: T) => void;",
				Options: option(3),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 4, 3)}},
			},
			{
				Code:    "interface I { field: (a: number, b: number, c: number, d: number) => void; }",
				Options: option(3),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 4, 3)}},
			},
			{
				Code:    "declare const handler: (event: Event, source: string, retry: boolean, meta: unknown) => void;",
				Options: option(3),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 4, 3)}},
			},
			{
				Code:    "declare namespace API { export function request(a: string, b: string, c: string, d: string): void; }",
				Options: option(3),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'request'", 4, 3)}},
			},
			{
				Code: "function overload(a: string, b: string, c: string, d: string): void;\n" +
					"function overload(a: string, b: string, c: string, d: string) {}",
				Options: option(3),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: exceedMessage("Function 'overload'", 4, 3)},
					{MessageId: "exceed", Message: exceedMessage("Function 'overload'", 4, 3)},
				},
			},

			// Locks in upstream parseOptions arm 6: empty object defaults to max=3.
			{
				Code:    "function f(a, b, c, d) {}",
				Options: option(map[string]interface{}{}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'f'", 4, 3), Line: 1, Column: 1}},
			},
			// Locks in upstream countThis arm 3: "always" counts this: void.
			{
				Code:    "function f(this: void, a) {}",
				Options: option(map[string]interface{}{"max": 1, "countThis": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'f'", 2, 1), Line: 1, Column: 1}},
			},
			// Locks in upstream countVoidThis arm 2: true maps to always.
			{
				Code:    "function f(this: void, a) {}",
				Options: option(map[string]interface{}{"max": 1, "countVoidThis": true}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'f'", 2, 1), Line: 1, Column: 1}},
			},
			// Locks in upstream checkFunction arm 1: non-this first parameter with void type still counts.
			{
				Code:    "function f(value: void, a) {}",
				Options: option(map[string]interface{}{"max": 1, "countThis": "except-void"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'f'", 2, 1), Line: 1, Column: 1}},
			},
			// Locks in upstream checkFunction arm 2: only an exact `this: void` is stripped.
			{
				Code:    "function f(this: undefined, a, b, c) {}",
				Options: option(map[string]interface{}{"max": 3, "countThis": "except-void"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'f'", 4, 3), Line: 1, Column: 1}},
			},
			// Locks in upstream checkFunction arm 2: unions containing void still count.
			{
				Code:    "function f(this: void | undefined, a, b, c) {}",
				Options: option(map[string]interface{}{"max": 3, "countThis": "except-void"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'f'", 4, 3), Line: 1, Column: 1}},
			},

			// ---- Real-user: eslint/eslint#798 RequireJS define callback is still a normal function expression ----
			{
				Code: `define([
  'dependency1',
  'dependency2',
  'dependency3',
  'dependency4'
], function(d1, d2, d3, d4) {
  use(d1, d2, d3, d4);
});`,
				Options: option(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 4, 3), Line: 6, Column: 4}},
			},
			// ---- Real-user: Express-style route callbacks are normal function expressions ----
			{
				Code:    `app.get("/users/:id", function routeHandler(req, res, next, logger) { res.end(); });`,
				Options: option(map[string]interface{}{"max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'routeHandler'", 4, 3)}},
			},
		},
	)
}
