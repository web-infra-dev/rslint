package no_labels

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoLabelsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLabelsRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// ================================================================
			// Non-label contexts — should never trigger
			// ================================================================
			{Code: `var f = { label: foo() }`},
			{Code: `while (true) {}`},
			{Code: `while (true) { break; }`},
			{Code: `while (true) { continue; }`},
			{Code: `for (;;) { break; continue; }`},
			{Code: `do { break; } while (true)`},
			{Code: `switch (a) { case 0: break; }`},

			// ================================================================
			// allowLoop: all iteration statement types
			// ================================================================
			{
				Code:    `A: while (a) { break A; }`,
				Options: map[string]interface{}{"allowLoop": true},
			},
			{
				Code:    `A: do { if (b) { break A; } } while (a);`,
				Options: map[string]interface{}{"allowLoop": true},
			},
			{
				Code:    `A: for (;;) { break A; }`,
				Options: map[string]interface{}{"allowLoop": true},
			},
			{
				Code:    `A: for (var x in obj) { break A; }`,
				Options: map[string]interface{}{"allowLoop": true},
			},
			{
				Code:    `A: for (var x of arr) { break A; }`,
				Options: map[string]interface{}{"allowLoop": true},
			},
			// allowLoop: continue targeting labeled loop
			{
				Code:    `A: while (a) { continue A; }`,
				Options: map[string]interface{}{"allowLoop": true},
			},
			// allowLoop: labeled loop with nested switch using continue to outer loop
			{
				Code:    `A: for (var a in obj) { for (;;) { switch (a) { case 0: continue A; } } }`,
				Options: map[string]interface{}{"allowLoop": true},
			},

			// ================================================================
			// allowSwitch
			// ================================================================
			{
				Code:    `A: switch (a) { case 0: break A; }`,
				Options: map[string]interface{}{"allowSwitch": true},
			},

			// ================================================================
			// Both options true
			// ================================================================
			{
				Code:    `A: while (a) { break A; }`,
				Options: map[string]interface{}{"allowLoop": true, "allowSwitch": true},
			},
			{
				Code:    `A: switch (a) { case 0: break A; }`,
				Options: map[string]interface{}{"allowLoop": true, "allowSwitch": true},
			},
			// Both options: nested loop + switch, break/continue targeting outer loop
			{
				Code:    `A: while (a) { switch (x) { case 0: break A; continue A; } }`,
				Options: map[string]interface{}{"allowLoop": true, "allowSwitch": true},
			},
			// Both options: nested loops with labels — all break/continue target loops
			{
				Code:    `A: for (;;) { B: while (a) { break A; continue A; break B; continue B; } }`,
				Options: map[string]interface{}{"allowLoop": true, "allowSwitch": true},
			},
			// allowLoop: multiple break/continue targeting different labels — all loops, all allowed
			{
				Code:    `A: while (true) { B: while (true) { break A; break B; continue A; continue B; } }`,
				Options: map[string]interface{}{"allowLoop": true},
			},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// ================================================================
			// Dimension 1: All body types — default options (only unexpectedLabel)
			// ================================================================
			// while
			{
				Code: `label: while(true) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// do-while
			{
				Code: `A: do {} while (true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// for
			{
				Code: `A: for (;;) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// for-in
			{
				Code: `A: for (var x in obj) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// for-of
			{
				Code: `A: for (var x of arr) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// switch
			{
				Code: `A: switch (a) { case 0: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// block
			{
				Code: `A: {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// if
			{
				Code: `A: if (true) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// variable declaration
			{
				Code: `A: var foo = 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// expression statement
			{
				Code: `A: foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 2: Label + break/continue — error ordering
			// ================================================================
			// break fires before the labeled statement exit
			{
				Code: `label: while (true) { break label; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 23},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// continue fires before the labeled statement exit
			{
				Code: `label: while (true) { continue label; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInContinue", Line: 1, Column: 23},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// both break and continue on same label
			{
				Code: `A: while (true) { break A; continue A; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 19},
					{MessageId: "unexpectedLabelInContinue", Line: 1, Column: 28},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// label targeting itself (break in label body)
			{
				Code: `A: break A;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 4},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 3: Nested labels — scope chain correctness
			// ================================================================
			// Nested: break targets outer label through inner label
			{
				Code: `A: { if (foo()) { break A; } bar(); };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 19},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			{
				Code: `A: if (a) { if (foo()) { break A; } bar(); };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 26},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Nested: labeled switch with break targeting label
			{
				Code: `A: switch (a) { case 0: break A; default: break; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 25},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Double-nested labels in switch
			{
				Code: `A: switch (a) { case 0: B: { break A; } default: break; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 30},
					{MessageId: "unexpectedLabel", Line: 1, Column: 25},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Two nested labels both loops — inner exits before outer
			{
				Code: `A: while (true) { B: for (;;) { break A; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 33},
					{MessageId: "unexpectedLabel", Line: 1, Column: 19},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// continue targeting outer loop from inner label block
			{
				Code: `A: while (true) { B: { continue A; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInContinue", Line: 1, Column: 24},
					{MessageId: "unexpectedLabel", Line: 1, Column: 19},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// 3 levels deep: break targets outermost
			{
				Code: `A: { B: { C: while (true) { break A; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 29},
					{MessageId: "unexpectedLabel", Line: 1, Column: 11},
					{MessageId: "unexpectedLabel", Line: 1, Column: 6},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 4: Chained labels — getBodyKind on LabeledStatement body
			// A: B: while(true) {} → A's body is LabeledStatement, kind = "other"
			// ================================================================
			// Chained labels: only inner directly labels the loop
			{
				Code: `A: B: while (true) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 4},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Chained labels with break targeting each
			{
				Code: `A: B: while (true) { break A; break B; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 22},
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 31},
					{MessageId: "unexpectedLabel", Line: 1, Column: 4},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Chained labels with allowLoop: B (loop) is allowed, A (other) is not
			{
				Code:    `A: B: while (true) { break B; }`,
				Options: map[string]interface{}{"allowLoop": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Chained labels with allowLoop: break A still errors (A is "other")
			{
				Code:    `A: B: while (true) { break A; }`,
				Options: map[string]interface{}{"allowLoop": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 22},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Triple chained: A: B: C: while(true) {}
			{
				Code: `A: B: C: while (true) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 7},
					{MessageId: "unexpectedLabel", Line: 1, Column: 4},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Chained labels on switch with allowSwitch: same principle
			{
				Code:    `A: B: switch (a) { case 0: break B; }`,
				Options: map[string]interface{}{"allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 5: Same-name label shadowing
			// ================================================================
			// Inner label shadows outer — break targets inner
			{
				Code: `A: { A: while (true) { break A; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 24},
					{MessageId: "unexpectedLabel", Line: 1, Column: 6},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Shadowing with allowLoop: inner A (loop) allowed, break A allowed, outer A (other) errors
			{
				Code:    `A: { A: while (true) { break A; } }`,
				Options: map[string]interface{}{"allowLoop": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 6: Option combinations
			// ================================================================
			// allowLoop: variable declaration label — still "other"
			{
				Code:    `A: var foo = 0;`,
				Options: map[string]interface{}{"allowLoop": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// allowLoop: break label on block — still "other"
			{
				Code:    `A: break A;`,
				Options: map[string]interface{}{"allowLoop": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 4},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// allowLoop: block label with break
			{
				Code:    `A: { if (foo()) { break A; } bar(); };`,
				Options: map[string]interface{}{"allowLoop": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 19},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// allowLoop: if label with break
			{
				Code:    `A: if (a) { if (foo()) { break A; } bar(); };`,
				Options: map[string]interface{}{"allowLoop": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 26},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// allowLoop: switch label — switch is not "loop", still errors
			{
				Code:    `A: switch (a) { case 0: break A; default: break; };`,
				Options: map[string]interface{}{"allowLoop": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 25},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// allowSwitch: variable declaration
			{
				Code:    `A: var foo = 0;`,
				Options: map[string]interface{}{"allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// allowSwitch: break label on other
			{
				Code:    `A: break A;`,
				Options: map[string]interface{}{"allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 4},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// allowSwitch: block label with break
			{
				Code:    `A: { if (foo()) { break A; } bar(); };`,
				Options: map[string]interface{}{"allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 19},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// allowSwitch: if label
			{
				Code:    `A: if (a) { if (foo()) { break A; } bar(); };`,
				Options: map[string]interface{}{"allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 26},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// allowSwitch: while label — loop is not "switch", still errors
			{
				Code:    `A: while (a) { break A; }`,
				Options: map[string]interface{}{"allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 16},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// allowSwitch: do-while label
			{
				Code:    `A: do { if (b) { break A; } } while (a);`,
				Options: map[string]interface{}{"allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 18},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// allowSwitch: for-in label with break targeting labeled loop
			{
				Code:    `A: for (var a in obj) { for (;;) { switch (a) { case 0: break A; } } }`,
				Options: map[string]interface{}{"allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 57},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Both options true: block label — "other" is never allowed
			{
				Code:    `A: { break A; }`,
				Options: map[string]interface{}{"allowLoop": true, "allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 6},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Both options true: if label — still "other"
			{
				Code:    `A: if (true) { break A; }`,
				Options: map[string]interface{}{"allowLoop": true, "allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 16},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Both options true: variable declaration
			{
				Code:    `A: var foo = 0;`,
				Options: map[string]interface{}{"allowLoop": true, "allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 7: Cross-label break/continue with options
			// ================================================================
			// allowLoop: break targets outer loop from inner block label — outer is loop, allowed
			{
				Code:    `A: while (true) { B: { break A; } }`,
				Options: map[string]interface{}{"allowLoop": true},
				Errors: []rule_tester.InvalidTestCaseError{
					// B (block) is not allowed
					{MessageId: "unexpectedLabel", Line: 1, Column: 19},
				},
			},
			// allowLoop: continue targets outer loop from inner block label
			{
				Code:    `A: while (true) { B: { continue A; } }`,
				Options: map[string]interface{}{"allowLoop": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 19},
				},
			},
			// allowSwitch: break targets outer switch from inner block label
			{
				Code:    `A: switch (a) { case 0: B: { break A; } }`,
				Options: map[string]interface{}{"allowSwitch": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 25},
				},
			},

			// ================================================================
			// Dimension 8: Multi-line
			// ================================================================
			{
				Code: "A:\n  while (true) {\n    break A;\n  }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 3, Column: 5},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 9: Sequential labels — scope cleanup
			// After exiting A's scope, scopeInfo must be restored to nil/previous.
			// If scope leaks, the second label or its break would see stale data.
			// ================================================================
			// Two sequential labels at same level — each independent
			{
				Code: `A: while (true) {} B: for (;;) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
					{MessageId: "unexpectedLabel", Line: 1, Column: 20},
				},
			},
			// Sequential: first label with break, second clean
			{
				Code: `A: { break A; } B: while (true) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 6},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
					{MessageId: "unexpectedLabel", Line: 1, Column: 17},
				},
			},
			// Sequential with allowLoop: first (block) errors, second (loop) allowed
			{
				Code:    `A: { break A; } B: while (true) { break B; }`,
				Options: map[string]interface{}{"allowLoop": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 6},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 10: Multiple break/continue targeting different labels
			// Tests that getKind correctly distinguishes labels in the chain.
			// ================================================================
			{
				Code: `A: while (true) { B: while (true) { break A; break B; continue A; continue B; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 37},
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 46},
					{MessageId: "unexpectedLabelInContinue", Line: 1, Column: 55},
					{MessageId: "unexpectedLabelInContinue", Line: 1, Column: 67},
					{MessageId: "unexpectedLabel", Line: 1, Column: 19},
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 11: Labels inside function/class bodies
			// The rule tracks scope via enter/exit, which is per-node.
			// Labels inside nested functions still push/pop correctly.
			// ================================================================
			// Label inside function body
			{
				Code: `function f() { A: while (true) { break A; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 34},
					{MessageId: "unexpectedLabel", Line: 1, Column: 16},
				},
			},
			// Label inside arrow function
			{
				Code: `var f = () => { A: while (true) { break A; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 35},
					{MessageId: "unexpectedLabel", Line: 1, Column: 17},
				},
			},
			// Label inside class method
			{
				Code: `class C { method() { A: while (true) { break A; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabelInBreak", Line: 1, Column: 40},
					{MessageId: "unexpectedLabel", Line: 1, Column: 22},
				},
			},

			// ================================================================
			// Dimension 12: Rare body types
			// ================================================================
			// Label on function declaration — "other"
			{
				Code: `A: function f() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Label on class declaration — "other"
			{
				Code: `A: class C {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Label on empty statement — "other"
			{
				Code: `A: ;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
			// Label on try-catch — "other"
			{
				Code: `A: try {} catch (e) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLabel", Line: 1, Column: 1},
				},
			},
		},
	)
}
