package no_else_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoElseReturnExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row it covers, so future refactors can't
// silently regress them without breaking a named lock-in.
func TestNoElseReturnExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoElseReturnRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver / expression wrappers ----
			// N/A: the rule does not inspect receiver, member, call, or literal
			// expression children; only statement/control-flow shape matters.

			// ---- Dimension 4: access / key forms ----
			// N/A: the rule does not inspect object/class property keys.

			// ---- Dimension 4: declaration/container forms ----
			{Code: `const f = () => { if (a) { foo(); } else { return 1; } };`},
			{Code: `class C { m() { if (a) { foo(); } else { return 1; } } }`},
			{Code: `async function f() { if (a) { foo(); } else { return 1; } }`},
			{Code: `function *f() { if (a) { foo(); } else { return 1; } }`},

			// ---- Dimension 4: nesting / traversal boundaries ----
			// Locks in upstream checkIfWithoutElse parent guard: an if in a
			// single-statement position cannot be safely split into two statements.
			{Code: `function f() { while (foo) if (bar) return; else baz(); }`},
			// Locks in upstream checkIfWithoutElse early return: no alternate
			// means the rule must not continue walking the chain.
			{Code: `function f() { if (bar) return; }`},

			// ---- Dimension 4: graceful degradation ----
			{Code: `function f() { if (bar) {} else {} }`},
			{Code: `function f() { if (bar) { return; } }`},
			{Code: `declare function f(): void;`},
			{Code: `abstract class C { abstract m(): void }`},

			// ---- Real-user: eslint/eslint#3015 else-if false positive ----
			{Code: `function f() { if (x) { doOneThing(); } else if (y) { return true; } else { doAnotherThing(); } }`},
			// ---- Real-user: eslint/eslint#9228 default allowElseIf behavior ----
			{Code: `const res = (() => { if (error) { return "It failed"; } else if (loading) { return "Still loading"; } return result; })();`},
			// ---- Real-user: eslint/eslint#15496 break/continue are intentionally not return ----
			{Code: `function f(list) { for (const item of list) { if (foo) { if (a) { break; } else { return item; } } if (bar) { if (a) { continue; } else { return item; } } } }`},

			// Locks in upstream checkIfWithElse arm: no alternate means no report
			// even when allowElseIf is disabled.
			{Code: `function f() { if (bar) return; }`, Options: map[string]interface{}{"allowElseIf": false}},
			// Locks in upstream alwaysReturns false arm: the consequent contains
			// no return-like statement, so the else is still necessary.
			{Code: `function f() { if (bar) { foo(); } else { baz(); } }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: declaration/container forms ----
			{
				Code:   `const f = () => { if (a) { return 1; } else { return 2; } };`,
				Output: []string{`const f = () => { if (a) { return 1; }  return 2;  };`},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt(`const f = () => { if (a) { return 1; } else { return 2; } };`, `{ return 2; }`),
				},
			},
			{
				Code:   `class C { m() { if (a) return 1; else return 2; } }`,
				Output: []string{`class C { m() { if (a) return 1; return 2; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt(`class C { m() { if (a) return 1; else return 2; } }`, `return 2;`),
				},
			},
			{
				Code:   `async function f() { if (a) return 1; else return 2; }`,
				Output: []string{`async function f() { if (a) return 1; return 2; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt(`async function f() { if (a) return 1; else return 2; }`, `return 2;`),
				},
			},
			{
				Code:   `function *f() { if (a) { return 1; } else { return 2; } }`,
				Output: []string{`function *f() { if (a) { return 1; }  return 2;  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt(`function *f() { if (a) { return 1; } else { return 2; } }`, `{ return 2; }`),
				},
			},

			// Locks in upstream alwaysReturns BlockStatement arm: any return-like
			// statement in the consequent block is enough, not only the final one.
			{
				Code:   `function f() { if (a) { return 1; foo(); } else { foo(); } }`,
				Output: []string{`function f() { if (a) { return 1; foo(); }  foo();  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt(`function f() { if (a) { return 1; foo(); } else { foo(); } }`, `{ foo(); }`),
				},
			},
			// Locks in upstream checkIfWithElse arm: allowElseIf false reports
			// the else-if statement itself. This uses array-wrapped options to
			// exercise the JSON option path.
			{
				Code:    `function f() { if (a) return 1; else if (b) return 2; }`,
				Output:  []string{`function f() { if (a) return 1; if (b) return 2; }`},
				Options: []interface{}{map[string]interface{}{"allowElseIf": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt(`function f() { if (a) return 1; else if (b) return 2; }`, `if (b) return 2;`),
				},
			},
			// Locks in upstream isSafeFromNameCollisions branch: a conditional
			// function declaration alternate is reported but not fixed.
			{
				Code: `function f() { if (a) { return 1; } else function g() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt(`function f() { if (a) { return 1; } else function g() {} }`, `function g() {}`),
				},
			},
		},
	)
}
