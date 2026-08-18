// TestNoContinueExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
//
// Dimension walk notes for no-continue:
//   - Dimension 3 (autofix boundaries): N/A — the rule has no autofix.
//   - Dimension 4 (receiver/expression wrappers, access/key forms,
//     declaration/container forms): N/A — ContinueStatement's only child is
//     an optional label Identifier that is never dereferenced through
//     parens, optional chaining, type-expression wrappers, or property/key
//     access, and the rule never inspects functions or classes.
//   - Dimension 4 (graceful degradation: SpreadAssignment/RestElement, empty
//     bodies, overload signatures): N/A — none of these shapes can contain
//     or affect a ContinueStatement.
package no_continue

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoContinueExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoContinueRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: loop kinds without continue ----
			{Code: `for (const k in obj) { doStuff(k); }`},
			{Code: `for (const v of arr) { doStuff(v); }`},
			{Code: `do { doStuff(); } while (x);`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 1: for-in statement ----
			{
				Code: `for (const k in obj) { continue; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24, EndLine: 1, EndColumn: 33},
				},
			},
			// ---- Dimension 1: for-of statement ----
			{
				Code: `for (const v of arr) { continue; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24, EndLine: 1, EndColumn: 33},
				},
			},
			// ---- Dimension 1: do-while statement ----
			{
				Code: `do { continue; } while (x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6, EndLine: 1, EndColumn: 15},
				},
			},
			// ---- Dimension 1: for-await-of statement (TS/async-only syntax form) ----
			{
				Code: `async function f() { for await (const x of gen()) { if (!x) continue; use(x); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 61, EndLine: 1, EndColumn: 70},
				},
			},
			// ---- Dimension 2: labeled continue targeting the outer loop of a 2-level nest ----
			{
				Code: `outer: for (const v of arr) { for (const w of v) { continue outer; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 52, EndLine: 1, EndColumn: 67},
				},
			},
			// ---- Dimension 2: labeled continue targeting the outermost loop of a 3-level nest ----
			{
				Code: `outer: for (let i = 0; i < 3; i++) { for (let j = 0; j < 3; j++) { for (let k = 0; k < 3; k++) { continue outer; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 98, EndLine: 1, EndColumn: 113},
				},
			},
			// ---- Dimension 2: continue inside a switch statement nested inside a loop ----
			{
				Code: `for (let i = 0; i < 3; i++) { switch (i) { case 0: continue; default: break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 52, EndLine: 1, EndColumn: 61},
				},
			},
			// ---- Dimension 2: nested-loop traversal boundary — a loop inside an arrow-function
			// callback nested inside an outer loop still reports the inner continue ----
			{
				Code: `for (const v of arr) { [1, 2].forEach(() => { for (const w of v) { continue; } }); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 68, EndLine: 1, EndColumn: 77},
				},
			},
			// ---- Real-user: skipping falsy values while iterating Object.entries ----
			{
				Code: `for (const [key, value] of Object.entries(obj)) { if (!value) continue; use(key, value); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 63, EndLine: 1, EndColumn: 72},
				},
			},
			// ---- Real-user: skipping invalid items inside a try/catch during async iteration ----
			{
				Code: `async function* run(items) {
  for await (const item of items) {
    try {
      if (!isValid(item)) {
        continue;
      }
      yield item;
    } catch (err) {
      continue;
    }
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 5, Column: 9, EndLine: 5, EndColumn: 18},
					{MessageId: "unexpected", Line: 9, Column: 7, EndLine: 9, EndColumn: 16},
				},
			},
			// ---- Locks in upstream ContinueStatement() sole branch: the listener reports
			// unconditionally, so every continue statement in the same loop is reported
			// independently (no dedupe / early-exit after the first match) ----
			{
				Code: `for (let i = 0; i < 3; i++) { if (i === 0) continue; if (i === 1) continue; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 44, EndLine: 1, EndColumn: 53},
					{MessageId: "unexpected", Line: 1, Column: 67, EndLine: 1, EndColumn: 76},
				},
			},
		},
	)
}
