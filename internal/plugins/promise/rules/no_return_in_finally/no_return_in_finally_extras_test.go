// TestNoReturnInFinallyExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
package no_return_in_finally_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_return_in_finally"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoReturnInFinallyExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_return_in_finally.NoReturnInFinallyRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: nesting / traversal boundary ----
			// Return inside a nested function inside finally callback: only the
			// callback's own direct top-level statements are scanned, so the
			// inner function's return is not seen.
			{Code: `myPromise.finally(function() { var f = function() { return 2; }; f(); })`},
			// Return inside a method inside finally callback: the callback's
			// direct statement is a class declaration, not a return.
			{Code: `myPromise.finally(function() { class C { m() { return 2; } } })`},
			// Return inside getter inside finally callback.
			{Code: `myPromise.finally(function() { class C { get x() { return 2; } } })`},
			// Return inside setter inside finally callback.
			{Code: `myPromise.finally(function() { class C { set x(v) { return v; } } })`},
			// Return inside constructor inside finally callback.
			{Code: `myPromise.finally(function() { class C { constructor() { return; } } })`},
			// Arrow function with expression body (no explicit return): valid.
			// N/A for return-statement analysis since expression body has no ReturnStatement.
			{Code: `myPromise.finally(() => 2)`},

			// Ensure the callback is only checked when it is the first argument of finally()
			{Code: `Promise.finally(arg1, () => { return 1; })`},

			// Ensure only top-level returns in the function's block are checked (ESLint parity).
			{Code: `Promise.finally(() => { if (true) { return 1; } })`},
			{Code: `Promise.finally(() => { try { return 1; } catch(e) {} })`},

			// ---- Dimension 4: access / key forms ----
			// Computed access: promise['finally'](fn) is NOT matched by IsMemberCall.
			// N/A: IsMemberCall only checks PropertyAccessExpression (dotted access).
			{Code: `myPromise['finally'](() => { return 2 })`},

			// ---- Dimension 4: receiver / expression wrappers ----
			// No `.finally()` call at all: return at top level is not inside
			// any function → no report. N/A: ReturnStatement at module level is a
			// parse error; covered implicitly since no listener fires.

			// ---- Locks in IsMemberCall(node, "finally") gate: no .finally() call ----
			// A return not inside any .finally() callback produces no error.
			// (The rule listener fires on CallExpression and exits early unless
			// IsMemberCall(node, "finally") matches.)
			{Code: `var x = 1`},

			// ---- Locks in IsMemberCall(node, "finally") gate: non-finally callee ----
			// Return inside a .then() callback is not a finally callback.
			{Code: `myPromise.then(function() { return 2; })`},
			// Return inside a .catch() callback is not a finally callback.
			{Code: `myPromise.catch(function() { return 2; })`},
			// Return inside a plain function call is not a finally callback.
			{Code: `doThing(function() { return 2; })`},
			// Return inside a function declaration is not inside any finally callback.
			{Code: `function foo() { return 2; }`},

			// ---- Real-user: cleanup without returning ----
			// Common pattern: side-effect only in finally, no return.
			{Code: `fetch('/api').finally(function() { cleanup(); })`},
			// Chained promise with finally doing side-effects only.
			{Code: `Promise.resolve(1).then(fn).finally(() => { console.log('done'); })`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: receiver / expression wrappers (invalid) ----
			// Parenthesized callback must still be flagged (SkipOuterExpressions).
			{
				Code:   `myPromise.finally((function() { return 2; }))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 11}},
			},
			// Double parens around callback.
			{
				Code:   `myPromise.finally(((function() { return 2; })))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 11}},
			},
			// Parenthesized arrow function.
			{
				Code:   `myPromise.finally((() => { return 2; }))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 11}},
			},

			// ---- Dimension 4: optional chain on the callee ----
			// Optional-chain finally: promise?.finally(fn). tsgo represents ?. as
			// a QuestionDotToken flag on the CallExpression; IsMemberCall checks
			// PropertyAccessExpression name which still matches.
			{
				Code:   `myPromise?.finally(() => { return 2 })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 12}},
			},

			// ---- Dimension 4: declaration / container forms ----
			// Async function expression callback.
			{
				Code:   `myPromise.finally(async function() { return 2; })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 11}},
			},
			// Named function expression callback.
			{
				Code:   `myPromise.finally(function cleanup() { return 2; })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 11}},
			},

			// ---- Dimension 4: nesting / traversal boundary ----
			// Outer return is a direct top-level statement and is flagged;
			// the inner function's return is not a top-level statement of the
			// callback, so it's ignored. Reports once on `.finally`, not on
			// either return.
			{
				Code:   `myPromise.finally(function() { var f = function() { return 2; }; return 3; })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 11}},
			},
			// FunctionDeclaration's return isn't a top-level statement of the
			// callback either; the outer return still flags the call once.
			{
				Code:   `myPromise.finally(() => { function a() { return 1; } return a(); })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 11}},
			},
			// Empty returns at the top level should still be caught.
			{
				Code:   `Promise.finally(() => { return; })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 9}},
			},
			// Multiple top-level returns still produce a single report (upstream
			// uses .some(), not .forEach()).
			{
				Code:   `myPromise.finally(() => { return 1; return 2; })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 1, Column: 11}},
			},

			// ---- Real-user: return in multiline finally callback ----
			{
				Code: `
fetch('/api')
  .then(process)
  .finally(function() {
    return cleanup();
  });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReturnInFinally", Message: noReturnMsg, Line: 4, Column: 4}},
			},
		},
	)
}
