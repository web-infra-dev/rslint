// cspell:ignore fallsthrough
package no_fallthrough

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoFallthroughRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoFallthroughRule,
		[]rule_tester.ValidTestCase{
			// Break exits the case
			{Code: `switch(foo) { case 0: a(); break; case 1: b(); }`},
			// Empty case (no statements), allowed
			{Code: `switch(foo) { case 0: case 1: a(); break; }`},
			// Comment suppresses warning (various patterns matching /falls?\s?through/i)
			{Code: `switch(foo) { case 0: a(); /* falls through */ case 1: b(); }`},
			{Code: `switch(foo) { case 0: a(); /* fall through */ case 1: b(); }`},
			{Code: `switch(foo) { case 0: a(); /* fallsthrough */ case 1: b(); }`},
			// Return exits the case
			{Code: `function foo() { switch(bar) { case 0: a(); return; case 1: b(); } }`},
			// Throw exits the case
			{Code: `switch(foo) { case 0: a(); throw e; case 1: b(); }`},
			// Continue exits the case
			{Code: `while(a) { switch(foo) { case 0: a(); continue; case 1: b(); } }`},
			// "fall through" also works (case-insensitive)
			{Code: `switch(foo) { case 0: a(); // Fall Through` + "\n" + `case 1: b(); }`},
			// Last case doesn't need break
			{Code: `switch(foo) { case 0: a(); }`},
			// Multiple empty cases before a case with break
			{Code: `switch(foo) { case 0: case 1: case 2: a(); break; }`},
			// If/else both branches terminate
			{Code: `switch(foo) { case 0: if (a) { break; } else { break; } case 1: b(); }`},
			// Block with break at the end
			{Code: `switch(foo) { case 0: { a(); break; } case 1: b(); }`},
			// Default with break
			{Code: `switch(foo) { case 0: a(); break; default: b(); break; case 1: c(); }`},
			// Try/catch where both branches terminate
			{Code: `switch(foo) { case 0: try { break; } catch(e) { break; } case 1: b(); }`},
			// Nested switch: break in inner switch only exits inner, but outer has its own break
			{Code: `switch(foo) { case 0: switch(bar) { case 1: break; } break; case 2: b(); }`},
			// FALLS THROUGH (all caps)
			{Code: `switch(foo) { case 0: a(); /* FALLS THROUGH */ case 1: b(); }`},
			// Try/finally where finally breaks — terminal
			{Code: `switch(foo) { case 0: try { a(); } finally { break; } case 1: b(); }`},
			// Try/finally where finally returns — terminal
			{Code: `function f1() { switch(foo) { case 0: try { a(); } finally { return; } case 1: b(); } }`},
			// Try/catch/finally where finally breaks — terminal regardless of catch
			{Code: `switch(foo) { case 0: try { a(); } catch(e) { b(); } finally { break; } case 1: c(); }`},
			// Labeled break — terminal
			{Code: `switch(foo) { case 0: label1: break; case 1: b(); }`},
			// If/else if/else all terminate
			{Code: `switch(foo) { case 0: if(a) { break; } else if(b) { break; } else { break; } case 1: c(); }`},
			// Deeply nested if/else
			{Code: `switch(foo) { case 0: if(a) { if(b) { break; } else { break; } } else { break; } case 1: c(); }`},
			// Switch with only default — last case doesn't need break
			{Code: `switch(foo) { default: a(); }`},
			// All cases have breaks
			{Code: `switch(foo) { case 0: a(); break; case 1: b(); break; default: c(); break; }`},
			// Multi-line comment between cases containing "falls through"
			{Code: "switch(foo) { case 0: a();\n/* This falls through intentionally */\ncase 1: b(); }"},
			// Custom commentPattern option
			{
				Code:    "switch(foo) { case 0: a();\n/* break omitted */\ncase 1: b(); }",
				Options: []any{map[string]any{"commentPattern": `break[\s\w]*omitted`}},
			},
			// allowEmptyCase with empty statement
			{
				Code:    `switch(foo) { case 0: ; case 1: a(); break; }`,
				Options: []any{map[string]any{"allowEmptyCase": true}},
			},
			// Infinite loops — terminal
			{Code: `switch(foo) { case 0: while(true) {} case 1: b(); }`},
			{Code: `switch(foo) { case 0: while("x") {} case 1: b(); }`},
			// Infinite loop: for(;;) {} — terminal
			{Code: `switch(foo) { case 0: for(;;) {} case 1: b(); }`},
			// Infinite loop: do {} while(true) — terminal
			{Code: `switch(foo) { case 0: do {} while(true); case 1: b(); }`},
			// Nested try/finally — inner finally break swallows exceptions, outer catch unreachable
			{Code: `switch(foo) { case 0: try { try { a(); } finally { break; } } catch(e) { b(); } case 1: c(); }`},
			// Try with only break in try block — catch is unreachable
			{Code: `switch(foo) { case 0: try { break; } catch(e) { a(); } case 1: b(); }`},
			// Try with only continue in try block — catch is unreachable
			{Code: `while(a) { switch(foo) { case 0: try { continue; } catch(e) { a(); } case 1: b(); } }`},
			// Try with bare return — catch is unreachable
			{Code: `function f2() { switch(foo) { case 0: try { return; } catch(e) { a(); } case 1: b(); } }`},
			// while(true) with break in nested switch — still infinite (break captured)
			{Code: `switch(foo) { case 0: while(true) { switch(bar) { case 1: break; } } case 1: b(); }`},
			// while(true) with continue — still infinite
			{Code: `switch(foo) { case 0: while(true) { continue; } case 1: b(); }`},
			// while(true) with break in nested for — still infinite (break captured by for)
			{Code: `switch(foo) { case 0: while(true) { for(var j=0;j<1;j++) { break; } } case 1: b(); }`},
			// Try/catch where catch has throw — both terminate
			{Code: `switch(foo) { case 0: try { break; } catch(e) { throw e; } case 1: b(); }`},
			// Triple nested blocks with break
			{Code: `switch(foo) { case 0: { { { break; } } } case 1: b(); }`},
			// Labeled break targeting label wrapping the switch — terminal
			{Code: `outer3: switch(foo) { case 0: a(); break outer3; case 1: b(); }`},
			// Inner switch with all branches terminal (return/throw + default) — terminal
			{Code: "function f3() { switch(a) { case 0: switch(b) { case 1: return; default: throw e; }\ndefault: throw e; } }"},
			// return followed by unreachable code — still terminal
			{Code: `function f6() { switch(a) { case 0: return; a(); case 1: x(); } }`},
			// break followed by unreachable code — still terminal
			{Code: `switch(a) { case 0: break; a(); case 1: x(); }`},
			// unreachable code inside block after return
			{Code: `function f7() { switch(a) { case 0: if(x) { return; a(); } else { return; } case 1: x(); } }`},
			// while(true) with labeled break to inner block — loop is still infinite
			{Code: `switch(a) { case 0: while(true) { inner: { break inner; } } case 1: b(); }`},
			// Inner switch with breaks + outer break — terminal due to outer break
			{Code: `switch(a) { case 0: switch(b) { case 1: break; default: break; } break; case 1: x(); }`},
		},
		[]rule_tester.InvalidTestCase{
			// Fallthrough from case to case
			{
				Code: "switch(foo) { case 0: a();\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Fallthrough from case to default
			{
				Code: "switch(foo) { case 0: a();\ndefault: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "default", Line: 2, Column: 1},
				},
			},
			// Multiline fallthrough
			{
				Code: "switch(foo) { case 0:\n  a();\ncase 1:\n  b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 3, Column: 1},
				},
			},
			// Nested switch: break in inner switch does NOT prevent outer fallthrough
			{
				Code: "switch(foo) { case 0: switch(bar) { case 1: break; }\ncase 2: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// If without else: not terminal
			{
				Code: "switch(foo) { case 0: if (a) { break; }\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// If/else where only one branch terminates
			{
				Code: "switch(foo) { case 0: if (a) { break; } else { c(); }\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Empty block is not terminal
			{
				Code: "switch(foo) { case 0: { }\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Default in middle falls through
			{
				Code: "switch(foo) { case 0: break; default: a();\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Comment not matching pattern
			{
				Code: "switch(foo) { case 0: a();\n/* intentional */\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 3, Column: 1},
				},
			},
			// If/else if without else — not terminal
			{
				Code: "switch(foo) { case 0: if(a) { break; } else if(b) { break; }\ncase 1: c(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Labeled statement wrapping non-terminal
			{
				Code: "switch(foo) { case 0: label1: a();\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// For-in loop (not infinite) — not terminal
			{
				Code: "switch(foo) { case 0: for(var x in obj) { break; }\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// For-of loop (not infinite) — not terminal
			{
				Code: "switch(foo) { case 0: for(var x of arr) { break; }\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// while loop with break — NOT infinite, falls through
			{
				Code: "switch(foo) { case 0: while(a) { break; }\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Try with expression before break — catch is reachable
			{
				Code: "switch(foo) { case 0: try { foo(); break; } catch(e) { a(); }\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Try with return expr — expression might throw
			{
				Code: "function f3() { switch(foo) { case 0: try { return foo(); } catch(e) { a(); }\ncase 1: b(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// try/finally where finally does NOT terminate — falls through
			{
				Code: "switch(foo) { case 0: try { a(); } finally { b(); }\ncase 1: c(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// while(true) with conditional break — NOT infinite
			{
				Code: "switch(foo) { case 0: while(true) { if(x) break; }\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// labeled break targeting label INSIDE switch — NOT terminal for switch
			{
				Code: "switch(foo) { case 0: inner1: { break inner1; }\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Nested if where inner doesn't terminate
			{
				Code: "switch(foo) { case 0: if(a) { if(b) { break; } } else { break; }\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// variable declaration — not terminal
			{
				Code: "switch(foo) { case 0: var y = 1;\ncase 1: x(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Multiple consecutive cases falling through
			{
				Code: "switch(foo) { case 0: a();\ncase 1: b();\ncase 2: c(); break; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
					{MessageId: "case", Line: 3, Column: 1},
				},
			},
			// nested labels — break outer exits both labels AND the while loop
			{
				Code: "switch(a) { case 0: outer: inner: while(true) { break outer; }\ncase 1: x(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// labeled break exits the while loop → not infinite, falls through
			{
				Code: "function t1() { switch(a) { case 0: label: while(true) { break label; }\ncase 1: x(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// while(0x00) — zero in hex, falsy, falls through
			{
				Code: "switch(foo) { case 0: while(0x00) {}\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Inner switch with trailing empty default — switch exits normally, falls through
			{
				Code: "function f5() { switch(a) { case 0: switch(b) { case 1: return; default: }\ncase 1: x(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Inner switch with all breaks, no outer break — break exits inner, falls through outer
			{
				Code: "switch(a) { case 0: switch(b) { case 1: break; default: break; }\ncase 1: x(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Inner switch without default — NOT terminal (can skip all cases)
			{
				Code: "function f4() { switch(a) { case 0: switch(b) { case 1: return; case 2: return; }\ncase 1: x(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Inner switch with default but one case doesn't terminate — 2 errors:
			// outer case 0→case 1 fallthrough + inner case 1→default fallthrough
			{
				Code: "switch(a) { case 0: switch(b) { case 1: x(); default: break; }\ncase 1: x(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
					{MessageId: "default", Line: 1, Column: 46},
				},
			},
			// Default first, falls through to case
			{
				Code: "switch(foo) { default: a();\ncase 0: b(); break; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Custom commentPattern — default pattern no longer matches
			{
				Code:    "switch(foo) { case 0: a(); /* falls through */\ncase 1: b(); }",
				Options: []any{map[string]any{"commentPattern": `break[\s\w]*omitted`}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// reportUnusedFallthroughComment — terminal case with comment
			{
				Code:    "switch(foo) { case 0: a(); break; /* falls through */\ncase 1: b(); }",
				Options: []any{map[string]any{"reportUnusedFallthroughComment": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedFallthroughComment", Line: 2, Column: 1},
				},
			},
		},
	)
}
