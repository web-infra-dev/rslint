package no_unreachable

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnreachableRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnreachableRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// --- Function hoisting ---
			{Code: `function foo() { function bar() { return 1; } return bar(); }`},
			{Code: `function foo() { return bar(); function bar() { return 1; } }`},

			// --- var without init after terminal (hoisted) ---
			{Code: `function foo() { return x; var x; }`},
			{Code: `function foo() { return; var x; }`},
			{Code: `function foo() { return; var x, y, z; }`},
			{Code: `while (true) { break; var x; }`},
			{Code: `while (true) { continue; var x, y; }`},
			{Code: `while (true) { throw 'message'; var x; }`},

			// --- Empty statement after terminal ---
			{Code: `function foo() { return; ; }`},

			// --- Function declaration after terminal ---
			{Code: `function foo() { throw new Error(); function bar() {} }`},

			// --- Conditional terminals (not all paths terminate) ---
			{Code: `function foo() { if (x) { return; } bar(); }`},
			{Code: `function foo() { if (x) { } else { return; } x = 2; }`},
			{Code: `function foo() { if (x) { return; } else { bar(); } baz(); }`},
			{Code: `function foo() { switch (x) { case 0: break; default: return; } x = 2; }`},
			{Code: `function foo() { while (x) { return; } x = 2; }`},
			{Code: `function foo() { for (x in {}) { return; } x = 2; }`},
			{Code: `function foo() { for (;;) { if (x) break; } x = 2; }`},

			// --- Try/finally where finally doesn't terminate ---
			{Code: `function foo() { try { return; } finally { x = 2; } }`},

			// --- Switch patterns ---
			{Code: `switch (x) { case 1: break; }`},
			{Code: `function foo() { switch(x) { case 1: return; } bar(); }`},
			{Code: `while (true) { switch (foo) { case 1: x = 1; x = 2; } }`},
			{Code: `switch (foo) { case 1: break; var x; }`},

			// --- Labeled block break ---
			{Code: `function foo() { A: { break A; } bar(); }`},

			// --- Top-level throw with var (no init) after ---
			{Code: `var x = 1; throw 'uh oh'; var y;`},

			// --- Single statement loops ---
			{Code: `while (true) continue;`},

			// --- Try/catch where try can throw (catch is reachable) ---
			{Code: `function foo() { try { bar(); return; } catch (err) { return err; } }`},
			{Code: `function foo() { try { a.b.c = 1; return; } catch (err) { return err; } }`},

			// --- TypeScript: type-only declarations after return (erased) ---
			{Code: `function foo() { return; type A = string; }`},
			{Code: `function foo() { return; interface B {} }`},

			// --- Nested function scopes (inner function is independent) ---
			{Code: `function foo() { return; function bar() { var x = 1; } }`},

			// --- Arrow function body is independent scope ---
			{Code: `function foo() { var f = () => { return 1; }; bar(); }`},

			// --- Binder evaluates false keyword → code after if(false){return} is reachable ---
			{Code: `function foo() { if (false) { return; } bar(); }`},

			// --- Binder does NOT evaluate numeric/string literals ---
			{Code: `function foo() { if (1) { return; } bar(); }`},
			{Code: `function foo() { if (0) { return; } bar(); }`},
			{Code: `function foo() { if ("x") { return; } bar(); }`},
			{Code: `function foo() { if (null) { return; } bar(); }`},

			// --- for-of/for-in: loop body may not execute → code after is reachable ---
			{Code: `function foo() { for (const x of arr) { return; } bar(); }`},
			{Code: `function foo() { for (const x in obj) { return; } bar(); }`},

			// --- do-while(false): body executes once then exits → code after reachable ---
			{Code: `function foo() { do { break; } while (false); bar(); }`},

			// --- Async/generator functions: inner return is scoped ---
			{Code: `async function foo() { var f = async () => { return 1; }; bar(); }`},
			{Code: `function foo() { var f = function*() { return 1; }; bar(); }`},

			// --- Class method: unreachable code is inside method scope ---
			{Code: `class C { foo() { return bar(); function bar() { return 1; } } }`},

			// --- Deeply nested: try inside if inside switch → still valid ---
			{Code: `function foo() { switch(x) { default: if (y) { try { return; } finally {} } } bar(); }`},

			// --- Dead branch from constant condition: binder marks as unreachable,
			// but we don't report entire dead branches (not unreachable "after" a terminal) ---
			{Code: `function foo() { if (false) { return; } bar(); }`},
			{Code: `function foo() { if (true) { return; } else { bar(); } }`},

			// --- Generator try/yield: catch IS reachable (yield can throw) ---
			{Code: `function* foo() { try { yield 1; return; } catch (err) { return err; } }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// --- Basic terminals ---
			{
				Code:   `function foo() { return; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 26}},
			},
			{
				Code:   `function foo() { throw new Error(); x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 37}},
			},
			{
				Code:   `while (true) { break; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 23}},
			},
			{
				Code:   `while (true) { continue; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 26}},
			},

			// --- var with initializer after return IS reported ---
			{
				Code:   `function foo() { return; var x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 26}},
			},
			// var with partial initializer (some declarators have init)
			{
				Code:   `function foo() { return x; var x, y = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// var with init after continue
			{
				Code:   `while (true) { continue; var x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Consecutive grouping (multiple statements → one report) ---
			{
				Code:   `function foo() { return; x = 1; y = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 26}},
			},
			// Multi-line consecutive grouping
			{
				Code:   "function foo() {\n  return;\n  a();\n  b();\n  c();\n}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 3, Column: 3}},
			},
			// Consecutive with if block included
			{
				Code:   "function foo() {\n  return;\n  a();\n  if (b()) {\n    c()\n  } else {\n    d()\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 3, Column: 3}},
			},

			// --- Switch case terminals ---
			{
				Code:   `switch (x) { case 1: return; foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 30}},
			},
			{
				Code:   `switch (x) { default: return; foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 31}},
			},
			// Switch case with throw
			{
				Code:   `function foo() { switch (foo) { case 1: throw e; x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// Switch in while: break exits switch, not loop
			{
				Code:   `while (true) { switch (foo) { case 1: break; x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// Switch in while: continue exits loop iteration
			{
				Code:   `while (true) { switch (foo) { case 1: continue; x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- let/const/class after return ---
			{
				Code:   `function foo() { return; let x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 26}},
			},
			{
				Code:   `function foo() { return; const x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 26}},
			},
			{
				Code:   `function foo() { return; class Bar {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 26}},
			},

			// --- Top-level throw with var WITH init ---
			{
				Code:   `var x = 1; throw 'uh oh'; var y = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- if/else both terminate ---
			{
				Code:   `function foo() { if (x) { return; } else { throw e; } x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// if/else without braces
			{
				Code:   `function foo() { if (x) return; else throw -1; x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// Unreachable after if/else where both branches return
			{
				Code:   "function foo() { if (x) { return 1; } else { return 2; } bar(); }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 58}},
			},

			// --- Try/finally patterns ---
			{
				Code:   `function foo() { try { return; } finally {} x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			{
				Code:   `function foo() { try { } finally { return; } x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// try/catch where both terminate
			{
				Code:   "function foo() { try { return 1; } catch(e) { return 2; } bar(); }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// try/finally where finally returns
			{
				Code:   `function foo() { try { console.log('x'); } finally { return 1; } bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Loop patterns ---
			// do-while with return
			{
				Code:   `function foo() { do { return; } while (x); x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// do-while: unreachable inside body after return
			{
				Code:   `function foo() { do { return; x = 1; } while(true); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 1, Column: 31}},
			},
			// while with if/else both break/continue → unreachable inside loop
			{
				Code:   `function foo() { while (x) { if (x) break; else continue; x = 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// for(;;) with only continue (no break) → infinite loop
			{
				Code:   `function foo() { for (;;) { if (x) continue; } x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// while(true) empty body → infinite loop
			{
				Code:   `function foo() { while (true) { } x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// while(true) infinite loop
			{
				Code:   `function foo() { while(true) { doSomething(); } bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// for(;;) infinite loop
			{
				Code:   `function foo() { for(;;) { doSomething(); } bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Labeled break ---
			{
				Code:   `outer: while(true) { while(true) { break outer; } bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Switch with default, all cases return ---
			{
				Code:   `function foo() { switch(x) { case 1: return 1; default: return 2; } bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Nested if/else all returning ---
			{
				Code:   `function foo() { if (a) { if (b) { return; } else { return; } } else { return; } bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Multiple unreachable blocks in same function ---
			{
				Code: "function foo() {\n  if (a) {\n    return\n    b();\n    c();\n  } else {\n    throw err\n    d();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 4},
					{MessageId: "unreachableCode", Line: 8},
				},
			},
			// Both branches unreachable + code after if/else
			// Note: outer block reports e() first, then inner blocks report b() and d()
			{
				Code: "function foo() {\n  if (a) {\n    return\n    b();\n  } else {\n    throw err\n    d();\n  }\n  e();\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 9},
					{MessageId: "unreachableCode", Line: 4},
					{MessageId: "unreachableCode", Line: 7},
				},
			},

			// --- Arrow function with unreachable ---
			{
				Code:   `var f = (arrow) => { switch (arrow) { default: throw new Error(); }; g() }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- TypeScript: enum after return (has runtime effect) ---
			{
				Code:   `function foo() { return; enum Color { Red } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Multiline with precise position ---
			{
				Code:   "function foo() {\n  return;\n  x = 1;\n}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode", Line: 3, Column: 3}},
			},

			// --- Binder constant condition: if(true) break inside while ---
			// Binder evaluates if(true) → break always executes → var x = 1 unreachable
			{
				Code:   `while (true) { if (true) break; var x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Class method/getter unreachable ---
			{
				Code:   `class C { foo() { return; bar(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			{
				Code:   `class C { get x() { return 1; bar(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Async/generator unreachable ---
			{
				Code:   `async function foo() { return; bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			{
				Code:   `function* foo() { return; bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			{
				Code:   `async function* foo() { return; bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Deeply nested: try inside if inside switch, all terminate ---
			{
				Code:   `function foo() { switch(x) { default: if (y) { return; } else { try { throw e; } catch(e) { return; } } } bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Namespace/module block unreachable ---
			{
				Code:   "namespace Foo { throw new Error(); var x = 1; }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Multiple switch cases all returning with default ---
			{
				Code:   `function foo() { switch(x) { case 1: return 1; case 2: return 2; case 3: return 3; default: return 0; } bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},

			// --- Catch unreachable: try block can't throw ---
			{
				Code: "function foo() {\n  try {\n    return;\n  } catch (err) {\n    return err;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 4},
				},
			},
			// Generator: try { return; } catch → catch unreachable
			{
				Code: "function* foo() {\n  try {\n    return;\n  } catch (err) {\n    return err;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 4},
				},
			},
			// try { return; let a = 1; } catch → both let and catch unreachable
			// TryStatement listener reports catch (line 5) first, then Block reports let (line 4)
			{
				Code: "function foo() {\n  try {\n    return;\n    let a = 1;\n  } catch (err) {\n    return err;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 5},
					{MessageId: "unreachableCode", Line: 4},
				},
			},

			// --- Catch reachability edge cases ---
			// try { ; return; } catch → empty stmt before return, still can't throw
			{
				Code:   `function foo() { try { ; return; } catch (err) { return err; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// try { break; } catch inside while → catch unreachable (break can't throw)
			{
				Code:   `function foo() { while (true) { try { break; } catch (err) { return err; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// try { { return; } } catch → nested block, still can't throw
			{
				Code:   `function foo() { try { { return; } } catch (err) { return err; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
			// try with finally that terminates, catch unreachable
			{
				Code:   `function foo() { try { foo(); } catch (err) {} finally { return; } bar(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unreachableCode"}},
			},
		},
	)
}
