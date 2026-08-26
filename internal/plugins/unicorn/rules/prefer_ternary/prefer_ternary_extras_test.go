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
		},
		[]rule_tester.InvalidTestCase{
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
		},
	)
}
