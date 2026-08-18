// TestNoUseBeforeDefineExtrasScopeRegressions locks in shared-scope behavior
// that affects the ESLint core rule: overload anchors, computed signature keys,
// type-predicate value references, and TypeScript signature type references.
// The migrated upstream suites and other edge cases live in the sibling
// no_use_before_define_*_test.go files.
package no_use_before_define

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUseBeforeDefineExtrasScopeRegressions(t *testing.T) {
	checkTypes := func() map[string]any {
		return map[string]any{"ignoreTypeReferences": false}
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUseBeforeDefineRule,
		[]rule_tester.ValidTestCase{
			// The first overload signature defines the binding before this call.
			{Code: `function f(x: string): void; f(); function f(x: any) {}`},
			// Direct TSTypeReference children remain ignored by default.
			{Code: `let f: (x: Later) => void; type Later = string;`},
			{Code: `interface I { m(x: Later): void } type Later = string;`},
		},
		[]rule_tester.InvalidTestCase{
			// ESLint core's default type-reference exemption is deliberately
			// narrow: heritage/qualified/export positions remain checked.
			{
				Code: `class C implements Later {} interface Later {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 20},
				},
			},
			{
				Code: `class C implements NS.Later {} namespace NS { export interface Later {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 20},
				},
			},
			{
				Code: `interface C extends NS.Later {} namespace NS { export interface Later {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 21},
				},
			},
			{
				Code: `let x: NS.Later; namespace NS { export type Later = string; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 8},
				},
			},
			{
				Code: `export = X; const X = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 10},
				},
			},

			// ESLint core checks signature type references when the option is off.
			{
				Code:    `let f: (x: Later) => void; type Later = string;`,
				Options: checkTypes(),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 12}},
			},
			{
				Code:    `let f: (this: Later) => void; type Later = string;`,
				Options: checkTypes(),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code:    `let f: new (x: Later) => void; type Later = string;`,
				Options: checkTypes(),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 16}},
			},
			{
				Code:    `let f: <T extends Later>() => void; type Later = string;`,
				Options: checkTypes(),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 19}},
			},
			{
				Code:    `interface I { (x: Later): void } type Later = string;`,
				Options: checkTypes(),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 19}},
			},
			{
				Code:    `interface I { new (x: Later): void } type Later = string;`,
				Options: checkTypes(),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 23}},
			},
			{
				Code:    `interface I { m(x: Later): void } type Later = string;`,
				Options: checkTypes(),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 20}},
			},

			// A method-signature computed key is evaluated in the outer scope.
			{
				Code: `interface I { [key](): void } declare const key: unique symbol;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 16},
				},
			},

			// Parentheses do not stop a default export from resolving both spaces.
			{
				Code: `export default ((T)); type T = number;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 18},
				},
			},

			// A type-predicate parameter name references a value binding.
			{
				Code: `function f(): x is string { return true; } const x = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 15},
				},
			},

			// Core measures an anonymous enum definition from its literal name.
			{
				Code: `enum E { b = a, "a" = 1, a = 2 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 14},
				},
			},
		},
	)
}
