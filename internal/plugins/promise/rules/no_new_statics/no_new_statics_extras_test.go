package no_new_statics_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_new_statics"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoNewStaticsExtras covers edge-shape augmentation (Layer 2) and
// upstream branch lock-ins (Layer 3).
func TestNoNewStaticsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_new_statics.NoNewStaticsRule,
		[]rule_tester.ValidTestCase{

			// ---- Dimension 4: Receiver / expression wrappers ----

			// N/A: optional chain (`new Promise?.resolve()`) is a parse error in JS/TS.

			// TS type-assertion wrapper on object: tsgo preserves the AsExpression,
			// SkipOuterExpressions(OEKParentheses) does not unwrap it, so the object
			// is not seen as Identifier("Promise") — not flagged.
			// (Type B divergence from ESLint/@typescript-eslint/parser which strips the assertion.)
			{Code: `new (Promise as any).resolve()`},
			{Code: `new (<any>Promise).resolve()`},
			// TS non-null assertion on object: same — not unwrapped, not flagged.
			// N/A in practice since `Promise!` is unusual, but verifies graceful skip.
			{Code: `new (Promise!).resolve()`},

			// ---- Dimension 4: Access / key forms ----

			// Computed / bracket access is ElementAccessExpression (not
			// PropertyAccessExpression), so it never matches the listener guard.
			{Code: `new Promise['resolve']()`},
			{Code: `new Promise['reject']()`},

			// ---- Dimension 4: Parenthesized receiver (tsgo preserves, ESTree flattens) ----

			// Single parenthesis around the whole expression: valid because the
			// callee is now a ParenthesizedExpression wrapping the
			// PropertyAccessExpression, not a bare PropertyAccessExpression.
			// After SkipOuterExpressions on node.Expression, we'd get the PAE,
			// but in tsgo `new (Promise.resolve)()` has Expression = paren(PAE)
			// and SkipOuterExpressions would unwrap the paren to PAE which IS a PAE.
			// However, `new (Promise.resolve)()` is flagged if the PAE's object is Promise —
			// so include it here as an *invalid* case (see below).
			// What is NOT flagged: `new (Promise as any).resolve()` (type wrapper, covered above).

			// Property is not in PROMISE_STATICS — branch lock-in (Condition 3).
			{Code: `new Promise.unknown()`},
			{Code: `new Promise.then()`},
			{Code: `new Promise.catch()`},
			{Code: `new Promise.finally()`},

			// Object is not Promise — branch lock-in (Condition 2).
			{Code: `new MyPromise.resolve()`},
			{Code: `new globalThis.Promise.resolve()`},

			// Callee is not a PropertyAccessExpression — branch lock-in (Condition 1).
			{Code: `new Promise()`},
			{Code: `new Foo()`},

			// ---- Real-user: nested in async functions ----
			{Code: `async function f() { return await Promise.resolve(1) }`},

			// ---- Real-user: used as default export ----
			{Code: `export default Promise.resolve(42)`},
		},
		[]rule_tester.InvalidTestCase{

			// ---- Dimension 4: Parenthesized receiver ----

			// `new (Promise).resolve()` — parens around the Promise identifier:
			// `node.Expression` is PAE, PAE.Expression is paren(Identifier(Promise)),
			// SkipOuterExpressions unwraps to Identifier(Promise) → flagged.
			// Fix removes only "new ", leaving the parens in place.
			{
				Code:   `new (Promise).resolve()`,
				Output: []string{`(Promise).resolve()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNewStatic", Message: "Avoid calling 'new' on 'Promise.resolve()'"},
				},
			},
			// Double parens on Promise identifier — fix removes "new ", parens stay.
			{
				Code:   `new ((Promise)).resolve()`,
				Output: []string{`((Promise)).resolve()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNewStatic", Message: "Avoid calling 'new' on 'Promise.resolve()'"},
				},
			},
			// Parens around the whole callee PropertyAccessExpression:
			// node.Expression is paren(PAE), SkipOuterExpressions unwraps to PAE,
			// PAE.Expression is Identifier(Promise) → flagged.
			// Fix removes "new ", parens around callee stay.
			{
				Code:   `new (Promise.resolve)()`,
				Output: []string{`(Promise.resolve)()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNewStatic", Message: "Avoid calling 'new' on 'Promise.resolve()'"},
				},
			},

			// ---- Dimension 3: Autofix — arguments with values ----
			{
				Code:   `new Promise.resolve(42)`,
				Output: []string{`Promise.resolve(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNewStatic", Message: "Avoid calling 'new' on 'Promise.resolve()'"},
				},
			},
			{
				Code:   `new Promise.all([p1, p2])`,
				Output: []string{`Promise.all([p1, p2])`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNewStatic", Message: "Avoid calling 'new' on 'Promise.all()'"},
				},
			},

			// ---- Dimension 3: Autofix — leading whitespace / indentation preserved ----
			{
				Code:   "  new Promise.resolve()",
				Output: []string{"  Promise.resolve()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNewStatic", Message: "Avoid calling 'new' on 'Promise.resolve()'"},
				},
			},

			// ---- Dimension 3: Autofix — comment between `new` and callee ----
			// Only whitespace after `new` is stripped; comments are preserved.
			{
				Code:   "new/*comment*/Promise.resolve()",
				Output: []string{"/*comment*/Promise.resolve()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNewStatic", Message: "Avoid calling 'new' on 'Promise.resolve()'"},
				},
			},
			// Newline between `new` and callee.
			{
				Code:   "new\nPromise.resolve()",
				Output: []string{"Promise.resolve()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNewStatic", Message: "Avoid calling 'new' on 'Promise.resolve()'"},
				},
			},
			// Multiple spaces between `new` and callee.
			{
				Code:   "new  Promise.resolve()",
				Output: []string{"Promise.resolve()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNewStatic", Message: "Avoid calling 'new' on 'Promise.resolve()'"},
				},
			},

			// ---- Real-user: flagged in class method ----
			{
				Code:   `class A { m() { return new Promise.resolve(1) } }`,
				Output: []string{`class A { m() { return Promise.resolve(1) } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNewStatic", Message: "Avoid calling 'new' on 'Promise.resolve()'"},
				},
			},
		},
	)
}
