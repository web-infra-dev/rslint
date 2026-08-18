// TestNoPlusplusExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
//
// Dimension 4 walk (rows that don't apply to no-plusplus, with reasons):
//   - N/A access / key forms (identifier, string, numeric, private,
//     computed, element access): the rule inspects only the operator of a
//     Prefix/PostfixUnaryExpression, never a property key.
//   - N/A optional chain (X?.y, X?.()): `++`/`--` require a simple
//     assignment target; an optional-chain expression is never a valid
//     UpdateExpression operand.
//   - N/A TS type-expression wrappers on the operand (`(X as any).y`,
//     `X satisfies T`): the rule never inspects the operand's shape, only
//     the enclosing UpdateExpression's own operator and ancestor chain.
//   - N/A declaration/container forms (class/function declaration vs
//     expression, async/generator variants): the rule fires the same way in
//     every enclosing container; no listener targets a declaration.
//   - N/A autofix boundaries: the rule has neither an autofix nor a
//     suggestion.
//   - N/A graceful degradation around SpreadAssignment / RestElement / empty
//     bodies / overload signatures: the rule never inspects object, array,
//     or class-member shapes.
package no_plusplus

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoPlusplusExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoPlusplusRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized wrapper directly around the update expression ----
			// ESTree has no ParenthesizedExpression node, so `(i++)` as a for
			// loop's afterthought sees ForStatement as i++'s direct parent. tsgo
			// interposes a ParenthesizedExpression node instead; the rule must
			// walk past it (and past multiple nested wrappers) to match upstream.
			{Code: `for (;; (i++));`, Options: allowForLoopAfterthoughts},
			{Code: `for (;; ((i++)));`, Options: allowForLoopAfterthoughts},

			// ---- Dimension 4: parens around one member of the comma chain ----
			// A ParenthesizedExpression can wrap either operand of the comma
			// BinaryExpression without changing which chain the operand belongs
			// to; the walk must skip parens found *between* comma levels too.
			{Code: `for (;; foo(), (i++));`, Options: allowForLoopAfterthoughts},
			{Code: `for (;; (i++), foo());`, Options: allowForLoopAfterthoughts},

			// ---- Dimension 4: parens around the whole comma chain ----
			{Code: `for (;; (foo(), i++));`, Options: allowForLoopAfterthoughts},

			// ---- Real-user: eslint/eslint#13005 (two-pointer variant, mixed ++/--) ----
			// The comma-chain climb must treat ++ and -- identically — this is
			// the common two-pointer-swap idiom that motivated upstream to widen
			// allowForLoopAfterthoughts to arbitrary comma-operand position.
			{Code: `for (let i = 0, j = arr.length - 1; i < j; i++, j--) { swap(arr, i, j); }`, Options: allowForLoopAfterthoughts},

			// ---- Dimension 2: nested for loops — each loop's own incrementor is independently allowed ----
			// Locks in that the Incrementor identity check (parent.AsForStatement().Incrementor == current)
			// is scoped per-ForStatement and doesn't bleed across nesting: the
			// outer loop's `i++` and the inner loop's `j++` are each their own
			// loop's afterthought.
			{Code: `for (let i = 0;; i++) { for (let j = 0;; j++) { doSomething(i, j); } }`, Options: allowForLoopAfterthoughts},

			// ---- Options contract: explicit {} uses the schema default (false is a no-op distinction here since these are valid regardless) ----
			{Code: `var foo = 0; foo=+1;`, Options: []any{map[string]any{}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Real-user: eslint/eslint#11343 — ++ nested inside a LogicalExpression, not a for-loop context ----
			{
				Code: `let c = b || a++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 14, EndLine: 1, EndColumn: 17},
				},
			},

			// ---- Dimension 2: nesting boundary — a statement inside the loop body that merely looks like an afterthought is still reported ----
			// Locks in that only the exact node stored in ForStatement.Incrementor
			// is exempted; an `i++` statement in the loop body (even one that
			// mirrors the loop variable) is not.
			{
				Code:    `for (let i = 0; i < 5;) { i++; }`,
				Options: allowForLoopAfterthoughts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 27, EndLine: 1, EndColumn: 30},
				},
			},

			// Locks in the same nesting-boundary guarantee from the invalid side:
			// a bare `i++;` statement inside the inner loop's body is reported
			// even though both enclosing loops have their own allowed afterthoughts.
			{
				Code:    `outer: for (let i = 0;; i++) { for (let j = 0;; j++) { i++; break outer; } }`,
				Options: allowForLoopAfterthoughts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 56, EndLine: 1, EndColumn: 59},
				},
			},

			// ---- Options contract: schema default is false; explicit false is not a no-op difference from omitting the option ----
			{
				Code:    `for (;; i++);`,
				Options: []any{map[string]any{"allowForLoopAfterthoughts": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:    `for (;; i++);`,
				Options: []any{map[string]any{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12},
				},
			},
		},
	)
}
