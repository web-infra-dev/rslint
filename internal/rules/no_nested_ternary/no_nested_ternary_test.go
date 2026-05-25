package no_nested_ternary

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNestedTernaryRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNestedTernaryRule,
		[]rule_tester.ValidTestCase{
			// Baseline: plain ternaries from the original ESLint test suite.
			{Code: `foo ? doBar() : doBaz();`},
			{Code: `var foo = bar === baz ? qux : quxx;`},

			// Two independent (non-nested) ternaries in sequence.
			{Code: `a ? b : c; d ? e : f;`},

			// Ternary whose test is itself a ternary — ESLint does NOT flag this
			// because only consequent/alternate are checked.
			{Code: `var x = (a ? b : c) ? d : e;`},

			// Ternary branch is an arrow whose body contains a ternary — the ternary
			// is inside the arrow, not a direct branch of the outer.
			{Code: `var x = a ? () => b : () => (c ? d : e);`},

			// Ternary branch is an object/array/call containing a ternary — the
			// inner ternary is not a direct branch.
			{Code: `var x = a ? { k: b ? c : d } : e;`},
			{Code: `var x = a ? [b ? c : d] : e;`},
			{Code: `var x = a ? foo(b ? c : d) : e;`},

			// Ternary branch wrapped in a TypeScript-only outer expression that
			// ESTree does NOT strip (TSAsExpression / TSNonNullExpression /
			// TSSatisfiesExpression / TSTypeAssertion). ESLint does not flag these.
			{Code: `var x = a ? (b ? c : d) as any : e;`},
			{Code: `var x = a ? (b ? c : d)! : e;`},
			{Code: `var x = a ? (b ? c : d) satisfies unknown : e;`},
			{Code: `var x = a ? <any>(b ? c : d) : e;`},

			// TypeScript conditional TYPE — a different AST kind, not flagged.
			{Code: `type T = A extends B ? C : D extends E ? F : G;`},

			// Composite TS wrapper around a nested ternary branch — `as` and `!`
			// are not stripped, so the direct branch is NonNullExpression.
			{Code: `var x = a ? ((b ? c : d) as any)! : e;`},

			// JSX element as a ternary branch — single-level ternary, no nesting.
			{Code: `var el = a ? <A/> : <B/>;`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Nested in alternate.
			{
				Code: `foo ? bar : baz === qux ? quxx : foobar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 1},
				},
			},
			// Nested in consequent.
			{
				Code: `foo ? baz === qux ? quxx : foobar : bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 1},
				},
			},
			// Parenthesized consequent / alternate / double parens.
			{
				Code: `var a = foo ? (bar ? baz : qux) : quux;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 9},
				},
			},
			{
				Code: `var a = foo ? bar : (baz ? qux : quux);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 9},
				},
			},
			{
				Code: `var a = foo ? ((bar ? baz : qux)) : quux;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 9},
				},
			},
			// Both branches are ternaries — parses as `a ? (b?c:d) : (e?f:g)`;
			// only the outer gets a single report (rule reports at most once per parent).
			{
				Code: `var a = a ? b ? c : d : e ? f : g;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 9},
				},
			},
			// Right-associative chain — flags the outer AND the middle.
			{
				Code: `var a = a ? b : c ? d : e ? f : g;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 9},
					{MessageId: "noNestedTernary", Line: 1, Column: 17},
				},
			},
			// Comments inside a parenthesized branch should not hide the nesting.
			{
				Code: `var a = foo ? /* x */ (bar ? baz : qux) /* y */ : quux;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 9},
				},
			},
			// Rule fires wherever the parent ternary lives — call argument, template,
			// return, arrow body — ensure position bookkeeping is driven by the
			// parent ternary and not its container.
			{
				Code: `foo(a ? b ? c : d : e);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 5},
				},
			},
			{
				Code: "var s = `${a ? b : c ? d : e}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 12},
				},
			},
			{
				Code: `function f() { return a ? b : c ? d : e; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 23},
				},
			},
			{
				Code: `const f = () => a ? b : c ? d : e;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 17},
				},
			},
			// Multi-line formatting — still one parent, still one report.
			{
				Code: "var a = foo\n  ? bar\n  : baz ? qux : quux;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 9},
				},
			},
			// Opaque wrappers (unary / statement heads / computed property name)
			// do not hide a nested ternary one level below — the listener fires
			// on every ConditionalExpression regardless of its container.
			{
				Code: `!(a ? b : c ? d : e);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 3},
				},
			},
			{
				Code: `if (a ? b : c ? d : e) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 5},
				},
			},
			{
				Code: `var o = { [a ? b : c ? d : e]: 1 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 12},
				},
			},
			{
				Code: `throw a ? b : c ? d : e;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 7},
				},
			},
			// Spread element container.
			{
				Code: `foo(...(a ? b : c ? d : e));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 9},
				},
			},
			// Element access index position.
			{
				Code: `obj[a ? b : c ? d : e];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 5},
				},
			},
			// JSX expression container wrapping a nested ternary.
			{
				Code: `var el = <div>{a ? b : c ? d : e}</div>;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 16},
				},
			},
			// Ternary whose alternate is a nested ternary, with JSX branches.
			{
				Code: `var el = a ? <A/> : b ? <B/> : <C/>;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 10},
				},
			},
			// JSX attribute value — computed via JSX expression container.
			{
				Code: `var el = <div x={a ? b : c ? d : e} />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNestedTernary", Line: 1, Column: 18},
				},
			},
		},
	)
}
