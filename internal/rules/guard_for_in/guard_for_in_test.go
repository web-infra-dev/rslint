package guard_for_in

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestGuardForInRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&GuardForInRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// ================================================================
			// ESLint upstream valid cases
			// ================================================================
			{Code: `for (var x in o);`},
			{Code: `for (var x in o) {}`},
			{Code: `for (var x in o) if (x) f();`},
			{Code: `for (var x in o) { if (x) { f(); } }`},
			{Code: `for (var x in o) { if (x) continue; f(); }`},
			{Code: `for (var x in o) { if (x) { continue; } f(); }`},

			// ================================================================
			// Declaration forms
			// ================================================================
			{Code: `for (let x in o) if (x) f();`},
			{Code: `for (const x in o) if (x) f();`},
			{Code: `for (x in o) if (x) f();`},

			// ================================================================
			// Destructuring initializer (body shape is what matters)
			// ================================================================
			{Code: `for (const { a } in o) if (a) f();`},
			{Code: `for (const [a, b] in o) if (a) f();`},

			// ================================================================
			// Labeled continue still counts as a continue-guard
			// ================================================================
			{Code: `outer: for (var x in o) { if (x) continue outer; f(); }`},
			{Code: `outer: for (var x in o) { if (x) { continue outer; } f(); }`},

			// ================================================================
			// If-else: ESLint only inspects consequent. A continue-consequent
			// guards the rest of the body even when an else branch is present.
			// ================================================================
			{Code: `for (var x in o) { if (x) continue; else f(); g(); }`},
			{Code: `for (var x in o) { if (x) { continue; } else { f(); } g(); }`},

			// ================================================================
			// for-of must never be flagged (separate AST kind)
			// ================================================================
			{Code: `for (const x of arr) f();`},
			{Code: `for (const x of arr) { f(); g(); }`},

			// ================================================================
			// Nested for-in, both guarded
			// ================================================================
			{Code: `for (var x in o) { if (x) continue; for (var y in o2) if (y) g(); }`},
			{Code: `for (var x in o) if (x) { for (var y in o2) if (y) g(); }`},

			// ================================================================
			// Wrapped in various scope contexts — rule fires per for-in
			// regardless of enclosing scope.
			// ================================================================
			{Code: `function f() { for (var x in o) if (x) g(); }`},
			{Code: `const f = () => { for (var x in o) if (x) g(); }`},
			{Code: `class A { static { for (var x in o) if (x) g(); } }`},
			{Code: `class A { m() { for (var x in o) if (x) g(); } }`},
			{Code: `(function () { for (var x in o) if (x) g(); })();`},

			// ================================================================
			// Block with only empty if (still matches "block with just if")
			// ================================================================
			{Code: `for (var x in o) { if (x); }`},

			// ================================================================
			// Comments between tokens are trivia — body shape unchanged
			// ================================================================
			{Code: `for (var x in o) /* c */ if (x) f();`},
			{Code: `for (var x in o) { /* c1 */ if (x) /* c2 */ continue; /* c3 */ f(); }`},

			// ================================================================
			// TypeScript expressions on the iterated value don't affect body
			// ================================================================
			{Code: `for (var x in (o)) if (x) f();`},
			{Code: `for (var x in o!) if (x) f();`},
			{Code: `for (var x in o as any) if (x) f();`},

			// ================================================================
			// 3-level nesting, every level guarded
			// ================================================================
			{Code: `for (var a in x) { if (a) continue; for (var b in y) { if (b) continue; for (var c in z) if (c) g(); } }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// ================================================================
			// ESLint upstream invalid cases
			// ================================================================
			{
				Code: `for (var x in o) { if (x) { f(); continue; } g(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				Code: `for (var x in o) { if (x) { continue; f(); } g(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				Code: `for (var x in o) { if (x) { f(); } g(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				Code: `for (var x in o) { if (x) f(); g(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				Code: `for (var x in o) { foo() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				Code: `for (var x in o) foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Single-statement body that isn't EmptyStatement/IfStatement
			// ================================================================
			{
				Code: `for (var x in o) throw x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				Code: `for (var x in o) for (var y in o2) g();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
					{MessageId: "wrap", Line: 1, Column: 18},
				},
			},
			{
				Code: `for (var x in o) lbl: continue;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Block variants that must NOT count as guarded
			// ================================================================
			{
				// First statement is not an IfStatement
				Code: `for (var x in o) { f(); if (x) continue; g(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				// First statement is an empty statement, not an if
				Code: `for (var x in o) { ; if (x) continue; f(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				// First statement is a block, not an if
				Code: `for (var x in o) { { if (x) continue; } f(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				// If-consequent is an EmptyStatement, not a continue
				Code: `for (var x in o) { if (x); f(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				// If-consequent is an expression, not a continue
				Code: `for (var x in o) { if (x) f(); else continue; g(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				// If-consequent is a multi-statement block
				Code: `for (var x in o) { if (x) { continue; continue; } f(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Nested: only the unguarded loop should be reported
			// ================================================================
			{
				// Outer guarded, inner unguarded → inner only
				Code: `for (var x in o) { if (x) continue; for (var y in o2) g(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 37},
				},
			},
			{
				// Outer unguarded, inner guarded → outer only
				Code: `for (var x in o) { for (var y in o2) if (y) g(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				// Both unguarded → two reports
				Code: `for (var x in o) { for (var y in o2) { g(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
					{MessageId: "wrap", Line: 1, Column: 20},
				},
			},

			// ================================================================
			// Multi-line: line/column must reflect the `for` keyword position
			// ================================================================
			{
				Code: "function run() {\n  for (var x in o) {\n    bar();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 2, Column: 3},
				},
			},

			// ================================================================
			// Block first statement is not an IfStatement
			// ================================================================
			{
				// VariableStatement first
				Code: `for (var x in o) { var y = 1; if (x) continue; f(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				// LabeledStatement wrapping an If — still not a plain IfStatement
				Code: `for (var x in o) { lbl: if (x) continue; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// IfStatement with an empty-block consequent — not a continue-guard
			// ================================================================
			{
				Code: `for (var x in o) { if (x) {} else { continue; } f(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Non-If/Empty single-statement bodies
			// ================================================================
			{
				Code: `for (var x in o) switch (x) { case 1: f(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},
			{
				Code: `for (var x in o) do f(); while (false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// 3-level nested for-in where only the deepest is unguarded
			// ================================================================
			{
				Code: `for (var a in x) { if (a) continue; for (var b in y) { if (b) continue; for (var c in z) { g(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "wrap", Line: 1, Column: 73},
				},
			},
		},
	)
}
