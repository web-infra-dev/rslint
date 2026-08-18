package no_negated_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoNegatedConditionExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
func TestNoNegatedConditionExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNegatedConditionRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver / expression wrappers ----
			// A TS type-expression wrapper on the *whole* test changes its
			// top-level Kind away from Unary/Binary, so it no longer matches
			// — mirrors how ESLint's own ESTree check (test.type === ...)
			// would also miss these once the parser records the cast node.
			{Code: `if ((a != b) as any) {} else {}`},
			{Code: `if ((a != b) satisfies boolean) {} else {}`},

			// Locks in isNegatedUnaryExpression() operator gate: any prefix
			// unary operator other than `!` must not match.
			{Code: `if (-a) {} else {}`},
			{Code: `if (typeof a) {} else {}`},

			// Locks in isNegatedBinaryExpression() operator gate: tsgo
			// collapses ESTree's LogicalExpression (`&&`/`||`) into the same
			// BinaryExpression kind as equality operators, unlike ESLint's
			// ESTree where these never reach the isNegatedBinaryExpression
			// check at all (different node type). Must still fall through.
			{Code: `if (a && b) {} else {}`},
			{Code: `if (a || b) {} else {}`},

			// ---- Dimension 4: access / key forms ----
			// N/A: the rule inspects only the shape of the test expression
			// (unary/binary operator), never object/class property keys.

			// ---- Dimension 4: declaration/container forms ----
			// N/A: the rule matches IfStatement/ConditionalExpression nodes
			// directly and does not special-case the enclosing function or
			// class container.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: receiver / expression wrappers ----
			// Single- and multi-level parenthesized test: tsgo preserves
			// ParenthesizedExpression as an explicit node (ESTree flattens
			// it away), so the top-level Kind check must SkipParentheses
			// first or these silently stop matching.
			{
				Code: "if ((!a)) {} else {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: "if (((!a))) {} else {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: "if ((a != b)) {} else {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			// TS non-null assertion on the unary operand: `!a!` parses as
			// `!(a!)` — the outer PrefixUnaryExpression operator is still
			// `!`, so this must match regardless of what the operand is.
			{
				Code: "if (!a!) {} else {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// Optional chain as the unary operand: the `?.` flag lives on
			// the inner PropertyAccessExpression, not the outer `!`.
			{
				Code: "if (!a?.b) {} else {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},

			// ---- Dimension 4: nesting / traversal boundaries ----
			// Locks in upstream's per-IfStatement visitor semantics: an
			// `else if` link is itself a distinct IfStatement node, visited
			// independently of its parent. The outer if's alternate is an
			// IfStatement (skipped by hasElseWithoutCondition), but the
			// inner else-if link's own alternate is a plain `else`, so only
			// the inner one reports.
			{
				Code: "if (a) {} else if (!b) {} else {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 16, EndLine: 1, EndColumn: 34},
				},
			},
			// Same lock-in with a longer else-if chain: only the last link
			// (whose own alternate is the final plain `else`) reports.
			{
				Code: "if (a) {} else if (b) {} else if (!c) {} else {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 31, EndLine: 1, EndColumn: 49},
				},
			},
			// Nested (non-else-if) if inside a negated outer if: the rule
			// evaluates each IfStatement independently, so the outer
			// reports on its own test/alternate without being affected by
			// the inner (non-negated, so non-matching) if-else.
			{
				Code: "if (!a) { if (b) {} else {} } else {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},
			// Nested ConditionalExpression: only the inner ternary (whose
			// own test is negated) reports, not the outer one.
			{
				Code: "a ? b : !c ? d : e",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 9, EndLine: 1, EndColumn: 19},
				},
			},

			// ---- Dimension 1: AST node types — non-block statement bodies ----
			// `if`/`else` without braces still produce IfStatement /
			// ElseStatement nodes; the rule doesn't inspect body shape.
			{
				Code: "if (!a); else;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},

			// ---- Real-user: negated optional-chain guard ----
			// A common production pattern: guarding on a negated optional
			// property access before an else branch.
			{
				Code: `if (!user?.isActive) { markInactive(); } else { markActive(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 64},
				},
			},
			// ---- Real-user: negated member access off a parenthesized TS cast ----
			// The `as` cast here wraps only a sub-expression (the receiver
			// of `.isValid`), not the whole test, so it does not shield the
			// outer negation the way wrapping the entire test does above.
			{
				Code: `if (!(value as MyType).isValid) { reject(); } else { accept(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegated", Line: 1, Column: 1, EndLine: 1, EndColumn: 65},
				},
			},
		},
	)
}
