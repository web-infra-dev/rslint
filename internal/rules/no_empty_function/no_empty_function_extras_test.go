// TestNoEmptyFunctionExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
//
// Dimension 4 walk (rows that don't apply to no-empty-function, with reasons):
//   - N/A member receiver wrappers ((X).y, X!.y, X as T, X satisfies T, X?.y):
//     the rule inspects function-like declarations and their block body, not a
//     member receiver expression. Expression wrappers around function values
//     are covered below because tsgo preserves those wrappers.
//   - N/A element access forms (X['y'], X[`y`], X[0]): the rule does not
//     inspect member expressions.
//   - N/A autofix boundaries: the rule has suggestions only, not an autofix.
package no_empty_function

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyFunctionExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyFunctionRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: non-block arrow body is not an empty function ----
			{Code: `const fn = () => value;`},

			// ---- Dimension 4: expression wrappers around function values preserve allow matching ----
			{Code: `const fn = ((() => {})) as Fn;`, Options: []any{map[string]any{"allow": []any{"arrowFunctions"}}}},
			{Code: `const fn = (function named() {}) satisfies Fn;`, Options: []any{map[string]any{"allow": []any{"functions"}}}},
			{Code: `const fn = (function named() {})!;`, Options: []any{map[string]any{"allow": []any{"functions"}}}},

			// ---- Dimension 4: comments inside an otherwise empty body are allowed ----
			{Code: `function f() { /* intentionally empty */ }`},
			{Code: "const f = () => {\n  // intentionally empty\n};"},
			{Code: `class C { method() { /* intentionally empty */ } }`},

			// Locks in upstream reportIfEmpty() arm: non-empty body short-circuits before options.
			{Code: `function f() { sideEffect(); }`},
			{Code: `function f() { ; }`},
			{Code: `function f() { "use strict"; }`},
			{Code: `class C { method() { sideEffect(); } }`},

			// Locks in upstream isAllowedEmptyFunction() constructors arm: TS parameter properties are allowed.
			{Code: `class C { constructor(public value: string) {} }`},
			{Code: `class C { constructor(private value: string) {} }`},
			{Code: `class C { constructor(protected value: string) {} }`},
			{Code: `class C { constructor(readonly value: string) {} }`},

			// Locks in upstream isAllowedEmptyFunction() private/protected constructor arms.
			{Code: `class C { private constructor() {} }`, Options: []any{map[string]any{"allow": []any{"privateConstructors"}}}},
			{Code: `class C { protected constructor() {} }`, Options: []any{map[string]any{"allow": []any{"protectedConstructors"}}}},

			// Locks in upstream isAllowedEmptyFunction() decoratedFunctions arm.
			{Code: `class C { @Log("This is a contrived example.") blah(): void { } }`, Options: []any{map[string]any{"allow": []any{"decoratedFunctions"}}}},

			// ---- Real-user: typescript-eslint#2838 decorated methods may be intentionally empty ----
			{Code: `class C { @Log("This is a contrived example.") blah(): void {} }`, Options: []any{map[string]any{"allow": []any{"decoratedFunctions"}}}},

			// ---- Real-user: typescript-eslint#2278 decorated private methods may be intentionally empty ----
			{Code: `class C { @Emit("click") private onClick() {} }`, Options: []any{map[string]any{"allow": []any{"decoratedFunctions"}}}},

			// Locks in upstream isAllowedEmptyFunction() overrideMethods arm.
			{Code: `class C extends B { override method() {} }`, Options: []any{map[string]any{"allow": []any{"overrideMethods"}}}},
			{Code: `class C extends B { static override async method() {} }`, Options: []any{map[string]any{"allow": []any{"overrideMethods"}}}},

			// ---- Dimension 4: class-field arrows are named as methods but allowed by arrowFunctions ----
			{Code: `class C { field = () => {} }`, Options: []any{map[string]any{"allow": []any{"arrowFunctions"}}}},
			{Code: `class C { field = (() => {}) }`, Options: []any{map[string]any{"allow": []any{"arrowFunctions"}}}},

			// Locks in upstream getKind() parent PropertyDefinition fallback: class-field function expressions are functions.
			{Code: `class C { field = function named() {} }`, Options: []any{map[string]any{"allow": []any{"functions"}}}},
			{Code: `class C { field = async function named() {} }`, Options: []any{map[string]any{"allow": []any{"asyncFunctions"}}}},

			// Locks in upstream getKind() parent Property fallback: object property arrows/functions keep their function kind.
			{Code: `const obj = { foo: () => {} };`, Options: []any{map[string]any{"allow": []any{"arrowFunctions"}}}},
			{Code: `const obj = { foo: async function () {} };`, Options: []any{map[string]any{"allow": []any{"asyncFunctions"}}}},

			// ---- Dimension 4: async generator functions take the generatorFunctions kind, matching upstream prefix priority ----
			{Code: `async function* f() {}`, Options: []any{map[string]any{"allow": []any{"generatorFunctions"}}}},
			{Code: `class C { async *m() {} }`, Options: []any{map[string]any{"allow": []any{"generatorMethods"}}}},

			// Locks in option parsing for combined arrays.
			{Code: `function f() {} const g = () => {}; class C { method() {} }`, Options: []any{map[string]any{"allow": []any{"functions", "arrowFunctions", "methods"}}}},

			// ---- Dimension 4: declaration/body-absent forms do not report or crash ----
			{Code: `declare function f(): void;`},
			{Code: `abstract class C { abstract method(): void }`},
			{Code: `class C { method(): void; }`},
			{Code: `class C {}`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: class-field arrow name comes from the property definition ----
			generatedInvalidCase(`class C { field = () => {} }`, "method 'field'", map[string]interface{}{"allow": []interface{}{"methods"}}),
			generatedInvalidCase(`class C { field = (() => {}) }`, "method 'field'", map[string]interface{}{"allow": []interface{}{"methods"}}),
			generatedInvalidCase(`class C { field = ((function named() {})) }`, "method 'field'", nil),
			generatedInvalidCase(`class C { field = function named() {} }`, "method 'field'", map[string]interface{}{"allow": []interface{}{"methods"}}),

			// ---- Dimension 4: expression wrappers around function values do not hide empty bodies ----
			generatedInvalidCase(`const fn = (() => {}) as Fn;`, "arrow function", nil),
			generatedInvalidCase(`const fn = (function named() {}) satisfies Fn;`, "function 'named'", nil),
			generatedInvalidCase(`const fn = (function named() {})!;`, "function 'named'", nil),

			// ---- Dimension 4: private class-field arrow uses raw PrivateIdentifier name ----
			generatedInvalidCase(`class C { static #field = () => {} }`, "static private method #field", nil),
			generatedInvalidCase(`class C { static #field = (async () => {}) }`, "static private async method #field", nil),
			generatedInvalidCase(`class C { "#field" = () => {} }`, "method '#field'", nil),
			generatedInvalidCase(`class C { static #field() {} }`, "static private method #field", nil),
			generatedInvalidCase(`class C { static async #field() {} }`, "static private async method #field", nil),

			// ---- Dimension 4: string / numeric / computed-static / private / dynamic keys ----
			generatedInvalidCase(`const obj = {"quoted"() {}};`, "method 'quoted'", nil),
			generatedInvalidCase(`const obj = {0() {}};`, "method '0'", nil),
			generatedInvalidCase(`const obj = { ["computed"]() {} };`, "method 'computed'", nil),
			generatedInvalidCase("const obj = { [`computed`]() {} };", "method 'computed'", nil),
			generatedInvalidCase(`class C { [0x10]() {} }`, "method '16'", nil),
			generatedInvalidCase(`class C { [1n]() {} }`, "method '1'", nil),
			generatedInvalidCase(`class C { ["computed"]() {} }`, "method 'computed'", nil),
			generatedInvalidCase(`class C { #privateMethod() {} }`, "private method #privateMethod", nil),
			generatedInvalidCase(`class C { [dynamicName]() {} }`, "method", nil),
			generatedInvalidCase(`const obj = {foo: (function () {})};`, "method 'foo'", nil),
			generatedInvalidCase(`const obj = {foo: (() => {})};`, "method 'foo'", nil),
			generatedInvalidCase(`const obj = {[dynamicName]: (function () {})};`, "method", nil),

			// Locks in upstream getKind() parent Property/PropertyDefinition fallbacks.
			generatedInvalidCase(`const obj = { foo: () => {} };`, "method 'foo'", map[string]interface{}{"allow": []interface{}{"methods"}}),
			generatedInvalidCase(`const obj = { foo: async function () {} };`, "async method 'foo'", map[string]interface{}{"allow": []interface{}{"asyncMethods"}}),

			// ---- Dimension 4: same-kind nesting reports only the empty inner function ----
			generatedInvalidCase(`function outer() { function inner() {} }`, "function 'inner'", nil),
			generatedInvalidCase(`function outer() { const inner = (() => {}); }`, "arrow function", nil),
			generatedInvalidCase(`class C { static { function nested() {} } }`, "function 'nested'", nil),
			generatedInvalidCaseWithNames("class C {\n  method() {}\n  field = () => {}\n}\nconst obj = { nested: function named() {} };\n", "method 'method'", "method 'field'", "method 'nested'"),

			// ---- Dimension 4: spread/rest shapes do not mask sibling empty functions ----
			generatedInvalidCase(`const obj = { ...extra, method() {} };`, "method 'method'", nil),
			generatedInvalidCase(`function f(...args) {}`, "function 'f'", nil),
			generatedInvalidCase(`function f() /* outside body */ {}`, "function 'f'", nil),
			locationCase("class C {\n  field = () => {\n  }\n}", "method 'field'", 2, 17, 3, 4, "class C {\n  field = () => { /* empty */ }\n}"),

			// Locks in upstream getKind() prefix priority: async generator is not allowed by asyncFunctions.
			generatedInvalidCase(`async function* f() {}`, "async generator function 'f'", []any{map[string]any{"allow": []any{"asyncFunctions"}}}),
			generatedInvalidCase(`class C { async *m() {} }`, "async generator method 'm'", []any{map[string]any{"allow": []any{"asyncMethods"}}}),

			// Locks in option parsing for empty object and empty allow.
			generatedInvalidCase(`const fn = () => {};`, "arrow function", []any{map[string]any{"allow": []any{"functions"}}}),
			generatedInvalidCase(`function f() {}`, "function 'f'", []any{map[string]any{}}),
			generatedInvalidCase(`function f() {}`, "function 'f'", []any{map[string]any{"allow": []any{}}}),

			// ---- Dimension 4: overload/body-absent declarations do not mask the implementation body ----
			generatedInvalidCase(`function f(value: string): void; function f(value: number): void; function f(value: string | number) {}`, "function 'f'", nil),
			generatedInvalidCase(`class C { method(value: string): void; method(value: number): void; method(value: string | number) {} }`, "method 'method'", nil),

			// ---- Real-user: typescript-eslint#2278 still reports decorated methods without the option ----
			generatedInvalidCase(`class C { @Emit("click") private onClick() {} }`, "method 'onClick'", nil),

			// Locks in upstream isAllowedEmptyFunction() method/accessor-only arm: decorated/override fields stay function/arrow kinds.
			generatedInvalidCase(`class C { @dec field = () => {} }`, "method 'field'", []any{map[string]any{"allow": []any{"decoratedFunctions"}}}),
			generatedInvalidCase(`class C extends B { override field = () => {} }`, "method 'field'", []any{map[string]any{"allow": []any{"overrideMethods"}}}),
		},
	)
}
