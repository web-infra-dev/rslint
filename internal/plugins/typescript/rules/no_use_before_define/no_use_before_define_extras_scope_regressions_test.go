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
