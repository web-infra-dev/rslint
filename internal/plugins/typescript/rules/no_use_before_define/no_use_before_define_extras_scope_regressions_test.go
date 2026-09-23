// TestNoUseBeforeDefineExtrasScopeRegressions locks in scope-manager details
// that are easy to lose when consuming the shared scope model: merged binding
// identifiers, function-type boundaries, and independent value/type reference
// spaces. The migrated upstream suite lives in
// no_use_before_define_upstream_test.go; other rslint-specific cases live in
// no_use_before_define_extras_test.go.
package no_use_before_define

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUseBeforeDefineExtrasScopeRegressions(t *testing.T) {
	checkTypes := func() map[string]interface{} {
		return map[string]interface{}{"ignoreTypeReferences": false}
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUseBeforeDefineRule,
		[]rule_tester.ValidTestCase{
			// ---- Merged declarations: the first overload is already a definition ----
			{Code: `function f(x: string): void; f(); function f(x: any) {}`},
			{Code: `function f(x: string): void; f(); function f(x: number): void; function f(x: any) {}`},

			// ---- True functionType scopes remain exempt with type checks enabled ----
			{Code: `type F = (x: Later) => void; interface Later {}`, Options: checkTypes()},
			{Code: `type F = () => Later; interface Later {}`, Options: checkTypes()},
			{Code: `type F = new (x: Later) => Result; interface Later {} interface Result {}`, Options: checkTypes()},
			{Code: `interface I { (x: Later): void } interface Later {}`, Options: checkTypes()},
			{Code: `interface I { new (x: Later): Result } interface Later {} interface Result {}`, Options: checkTypes()},
			{Code: `interface I { method(x: Later): void } interface Later {}`, Options: checkTypes()},

			// ---- A binding with no declaration identifier is skipped ----
			{Code: `enum E { b = a, "a" = 1 }`},
			{Code: `enum E { b = a, "a" = 1, "a" = 2 }`},
			{Code: `enum E { "a" = 1, a = 2, b = a }`},

			// ---- ignoreTypeReferences still exempts dual type/value references ----
			{Code: `const x = 1 as typeof x;`},
			{Code: `export default T; type T = number;`},

			// ---- A type query outside the binding's evaluated initializer is safe ----
			{Code: `const x: typeof x = 1;`, Options: checkTypes()},
			{Code: `const x = () => 1 as typeof x;`, Options: checkTypes()},

			// ---- A predicate naming its own parameter is defined before the use ----
			{Code: `function isString(value: unknown): value is string { return typeof value === "string"; }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Merged declarations: a use before the first overload still reports ----
			{
				Code: `f(); function f(x: string): void; function f(x: any) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 1},
				},
			},
			{
				Code:    `function f(x: Later): void; interface Later {} function f(x: any) {}`,
				Options: checkTypes(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 15},
				},
			},

			// ---- Body-less runtime declarations are ordinary function scopes ----
			{
				Code:    `declare function f(x: Later): void; interface Later {}`,
				Options: checkTypes(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 23},
				},
			},
			{
				Code:    `abstract class C { abstract m(x: Later): void } interface Later {}`,
				Options: checkTypes(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 34},
				},
			},
			{
				Code:    `declare class C { m(x: Later): void } interface Later {}`,
				Options: checkTypes(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 24},
				},
			},
			{
				Code:    `declare namespace N { function f(x: Later): void } interface Later {}`,
				Options: checkTypes(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 37},
				},
			},
			{
				Code:    `class C { m(x: Later): void; m(x: any): void {} } interface Later {}`,
				Options: checkTypes(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 16},
				},
			},

			// ---- Type queries retain their value-reference semantics ----
			{
				Code:    `const x = 1 as typeof x;`,
				Options: checkTypes(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 23},
				},
			},
			{
				Code:    `const [x = 1 as typeof x] = [];`,
				Options: checkTypes(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 24},
				},
			},

			// ---- Skip only bindings with no identifier, not one anonymous definition ----
			{
				Code: `enum E { b = a, "a" = 1, a = 2 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 14},
				},
			},

			// ---- A method-signature computed key is evaluated in the outer scope ----
			{
				Code: `interface I { [key](): void } declare const key: unique symbol;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 16},
				},
			},
			{
				Code: `interface I { [((key))](): void } declare const key: unique symbol;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 18},
				},
			},
			{
				Code: `interface I { [makeKey()](): void } declare function makeKey(): unique symbol;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 16},
				},
			},
			{
				Code: `interface I { [keys.x](): void } declare const keys: { x: unique symbol };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 16},
				},
			},

			// ---- A default-export identifier can resolve to the type space ----
			{
				Code:    `export default T; type T = number;`,
				Options: checkTypes(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 16},
				},
			},
			{
				Code:    `export default ((T)); type T = number;`,
				Options: checkTypes(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 18},
				},
			},
			{
				Code:    `export = ((T)); type T = number;`,
				Options: checkTypes(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 12},
				},
			},

			// ---- A type-predicate parameter name is a value reference ----
			{
				Code: `function f(): x is string { return true; } const x = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 15},
				},
			},
			{
				Code: `function f(): asserts x is string {} const x = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 23},
				},
			},
		},
	)
}

// Expected positions were checked against typescript-eslint v8.70.1.
func TestNoUseBeforeDefineExtrasOptionAndPatternRegressions(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, tc := range []struct {
		code    string
		options any
		ranges  [][2]int
	}{
		// Named exports still use the ordinary options when allowed.
		{`export { a }; const a = 1;`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export { a as b }; let a = 1;`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export { a }; function a() {}`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export { a }; class a {}`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export { a }; enum a { X }`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export { a }; type a = number;`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export type { a }; interface a {}`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{15, 16}}},
		{`export { type a }; type a = number;`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{15, 16}}},
		{`export { a }; namespace a {}`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`const a = 1; export { a };`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, nil},
		{`export { missing };`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, nil},
		{`export { a } from "x";`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, nil},
		{`export { a }; import { a } from "x";`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export { a }; type a = number; function a() {}`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export { a }; function a() {} type a = number;`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export default a; const a = 1;`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{16, 17}}},
		{`export = a; type a = number;`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export { a }; const a = 1;`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false, "variables": false}, [][2]int{{10, 11}}},
		{`export { a }; function a() {}`, map[string]any{"allowNamedExports": true, "functions": false, "ignoreTypeReferences": false}, nil},
		{`export { a }; class a {}`, map[string]any{"allowNamedExports": true, "classes": false, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export { a }; enum a { X }`, map[string]any{"allowNamedExports": true, "enums": false, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export { a }; type a = number;`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false, "typedefs": false}, nil},
		{`export type { a }; interface a {}`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false, "typedefs": false}, nil},
		{`export { type a }; type a = number;`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false, "typedefs": false}, nil},
		{`export { a }; type a = number; function a() {}`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false, "typedefs": false}, nil},
		{`export { a }; type a = number; function a() {}`, map[string]any{"allowNamedExports": true, "functions": false, "ignoreTypeReferences": false}, [][2]int{{10, 11}}},
		{`export { a }; function a() {} type a = number;`, map[string]any{"allowNamedExports": true, "ignoreTypeReferences": false, "typedefs": false}, [][2]int{{10, 11}}},
		{`export { a }; function a() {} type a = number;`, map[string]any{"allowNamedExports": true, "functions": false, "ignoreTypeReferences": false}, nil},
		{`export { a }; function a() {}`, map[string]any{"functions": false}, [][2]int{{10, 11}}},
		{`export { a }; type a = number;`, map[string]any{"typedefs": false}, [][2]int{{10, 11}}},
		{`export { a }; const a = 1;`, map[string]any{}, [][2]int{{10, 11}}},
		{`export { a }; function a() {}`, "nofunc", [][2]int{{10, 11}}},
		// Generic and inferred parameters are upstream Type definitions.
		{`type T<U = V, V = U> = [U, V];`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{12, 13}}},
		{`type T<U = V, V = U> = [U, V];`, map[string]any{"ignoreTypeReferences": false, "typedefs": false}, nil},
		{`interface I<T extends U, U> {}`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{23, 24}}},
		{`interface I<T extends U, U> {}`, map[string]any{"ignoreTypeReferences": false, "typedefs": false}, nil},
		{`class C<T extends U, U> {}`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{19, 20}}},
		{`class C<T extends U, U> {}`, map[string]any{"ignoreTypeReferences": false, "typedefs": false}, nil},
		{`function f<T extends U, U>() {}`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{22, 23}}},
		{`function f<T extends U, U>() {}`, map[string]any{"ignoreTypeReferences": false, "typedefs": false}, nil},
		{`const f = <T extends U, U>() => {};`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{22, 23}}},
		{`const f = <T extends U, U>() => {};`, map[string]any{"ignoreTypeReferences": false, "typedefs": false}, nil},
		{`type T = {[K in keyof U]: U[K]}; type U = {};`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{23, 24}, {27, 28}}},
		{`type T = {[K in keyof U]: U[K]}; type U = {};`, map[string]any{"ignoreTypeReferences": false, "typedefs": false}, nil},
		{`type T = U extends infer U ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{10, 11}, {34, 35}}},
		{`type T = U extends infer U ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false, "typedefs": false}, nil},
		{`type T<X> = X extends () => infer U ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{43, 44}}},
		{`type T<X> = X extends new () => infer U ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{47, 48}}},
		{`type T<X> = X extends (x: infer U) => void ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{50, 51}}},
		{`type T<X> = X extends { m(): infer U } ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{46, 47}}},
		{`type T<X> = X extends { [K in keyof X]: infer U } ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{57, 58}}},
		{`type T<X> = X extends () => { [K in keyof X]: () => infer U } ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{69, 70}}},
		{`type T<X> = X extends (x: infer U) => (x: infer U) => infer U ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{69, 70}}},
		{`type T<X> = X extends (X extends infer U ? U : never) ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{57, 58}, {61, 62}}},
		{`type T<X> = X extends { p: X extends infer U ? U : never } ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{62, 63}, {66, 67}}},
		// Each destructuring default contributes a separate write reference.
		{`([a = a] = []); let a = 1;`, nil, [][2]int{{3, 4}, {3, 4}, {7, 8}}},
		{`let a = 1; ([a = a] = []);`, nil, nil},
		{`for ([a = a] of []) {} let a = 1;`, nil, [][2]int{{7, 8}, {7, 8}, {11, 12}}},
		{`({a = a} = []); let a = 1;`, nil, [][2]int{{3, 4}, {3, 4}, {7, 8}}},
		{`let a = 1; ({a = a} = []);`, nil, nil},
		{`for ({a = a} of []) {} let a = 1;`, nil, [][2]int{{7, 8}, {7, 8}, {11, 12}}},
		{`({p: a = a} = []); let a = 1;`, nil, [][2]int{{6, 7}, {6, 7}, {10, 11}}},
		{`let a = 1; ({p: a = a} = []);`, nil, nil},
		{`for ({p: a = a} of []) {} let a = 1;`, nil, [][2]int{{10, 11}, {10, 11}, {14, 15}}},
		{`([{a = a} = a] = []); let a = 1;`, nil, [][2]int{{4, 5}, {4, 5}, {4, 5}, {8, 9}, {13, 14}}},
		{`let a = 1; ([{a = a} = a] = []);`, nil, nil},
		{`for ([{a = a} = a] of []) {} let a = 1;`, nil, [][2]int{{8, 9}, {8, 9}, {8, 9}, {12, 13}, {17, 18}}},
		{`({p: [a = a] = a} = []); let a = 1;`, nil, [][2]int{{7, 8}, {7, 8}, {7, 8}, {11, 12}, {16, 17}}},
		{`let a = 1; ({p: [a = a] = a} = []);`, nil, nil},
		{`for ({p: [a = a] = a} of []) {} let a = 1;`, nil, [][2]int{{11, 12}, {11, 12}, {11, 12}, {15, 16}, {20, 21}}},
		{`({[a]: a = a} = []); let a = 1;`, nil, [][2]int{{4, 5}, {8, 9}, {8, 9}, {12, 13}}},
		{`let a = 1; ({[a]: a = a} = []);`, nil, nil},
		{`for ({[a]: a = a} of []) {} let a = 1;`, nil, [][2]int{{8, 9}, {12, 13}, {12, 13}, {16, 17}}},
		{`([a.x = a] = []); let a = 1;`, nil, [][2]int{{3, 4}, {9, 10}}},
		{`let a = 1; ([a.x = a] = []);`, nil, nil},
		{`for ([a.x = a] of []) {} let a = 1;`, nil, [][2]int{{7, 8}, {13, 14}}},
		{`([a[a] = a] = []); let a = 1;`, nil, [][2]int{{3, 4}, {5, 6}, {10, 11}}},
		{`let a = 1; ([a[a] = a] = []);`, nil, nil},
		{`for ([a[a] = a] of []) {} let a = 1;`, nil, [][2]int{{7, 8}, {9, 10}, {14, 15}}},
		{`([(a as any) = a] = []); let a = 1;`, nil, [][2]int{{4, 5}, {4, 5}, {16, 17}}},
		{`let a = 1; ([(a as any) = a] = []);`, nil, nil},
		{`for ([(a as any) = a] of []) {} let a = 1;`, nil, [][2]int{{8, 9}, {8, 9}, {20, 21}}},
		{`([a! = a] = []); let a = 1;`, nil, [][2]int{{3, 4}, {3, 4}, {8, 9}}},
		{`let a = 1; ([a! = a] = []);`, nil, nil},
		{`for ([a! = a] of []) {} let a = 1;`, nil, [][2]int{{7, 8}, {7, 8}, {12, 13}}},
		{`const [a = a] = [];`, nil, [][2]int{{12, 13}}},
		{`const {a = a} = [];`, nil, [][2]int{{12, 13}}},
		{`const {p: a = a} = [];`, nil, [][2]int{{15, 16}}},
		{`const [{a = a} = a] = [];`, nil, [][2]int{{13, 14}, {18, 19}}},
		{`const {p: [a = a] = a} = [];`, nil, [][2]int{{16, 17}, {21, 22}}},
		{`const {[a]: a = a} = [];`, nil, [][2]int{{9, 10}, {17, 18}}},
		// Nested conditional branches and pattern assertion reference spaces.
		{`type T<X> = X extends (X extends string ? never : infer U) ? U : never; type U = number;`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`type T<X> = X extends any ? [infer U, U] : never; type U = number;`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`type T<X> = infer U extends U ? U : never; type U = number;`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`type T<X> = X extends <U>() => infer U ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{46, 47}}},
		{`type T<X> = X extends { [U in keyof X]: infer U } ? U : U; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{57, 58}}},
		{`([(a as U) = a] = []); let a = 1; type U = number;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{4, 5}, {4, 5}, {14, 15}}},
		{`({[a = a]: b = a} = {}); let a = 1; let b;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{4, 5}, {8, 9}, {12, 13}, {12, 13}, {16, 17}}},
		{`([b = (a = a)] = []); let a = 1; let b;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{3, 4}, {3, 4}, {8, 9}, {12, 13}}},
		{`([b = get({a})] = []); let a = 1; let b;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{3, 4}, {3, 4}, {12, 13}}},
		{`([{b = cond ? {a} : {}} = {}] = []); let a = 1; let b;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{4, 5}, {4, 5}, {4, 5}, {16, 17}}},
		{`([...{a = a}] = []); let a = 1;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{7, 8}, {7, 8}, {11, 12}}},
		{`let [a = ([b = a] = [])] = []; let b;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{12, 13}, {12, 13}, {16, 17}}},
		{`let a = ([a = 1] = []);`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{11, 12}, {11, 12}}},
		{`function f<T extends U>(p = a) { b; } const z = a; type U = number; let a; let b;`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{22, 23}, {29, 30}, {34, 35}, {49, 50}}},
		{`let a; ([(a as U) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as U) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{16, 17}, {16, 17}}},
		{`let a; ([(a as U<V>) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as U<V>) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{16, 17}, {16, 17}, {18, 19}, {18, 19}}},
		{`let a; ([(a as ns.U<V>) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{16, 18}, {16, 18}}},
		{`let a; ([(a as ns.U<V>) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{16, 18}, {16, 18}, {19, 20}, {19, 20}, {21, 22}, {21, 22}}},
		{`let a; ([(a as typeof U) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as typeof U) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{23, 24}, {23, 24}}},
		{`let a; ([(a as {a: U}) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as {a: U}) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as {[U(V)]: V}) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as {[U(V)]: V}) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{18, 19}, {18, 19}, {20, 21}}},
		{`let a; ([(a as { f(x: U): V }) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as { f(x: U): V }) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as (x: U) => V) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as (x: U) => V) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as <T extends U>() => V) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as <T extends U>() => V) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{27, 28}, {27, 28}}},
		{`let a; ([(a as { [K in keyof U]: V }) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as { [K in keyof U]: V }) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{30, 31}, {30, 31}, {34, 35}, {34, 35}}},
		{`let a; ([(a as U extends infer V ? V : U) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as U extends infer V ? V : U) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{16, 17}, {16, 17}, {32, 33}, {32, 33}, {36, 37}, {36, 37}, {40, 41}, {40, 41}}},
		{`let a; ([(a as import("x").U<V>) = a] = []); type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, nil},
		{`let a; ([(a as import("x").U<V>) = a] = []); class U {} class V {} const ns = {};`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{28, 29}, {28, 29}, {30, 31}, {30, 31}}},
		{`let a; ([(a as U) = a] = []); class U {} class V {} const ns = {};`, nil, [][2]int{{16, 17}, {16, 17}}},
		{`let a; (a as U) = a; type U = number; type V = number; namespace ns {}`, map[string]any{"ignoreTypeReferences": false}, [][2]int{{14, 15}}},
		{`\u0061; const a = 1;`, nil, [][2]int{{1, 7}}},
	} {
		if len(tc.ranges) == 0 {
			valid = append(valid, rule_tester.ValidTestCase{Code: tc.code, Options: tc.options})
			continue
		}
		errors := make([]rule_tester.InvalidTestCaseError, 0, len(tc.ranges))
		for _, span := range tc.ranges {
			errors = append(errors, rule_tester.InvalidTestCaseError{
				MessageId: "noUseBeforeDefine", Line: 1, Column: span[0], EndLine: 1, EndColumn: span[1],
			})
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: tc.code, Options: tc.options, Errors: errors})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUseBeforeDefineRule, valid, invalid)
}
