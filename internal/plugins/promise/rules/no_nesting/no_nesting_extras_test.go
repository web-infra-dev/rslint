// TestNoNestingExtras locks in branches and edge shapes that the upstream test suite
// doesn't exercise. Each case carries an inline comment pointing at the specific
// branch / Dimension 4 row / tsgo AST quirk it covers, so future refactors can't
// silently regress them without breaking a named lock-in.

package no_nesting_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_nesting"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNestingExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_nesting.NoNestingRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized receiver on outer .then call ----
			// The receiver of .then is parenthesized; the function IS still detected
			// as a promise callback, and the inner .then is exempt because it uses `a`.
			{Code: `(doThing()).then(a => innerThing().then(b => getC(a, b)))`},

			// ---- Dimension 4: element access ['then'] is not detected ----
			// IsMemberCall requires a PropertyAccessExpression; computed access is ignored
			// so the function is NOT pushed to the callback stack.
			{Code: `doThing()['then'](function() { a.then() })`},

			// ---- Dimension 4: optional-chain on outer .then call ----
			// IsMemberCall checks the property name regardless of QuestionDotToken.
			// The function is pushed; its inner .then uses `a` from the outer scope → skip.
			{Code: `doThing()?.then(a => innerThing().then(b => getC(a, b)))`},

			// ---- Dimension 4: parenthesized callback function ----
			// isPromiseCallback skips outer expressions (parentheses), so (fn) still
			// detects the outer function as a promise callback. Inner .then uses `a` → skip.
			{Code: `doThing().then((a => innerThing().then(b => getC(a, b))))`},

			// ---- Dimension 4: async arrow function callback ----
			// Async modifier doesn't affect isThenOrCatchCall or isPromiseCallback.
			{Code: `doThing().then(async a => innerThing().then(b => getC(a, b)))`},

			// ---- Dimension 4: local const — closure from const declaration ----
			// A local `const` inside the callback body is collected by collectScopeBindings;
			// the inner .then references it → skip.
			{Code: `doThing().then(a => {
  const result = doInner(a);
  return result.then(b => getC(result, b));
})`},

			// ---- Dimension 4: local var inside if-block — collected across blocks ----
			{Code: `doThing().then(function() {
  var x = 1;
  if (cond) { var y = 2; }
  return inner().then(b => getC(x, b));
})`},

			// ---- Dimension 4: local var in for-loop initializer is collected ----
			// collectDeclsInStmt handles KindVariableDeclarationList nodes which appear
			// as for-loop initializers.
			{Code: `doThing().then(function() {
  for (var i = 0; i < 10; i++) {}
  return inner().then(b => done(i, b));
})`},

			// ---- Dimension 4: flat chain — second .then callback has no nested call ----
			// The second callback `() => b` contains only an identifier reference, not a
			// .then/.catch call, so there is nothing to report.
			{Code: `doThing().then(() => a).then(() => b)`},

			// ---- Dimension 4: non-arrow, non-function argument — no push ----
			// A variable reference passed to .then is invisible to isPromiseCallback.
			{Code: `doThing().then(handler)`},

			// ---- Dimension 4: .then with no arguments — no callback pushed ----
			{Code: `doThing().then()`},

			// ---- Real-user: promise returned directly without nesting ----
			{Code: `doThing().then(() => otherThing())`},

			// ---- Dimension 4: destructured params are collected ----
			// CollectBindingNames recurses into ObjectBindingPattern; `x` from the
			// destructured param is found in the inner .then's arg → skip.
			{Code: `doThing().then(({ x }) => inner(x).then(b => done(x, b)))`},

			// ---- Dimension 4: rest parameter is collected ----
			{Code: `doThing().then((...args) => inner(args).then(b => done(args, b)))`},

			// ---- Dimension 4: function declaration inside callback uses closure var ----
			// FunctionDeclaration is NOT pushed to the callback stack (isPromiseCallback
			// only fires for FunctionExpression/ArrowFunction). The outer callback has `x`
			// as a param; `a.then(x)` inside the FunctionDeclaration references `x` → skip.
			{Code: `doThing().then(function(x) {
  function inner() { return a.then(x) }
  return inner();
})`},

			// ---- Dimension 4: getter inside callback uses closure var ----
			// Same reasoning: class methods / accessors are not pushed to the stack.
			{Code: `doThing().then(function(x) {
  return { get val() { return a.then(x) } };
})`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized callback — still reports inner .then ----
			// The outer function is correctly detected (parens are skipped). Inner .then
			// has no args referencing the outer scope → report.
			{
				Code:   `doThing().then((function() { a.then() }))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Line: 1}},
			},

			// ---- Dimension 4: optional-chain outer call — inner .then reported ----
			{
				Code:   `doThing()?.then(function() { a.then() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Line: 1}},
			},

			// ---- Dimension 4: async function callback ----
			{
				Code:   `doThing().then(async function() { a.then() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Line: 1}},
			},

			// ---- Dimension 4: named function expression callback ----
			{
				Code:   `doThing().then(function named() { a.then() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Line: 1}},
			},

			// ---- Dimension 4: catch as outer callback ----
			{
				Code:   `doThing().catch(function() { a.then() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Line: 1}},
			},

			// ---- Dimension 4: function declaration inside callback — no closure var used ----
			// The FunctionDeclaration is not pushed to the stack, but the outer callback IS
			// on the stack. a.then() inside the declaration has no args referencing the outer
			// scope → report. This matches ESLint's behavior.
			{
				Code: `doThing().then(function() {
  function inner() { return a.then() }
  return inner();
})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Line: 2}},
			},

			// ---- Dimension 4: getter inside callback — no closure var used ----
			// The accessor method is not pushed to the stack; a.then() with no closure
			// refs → report.
			{
				Code: `doThing().then(function() {
  return { get x() { return a.then() } };
})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Line: 2}},
			},

			// ---- Dimension 4: chained .then callback itself contains a nested call ----
			// The second .then's callback `() => b.catch()` has b.catch() inside it.
			// That b.catch() is nested in the second callback scope with no closure refs → report.
			{
				Code:   `doThing().then(() => a).then(() => b.catch())`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Line: 1}},
			},

			// ---- Dimension 4: 3-level nesting reports both foldable levels ----
			// Both a.then() and b.then() are independently reportable:
			// a.then(b=>..) → outer callback has `a` but inner arg has no `a` usage → report
			// b.then(c=>c) → middle callback has `b` but inner arg has no `b` → report
			{
				Code: `doThing().then(a => a.then(b => b.then(c => c)))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNesting", Line: 1},
					{MessageId: "avoidNesting", Line: 1},
				},
			},

			// Locks in upstream arm: callbackScopes non-empty and inner args don't ref closure.
			{
				Code:   `doThing().then(() => a.catch())`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Line: 1}},
			},

			// ---- Real-user: rejection handler (second arg) contains nested .then ----
			{
				Code:   `doThing().then(null, function() { a.then() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Line: 1}},
			},

			// ---- Real-user: .catch rejection handler contains nested .catch ----
			{
				Code:   `doThing().catch(function() { a.catch() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Line: 1}},
			},
		},
	)
}
