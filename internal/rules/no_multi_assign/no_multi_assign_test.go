package no_multi_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMultiAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoMultiAssignRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint upstream valid cases ----
			{Code: "var a, b, c,\nd = 0;"},
			{Code: "var a = 1; var b = 2; var c = 3;\nvar d = 0;"},
			{Code: "var a = 1 + (b === 10 ? 5 : 4);"},
			{Code: "const a = 1, b = 2, c = 3;"},
			{Code: "const a = 1;\nconst b = 2;\n const c = 3;"},
			{Code: "for(var a = 0, b = 0;;){}"},
			{Code: "for(let a = 0, b = 0;;){}"},
			{Code: "for(const a = 0, b = 0;;){}"},
			{Code: "export let a, b;"},
			{Code: "export let a,\n b = 0;"},
			{
				Code:    "const x = {};const y = {};x.one = y.one = 1;",
				Options: map[string]interface{}{"ignoreNonDeclaration": true},
			},
			{
				Code:    "let a, b;a = b = 1",
				Options: map[string]interface{}{"ignoreNonDeclaration": true},
			},
			{Code: "class C { [foo = 0] = 0 }"},

			// ---- Additional edge cases (not in upstream suite) ----

			// Assignment in a computed object key — outside the rule's scope
			{Code: "({ [a = 1]: 1 })"},
			// Assignment as a property value (object literal, not class field)
			{Code: "({ a: b = 1 })"},
			// Assignment as default in object pattern destructuring
			{Code: "let { a = 1 } = {};"},
			// Plain assignment with one target — not chained
			{Code: "a = 1;"},
			{Code: "var a = 1;"},
			// Compound assignment alone — not chained
			{Code: "a += 1;"},
			// Logical assignment alone
			{Code: "a ||= 1;"},
			// Destructuring assignment alone
			{Code: "[a, b] = [1, 2];"},
			{Code: "({ a, b } = { a: 1, b: 2 });"},
			// TypeScript type annotation with single value
			{Code: "let x: number = 1;"},
			// Class field without initializer
			{Code: "class C { x; }"},
			// Class field with non-assignment init
			{Code: "class C { x = 1; }"},
			// Skipped declarations
			{Code: "class C { declare x: number; }"},
			// `ignoreNonDeclaration: true` allows property-target assignment chains
			{
				Code:    "let a; let b; a = b = 'baz';",
				Options: map[string]interface{}{"ignoreNonDeclaration": true},
			},
			// `ignoreNonDeclaration: true` with logical assignment in non-declaration
			{
				Code:    "let a, b; a = b ||= 1;",
				Options: map[string]interface{}{"ignoreNonDeclaration": true},
			},
			// Default value inside destructuring pattern is not an init
			{Code: "var { a = 1 } = obj;"},
			{Code: "var { a = b = c } = obj;"},
			{Code: "var [ a = b = c ] = arr;"},
			// Assignment as RHS of a non-assignment binary operator
			{Code: "var a = b + (c = d);"},
			{Code: "var a = (b = c) + d;"},
			// Chain inside computed key of a method/getter (not init / value)
			{Code: "class C { [foo = bar] = 1 }"},
			{Code: "({ [a = b]: 1 })"},
			// Multi-declarator: only the relevant declarator carries a chain
			{Code: "var a = 1, b = 2;"},
			// Logical assignment alone (no chain)
			{Code: "a ??= 1;"},
			// Assignments scattered across conditional branches — each branch
			// is a standalone assignment, not a chain.
			{Code: "var a = cond ? (b = c) : (d = e);"},
			// Assignment as a conditional branch — the conditional sits between
			// the two assignments, so neither side is the `.right` of the other.
			{Code: "a = cond ? b = c : d;"},
			// Sequence (comma) operator wrapping standalone assignments.
			{Code: "var a = (b = 1, c = 2);"},
			// Assignment as a function argument — argument position, not chain.
			{Code: "fn(a = 1, b = 2);"},
			// Assignment inside `return` — not a chain when single.
			{Code: "function f() { return a = 1; }"},
			// Standalone yield with assignment.
			{Code: "function* g() { yield a = 1; }"},
			// Update / unary expressions on declared values
			{Code: "var a = b++;"},
			{Code: "var a = !b;"},
			// Class field whose value is a function — body never counts as init
			{Code: "class C { handler = () => { let a; a = 1; }; }"},
			// Class field whose value is an arrow with chained assignment in
			// BODY — body of arrow is its own scope, the field's init is the
			// arrow node itself, not the inner assignment.
			{Code: "class C { handler = () => { var a; a = b; }; }"},
			// Type-only declaration in TS — no Initializer, can't fire.
			{Code: "let a: number;"},
			// Object property value assignment — not class field, not chained.
			{Code: "var obj = { a: b = 1 };"},
			// JSX-ish attribute literal value (still parseable as TS expr).
			{Code: "var x = { value: a = 1 };"},
			// Named export of a single declaration with no init.
			{Code: "export const PI = 3.14;"},
			// Comment between operator and RHS — should not change semantics.
			{Code: "var a = /* hi */ 1;"},
			// Type assertion / `as` / `satisfies` on init — not assignments.
			{Code: "var a = b as number;"},
			{Code: "var a = (b satisfies number);"},
			{Code: "var a = b!;"},
			// Optional chain on init — not assignment.
			{Code: "var a = b?.c;"},
			// Tagged template / call assignments — single, not chain.
			{Code: "var a = tag`x`;"},
			{Code: "var a = fn();"},
			// Class with multiple non-chained fields — sanity.
			{Code: "class C { a = 1; b = 2; static c = 3; }"},
			// Parameter property in class constructor — not a class field
			// initializer in our sense.
			{Code: "class C { constructor(public x = 1) {} }"},
			// IIFE returning an assignment.
			{Code: "var a = (() => b = 1)();"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ESLint upstream invalid cases ----
			{
				Code: "var a = b = c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a = b = c = d;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
					{MessageId: "unexpectedChain", Line: 1, Column: 13},
				},
			},
			{
				Code: "let foo = bar = cee = 100;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 11},
					{MessageId: "unexpectedChain", Line: 1, Column: 17},
				},
			},
			{
				Code: "a=b=c=d=e",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 3},
					{MessageId: "unexpectedChain", Line: 1, Column: 5},
					{MessageId: "unexpectedChain", Line: 1, Column: 7},
				},
			},
			{
				Code: "a=b=c",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 3},
				},
			},
			{
				Code: "a\n=b\n=c",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 2, Column: 2},
				},
			},
			{
				Code: "var a = (b) = (((c)))",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a = ((b)) = (c)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a = b = ( (c * 12) + 2)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a =\n((b))\n = (c)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 2, Column: 1},
				},
			},
			{
				Code: "a = b = '=' + c + 'foo';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 5},
				},
			},
			{
				Code: "a = b = 7 * 12 + 5;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 5},
				},
			},
			{
				Code:    "const x = {};\nconst y = x.one = 1;",
				Options: map[string]interface{}{"ignoreNonDeclaration": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 2, Column: 11},
				},
			},
			{
				Code:    "let a, b;a = b = 1",
				Options: map[string]interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 14},
				},
			},
			{
				Code:    "let x, y;x = y = 'baz'",
				Options: map[string]interface{}{"ignoreNonDeclaration": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 14},
				},
			},
			{
				Code:    "const a = b = 1",
				Options: map[string]interface{}{"ignoreNonDeclaration": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 11},
				},
			},
			{
				Code: "class C { field = foo = 0 }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 19},
				},
			},
			{
				Code:    "class C { field = foo = 0 }",
				Options: map[string]interface{}{"ignoreNonDeclaration": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 19},
				},
			},

			// ---- Additional edge cases ----

			// Paren-wrapped inner assignment in declaration: ESTree drops parens, so
			// the rule still matches. Locks in the parent walk through paren nodes.
			{
				Code: "var a = (b = c);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 10},
				},
			},
			// Multiple paren layers wrapping the inner assignment.
			{
				Code: "var a = ((b = c));",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 11},
				},
			},
			// Paren wrapping the inner of a non-declaration chain.
			{
				Code: "a = (b = c);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 6},
				},
			},
			// Compound assignment in declaration init: `+=` is still an
			// AssignmentExpression in ESTree, so the rule reports it.
			{
				Code: "var a = b += c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			// Logical assignment in declaration init.
			{
				Code: "var a = b ||= c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			// Logical assignment as inner of non-declaration chain.
			{
				Code: "a = b ??= c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 5},
				},
			},
			// Compound assignment as the OUTER of the chain still triggers
			// the .right selector.
			{
				Code: "a += b = c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 6},
				},
			},
			// For-loop init declaration with chain.
			{
				Code: "for (let i = j = 0;;) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 14},
				},
			},
			// Static class field with chain.
			{
				Code: "class C { static x = y = 1 }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 22},
				},
			},
			// Private class field with chain.
			{
				Code: "class C { #x = y = 1 }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 16},
				},
			},
			// TypeScript-typed declaration with chain.
			{
				Code: "let a: number = b = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 17},
				},
			},
			// `ignoreNonDeclaration: true` still reports declaration chains.
			{
				Code:    "let a = b = 1",
				Options: map[string]interface{}{"ignoreNonDeclaration": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			// Long chain across many operators: every link reported.
			{
				Code: "var a = b = c = d = e = f;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
					{MessageId: "unexpectedChain", Line: 1, Column: 13},
					{MessageId: "unexpectedChain", Line: 1, Column: 17},
					{MessageId: "unexpectedChain", Line: 1, Column: 21},
				},
			},
			// Multi-declarator: only the declarator with a chain reports.
			{
				Code: "var a = b = c, d = e = f;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
					{MessageId: "unexpectedChain", Line: 1, Column: 20},
				},
			},
			// Mix: chain + plain compound on RHS of outer.
			{
				Code: "a = b = c ??= d",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 5},
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			// Chain inside template literal expression — outer is non-declaration.
			{
				Code: "`${a = b = c}`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 8},
				},
			},
			// Chain inside arrow body.
			{
				Code: "() => a = b = c",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 11},
				},
			},
			// Class field with paren-wrapped chain.
			{
				Code: "class C { x = (y = 1); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 16},
				},
			},
			// Nested class chain — outer field's init is itself a class with a field chain.
			{
				Code: "class A { x = class { y = z = 1 } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 27},
				},
			},
			// Chain inside an IIFE init.
			{
				Code: "var a = (() => { return b = c = 1; })();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 29},
				},
			},
			// Chain across newlines / comments.
			{
				Code: "var a = b /* hi */ = c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a =\n  b =\n  c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 2, Column: 3},
				},
			},
			// Chain inside arrow with body block + return.
			{
				Code: "const f = () => { return a = b = 1; };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 30},
				},
			},
			// Chain on member-access targets.
			{
				Code: "obj.a = obj.b = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			// Chain on element-access target.
			{
				Code: "obj['a'] = obj['b'] = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 12},
				},
			},
			// Chain with `this.x`.
			{
				Code: "this.x = this.y = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 10},
				},
			},
			// Chain with computed-style target.
			{
				Code: "a[k] = b[k] = v;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 8},
				},
			},
			// Static class field with paren-wrapped chain.
			{
				Code: "class C { static x = (y = 1); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 23},
				},
			},
			// Chain inside parenthesized init: `var a = ((b) = c)` — left of inner has paren.
			{
				Code: "var a = ((b) = c);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 10},
				},
			},
			// Chain on an LHS through TS `as` cast — `(b as any) = c`. Even though
			// `b as any` cast as LHS is valid TS, it's still an assignment chain.
			{
				Code: "var a = (b as any) = c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			// Mixed-style chain crossing `=` and `**=`.
			{
				Code: "var a = b **= c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 9},
				},
			},
			// Chain assigned through optional class field with definite-assignment `!`.
			{
				Code: "class C { x!: number = y = 1; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 24},
				},
			},
			// `using` declaration with chain (TS 5.2+).
			{
				Code: "{ using a = b = c; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 13},
				},
			},
			// Export with chain.
			{
				Code: "export let a = b = c;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 16},
				},
			},
			// Block-scoped chain.
			{
				Code: "{ let a = b = 1; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 11},
				},
			},
			// Chain inside switch case body.
			{
				Code: "switch (x) { case 1: a = b = 2; break; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 26},
				},
			},
			// Chain with deeply nested parens on inner.
			{
				Code: "var a = ((((b = c))));",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 13},
				},
			},
			// Chain with paren around outer right side: `a = ((b = c))`.
			{
				Code: "a = ((b = c));",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 7},
				},
			},
			// `ignoreNonDeclaration: true` — paren-wrapped declaration init still reports.
			{
				Code:    "let a = (b = 1);",
				Options: map[string]interface{}{"ignoreNonDeclaration": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 10},
				},
			},
			// `ignoreNonDeclaration: true` — class field paren-wrapped still reports.
			{
				Code:    "class C { x = (y = 1); }",
				Options: map[string]interface{}{"ignoreNonDeclaration": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedChain", Line: 1, Column: 16},
				},
			},
		},
	)
}

// Locks in the message-text contract — the rule emits exactly
// "Unexpected chained assignment." on every variant, with no interpolation.
func TestNoMultiAssignMessageText(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoMultiAssignRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code: "a = b = c",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedChain",
						Message:   "Unexpected chained assignment.",
						Line:      1,
						Column:    5,
					},
				},
			},
			{
				Code: "var a = b = c",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedChain",
						Message:   "Unexpected chained assignment.",
						Line:      1,
						Column:    9,
					},
				},
			},
			{
				Code: "class C { x = y = 1 }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedChain",
						Message:   "Unexpected chained assignment.",
						Line:      1,
						Column:    15,
					},
				},
			},
		},
	)
}
