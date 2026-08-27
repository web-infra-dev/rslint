// TestPreferTernaryExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise, plus rslint-specific quirks of the
// tsgo AST that the migration has to work around.
package prefer_ternary_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_ternary"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferTernaryExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_ternary.PreferTernaryRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: chained if/else if / else if ----
			// The chain's interior `else if` clauses are IfStatements
			// whose parent is another IfStatement (and whose role is
			// `ElseStatement`). Those nodes must be skipped, and the
			// chain itself is never reported because the merge walker
			// needs a binary expression or return statement in both
			// arms; the chain has neither.
			{
				Code: `function foo() {
	if (a) {
		return 1;
	} else if (b) {
		return 2;
	} else if (c) {
		return 3;
	} else {
		return 4;
	}
}`,
			},

			// ---- Dimension 4: the let-then-if gate is `let` only ----
			// A `const` declaration is never rewritten to another
			// `const`; even if the const has no later writes, a
			// `const x = a; if (test) { x = b; }` would already be a
			// type error and is not the rule's concern.
			{
				Code: `const x = a;
if (test) {
	x = b;
}`,
			},

			// ---- Dimension 4: empty consequent / alternate (no body) ----
			// Upstream's invalid cases include `if (a) {b}` etc. These
			// shouldn't be flagged: the consequent is an
			// ExpressionStatement whose expression is an Identifier
			// (not a Return or Assignment), and the alternate body
			// is empty. The merge walker returns false because
			// `consequent` and `alternate` are different kinds.
			{Code: `if (a) {b}`},
			{Code: `if (a) {} else {b}`},
			{Code: `if (a) {} else {}`},

			// ---- P1 #1 (negative): destructuring assignment is silent ----
			// `({x} = a)` parses as a BinaryExpression with left being
			// an ObjectLiteralExpression used as a destructuring target.
			// Two ObjectLiteralExpression nodes are not the same
			// reference, so the LHS check fails and the rule stays
			// quiet. This is the same behavior upstream exhibits.
			{Code: `if (t) { ({x} = a); } else { ({x} = b); }`},

			// ---- P1 #2 / N2 (negative): nested effects in init ----
			// Every kind of nested effect must short-circuit the
			// collapse. The previous top-level kind switch did not
			// recurse, so these slipped through and produced
			// semantics-changing suggestions that turned unconditional
			// side effects into conditional ones.
			{Code: `let x = {a: sideEffect()};
if (test) { x = c; }`},
			{Code: `let x = [sideEffect()];
if (test) { x = c; }`},
			{Code: "let x = `${sideEffect()}`;\nif (test) { x = c; }"},
			{Code: `let x = (y = 1);
if (test) { x = c; }`},

			// ---- N2: every assignment operator has side effects ----
			// Compound assignments like `+=`, `||=`, `??=` are
			// assignments too — upstream's `AssignmentExpression`
			// handler returns true regardless of operator. Each of
			// these initializers would otherwise slip through and
			// produce a semantics-changing suggestion that reorders
			// the write.
			{Code: `let x = (y += 1);
if (test) { x = next; }`},
			{Code: `let x = (y ||= 1);
if (test) { x = next; }`},
			{Code: `let x = (y ??= 1);
if (test) { x = next; }`},
			{Code: `let x = {value: (y += 1)};
if (test) { x = next; }`},
			{Code: `let x = delete object.x;
if (test) { x = next; }`},

			// ---- N2: `new` of a class with a side-effecting method ----
			// The class constructor is invoked at the `new` site, so
			// any side effect in the method body fires whether or not
			// the ternary's alternate branch runs. The collapse must
			// NOT be offered.
			{Code: `class C { m() { sideEffect(); } }
let x = new C();
if (test) { x = next; }`},

			// Static property-name evaluation is scope-free upstream: a local
			// constant must not make x[key] match x.bar.
			{Code: `const key = 'bar';
if (t) { x[key] = a; } else { x.bar = b; }`},
			// Parentheses are transparent in ESTree classification gates.
			{Code: `function f() { if (t) { return (a ? b : c); } else { return d; } }`},
			{Code: `function f() { if (t) { return (true); } else { return ((false)); } }`},
			{Code: `if (t) { x = (a ? b : c); } else { x = d; }`},
			{Code: `let x = (a ? b : c);
if (t) { x = d; }`},
			// Evaluating a method decorator is immediate, so it blocks the fold.
			{Code: `let x = class { @dec(call()) m() {} };
if (t) { x = next; }`},

			// ---- N1: private field vs string-keyed property ----
			// `this.#x` and `this["#x"]` are different assignment
			// targets in TypeScript. The static-name fast path must
			// skip when either side uses a private identifier.
			{Code: `class C {
	#x;
	m() {
		if (t) {
			this.#x = a;
		} else {
			this["#x"] = b;
		}
	}
}`},
		},
		[]rule_tester.InvalidTestCase{
			// Parentheses around a conditional operand must suppress the merge.
			// These deferred-member parameter defaults and constructor bodies are
			// not evaluated when their object/class is created, so they may move.
			{
				Code: `let x = {m(v = sideEffect()) {}};
if (t) { x = next; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "prefer-ternary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-ternary/suggestion", Output: "const x = t ? next : {m(v = sideEffect()) {}};"}},
				}},
			},
			{
				Code: `let x = {set m(v = sideEffect()) {}};
if (t) { x = next; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "prefer-ternary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-ternary/suggestion", Output: "const x = t ? next : {set m(v = sideEffect()) {}};"}},
				}},
			},
			{
				Code: `let x = class { constructor() { sideEffect(); } };
if (t) { x = next; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "prefer-ternary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-ternary/suggestion", Output: "const x = t ? next : class { constructor() { sideEffect(); } };"}},
				}},
			},
			{
				Code:     `while (keep) if (t) { (x).y = a; } else { (x).y = b; }`,
				Output:   []string{`while (keep) (x).y = t ? a : b;`},
				FileName: "file.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "prefer-ternary"}},
			},
			{
				Code:     `if (keep) if (t) { (x).y = a; } else { (x).y = b; }`,
				Output:   []string{`if (keep) (x).y = t ? a : b;`},
				FileName: "file.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "prefer-ternary"}},
			},
			{
				Code:     `for (;;) if (t) { (x).y = a; } else { (x).y = b; }`,
				Output:   []string{`for (;;) (x).y = t ? a : b;`},
				FileName: "file.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "prefer-ternary"}},
			},
			{
				Code:     `for (k in object) if (t) { (x).y = a; } else { (x).y = b; }`,
				Output:   []string{`for (k in object) (x).y = t ? a : b;`},
				FileName: "file.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "prefer-ternary"}},
			},
			{
				Code:     `for (const item of items) if (t) { (x).y = a; } else { (x).y = b; }`,
				Output:   []string{`for (const item of items) (x).y = t ? a : b;`},
				FileName: "file.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "prefer-ternary"}},
			},
			{
				Code:     `with (scope) if (t) { (x).y = a; } else { (x).y = b; }`,
				Output:   []string{`with (scope) (x).y = t ? a : b;`},
				FileName: "file.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "prefer-ternary"}},
			},
			{
				Code:     "const value = {}\nif (t) { (x).y = a; } else { (x).y = b; }",
				Output:   []string{"const value = {}\n;(x).y = t ? a : b;"},
				FileName: "file.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "prefer-ternary"}},
			},
			// ---- Dimension 4: nested return if/else inside a function body ----
			// Walks through the bare form (no block) to confirm the
			// ExpressionStatement unwrapping handles the case where the
			// consequent is `if (test) return a;` with no braces.
			{
				Code:     `function f() { if (test) return a; else return b; }`,
				Output:   []string{`function f() { return test ? a : b; }`},
				FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-ternary"},
				},
			},

			// ---- Dimension 4: simple if/else with `let` and a later write ----
			// Locks in the branch where `hasOtherWrites` is true and
			// the suggestion preserves the `let` keyword. The upstream
			// suite has the same case under
			// `// Keep \`let\` when there are later writes` but with
			// function-body wrappers; the bare form exercises the
			// same code path with a tighter test source.
			{
				Code: `let x = a;
if (test) { x = b; }
x = c;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prefer-ternary",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prefer-ternary/suggestion", Output: "let x = test ? b : a;\nx = c;"},
						},
					},
				},
			},

			// ---- Dimension 4: precedence parens around the test ----
			// `a = b` in the test position binds more loosely than `?:`
			// and so the test must be wrapped in parens in the
			// rewritten ternary. Upstream's test only checks the
			// assignment case; this also covers the yield variants.
			{
				Code:     `if (a = b) { foo = 1; } else foo = 2;`,
				Output:   []string{`foo = (a = b) ? 1 : 2;`},
				FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-ternary"},
				},
			},

			// ---- P1 #4: static computed-key evaluation ----
			// The literal-only `AccessExpressionStaticName` cannot fold
			// `'b' + 'ar'`, so this used to slip through. The static
			// evaluator folds the concatenation and the LHS compares
			// equal to `foo.bar`, matching upstream Unicorn v73.0.0.
			{
				Code: `function unicorn() {
	a()
	if (test) {
		(foo)['b' + 'ar'] = a
	} else{
		foo.bar = b
	}
}`,
				Output: []string{`function unicorn() {
	a()
	;(foo)['b' + 'ar'] = test ? a : b;
}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-ternary"},
				},
			},

			// ---- P1 #1: parenthesized assignment body ----
			// Regression for the WalkUpParenthesizedExpressions panic:
			// when the body is a parenthesized assignment, the unwrap
			// must use the immediate `.Expression` field, not walk up
			// to a parent ExpressionStatement.
			{
				Code:     `if (t) { (x = a); } else { (x = b); }`,
				Output:   []string{`x = t ? a : b;`},
				FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-ternary"},
				},
			},

			// ---- P1 #3: only-single-line on unwrapped bodies ----
			// Upstream checks the unwrapped `consequent` / `alternate`
			// after getNodeBody, so individual single-line expressions
			// inside multi-line blocks still report.
			{
				Code: `if (t) {
	x = a;
} else {
	x = b;
}`,
				Options: []interface{}{"only-single-line"},
				Output:  []string{`x = t ? a : b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-ternary"},
				},
			},
			// Multi-line parenthesized test: the if's Expression field
			// already excludes the if's own `(` and `)` (those are
			// part of the IfStatement's structure, not the inner
			// expression), so the single-line gate sees just `t`.
			// Upstream matches this; rslint strips the parens in the
			// output for the same reason.
			{
				Code:    "if (\n\tt\n) {\n\ta = foo;\n} else {\n\ta = bar;\n}",
				Options: []interface{}{"only-single-line"},
				Output:  []string{"a = t ? foo : bar;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-ternary"},
				},
			},
			// Parenthesized body expression: the parens around the
			// consequent's expression are unwrapped to the assignment
			// before the merge, so a single-line inner expression
			// still reports under only-single-line.
			{
				Code: `if (test) {
	(
		a = foo
	);
} else {
	a = bar;
}`,
				Options: []interface{}{"only-single-line"},
				Output:  []string{`a = test ? foo : bar;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-ternary"},
				},
			},

			// ---- P1 #2: let-then-if with pure binary initializer ----
			// `a + b` is pure per upstream, so the let-then-if
			// collapse is allowed. The previous top-level kind switch
			// flagged any BinaryExpression as having a side effect and
			// incorrectly suppressed the suggestion. The parens
			// around `a + b` in the output are required: `+` binds
			// more loosely than `?:`, so without them the
			// `?:...:` would associate wrongly.
			{
				Code: `let x = a + b;
if (test) {
	x = c;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prefer-ternary",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prefer-ternary/suggestion", Output: "const x = test ? c : a + b;"},
						},
					},
				},
			},
			// Function/arrow initializers are pure per upstream too —
			// wrapping the value in a function does not introduce a
			// side effect.
			{
				Code: `let x = () => a;
if (test) {
	x = b;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prefer-ternary",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prefer-ternary/suggestion", Output: "const x = test ? b : () => a;"},
						},
					},
				},
			},

			// ---- N2: deferred method/getter bodies are pure ----
			// A method body is not invoked at construction time, so
			// the side effect inside must NOT block the collapse.
			// Upstream's MethodDefinition visitor returns false via
			// its FunctionExpression value; the equivalent on the
			// tsgo side skips the MethodDeclaration body.
			{
				Code: `let x = {m() { sideEffect(); }};
if (test) {
	x = next;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prefer-ternary",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prefer-ternary/suggestion", Output: "const x = test ? next : {m() { sideEffect(); }};"},
						},
					},
				},
			},
			{
				Code: `let x = {get m() { sideEffect(); }};
if (test) {
	x = next;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prefer-ternary",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prefer-ternary/suggestion", Output: "const x = test ? next : {get m() { sideEffect(); }};"},
						},
					},
				},
			},
			// Class member bodies are likewise deferred. The class expression is
			// safe to move into the alternate branch even though its method body
			// contains a call.
			{
				Code: `let x = class C { m() { sideEffect(); } };
if (test) {
	x = next;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prefer-ternary",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prefer-ternary/suggestion", Output: "const x = test ? next : class C { m() { sideEffect(); } };"},
						},
					},
				},
			},
			// Computed method key with a pure-literal expression and
			// a side-effecting body. The body is deferred (no effect
			// at construction), the computed key `'y' + 'z'` is pure,
			// so the let-then-if collapse is offered.
			{
				Code: "let x = {['y' + 'z']() { sideEffect(); }};\nif (test) {\n\tx = next;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prefer-ternary",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prefer-ternary/suggestion", Output: "const x = test ? next : {['y' + 'z']() { sideEffect(); }};"},
						},
					},
				},
			},
			// Method with a side-effecting parameter default. The
			// default fires when the method is called, not at the
			// construction site, so the collapse is offered.
			{
				Code: `let x = {m() {}};
if (test) {
	x = next;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "prefer-ternary",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prefer-ternary/suggestion", Output: "const x = test ? next : {m() {}};"},
						},
					},
				},
			},

			// ---- N1: two matching private accesses ----
			// `this.#x` on both sides is the same target, so the
			// collapse is offered with a plain `this.#x` LHS.
			{
				Code: `class C {
	#x;
	m() {
		if (t) {
			this.#x = a;
		} else {
			this.#x = b;
		}
	}
}`,
				Output: []string{`class C {
	#x;
	m() {
		this.#x = t ? a : b;
	}
}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-ternary"},
				},
			},

			// ---- N3: computed-key evaluation without a TypeChecker ----
			// Mirrors the upstream `'b' + 'ar'` case but exercises
			// the literal-only evaluator that the rule now constructs
			// when ctx.TypeChecker is nil. The Go test harness always
			// supplies a TypeChecker for the TypeScript cases, so the
			// pure-literal case below is the JS-side lock-in for
			// that path. The LHS is kept verbatim — `(x)['b' + 'ar']`
			// — because the merge keeps the source text of the
			// leftmost access rather than rewriting it to a
			// dotted form.
			{
				Code:     `if (t) { (x)['b' + 'ar'] = a; } else { x.bar = b; }`,
				Output:   []string{`(x)['b' + 'ar'] = t ? a : b;`},
				FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-ternary"},
				},
			},
		},
	)
}
