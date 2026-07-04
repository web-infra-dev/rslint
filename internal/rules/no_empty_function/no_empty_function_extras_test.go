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
			{Code: `const fn = ((() => {})) as Fn;`, Options: map[string]interface{}{"allow": []interface{}{"arrowFunctions"}}},
			{Code: `const fn = (function named() {}) satisfies Fn;`, Options: map[string]interface{}{"allow": []interface{}{"functions"}}},
			{Code: `const fn = (function named() {})!;`, Options: map[string]interface{}{"allow": []interface{}{"functions"}}},

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
			{Code: `class C { private constructor() {} }`, Options: map[string]interface{}{"allow": []interface{}{"privateConstructors"}}},
			{Code: `class C { protected constructor() {} }`, Options: map[string]interface{}{"allow": []interface{}{"protectedConstructors"}}},

			// Locks in upstream isAllowedEmptyFunction() decoratedFunctions arm.
			{Code: `class C { @Log("This is a contrived example.") blah(): void { } }`, Options: map[string]interface{}{"allow": []interface{}{"decoratedFunctions"}}},

			// ---- Real-user: typescript-eslint#2838 decorated methods may be intentionally empty ----
			{Code: `class C { @Log("This is a contrived example.") blah(): void {} }`, Options: map[string]interface{}{"allow": []interface{}{"decoratedFunctions"}}},

			// ---- Real-user: typescript-eslint#2278 decorated private methods may be intentionally empty ----
			{Code: `class C { @Emit("click") private onClick() {} }`, Options: map[string]interface{}{"allow": []interface{}{"decoratedFunctions"}}},

			// Locks in upstream isAllowedEmptyFunction() overrideMethods arm.
			{Code: `class C extends B { override method() {} }`, Options: map[string]interface{}{"allow": []interface{}{"overrideMethods"}}},
			{Code: `class C extends B { static override async method() {} }`, Options: map[string]interface{}{"allow": []interface{}{"overrideMethods"}}},

			// ---- Dimension 4: class-field arrows are named as methods but allowed by arrowFunctions ----
			{Code: `class C { field = () => {} }`, Options: map[string]interface{}{"allow": []interface{}{"arrowFunctions"}}},
			{Code: `class C { field = (() => {}) }`, Options: map[string]interface{}{"allow": []interface{}{"arrowFunctions"}}},

			// Locks in upstream getKind() parent PropertyDefinition fallback: class-field function expressions are functions.
			{Code: `class C { field = function named() {} }`, Options: map[string]interface{}{"allow": []interface{}{"functions"}}},
			{Code: `class C { field = async function named() {} }`, Options: map[string]interface{}{"allow": []interface{}{"asyncFunctions"}}},

			// Locks in upstream getKind() parent Property fallback: object property arrows/functions keep their function kind.
			{Code: `const obj = { foo: () => {} };`, Options: map[string]interface{}{"allow": []interface{}{"arrowFunctions"}}},
			{Code: `const obj = { foo: async function () {} };`, Options: map[string]interface{}{"allow": []interface{}{"asyncFunctions"}}},

			// ---- Dimension 4: async generator functions take the generatorFunctions kind, matching upstream prefix priority ----
			{Code: `async function* f() {}`, Options: map[string]interface{}{"allow": []interface{}{"generatorFunctions"}}},
			{Code: `class C { async *m() {} }`, Options: map[string]interface{}{"allow": []interface{}{"generatorMethods"}}},

			// Locks in option parsing for combined arrays and the typed []string branch.
			{Code: `function f() {} const g = () => {}; class C { method() {} }`, Options: map[string]interface{}{"allow": []interface{}{"functions", "arrowFunctions", "methods"}}},
			{Code: `function f() {}`, Options: map[string]interface{}{"allow": []string{"functions"}}},

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

			// Locks in upstream isAllowedEmptyFunction() constructors arm: ESLint uses camelCase option names.
			generatedInvalidCase(`class C { private constructor() {} }`, "constructor", map[string]interface{}{"allow": []interface{}{"private-constructors"}}),
			generatedInvalidCase(`class C { protected constructor() {} }`, "constructor", map[string]interface{}{"allow": []interface{}{"protected-constructors"}}),

			// Locks in upstream getKind() prefix priority: async generator is not allowed by asyncFunctions.
			generatedInvalidCase(`async function* f() {}`, "async generator function 'f'", map[string]interface{}{"allow": []interface{}{"asyncFunctions"}}),
			generatedInvalidCase(`class C { async *m() {} }`, "async generator method 'm'", map[string]interface{}{"allow": []interface{}{"asyncMethods"}}),

			// Locks in option parsing for bare object, empty object, empty allow, and malformed allow values.
			generatedInvalidCase(`const fn = () => {};`, "arrow function", map[string]interface{}{"allow": []interface{}{"functions"}}),
			generatedInvalidCase(`function f() {}`, "function 'f'", map[string]interface{}{}),
			generatedInvalidCase(`function f() {}`, "function 'f'", []interface{}{map[string]interface{}{}}),
			generatedInvalidCase(`function f() {}`, "function 'f'", map[string]interface{}{"allow": []interface{}{}}),
			generatedInvalidCase(`function f() {}`, "function 'f'", map[string]interface{}{"allow": "functions"}),
			generatedInvalidCase(`function f() {}`, "function 'f'", map[string]interface{}{"allow": []interface{}{"unknown"}}),

			// ---- Dimension 4: overload/body-absent declarations do not mask the implementation body ----
			generatedInvalidCase(`function f(value: string): void; function f(value: number): void; function f(value: string | number) {}`, "function 'f'", nil),
			generatedInvalidCase(`class C { method(value: string): void; method(value: number): void; method(value: string | number) {} }`, "method 'method'", nil),

			// ---- Real-user: typescript-eslint#2278 still reports decorated methods without the option ----
			generatedInvalidCase(`class C { @Emit("click") private onClick() {} }`, "method 'onClick'", nil),

			// Locks in upstream isAllowedEmptyFunction() method/accessor-only arm: decorated/override fields stay function/arrow kinds.
			generatedInvalidCase(`class C { @dec field = () => {} }`, "method 'field'", map[string]interface{}{"allow": []interface{}{"decoratedFunctions"}}),
			generatedInvalidCase(`class C extends B { override field = () => {} }`, "method 'field'", map[string]interface{}{"allow": []interface{}{"overrideMethods"}}),
		},
	)
}
