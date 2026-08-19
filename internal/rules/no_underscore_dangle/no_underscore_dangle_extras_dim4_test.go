package no_underscore_dangle

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnderscoreDangleExtrasDim4 locks in the universal edge shapes the
// upstream test suite doesn't exercise: parenthesized / non-null / type-assertion
// receivers, optional chains, every key form a member, method name and class
// field can take, the declaration and container forms the rule distinguishes,
// and the TS-only declaration shapes upstream's parser models as their own node
// kinds. Each case carries an inline comment naming the row it covers, so a
// future refactor can't silently regress one without breaking a named lock-in.
// Upstream-migrated cases live in no_underscore_dangle_upstream_test.go.
func TestNoUnderscoreDangleExtrasDim4(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnderscoreDangleRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized receiver, single and multi-level ----
			{Code: `(this)._bar;`, Options: map[string]any{"allowAfterThis": true}},
			{Code: `((this))._bar;`, Options: map[string]any{"allowAfterThis": true}},
			{Code: `(this.constructor)._bar;`, Options: map[string]any{"allowAfterThisConstructor": true}},
			{Code: `((this).constructor)._bar;`, Options: map[string]any{"allowAfterThisConstructor": true}},

			// ---- Dimension 4: optional chain ----
			{Code: `this?._bar;`, Options: map[string]any{"allowAfterThis": true}},
			{Code: `this?.constructor?._bar;`, Options: map[string]any{"allowAfterThisConstructor": true}},
			{Code: `foo?.bar?.__proto__;`},

			// ---- Dimension 4: member key forms ----
			{Code: `foo['_bar'];`},
			{Code: "foo[`_bar`];"},
			{Code: `foo[0];`},
			{Code: `foo[Symbol.iterator];`},

			// ---- Dimension 4: method key forms ----
			{Code: `const o = { '_m'() {} };`, Options: map[string]any{"enforceInMethodNames": true}},
			{Code: `const o = { 1_0() {} };`, Options: map[string]any{"enforceInMethodNames": true}},
			{Code: `const o = { ['_' + m]() {} };`, Options: map[string]any{"enforceInMethodNames": true}},
			{Code: `class A { #_m() {} }`, Options: map[string]any{"enforceInMethodNames": true, "allow": []any{"_m"}}},

			// ---- Dimension 4: class field key forms ----
			{Code: `class A { '_f' = 1; }`, Options: map[string]any{"enforceInClassFields": true}},
			{Code: `class A { 0 = 1; }`, Options: map[string]any{"enforceInClassFields": true}},
			{Code: `class A { ['_' + f] = 1; }`, Options: map[string]any{"enforceInClassFields": true}},
			{Code: `class A { #_f = 1; }`, Options: map[string]any{"enforceInClassFields": true, "allow": []any{"_f"}}},

			// ---- Dimension 4: declaration and container forms ----
			{Code: `class _Foo {}`},

			// ---- Dimension 4: graceful degradation on empty and rest/spread shapes ----
			{Code: `const {} = obj;`, Options: map[string]any{"allowInObjectDestructuring": false}},
			{Code: `const [] = arr;`, Options: map[string]any{"allowInArrayDestructuring": false}},
			{Code: `class A {}`, Options: map[string]any{"enforceInClassFields": true, "enforceInMethodNames": true}},
			{Code: `function foo() {}`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `const o = { ..._spread };`, Options: map[string]any{"enforceInMethodNames": true}},
			{Code: `function foo({} = {}) {}`, Options: map[string]any{"allowFunctionParams": false}},

			// ---- Dimension 4: body-less TS declaration forms ----
			{Code: `declare function _foo(): void;`},
			{Code: `abstract class A { abstract _m(): void; }`, Options: map[string]any{"enforceInMethodNames": true}},
			{Code: `abstract class A { abstract _f: string; }`, Options: map[string]any{"enforceInClassFields": true}},
			{Code: `class A { accessor _f = 1; }`, Options: map[string]any{"enforceInClassFields": true}},
			{Code: `interface I { _f: string; _m(_a: string): void; }`, Options: map[string]any{"enforceInClassFields": true, "enforceInMethodNames": true, "allowFunctionParams": false}},
			{Code: `type T = { _f: string; _m(_a: string): void };`, Options: map[string]any{"enforceInClassFields": true, "enforceInMethodNames": true, "allowFunctionParams": false}},
			{Code: `enum E { _A }`},

			// ---- Dimension 4: TS parameter forms ----
			{Code: `class A { constructor(private _x: string) {} }`, Options: map[string]any{"allowFunctionParams": false}},
			{Code: `class A { constructor(readonly _x: string) {} }`, Options: map[string]any{"allowFunctionParams": false}},

			// ---- Dimension 4: JSX tag names are not member expressions ----
			{Code: `const a = <Foo._bar />;`, Tsx: true},
			{Code: `const a = <Foo._bar></Foo._bar>;`, Tsx: true},
			{Code: `const a = <Foo.Bar._baz />;`, Tsx: true},
			{Code: `const a = <div _foo="x" {..._props} />;`, Tsx: true},

			// ---- Dimension 4: catch clause bindings are not variable declarators ----
			{Code: `try {} catch (_e) {}`, Options: map[string]any{"allowInObjectDestructuring": false, "allowInArrayDestructuring": false}},
			{Code: `try {} catch ({ _e }) {}`, Options: map[string]any{"allowInObjectDestructuring": false}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver, single and multi-level ----
			{
				Code: `(foo)._bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code: `((foo))._bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- Dimension 4: TS non-null assertion on the receiver ----
			{
				Code: `foo!._bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    `this!._bar;`,
				Options: map[string]any{"allowAfterThis": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},

			// ---- Dimension 4: TS type-expression wrappers on the receiver ----
			{
				Code: `(foo as any)._bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `(this as any)._bar;`,
				Options: map[string]any{"allowAfterThis": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: `(foo satisfies Foo)._bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Dimension 4: optional chain ----
			{
				Code: `foo?._bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: `foo?.[_bar];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},

			// ---- Dimension 4: member key forms ----
			{
				Code: `foo[_bar];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: `foo[bar._baz];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_baz'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: `class A { m() { this.#_foo; } #_foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 27},
				},
			},

			// ---- Dimension 4: method key forms ----
			{
				Code:    `const o = { _m() {} };`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `const o = { [_m]() {} };`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    `const o = { [(_m)]() {} };`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    `class A { #_m() {} }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '#_m'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 19},
				},
			},

			// ---- Dimension 4: `allow` matches the private name without its `#` ----
			{
				Code:    `class A { #_m() {} }`,
				Options: map[string]any{"enforceInMethodNames": true, "allow": []any{"#_m"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '#_m'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `class A { #_f = 1; }`,
				Options: map[string]any{"enforceInClassFields": true, "allow": []any{"#_f"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '#_f'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 19},
				},
			},
			// ---- Dimension 4: class field key forms ----
			{
				Code:    `class A { [_f] = 1; }`,
				Options: map[string]any{"enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_f'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 20},
				},
			},

			// ---- Dimension 4: declaration and container forms ----
			{
				Code: `const _foo = function _bar() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code: `const _foo = () => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: `const _X = class _Foo {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_X'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: `async function _foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: `function* _foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: `async function* _foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    `class A { async *_m() {} }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `class A { static _m() {} }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `class A { static _f = 1; }`,
				Options: map[string]any{"enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_f'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `class A { _f = () => {}; }`,
				Options: map[string]any{"enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_f'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Dimension 4: `export` keywords are not part of the reported declaration ----
			{
				Code: `export function _foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: `export default function _foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 16, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code: `export default async function* _foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 16, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code: `export async function _foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code: `export const _foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 22},
				},
			},

			// ---- Dimension 4: same-kind nesting must not bleed across the boundary ----
			{
				Code: `function _outer() { function _inner() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_outer'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_inner'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code:    `class _Outer { _f = class _Inner { _g = 1; }; }`,
				Options: map[string]any{"enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_f'.", Line: 1, Column: 16, EndLine: 1, EndColumn: 46},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_g'.", Line: 1, Column: 36, EndLine: 1, EndColumn: 43},
				},
			},
			{
				Code:    `const o = { m() { const p = { _n() {} }; } };`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_n'.", Line: 1, Column: 31, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:    `function foo(_a) { function bar(_b) {} }`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_b'.", Line: 1, Column: 33, EndLine: 1, EndColumn: 35},
				},
			},

			// ---- Dimension 4: graceful degradation on empty and rest/spread shapes ----
			{
				Code:    `const { ..._rest } = obj;`,
				Options: map[string]any{"allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_rest'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `const [ ..._rest ] = arr;`,
				Options: map[string]any{"allowInArrayDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_rest'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Dimension 4: body-less TS declaration forms ----
			{
				Code: `function _foo(_a: string): void;
function _foo(_a: any) {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_foo'.", Line: 2, Column: 1, EndLine: 2, EndColumn: 26},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 2, Column: 15, EndLine: 2, EndColumn: 22},
				},
			},
			{
				Code:    `class A { _m(_a: string): void; _m(_a: any) {} }`,
				Options: map[string]any{"allowFunctionParams": false, "enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 32},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 33, EndLine: 1, EndColumn: 47},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 36, EndLine: 1, EndColumn: 43},
				},
			},
			{
				Code:    `class A { declare _f: string; }`,
				Options: map[string]any{"enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_f'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    `declare class A { _f: string; _m(): void; }`,
				Options: map[string]any{"enforceInClassFields": true, "enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_f'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 30},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 31, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code: `declare const _x: number;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_x'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Dimension 4: JSX tag names are not member expressions ----
			{
				Code: `const a = <Foo bar={obj._baz} />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_baz'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 29},
				},
			},

			// ---- Dimension 4: TS parameter forms ----
			{
				Code:    `class A { constructor(_x: string) {} }`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_x'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:    `function foo(this: T, _a: string) {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:    `function foo(_a?: string) {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `class A { m(@dec _a) {} }`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `class A { m(@dec _a = 1) {} }`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    `class A { m(@dec ..._a: any[]) {} }`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    `function foo<_T>(_a: _T) {}`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_a'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
				},
			},
		},
	)
}
