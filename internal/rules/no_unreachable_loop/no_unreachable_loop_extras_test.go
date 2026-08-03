package no_unreachable_loop

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnreachableLoopExtras covers what the upstream suite cannot: the tsgo
// node shapes ESTree has no analog for, the code path roots and try/finally
// layouts internal/cfg lays out differently, every arm of the upstream source
// its own tests leave untouched, and code shapes real projects write. Every
// verdict below was taken from ESLint itself, except the four marked as
// intentional differences.
func TestNoUnreachableLoopExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnreachableLoopRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized and type-wrapped iterables ----
			// tsgo keeps parentheses and TS type expressions as nodes of their
			// own; the loop shape underneath has to be recognised through them.
			{Code: `for (const x of (arr)) foo(x);`},
			{Code: `for (const x of arr as any[]) foo(x);`},

			// ---- Dimension 4: optional chain ----
			{Code: `for (const x of obj?.items ?? []) foo(x);`},

			// ---- Dimension 4: graceful degradation ----
			{Code: `for (const x of arr);`},
			{Code: `for (const x of arr) {}`},

			// ---- Dimension 4: nesting / traversal boundaries ----
			// A loop reached only through a `finally` block, which internal/cfg
			// lays out twice — once for normal completion and once for the path
			// that leaves the statement abruptly.
			{Code: `function f() { try { g(); } finally { for (;;) foo(); } }`},
			{Code: `for (;;) { try { foo(); } finally { continue; } }`},
			{Code: `outer: { for (;;) foo(); }`},

			// Locks in upstream onCodePathSegmentLoop()'s third state: the loop
			// event has no entry in loopsByTargetSegments. A `continue` in a
			// `do`…`while` restarts it at its test, which is not the node the
			// next iteration begins at — the test passing is what repeats it.
			{Code: `outer: do { continue outer; } while (a);`},
			{Code: `do { continue; } while (false);`},
			{Code: `do { continue; } while (a);`},
			{Code: `do { foo(); } while (false);`},

			// Locks in upstream onCodePathSegmentLoop() arm `node ===
			// ContinueStatement`, reached from inside a `switch`, and the
			// unlabelled `break` that leaves only the `switch`.
			{Code: `for (;;) { switch (a) { case 1: break; } }`},
			{Code: `for (;;) { switch (a) { case 1: continue; } }`},

			// A suspended generator or awaited promise resumes in the loop, so
			// the body still completes.
			{Code: `function* g() { for (;;) { yield 1; } }`},
			{Code: `async function f() { for (;;) { await x; } }`},

			// Locks in upstream's `isAnySegmentReachable` gate: a loop the code
			// path never reaches is not a candidate, however its body looks.
			{Code: `function f() { return; for (;;) break; }`},
			{Code: `function f() { throw e; while (a) break; }`},
			{Code: `while (true); for (;;) break;`},

			// ---- Real-user: reading only the first entry of a collection ----
			// The idiom the `ignore` option exists for.
			{
				Code:    `function firstKey(obj) { for (const k in obj) { return k; } return null; }`,
				Options: map[string]interface{}{"ignore": []interface{}{"ForInStatement"}},
			},

			// ---- Real-user: a guard that only sometimes leaves the loop ----
			{Code: `for (const x of items) { if (x.done) break; visit(x); }`},
			{Code: `function find(items, t) { for (const x of items) { if (x === t) return x; } }`},
			{Code: `async function retry(n) { for (let i = 0; i < n; i++) { try { return await run(); } catch (e) { continue; } } }`},

			// ---- Intentional difference from ESLint ----
			// A `continue` naming a label the loop carries alongside another one
			// still starts a new iteration, so the loop is not limited to one.
			// ESLint reports these: its code path analysis attaches only the
			// innermost label to the loop, so the jump misses the loop it names.
			{Code: `outer: inner: while (a) { continue outer; }`},
			{Code: `outer: inner: do { continue outer; } while (a);`},
			{Code: `a: b: c: for (;;) { continue a; }`},
			{Code: `outer: inner: for (const x of arr) { continue outer; }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized and type-wrapped iterables ----
			{
				Code: `for (const x of (arr)) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code: `for (const x of ((arr))) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code: `while ((a)) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: `for (const x of arr!) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: `for (const x of arr as any[]) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code: `for (const x of (arr satisfies string[])) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 49},
				},
			},

			// ---- Dimension 4: optional chain ----
			{
				Code: `for (const x of obj?.items ?? []) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code: `for (const x of obj?.get?.() ?? []) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
				},
			},

			// ---- Dimension 4: binding and assignment target forms ----
			// N/A: identifier / string-literal / numeric-literal / private /
			// computed key forms — a loop has no key or member name to match on.
			{
				Code: `for (const [a, b] of pairs) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code: `for (const { a } of items) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code: `for (const { a = 1 } of items) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code: `for (const [a, ...rest] of items) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code: `for (obj[key] of items) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code: `for (obj.prop in items) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},

			// ---- Dimension 4: declaration / container forms ----
			// Every kind of code path root internal/cfg lays out separately.
			{
				Code: `const f = () => { for (;;) break; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 19, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code: `const f = async () => { for await (const x of s) break; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 25, EndLine: 1, EndColumn: 56},
				},
			},
			{
				Code: `const o = { m() { for (;;) break; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 19, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code: `class C { m() { for (;;) break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 17, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code: `class C { get p() { for (;;) break; return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 21, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code: `class C { static { for (;;) break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 20, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code: `class C { p = (() => { for (;;) break; })(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 24, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code: `function* g() { for (;;) break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 17, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code: `async function* g() { for await (const x of s) break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 23, EndLine: 1, EndColumn: 54},
				},
			},
			{
				Code: `const C = class { m() { for (;;) break; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 25, EndLine: 1, EndColumn: 40},
				},
			},

			// ---- Dimension 4: nesting / traversal boundaries ----
			// A loop in a nested function belongs to that function's code path,
			// not the enclosing loop's, so neither verdict leaks into the other.
			{
				Code: `for (;;) { (function () { for (;;) break; })(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 27, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code: `for (;;) { (function () { for (;;) foo(); })(); break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 57},
				},
			},
			// Reports from separate code paths are ordered by position, not by
			// the order the code paths were laid out.
			{
				Code: `function a() { for (;;) break; } for (;;) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 16, EndLine: 1, EndColumn: 31},
					{MessageId: "invalid", Line: 1, Column: 34, EndLine: 1, EndColumn: 49},
				},
			},
			{
				Code: `for (;;) { foo(); } function a() { for (;;) break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 36, EndLine: 1, EndColumn: 51},
				},
			},

			// ---- Dimension 4: graceful degradation ----
			{
				Code: `for (const {} of arr) break;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: `lbl: for (;;) break lbl;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 6, EndLine: 1, EndColumn: 25},
				},
			},
			// N/A: autofix boundaries (Dimension 3) — this rule only reports.

			// ---- A loop that a `finally` block lays out twice ----
			// The second copy runs on the path that leaves the `try` statement
			// abruptly; the loop is one candidate, not two.
			{
				Code: `function f() { try { g(); } finally { for (;;) break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 39, EndLine: 1, EndColumn: 54},
				},
			},
			{
				Code: `function f() { try { return g(); } finally { for (;;) break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 46, EndLine: 1, EndColumn: 61},
				},
			},
			{
				Code: `function f() { for (;;) { try { break; } catch (e) { continue; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 16, EndLine: 1, EndColumn: 67},
				},
			},
			{
				Code: `for (;;) { try { foo(); } finally { break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 47},
				},
			},

			// ---- Labelled jump targets ----
			{
				Code: `outer: { for (;;) break outer; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 10, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code: `outer: for (;;) { inner: for (;;) { continue outer; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 26, EndLine: 1, EndColumn: 54},
				},
			},
			{
				Code: `outer: for (;;) { for (;;) { break outer; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 8, EndLine: 1, EndColumn: 46},
					{MessageId: "invalid", Line: 1, Column: 19, EndLine: 1, EndColumn: 44},
				},
			},

			// Locks in upstream isLoopingTarget()'s DoWhileStatement arm: the
			// loop repeats from its test, so a body that never reaches the test
			// is the only way a `do`…`while` runs once.
			{
				Code: `do { break; } while (false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},

			// A `switch` every clause of which exits leaves nothing to fall
			// through to the end of the body.
			{
				Code: `function f() { for (;;) { switch (a) { case 1: return; default: throw e; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 16, EndLine: 1, EndColumn: 77},
				},
			},
			{
				Code: `function* g() { for (;;) { yield 1; break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 17, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code: `async function f() { for (;;) { await x; break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 22, EndLine: 1, EndColumn: 50},
				},
			},

			// ---- Every arm of the `ignore` option ----
			{
				Code:    `while (a) break;`,
				Options: map[string]interface{}{"ignore": []interface{}{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    `while (a) break; for (;;) break;`,
				Options: map[string]interface{}{"ignore": []interface{}{"ForStatement"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    `while (a) break; do break; while (b);`,
				Options: map[string]interface{}{"ignore": []interface{}{"WhileStatement"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 18, EndLine: 1, EndColumn: 38},
				},
			},
			// An ignored loop still shapes the code path around the loops that
			// are checked.
			{
				Code:    `for (;;) { while (a) break; break; }`,
				Options: map[string]interface{}{"ignore": []interface{}{"WhileStatement"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},

			// A loop after a `break` is unreachable and so not a candidate; the
			// loop around it still is.
			{
				Code: `for (;;) { break; while (a) break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},

			// ---- Real-user: reading only the first entry of a collection ----
			{
				Code: `function firstKey(obj) { for (const k in obj) { return k; } return null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 26, EndLine: 1, EndColumn: 60},
				},
			},
			{
				Code: `function hasAny(obj) { for (const k in obj) { return true; } return false; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 24, EndLine: 1, EndColumn: 61},
				},
			},
			{
				Code: `function first(gen) { for (const v of gen()) { return v; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalid", Line: 1, Column: 23, EndLine: 1, EndColumn: 59},
				},
			},
		},
	)
}
