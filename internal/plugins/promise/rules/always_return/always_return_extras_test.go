// TestAlwaysReturnExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.

package always_return_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/always_return"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestAlwaysReturnExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&always_return.AlwaysReturnRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized callback ----
			// Parenthesized FunctionExpression: SkipOuterExpressions unwraps it.
			{Code: `hey.then((function() { return 1; }))`},
			// Double-parenthesized arrow function with block body.
			{Code: `hey.then(((x) => { return x; }))`},

			// ---- Dimension 4: optional chain on .then() itself ----
			// hey?.then(cb) — optional method access; still a .then() call.
			{Code: `hey?.then(function() { return 1; })`},
			// Arrow function with block body through optional .then().
			{Code: `hey?.then(x => { return x; })`},

			// ---- Dimension 4: arrow function expression body (implicit return) ----
			// N/A: expression-body arrows are filtered out by isFunctionWithBlockBody —
			// they always return their expression value so no check is needed.

			// ---- Dimension 4: async function expression ----
			// Async callback that returns a value is valid.
			{Code: `hey.then(async function() { return 1; })`},
			// Async arrow with block body that returns.
			{Code: `hey.then(async x => { return x; })`},

			// ---- Dimension 4: nested function scope does not bleed ----
			// Inner function has no return, but outer returns the inner function.
			{Code: `hey.then(function() { var f = function() {}; return f; })`},
			// Inner arrow has no return; outer returns a value.
			{Code: `hey.then(function() { const g = () => {}; return g; })`},
			// Class with method inside then-callback; outer function returns.
			{Code: `hey.then(function() { class C { m() {} } return new C(); })`},

			// ---- Dimension 4: second argument to .then() is not checked ----
			// Error callback (second arg) is ignored; only first arg is checked.
			{Code: `hey.then(function() { return 1; }, function() {})`},
			// Only second arg present: no callback in first position -> no error.
			{Code: `hey.then(null, function() {})`},

			// ---- Dimension 4: .catch() / .finally() callback not checked ----
			// The rule only targets .then() first argument.
			{Code: `hey.catch(function() {})`},
			{Code: `hey.finally(function() {})`},

			// ---- Dimension 4: non-.then() call is ignored ----
			{Code: `hey.map(function() {})`},
			{Code: `fn(function() {})`},

			// ---- Dimension 4: TS type wrappers on callback ----
			// Type assertion on the whole callback: skipped by SkipOuterExpressions.
			// N/A: SkipOuterExpressions only unwraps OEKParentheses (not type assertions
			// as outer expressions). A callback wrapped in `as FunctionType` would not be
			// recognized as a function literal, so the rule silently passes — same as ESLint
			// which also doesn't match non-FunctionExpression/ArrowFunctionExpression types.

			// ---- Dimension 4: TS non-null assertion on return value ----
			{Code: `hey.then(x => { return x!; })`},

			// ---- Real-user: issue #318 — ignoreLastCallback with then().catch() ----
			// Reported false-positive: callback followed by .catch() should be treated as last.
			{Code: `
hey
  .then(x => { doSomething(x) })
  .catch(e => { handleError(e) })`, Options: map[string]interface{}{"ignoreLastCallback": true}},

			// ---- Real-user: issue — assignment to deeply nested member of ignored var ----
			// globalThis.foo.bar = x should be recognized (root = globalThis).
			{Code: `hey.then(x => { globalThis.foo.bar = x })`},

			// ---- Branch lock-in: isFunctionWithBlockBody — FunctionExpression always true ----
			// Locks in upstream isFunctionWithBlockStatement: FunctionExpression always has block body.
			{Code: `hey.then(function named() { return 1; })`},

			// ---- Branch lock-in: isFunctionWithBlockBody — ArrowFunction with expression body excluded ----
			// N/A: expression-body arrows never reach the block-body branch check.

			// ---- Branch lock-in: thenBlockCallback — non-.then() call short-circuits ----
			// Locks in IsMemberCall returning false for non-"then" methods.
			{Code: `hey.catch(x => { /* no return */ })`},

			// ---- Branch lock-in: thenBlockCallback — first arg is not a function expression ----
			// Passing a reference (not inline function) is ignored.
			{Code: `hey.then(someHandler)`},
			// Passing null as first arg: no callback.
			{Code: `hey.then(null)`},

			// ---- Branch lock-in: isLastCallback — VoidExpression ----
			// void hey.then(cb): VoidExpression parent -> isLastCallback = true.
			{Code: `void hey.then(x => { doSomething(x) })`, Options: map[string]interface{}{"ignoreLastCallback": true}},

			// ---- Branch lock-in: isLastCallback — comma sequence, target is NOT rightmost ----
			// (hey.then(cb), 42): target is Left of comma BinaryExpression -> isLastCallback = true.
			{Code: `(hey.then(x => { doSomething(x) }), 42)`, Options: map[string]interface{}{"ignoreLastCallback": true}},

			// ---- Branch lock-in: isLastCallback — await ----
			// await hey.then(cb) inside async function: AwaitExpression parent -> continue walking.
			{Code: `async function f() { await hey.then(x => { doSomething(x) }) }`, Options: map[string]interface{}{"ignoreLastCallback": true}},

			// ---- Branch lock-in: allPathsTerminate — if with else, both branches terminate ----
			{Code: `hey.then(function(x) { if (x) { return x; } else { throw new Error(); } })`},

			// ---- Branch lock-in: allPathsTerminate — block statement recursion ----
			// Nested block: { { return x; } } — stmtTerminates recurses into inner block.
			{Code: `hey.then(function(x) { { return x; } })`},

			// ---- Branch lock-in: getRootObjectName — ElementAccess chains ----
			// globalThis[key] = x: root via ElementAccessExpression -> "globalThis".
			{Code: `hey.then(x => { globalThis[x] = x })`},

			// ---- Branch lock-in: ignoreAssignmentVariable empty array ----
			// Empty ignoreAssignmentVariable: check is skipped entirely.
			// N/A: when the array is empty, the condition `len(opts.IgnoreAssignmentVariable) > 0` is false.

			// ---- Branch lock-in: isIgnoredAssignment — non-assignment operator ----
			// `globalThis += x` uses += (assignment operator): still matches.
			{Code: `hey.then(x => { globalThis += x })`},

			// ---- switch: all cases + default return ----
			{Code: `hey.then(function(x) { switch(x){case 1:return 1;default:return 2} })`},
			// empty fallthrough case — case 1 has no stmts so it is skipped in the check
			{Code: `hey.then(function(x) { switch(x){case 1:case 2:return 1;default:return 2} })`},
			// case with throw instead of return
			{Code: `hey.then(function(x) { switch(x){case 1:throw new Error();default:return 2} })`},

			// ---- try/catch: both branches terminate ----
			{Code: `hey.then(function() { try{return 1}catch(e){return 2} })`},
			// no catch clause — only try+finally; if try terminates the statement terminates
			{Code: `hey.then(function() { try{return 1}finally{} })`},
			// catch throws instead of returning
			{Code: `hey.then(function() { try{return 1}catch(e){throw e} })`},

			// ---- loops: body terminates ----
			{Code: `hey.then(function() { while(true){return 1} })`},
			{Code: `hey.then(function() { for(;;){return 1} })`},
			{Code: `hey.then(function() { do{return 1}while(x) })`},
			// loop body throws
			{Code: `hey.then(function() { while(true){throw new Error()} })`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized callback — still reports ----
			{
				Code:   `hey.then((function() {}))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- Dimension 4: async function without return ----
			{
				Code:   `hey.then(async function() { doSomething(); })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code:   `hey.then(async x => { doSomething(x); })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- Dimension 4: optional .then() without return ----
			{
				Code:   `hey?.then(function() { doSomething(); })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- Dimension 4: process.exit in only one if-branch — other branch falls through ----
			// Locks in upstream arm: process.exit only marks one branch good; else-path is bad.
			{
				Code:   `hey.then(function(x) { if (x) { process.exit(0); } })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			// process.abort in one branch, no else -> other path falls through.
			{
				Code:   `hey.then(function(x) { if (x) { process.abort(); } })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			// process?.exit() — optional chain on process itself; not a terminator.
			{
				Code:   `hey.then(function(x) { process?.exit(0); })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			// process.exit?.() — optional call on exit; not a terminator.
			{
				Code:   `hey.then(function(x) { process.exit?.(0); })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- Branch lock-in: isLastCallback — returned from function (not last) ----
			// return hey.then(cb): parent is ReturnStatement in an outer function -> isLastCallback = false.
			{
				Code:    `function f() { return hey.then(x => { doSomething(x) }) }`,
				Options: map[string]interface{}{"ignoreLastCallback": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- Branch lock-in: isLastCallback — assigned to variable (not last) ----
			{
				Code:    `const p = hey.then(x => { doSomething(x) })`,
				Options: map[string]interface{}{"ignoreLastCallback": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- Branch lock-in: isLastCallback — passed as argument (not .catch/.finally) ----
			{
				Code:    `fn(hey.then(x => { doSomething(x) }))`,
				Options: map[string]interface{}{"ignoreLastCallback": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- Branch lock-in: isLastCallback — comma sequence, target IS rightmost, but sequence is assigned ----
			// `const x = (42, hey.then(cb))`: comma is rightmost (isLastCallback continues up),
			// but then parent is VariableDeclarator -> isLastCallback = false.
			{
				Code:    `const x = (42, hey.then(cb => { doSomething(cb) }))`,
				Options: map[string]interface{}{"ignoreLastCallback": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- Branch lock-in: ignoreAssignmentVariable not applies when NOT isLastCallback ----
			// Assigned to variable -> isLastCallback = false -> ignoreAssignmentVariable not checked.
			{
				Code:   `const p = hey.then(x => { globalThis = x })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- Real-user: callback with only side-effect call (no return) ----
			{
				Code:   `fetchData().then(function(data) { console.log(data) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- Real-user: missing return in one if branch of a chain ----
			{
				Code: `
p.then(function(x) {
  if (x > 0) {
    return x * 2;
  }
})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- switch: no default clause — execution can fall through ----
			{
				Code:   `hey.then(function(x) { switch(x){case 1:return 1} })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			// one case has statements but does not terminate
			{
				Code:   `hey.then(function(x) { switch(x){case 1:doSomething();default:return 2} })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			// default clause itself does not terminate
			{
				Code:   `hey.then(function(x) { switch(x){case 1:return 1;default:doSomething()} })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- try/catch: one branch does not terminate ----
			// try terminates but catch is empty
			{
				Code:   `hey.then(function() { try{return 1}catch(e){} })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			// try is empty, catch returns
			{
				Code:   `hey.then(function() { try{}catch(e){return 1} })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},

			// ---- loops: body does not terminate ----
			{
				Code:   `hey.then(function() { while(true){} })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code:   `hey.then(function() { for(;;){} })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code:   `hey.then(function() { do{}while(x) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
		},
	)
}
