// TestNoEqNullExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 shape / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
package no_eq_null

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEqNullExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEqNullRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TS non-null assertion wraps the null operand ----
			// `null!` is a TSNonNullExpression around the null keyword, not the
			// null keyword itself — matches typescript-eslint's parser, whose
			// `.type` for the wrapped node is TSNonNullExpression, not Literal,
			// so upstream's own `.type === "Literal"` check would also miss it.
			{Code: `x == null!`},

			// ---- Dimension 4: TS type-expression wrappers around the null
			//      operand (`as`, `satisfies`) ----
			// Same rationale as above: the wrapping AsExpression/SatisfiesExpression
			// means the operand's node kind is no longer NullKeyword, mirroring
			// typescript-eslint's own `.type !== "Literal"` outcome.
			{Code: `x == (null as any)`},
			{Code: `x == (null satisfies unknown)`},

			// ---- Dimension 4: neither operand is a null literal ----
			// Locks in the "no null operand" arm — a loose-equality comparison
			// between two non-null operands must not report.
			{Code: `1 == 2`},
			{Code: `a == b`},

			// ---- Locks in upstream badOperator arm: === / !== never report,
			//      even with a null operand on either side ----
			{Code: `x !== null`},
			{Code: `null !== x`},
			{Code: `null === null`},

			// ---- Locks in the operator-kind gate: non-equality BinaryExpression
			//      operators (&&, <) with a null operand must not report. tsgo
			//      represents &&/|| as BinaryExpression (no separate
			//      LogicalExpression kind), so this exercises the same listener. ----
			{Code: `x && null`},
			{Code: `null < x`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver, single and multi-level ----
			// tsgo keeps ParenthesizedExpression as an explicit wrapper node;
			// ESLint's ESTree has no such node at all, so `(null)` is
			// indistinguishable from `null` there. utils.IsNullLiteral unwraps
			// parentheses to stay aligned with that ESLint behavior.
			{
				Code: `x == (null)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code: `x == ((null))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},

			// ---- Dimension 4: optional chain on the non-null operand is inert
			//      to the null-literal check on the other operand ----
			{
				Code: `a?.b == null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- Dimension 4: same-kind nesting — only the inner
			//      BinaryExpression (a == null) matches; the outer
			//      (BinaryExpression == b) does not, since its left operand is
			//      itself a BinaryExpression, not a null literal. Exactly one
			//      diagnostic, proving the listener doesn't bleed to the outer
			//      node or double-report. ----
			{
				Code: `a == null == b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},

			// ---- Locks in upstream OR-branch: left-side null with != (upstream
			//      only exercises left-side null with ==, via `null == x`) ----
			{
				Code: `null != x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},

			// ---- Locks in: both operands are null literals — exactly one
			//      diagnostic, not two (the left-null and right-null arms are
			//      combined with ||, not evaluated as separate reports). ----
			{
				Code: `null == null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- Dimension 4: multi-line report range — Line/Column/EndLine/
			//      EndColumn must track the BinaryExpression's actual span, not
			//      collapse to a single line. ----
			{
				Code: "if (\n  x\n  ==\n  null\n) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 3, EndLine: 4, EndColumn: 7},
				},
			},

			// ---- Real-user: defensive `typeof` + loose-null-check idiom, common
			//      in pre-optional-chaining codebases guarding against both
			//      undeclared and null/undefined values in the same condition. ----
			{
				Code: `if (typeof value !== "undefined" && value != null) { doSomething(value); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 37, EndLine: 1, EndColumn: 50},
				},
			},

			// ---- Real-user: array-filter idiom that strips null/undefined
			//      entries via a loose comparison, a very common pattern in
			//      codebases predating `.filter(Boolean)` conventions. ----
			{
				Code: `items.filter(function (item) { return item != null; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 39, EndLine: 1, EndColumn: 51},
				},
			},
		},
	)
}
