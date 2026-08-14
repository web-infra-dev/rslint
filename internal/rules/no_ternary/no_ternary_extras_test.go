// TestNoTernaryExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in. See no_ternary_upstream_test.go for the migrated upstream suite.
//
// no-ternary's upstream create() has a single unconditional report call with
// no internal branches, so Layer 3 (branch lock-ins) reduces to the one true
// decision the rule makes: which AST kind its listener is scoped to. That is
// locked in below alongside the Layer 2 (Dimension 4 + real-user) cases.
package no_ternary

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoTernaryExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoTernaryRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: Access/key forms — N/A: the rule matches
			// ---- unconditionally on ast.KindConditionalExpression and never
			// ---- inspects property access or key identity, so there is no
			// ---- per-key-form behavior to diverge. ----

			// ---- Dimension 4: Graceful degradation — N/A: every
			// ---- ConditionalExpression syntactically requires all three
			// ---- operands (condition, WhenTrue, WhenFalse), so there is no
			// ---- empty/partial-operand shape to degrade on. ----

			// ---- Locks in: the listener is scoped to
			// ---- ast.KindConditionalExpression and never fires for
			// ---- ast.KindConditionalType — a TS conditional type shares the
			// ---- `?:`-like reading but is a distinct AST kind. ----
			{Code: `type X<T> = T extends string ? "yes" : "no";`},
			{Code: `type Y<T> = T extends string ? (T extends "a" ? 1 : 2) : 3;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver — single-level wrap;
			// ---- reported range excludes the outer parentheses ----
			{
				Code: "var x = (a ? b : c);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				},
			},
			// ---- Dimension 4: `(X) as T` type-expression wrapper on the
			// ---- whole ternary ----
			{
				Code: "(a ? b : c) as any;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 2, EndLine: 1, EndColumn: 11},
				},
			},
			// ---- Dimension 4: `X!` non-null assertion wrapper on the whole
			// ---- ternary ----
			{
				Code: "(a ? b : c)!;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 2, EndLine: 1, EndColumn: 11},
				},
			},
			// ---- Dimension 4: `X?.y` optional chaining inside the condition
			// ---- operand does not change the reported range of the
			// ---- enclosing ConditionalExpression ----
			{
				Code: "a?.b ? c : d;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			// ---- Dimension 4: ternary as a computed property key ----
			{
				Code: "var o = { [a ? b : c]: 1 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 12, EndLine: 1, EndColumn: 21},
				},
			},
			// ---- Dimension 4: ternary as an object property value ----
			{
				Code: "var o = { key: a ? b : c };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
				},
			},
			// ---- Dimension 4: ternary as a class-field initializer ----
			{
				Code: "class C { x = a ? b : c; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
				},
			},
			// ---- Dimension 4: ternary as a default parameter value ----
			{
				Code: "function f(a = b ? c : d) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
				},
			},
			// ---- Dimension 4: ternary as an unbraced arrow-function body ----
			{
				Code: "const f = () => a ? b : c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 17, EndLine: 1, EndColumn: 26},
				},
			},
			// ---- Dimension 4: nesting/traversal boundary — a nested ternary
			// ---- in the WhenFalse position reports independently of the
			// ---- outer one; there is no dedup / boundary suppression ----
			{
				Code: "a ? b : c ? d : e;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
					{MessageId: "noTernaryOperator", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- Dimension 4: nesting/traversal boundary — a ternary inside
			// ---- a nested function body still reports; function boundaries
			// ---- carry no special meaning for this rule ----
			{
				Code: "function f() { function g() { return a ? b : c; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 38, EndLine: 1, EndColumn: 47},
				},
			},
			// ---- Position contract: multi-line ternary — Line/Column mark
			// ---- the start ('a'), EndLine/EndColumn mark the end ('c') on a
			// ---- different line ----
			{
				Code: "var x = a\n  ? b\n  : c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 9, EndLine: 3, EndColumn: 6},
				},
			},
			// ---- Real-user: JSX conditional rendering — the most common
			// ---- real-world shape a `no-ternary` config encounters; the
			// ---- reported range is just the ternary, not the surrounding
			// ---- JSXExpressionContainer or its JSX-element branches ----
			{
				Code: "const el = <div>{cond ? <A /> : <B />}</div>;",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 18, EndLine: 1, EndColumn: 38},
				},
			},
			// ---- Real-user: ternary directly in a default-export
			// ---- declaration ----
			{
				Code: "export default cond ? a : b;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noTernaryOperator", Line: 1, Column: 16, EndLine: 1, EndColumn: 28},
				},
			},
		},
	)
}
