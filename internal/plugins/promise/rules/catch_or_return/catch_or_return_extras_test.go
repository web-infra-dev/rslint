// TestCatchOrReturnExtras locks in branches and edge shapes that the upstream test suite
// doesn't exercise. Each case carries an inline comment pointing at the specific branch /
// Dimension 4 row / tsgo AST quirk it covers, so future refactors can't silently regress
// them without breaking a named lock-in.
package catch_or_return_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/catch_or_return"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCatchOrReturnExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&catch_or_return.CatchOrReturnRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized expression statement (tsgo preserves parens) ----
			// Parenthesized promise ending with .catch() is still valid.
			{Code: `(frank().then(go)).catch(doIt)`},
			// Double-parenthesized expression statement.
			{Code: `((frank().then(go))).catch(doIt)`},

			// ---- Dimension 4: parenthesized receiver in chain ----
			// Parens around the base call do not affect promise recognition.
			{Code: `(frank()).then(go).catch(doIt)`},

			// ---- Dimension 4: optional chain on receiver ----
			// frank()?.then() is syntactically promise-like; .catch() terminates it.
			{Code: `frank()?.then(go).catch(doIt)`},

			// ---- Dimension 4: computed element-access ['catch'] ----
			// IsPromiseLikeCall requires a PropertyAccessExpression callee, so
			// element-access calls like ["catch"]() are not recognised as promise
			// statements and are silently ignored — no error reported.
			{Code: `frank().then(go)["catch"](doIt)`},

			// ---- Dimension 4: nested function — inner returned promise ----
			// A returned promise inside a nested function is not a statement, so exempt.
			{Code: `function outer() { frank().then(go).catch(doIt); function inner() { return frank().then(go) } }`},

			// ---- Dimension 4: arrow function with expression body ----
			// Promise used as the return value of an arrow, not a statement.
			{Code: `const fn = () => frank().then(go)`},

			// ---- Dimension 4: Cypress — deeper chain ----
			// Cypress root is nested two levels deep.
			{Code: `cy.get("button").click().then(go)`},

			// ---- Dimension 4: allowThen + allowThenStrict both set — only null is valid ----
			// When both flags are set, allowThen && !allowThenStrict is false, so only null passes.
			{Code: `frank().then(go).then(null, doIt)`, Options: map[string]interface{}{"allowThen": true, "allowThenStrict": true}},

			// ---- Dimension 4: allowFinally with terminationMethod array ----
			// .finally() terminates via terminationMethod list (not via allowFinally recursion).
			{Code: `frank().then(go).finally()`, Options: map[string]interface{}{"terminationMethod": []interface{}{"catch", "finally"}}},

			// ---- Branch lock-in: terminationMethod explicit empty array falls back to default ----
			// An explicitly empty array must not disable the allowlist entirely; it falls
			// back to the documented default ["catch"], same as omitting the option.
			{Code: `frank().then(go).catch(doIt)`, Options: map[string]interface{}{"terminationMethod": []interface{}{}}},

			// ---- Dimension 4: terminationMethod array with custom name ----
			{Code: `frank().then(go).asCallback(fn)`, Options: map[string]interface{}{"terminationMethod": []interface{}{"catch", "asCallback"}}},

			// ---- Real-user: cy chains with multiple method calls ----
			// https://github.com/eslint-community/eslint-plugin-promise/issues — Cypress exemption
			{Code: `cy.get("ul").find("li").first().then(go)`},

			// ---- Real-user: non-promise call resembling promise ----
			// A plain function call that happens to be named "then" at the top level is not
			// a promise if it's not a member expression.
			{Code: `then(go)`},

			// ---- Dimension 4: element access with string key is NOT a promise (no dot) ----
			// frank()["then"](go) — callee is ElementAccessExpression, not PropertyAccessExpression.
			// IsPromiseLikeCall returns false → not reported (matches ESLint behavior).
			{Code: `frank()["then"](go)`},

			// ---- N/A: TS type-expression wrappers (X as any).then() ----
			// The outer property access name is still "then" regardless of the cast;
			// behavior is identical to uncast — not a distinct Dimension 4 case.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: full position assertion for expression statement ----
			// Line 1, Column 1: simple expression statement at the top level.
			{
				Code:   `frank().then(go)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1, Column: 1}},
			},
			// Line 1, Column 1: Promise static call.
			{
				Code:   `Promise.resolve(frank)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1, Column: 1}},
			},

			// ---- Dimension 4: parenthesized expression-statement (tsgo preserves parens) ----
			// The ExpressionStatement is reported even when the entire expression is parenthesized.
			{
				Code:   `(frank().then(go))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}},
			},

			// ---- Dimension 4: optional chain — no catch, still reported ----
			{
				Code:   `frank()?.then(go)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}},
			},

			// ---- Dimension 4: allowThen + allowThenStrict — non-null first arg is invalid ----
			{
				Code:    `frank().then(a, b)`,
				Options: map[string]interface{}{"allowThen": true, "allowThenStrict": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}},
			},

			// ---- Dimension 4: allowFinally — .then() after .finally() is not valid termination ----
			{
				Code:    `frank().then(go).catch(doIt).finally(fn).then(bar)`,
				Options: map[string]interface{}{"allowFinally": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}},
			},

			// ---- Branch lock-in: allowThen=true but args.length=1 → allowThen branch NOT entered ----
			// Locks in: allowThen=true, name="then", args.length=1 (not 2) → the 2-arg guard
			// prevents entering the allowThen / allowThenStrict block → falls through → reported.
			{
				Code:    `frank().then(go)`,
				Options: map[string]interface{}{"allowThen": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}},
			},

			// ---- Branch lock-in: isAllowedPromiseTermination — allowFinally branch falls through ----
			// Locks in upstream arm: allowFinally=true, name="finally", but receiver is NOT a
			// promise-like call → allowFinally if fails → falls through to terminationMethod
			// → "finally" not in default list → reports.
			{
				Code:    `frank().finally(fn)`,
				Options: map[string]interface{}{"allowFinally": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}},
			},

			// ---- Branch lock-in: terminationMethod single string ----
			// Locks in that a single string terminationMethod is normalized to a slice.
			{
				Code:    `frank().then(go)`,
				Options: map[string]interface{}{"terminationMethod": "done"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: doneMessage, Line: 1}},
			},

			// ---- Branch lock-in: terminationMethod array with 2 elements ----
			// Locks in the []interface{} parsing branch.
			{
				Code:    `frank().then(go)`,
				Options: map[string]interface{}{"terminationMethod": []interface{}{"done", "catch"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: "Expected done,catch() or return", Line: 1}},
			},

			// ---- Branch lock-in: ExpressionStatement expression is not promise — no report ----
			// The non-promise case is covered by the valid suite; lock in via invalid-side contrast:
			// a bare function call IS a promise if the callee is a member named 'then'.
			{
				Code:   `someObj.then(go)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}},
			},

			// ---- Real-user: fetch() without catch ----
			// Common real-user pattern: fetch promise not terminated.
			{
				Code:   `fetch("/api/data").then(response => response.json())`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}},
			},

			// ---- Real-user: Promise chain inside if block (not returned) ----
			{
				Code:   `if (cond) { frank().then(go) }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}},
			},
		},
	)
}
