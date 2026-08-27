// TestPreferReadonlyParameterTypesExtras locks in branches and edge shapes that
// the upstream test suite does not exercise. Each case points at the relevant
// branch, Dimension 4 row, or real-user regression. The 1:1 upstream migration
// lives in prefer_readonly_parameter_types_upstream_test.go.
package prefer_readonly_parameter_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func extraReadonlyError(column, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "shouldBeReadonly",
		Message:   "Parameter should be a read only type.",
		Line:      1,
		Column:    column,
		EndLine:   1,
		EndColumn: endColumn,
	}
}

func TestPreferReadonlyParameterTypesExtras(t *testing.T) {
	// N/A: Dimension 4 receiver/expression wrappers and optional chains. The
	// rule inspects parameter types, not member-access receivers.
	// N/A: Dimension 4 access/key forms. The rule does not match property keys.
	// N/A: Dimension 3 autofix boundaries. The rule has no fixes or suggestions.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferReadonlyParameterTypesRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: empty function and empty parameter list ----
		{Code: `function empty() {}`},
		// ---- Dimension 4: function expression and nested function boundary ----
		{Code: `const outer = function (arg: readonly string[]) { function inner(value: number) {} };`},
		// ---- Dimension 4: class expression, method, and class-field arrow ----
		{Code: `const C = class { method(arg: readonly string[]) {} field = (arg: number) => {}; };`},
		// ---- Dimension 4: async, generator, and async-generator containers ----
		{Code: `async function a(x: ReadonlyArray<string>) {} function* g(x: number) {} async function* ag(x: () => void) {}`},
		// ---- Dimension 4: getter/setter and body-absent abstract/declare forms ----
		{Code: `abstract class C { abstract m(x: readonly string[]): void; set value(x: Readonly<{ p: string }>) {} } declare function f(x: number): void;`},
		// ---- Dimension 4: rest element and empty binding patterns degrade gracefully ----
		{Code: `function f(...args: readonly string[]) {} function g({}: {}) {} function h([]: readonly []) {}`},
		// Locks in upstream listener arm 1: an inferred parameter is checked by default.
		{Code: `const f = (value = 1) => value;`},
		// Locks in upstream ignoreInferredTypes arm: inferred parameters are skipped.
		{Code: `const f = (value = []) => value;`, Options: map[string]any{"ignoreInferredTypes": true}},
		// An ESTree AssignmentPattern has no direct type annotation, even when its
		// left-hand parameter does, so ignoreInferredTypes skips typed defaults too.
		{Code: `function f(value: string[] = []) {}`, Options: map[string]any{"ignoreInferredTypes": true}},
		// Locks in getParameterType() annotation arm: alias identity survives a union.
		{Code: `type Allowed = string[] | null; function f(value: Allowed) {}`, Options: map[string]any{"allow": []any{"Allowed"}}},
		// Locks in isTypeReadonly() union arm: every constituent is readonly.
		{Code: `function f(value: string | readonly number[]) {}`},
		// Locks in isTypeReadonly() conditional-type arm: both outcomes are readonly.
		{Code: `type C<T> = T extends string ? readonly string[] : number; function f<T>(value: C<T>) {}`},
		// Conditional branches are inspected before the outer type arguments are
		// instantiated, matching upstream's walk through the branch type nodes.
		{Code: `type C<T, U> = T extends string ? U : number; function f<T>(value: C<T, string[]>) {}`},
		// Locks in isTypeReadonly() pure-call-signature arm.
		{Code: `function f(value: { (x: number): string }) {}`},
		// Locks in isTypeReadonlyObject() readonly string and number index arms.
		{Code: `function f(value: { readonly [key: string]: { readonly x: number }; readonly [key: number]: { readonly x: number } }) {}`},
		// Locks in the private-property exemption.
		{Code: `class Secret { value = 1; #mutable = []; } function f(value: Readonly<Secret>) {}`},
		// Private properties stay exempt during deep value traversal.
		{Code: `class State { #items: string[] = [] } function consume(value: State) {}`},
		// Upstream considers only string and number index signatures.
		{Code: `type T = { [key: symbol]: string[] }; function f(value: T) {}`},
		{Code: "type T = { [key: `prefix-${string}`]: string[] }; function f(value: T) {}"},
		// Locks in treatMethodsAsReadonly=true for mutable collection methods.
		{Code: `function f(value: ReadonlySet<string>) {}`, Options: map[string]any{"treatMethodsAsReadonly": true}},
		// Mapped property symbols retain their source method semantics.
		{Code: `interface S { method(): void } type M<T> = { [K in keyof T]: T[K] }; function f(value: M<S>) {}`, Options: map[string]any{"treatMethodsAsReadonly": true}},
		// Computed unique-symbol properties need symbol-aware type lookup.
		{Code: `declare const tag: unique symbol; interface S { readonly [tag]: string[] } function f(value: S) {}`},
		// Const declarations are readonly property declarations upstream.
		{Code: `namespace Constants { export const version = 1 } function consume(value: typeof Constants) {}`},
		// Readonly assignment declarations from checked JavaScript are also readonly.
		{Code: `const state = {}; Object.defineProperty(state, "items", { value: [], writable: false }); /** @param {typeof state} value */ function consume(value) {}`, FileName: "file.mjs", TSConfig: "tsconfig.allow-js.json"},
		// Locks in default option equivalence: omitted options and [{}] are identical.
		{Code: `function f(value: Readonly<{ x: string }>) {}`, Options: map[string]any{}},
		// ---- Real-user: #1790 computed unique-symbol brand ----
		{Code: `declare const tag: unique symbol; type Branded = string & { readonly [tag]: true }; function f(value: Branded) {}`},
		// ---- Real-user: #5875 mutually recursive readonly index signatures ----
		{Code: `interface A { readonly [key: string]: B } interface B { readonly [key: number]: A } function f(value: A) {}`},
		// ---- Real-user: #11725 top-level alias allowlist around an inline union ----
		{Code: `interface Item { value: string } type Input = Item | Item[] | null; function f(value: Input) {}`, Options: map[string]any{"allow": []any{map[string]any{"from": "file", "name": "Input"}}}},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: function expression reports its mutable parameter ----
		{Code: `const f = function (value: string[]) {};`, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(21, 36)}},
		// ---- Dimension 4: class-field arrow reports independently from its class ----
		{Code: `class C { field = (value: { x: string }) => {}; }`, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(20, 40)}},
		// ---- Dimension 4: async generator container ----
		{Code: `async function* f(value: string[]) {}`, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(19, 34)}},
		// ---- Dimension 4: rest binding parameter ----
		{Code: `function f(...args: string[]) {}`, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(12, 29)}},
		// ---- Dimension 4: object binding pattern does not mask a sibling ----
		{Code: `function f({ x }: { x: string }, sibling: string[]) {}`, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(12, 32), extraReadonlyError(34, 51)}},
		// Locks in checkParameterProperties=false: skip only the property, not its sibling.
		{Code: `class C { constructor(private kept: string[], checked: string[]) {} }`, Options: map[string]any{"checkParameterProperties": false}, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(47, 64)}},
		// Locks in checkParameterProperties=true and TSParameterProperty range unwrapping.
		{Code: `class C { constructor(public value: string[]) {} }`, Options: map[string]any{"checkParameterProperties": true}, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(30, 45)}},
		// Locks in ignoreInferredTypes=false's inferred-type branch.
		{Code: `const f = (value = []) => value;`, Options: map[string]any{"ignoreInferredTypes": false}, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(12, 22)}},
		// Locks in union arm: one mutable constituent makes the union mutable.
		{Code: `function f(value: string | number[]) {}`, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(12, 36)}},
		// Locks in default option equivalence: [{}] matches omitted options.
		{Code: `function f(value: string[]) {}`, Options: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(12, 27)}},
		// Locks in array-intersection arm: every array part must be readonly.
		{Code: `function f(value: string[] & { readonly tag: true }) {}`, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(12, 52)}},
		// Locks in conditional-type arm: one mutable outcome makes the type mutable.
		{Code: `type C<T> = T extends string ? string[] : number; function f<T>(value: C<T>) {}`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeReadonly", Message: "Parameter should be a read only type.", Line: 1, Column: 65, EndLine: 1, EndColumn: 76}}},
		// Locks in object-property deep recursion.
		{Code: `function f(value: { readonly nested: { mutable: string } }) {}`, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(12, 59)}},
		// Locks in mutable index-signature rejection.
		{Code: `function f(value: { [key: string]: string }) {}`, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(12, 44)}},
		// Locks in treatMethodsAsReadonly=false's method-property branch.
		{Code: `function f(value: ReadonlySet<string>) {}`, Options: map[string]any{"treatMethodsAsReadonly": false}, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(12, 38)}},
		// A mapped -readonly modifier applies to computed unique-symbol keys.
		{Code: `declare const tag: unique symbol; interface S { readonly [tag]: string } type M<T> = { -readonly [K in keyof T]: T[K] }; function f(value: M<S>) {}`, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(133, 144)}},
		// Key remapping does not inherit readonly from the source property.
		{Code: "interface S { readonly item: string } type M<T> = { [K in keyof T as `get${Capitalize<K & string>}`]: T[K] }; function f(value: M<S>) {}", Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(122, 133)}},
		// Explicit mapped modifiers also apply to well-known symbol properties.
		{Code: `interface S { readonly [Symbol.iterator]: string } type M<T> = { -readonly [K in keyof T]: T[K] }; function f(value: M<S>) {}`, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(111, 122)}},
		// Locks in allowlist provenance: a file type does not match a lib specifier.
		{Code: `interface Local { x: string } function f(value: Local) {}`, Options: map[string]any{"allow": []any{map[string]any{"from": "lib", "name": "Local"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeReadonly", Message: "Parameter should be a read only type.", Line: 1, Column: 42, EndLine: 1, EndColumn: 54}}},
		// ---- Real-user: #8013 mutable Set remains rejected beside ReadonlySet ----
		{Code: `function f(value: Set<string>) {}`, Options: map[string]any{"treatMethodsAsReadonly": false}, Errors: []rule_tester.InvalidTestCaseError{extraReadonlyError(12, 30)}},
	})
}
