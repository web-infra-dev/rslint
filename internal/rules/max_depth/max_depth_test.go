package max_depth

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMaxDepth covers two corpora:
//
//  1. ESLint parity — every case from eslint/tests/lib/rules/max-depth.js
//     migrated 1:1, with line / column added to every invalid case (upstream
//     omits them on most cases — we lock them in here so subsequent refactors
//     can't silently shift the report position).
//  2. Additional edge cases — tsgo AST shape variants (class/object methods,
//     getters/setters, constructor, arrow class fields) that ESLint's tests
//     don't enumerate because ESTree collapses them into FunctionExpression /
//     ArrowFunctionExpression. Plus statement-kind coverage (do / for-in /
//     with) and an `else if` chain with a trailing sibling that exercises
//     the asymmetric push/pop on IfStatement (see max_depth.go).
func TestMaxDepth(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxDepthRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// 1. ESLint parity — mirrors tests/lib/rules/max-depth.js
			// ============================================================
			{
				Code:    "function foo() { if (true) { if (false) { if (true) { } } } }",
				Options: 3,
			},
			{
				Code:    "function foo() { if (true) { } else if (false) { } else if (true) { } else if (false) {} }",
				Options: 3,
			},
			{
				Code:    "var foo = () => { if (true) { if (false) { if (true) { } } } }",
				Options: 3,
			},
			{
				Code: "function foo() { if (true) { if (false) { if (true) { } } } }",
			},

			// object property options
			{
				Code:    "function foo() { if (true) { if (false) { if (true) { } } } }",
				Options: map[string]interface{}{"max": 3},
			},

			{
				Code:    "class C { static { if (1) { if (2) {} } } }",
				Options: 2,
			},
			{
				Code:    "class C { static { if (1) { if (2) {} } if (1) { if (2) {} } } }",
				Options: 2,
			},
			{
				Code:    "class C { static { if (1) { if (2) {} } } static { if (1) { if (2) {} } } }",
				Options: 2,
			},
			{
				Code:    "if (1) { class C { static { if (1) { if (2) {} } } } }",
				Options: 2,
			},
			{
				Code:    "function foo() { if (1) { class C { static { if (1) { if (2) {} } } } } }",
				Options: 2,
			},
			{
				Code:    "function foo() { if (1) { if (2) { class C { static { if (1) { if (2) {} } if (1) { if (2) {} } } } } } if (1) { if (2) {} } }",
				Options: 2,
			},

			// ============================================================
			// 2. Additional edge cases
			// ============================================================

			// --- Default option (max=4) ---
			{
				Code: "if (a) { if (b) { if (c) { if (d) {} } } }",
			},

			// --- Options shapes — ensure JSON path is exercised ---
			// Number directly (not wrapped in array).
			{Code: "function foo() { if (a) {} }", Options: 1},
			// Bare object (single-option CLI shape).
			{
				Code:    "function foo() { if (a) { if (b) {} } }",
				Options: map[string]interface{}{"max": 2},
			},
			// Array-wrapped object (multi-element / rule_tester shape).
			{
				Code:    "function foo() { if (a) { if (b) {} } }",
				Options: []interface{}{map[string]interface{}{"max": 2}},
			},
			// Empty options array → defaults to 4.
			{Code: "if (a) { if (b) { if (c) { if (d) {} } } }", Options: []interface{}{}},
			// Legacy `maximum` key.
			{
				Code:    "function foo() { if (a) { if (b) { if (c) {} } } }",
				Options: map[string]interface{}{"maximum": 3},
			},

			// --- Inner function resets the depth counter ---
			{
				Code:    "function outer() { if (1) { function inner() { if (2) { if (3) {} } } } }",
				Options: 2,
			},
			{
				Code:    "function outer() { if (1) { var inner = function () { if (2) { if (3) {} } }; } }",
				Options: 2,
			},
			{
				Code:    "function outer() { if (1) { var inner = () => { if (2) { if (3) {} } }; } }",
				Options: 2,
			},

			// --- Class methods / object methods / getters / setters / ctor reset depth ---
			{
				Code:    "class C { method() { if (1) { if (2) { if (3) {} } } } }",
				Options: 3,
			},
			{
				Code:    "class C { get x() { if (1) { if (2) { if (3) {} } } } }",
				Options: 3,
			},
			{
				Code:    "class C { set x(v) { if (1) { if (2) { if (3) {} } } } }",
				Options: 3,
			},
			{
				Code:    "class C { constructor() { if (1) { if (2) { if (3) {} } } } }",
				Options: 3,
			},
			{
				Code:    "var obj = { method() { if (1) { if (2) { if (3) {} } } } }",
				Options: 3,
			},
			{
				Code:    "var obj = { get x() { if (1) { if (2) { if (3) {} } } } }",
				Options: 3,
			},
			// Class field initializer arrow — its own ArrowFunction scope.
			{
				Code:    "class C { handler = () => { if (1) { if (2) { if (3) {} } } } }",
				Options: 3,
			},

			// --- Statement kinds — every depth-increasing form, just under the limit ---
			{
				Code:    "function f() { for (;;) { while (true) { do { if (1) {} } while (1); } } }",
				Options: 4,
			},
			{
				Code:    "function f() { for (var k in o) { for (var v of a) { switch (k) {} } } }",
				Options: 3,
			},
			{
				Code:    "function f() { try { } catch (e) { } finally { } }",
				Options: 1,
			},

			// --- `with` statement — depth-increasing ---
			{
				Code:    "function f() { with (o) { if (true) {} } }",
				Options: 2,
			},

			// --- `else if` chain at the limit, no sibling ---
			{
				Code:    "function f() { if (a) {} else if (b) {} else if (c) {} else if (d) {} }",
				Options: 1,
			},

			// --- TypeScript syntax — type annotations don't add depth ---
			{
				Code:    "function f<T extends { x: number }>(a: T): void { if (a.x) { if (true) {} } }",
				Options: 2,
			},
			// Type guards / `as` / `satisfies` are expressions, not statements,
			// and don't contribute to depth.
			{
				Code:    "function f(a: unknown) { if (a as string) { if (true) {} } }",
				Options: 2,
			},

			// --- Lock-in: ESLint's asymmetric `IfStatement:exit` pop ---
			// ESLint pushes only on the outer `if` of an else-if chain but pops
			// on every chain link, leaving the depth counter negative after the
			// chain ends. Subsequent sibling code therefore has a "discount"
			// equal to the chain length. The if(f) below would be at depth 3
			// under a naive read, but ESLint sees it at depth 1 (-2 residual +
			// 3 nested ifs) and reports nothing. See max_depth.go for context.
			{
				Code:    "function f() { if (a) {} else if (b) {} else if (c) {} if (d) { if (e) { if (f) {} } } }",
				Options: 2,
			},
			// Same shape, longer chain — confirms the residual scales linearly.
			{
				Code:    "if (a) {} else if (b) {} else if (c) {} else if (d) {} if (e) { if (f) { if (g) { if (h) {} } } }",
				Options: 3,
			},

			// --- Function/scope boundary edge cases ---
			// IIFE — depth resets inside the immediately invoked function;
			// inner if(c)/if(d)/if(e) reach depth 3, which is at the limit.
			{
				Code:    "if (a) { if (b) { (function () { if (c) { if (d) { if (e) {} } } })(); } }",
				Options: 3,
			},
			// Arrow IIFE — same.
			{
				Code:    "if (a) { if (b) { (() => { if (c) { if (d) { if (e) {} } } })(); } }",
				Options: 3,
			},
			// Async / generator / async generator function-likes — same boundary.
			{
				Code:    "async function f() { if (a) { if (b) { if (c) {} } } }",
				Options: 3,
			},
			{
				Code:    "function* g() { if (a) { if (b) { if (c) {} } } }",
				Options: 3,
			},
			{
				Code:    "async function* g() { if (a) { if (b) { if (c) {} } } }",
				Options: 3,
			},
			// Async arrow.
			{
				Code:    "var f = async () => { if (a) { if (b) { if (c) {} } } }",
				Options: 3,
			},
			// Default-export anonymous function-likes.
			{
				Code:    "export default function () { if (a) { if (b) { if (c) {} } } }",
				Options: 3,
			},
			{
				Code:    "export default class { method() { if (a) { if (b) { if (c) {} } } } }",
				Options: 3,
			},
			// Decorator on a class method — body still resets depth.
			{
				Code:    "function dec() { return (a, b, c) => {}; } class C { @dec() method() { if (a) { if (b) { if (c) {} } } } }",
				Options: 3,
			},

			// --- Real-world nesting shapes (just under default limit of 4) ---
			// fetch + retry pattern.
			{
				Code: `async function fetchWithRetry(url) {
  for (let i = 0; i < 3; i++) {
    try {
      const r = await fetch(url);
      if (r.ok) {
        return r.json();
      }
    } catch (e) {
      if (i === 2) throw e;
    }
  }
}`,
			},
			// Validation / state-machine pattern.
			{
				Code: `function validate(input) {
  switch (input.kind) {
    case 'number':
      if (input.value > 0) {
        for (const c of input.constraints) {
          if (!c.test(input.value)) return false;
        }
      }
      break;
  }
  return true;
}`,
			},

			// --- Type-only constructs do NOT contain executable nesting ---
			// Interface body / type literal — no statements inside, no depth.
			{
				Code:    "interface I { x: { y: { z: { w: number } } } } function f() { if (a) {} }",
				Options: 1,
			},
			// Type alias with nested object types.
			{
				Code:    "type T = { a: { b: { c: { d: number } } } }; function f() { if (a) {} }",
				Options: 1,
			},
			// Conditional types are types, not statements.
			{
				Code:    "type T<X> = X extends Array<infer Y> ? (Y extends string ? Y : never) : never; function f() { if (a) {} }",
				Options: 1,
			},

			// --- TS-only declarations / declaration merging ---
			// Namespace body acts like a module — but rule doesn't reset on it.
			// Still, no top-level depth-increasing statements here.
			{
				Code:    "namespace N { export function f() { if (a) { if (b) {} } } }",
				Options: 2,
			},
			// Module declaration with nested function.
			{
				Code:    "declare module 'x' { export function f(): void; }",
				Options: 1,
			},
			// Abstract / overload signatures (no body) — must not crash.
			{
				Code:    "abstract class A { abstract method(): void; }",
				Options: 1,
			},
			{
				Code:    "function f(x: number): number; function f(x: string): string; function f(x: any): any { if (x) {} }",
				Options: 1,
			},

			// --- Empty / degenerate forms ---
			{Code: ""},
			{Code: "function f() {}", Options: 0},
			{Code: "var f = () => {};", Options: 0},
			{Code: "class C {}", Options: 0},
			{Code: "class C { static {} }", Options: 0},
			{Code: "function f() { try { } catch (e) { } finally { } }", Options: 1},

			// --- Lone blocks do NOT increase depth (only listed kinds do) ---
			{
				Code:    "function f() { { { { if (a) {} } } } }",
				Options: 1,
			},

			// --- Bare `else { if(){} }` (block alternate, not chained) ---
			// The inner `if` parent is BlockStatement, so it pushes — distinct
			// from `else if` chaining (parent: IfStatement, no push).
			{
				Code:    "function f() { if (a) {} else { if (b) {} } }",
				Options: 2,
			},

			// --- `for await (...of...)` — same KindForOfStatement, counted ---
			{
				Code:    "async function f() { for await (const x of y) { if (a) { if (b) {} } } }",
				Options: 3,
			},

			// --- LabeledStatement does NOT add depth — its child still counts ---
			{
				Code:    "function f() { outer: while (true) { if (a) {} } }",
				Options: 2,
			},

			// --- Statement bodies without braces ---
			{
				Code:    "function f() { while (true) if (a) for (;;) {} }",
				Options: 3,
			},

			// --- switch with default only / empty switch ---
			{
				Code:    "function f() { switch (x) {} }",
				Options: 1,
			},
			{
				Code:    "function f() { switch (x) { default: if (a) {} } }",
				Options: 2,
			},

			// --- Class expression (vs declaration) ---
			{
				Code:    "var C = class { method() { if (a) { if (b) { if (c) {} } } } }",
				Options: 3,
			},

			// --- Method parameter default with arrow — its own scope ---
			{
				Code:    "class C { method(cb = () => { if (a) { if (b) { if (c) {} } } }) {} }",
				Options: 3,
			},

			// --- Namespace body — does NOT reset depth (tsgo + ESLint align) ---
			{
				Code:    "namespace N { if (a) { if (b) {} } }",
				Options: 2,
			},

			// --- TS Enum body has no depth-increasing statements ---
			{
				Code:    "enum E { A, B, C } function f() { if (a) {} }",
				Options: 1,
			},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// 1. ESLint parity — mirrors tests/lib/rules/max-depth.js
			// ============================================================
			{
				Code:    "function foo() { if (true) { if (false) { if (true) { } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    43,
					},
				},
			},
			{
				Code:    "var foo = () => { if (true) { if (false) { if (true) { } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    44,
					},
				},
			},
			{
				Code:    "function foo() { if (true) {} else { for(;;) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    38,
					},
				},
			},
			{
				Code:    "function foo() { while (true) { if (true) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    33,
					},
				},
			},
			{
				Code:    "function foo() { for (let x of foo) { if (true) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    39,
					},
				},
			},
			{
				Code:    "function foo() { while (true) { if (true) { if (false) { } } } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    33,
					},
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 1.",
						Line:      1,
						Column:    45,
					},
				},
			},
			{
				Code: "function foo() { if (true) { if (false) { if (true) { if (false) { if (true) { } } } } } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (5). Maximum allowed is 4.",
						Line:      1,
						Column:    68,
					},
				},
			},

			// object property options
			{
				Code:    "function foo() { if (true) { if (false) { if (true) { } } } }",
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    43,
					},
				},
			},

			{
				Code:    "function foo() { if (a) { if (b) { if (c) { if (d) { if (e) {} } } } } }",
				Options: map[string]interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (5). Maximum allowed is 4.",
						Line:      1,
						Column:    54,
					},
				},
			},
			{
				Code:    "function foo() { if (true) {} }",
				Options: map[string]interface{}{"max": 0},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (1). Maximum allowed is 0.",
						Line:      1,
						Column:    18,
					},
				},
			},

			{
				Code:    "class C { static { if (1) { if (2) { if (3) {} } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    38,
					},
				},
			},
			{
				Code:    "if (1) { class C { static { if (1) { if (2) { if (3) {} } } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    47,
					},
				},
			},
			{
				Code:    "function foo() { if (1) { class C { static { if (1) { if (2) { if (3) {} } } } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    64,
					},
				},
			},
			{
				Code:    "function foo() { if (1) { class C { static { if (1) { if (2) {} } } } if (2) { if (3) {} } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    80,
					},
				},
			},

			// ============================================================
			// 2. Additional edge cases
			// ============================================================

			// --- Top-level (no enclosing function) ---
			{
				Code:    "if (a) { if (b) { if (c) {} } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    19,
					},
				},
			},

			// --- Multi-line code: line/column precision ---
			{
				Code: `function foo() {
    if (a) {
        if (b) {
            if (c) {
            }
        }
    }
}`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      4,
						Column:    13,
					},
				},
			},

			// --- Statement kinds: do / switch / try / with / for-in ---
			{
				Code:    "function f() { do { if (a) {} } while (b); }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    21,
					},
				},
			},
			{
				Code:    "function f() { switch (x) { case 1: if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    37,
					},
				},
			},
			{
				Code:    "function f() { try { if (a) {} } catch (e) {} }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    22,
					},
				},
			},
			{
				Code:    "function f() { with (o) { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    27,
					},
				},
			},
			{
				Code:    "function f() { for (var k in o) { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    35,
					},
				},
			},

			// --- Class method / getter / setter / constructor reset depth, but
			//     deep nesting inside still triggers ---
			{
				Code:    "class C { method() { if (1) { if (2) { if (3) {} } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    40,
					},
				},
			},
			{
				Code:    "class C { get x() { if (1) { if (2) { if (3) {} } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    39,
					},
				},
			},
			{
				Code:    "class C { constructor() { if (1) { if (2) { if (3) {} } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    45,
					},
				},
			},
			{
				Code:    "var obj = { method() { if (1) { if (2) { if (3) {} } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    42,
					},
				},
			},
			// Class field with arrow — ArrowFunction scope inside class.
			{
				Code:    "class C { handler = () => { if (1) { if (2) { if (3) {} } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    47,
					},
				},
			},

			// --- Legacy `maximum` key honoured ---
			{
				Code:    "function foo() { if (a) { if (b) { if (c) {} } } }",
				Options: map[string]interface{}{"maximum": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    36,
					},
				},
			},
			// `maximum` wins when both keys are present and `maximum` is truthy
			// (matching ESLint's `option.maximum || option.max`).
			{
				Code:    "function foo() { if (a) { if (b) { if (c) {} } } }",
				Options: map[string]interface{}{"maximum": 2, "max": 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    36,
					},
				},
			},
			// `{ maximum: 0 }` falls through to undefined in ESLint and disables
			// the check; `{ max: 0 }` instead reports every block. This case
			// uses `max: 0` to lock in the falsy-but-valid path.
			{
				Code:    "var x = () => { if (a) {} }",
				Options: []interface{}{map[string]interface{}{"max": 0}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (1). Maximum allowed is 0.",
						Line:      1,
						Column:    17,
					},
				},
			},

			// --- IIFE — body has its own scope, but interior nesting still triggers ---
			{
				Code:    "(function () { if (a) { if (b) { if (c) {} } } })()",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    34,
					},
				},
			},
			{
				Code:    "(() => { if (a) { if (b) { if (c) {} } } })()",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    28,
					},
				},
			},

			// --- async / generator / async generator function bodies ---
			{
				Code:    "async function f() { if (a) { if (b) { if (c) {} } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    40,
					},
				},
			},
			{
				Code:    "function* g() { if (a) { if (b) { if (c) {} } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    35,
					},
				},
			},

			// --- Async method / async arrow class field ---
			{
				Code:    "class C { async method() { if (a) { if (b) { if (c) {} } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    46,
					},
				},
			},
			{
				Code:    "class C { handler = async () => { if (a) { if (b) { if (c) {} } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    53,
					},
				},
			},

			// --- Catch / finally body still counts inside the surrounding try ---
			{
				Code:    "function f() { try { } catch (e) { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    36,
					},
				},
			},
			{
				Code:    "function f() { try { } finally { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    34,
					},
				},
			},

			// --- switch with multiple case clauses, deep inside a case ---
			{
				Code:    "function f() { switch (x) { case 1: { for (;;) { if (a) {} } break; } case 2: break; } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    50,
					},
				},
			},

			// --- Top-level (no enclosing function) deep nesting ---
			{
				Code:    "for (;;) { while (true) { if (a) { if (b) {} } } }",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (4). Maximum allowed is 3.",
						Line:      1,
						Column:    36,
					},
				},
			},

			// --- Multiple separate violations in one source ---
			{
				Code: `function a() { if (1) { if (2) { if (3) {} } } }
function b() { while (1) { while (2) { while (3) {} } } }`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    34,
					},
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      2,
						Column:    40,
					},
				},
			},

			// --- Sibling violations under shared parent ---
			{
				Code:    "function f() { if (1) { if (2) { if (3) {} } if (4) { if (5) {} } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    34,
					},
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    55,
					},
				},
			},

			// --- Violation inside a nested function — depth resets at the
			//     function boundary, then deepens again inside the inner body ---
			{
				Code:    "function outer() { if (1) { function inner() { if (2) { if (3) { if (4) {} } } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    66,
					},
				},
			},

			// --- `else { if (b) {} }` — block alternate, NOT chained;
			//     inner if pushes (parent is BlockStatement, not IfStatement) ---
			{
				Code:    "function f() { if (a) {} else { if (b) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    33,
					},
				},
			},

			// --- LabeledStatement — wrapper transparent, inner still counts ---
			{
				Code:    "function f() { outer: while (true) { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    38,
					},
				},
			},

			// --- `for await (...of...)` triggers like normal for-of ---
			{
				Code:    "async function f() { for await (const x of y) { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (2). Maximum allowed is 1.",
						Line:      1,
						Column:    49,
					},
				},
			},

			// --- Body without braces (single-statement form) ---
			{
				Code:    "function f() { while (true) if (a) for (;;) {} }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    36,
					},
				},
			},

			// --- Class expression with method — same boundary as declaration ---
			{
				Code:    "var C = class { method() { if (a) { if (b) { if (c) {} } } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    46,
					},
				},
			},

			// --- Method parameter default arrow — independent depth scope ---
			{
				Code:    "class C { method(cb = () => { if (a) { if (b) { if (c) {} } } }) {} }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    49,
					},
				},
			},

			// --- Namespace body shares depth scope with its enclosing scope ---
			{
				Code:    "namespace N { if (a) { if (b) { if (c) {} } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tooDeeply",
						Message:   "Blocks are nested too deeply (3). Maximum allowed is 2.",
						Line:      1,
						Column:    33,
					},
				},
			},

			// --- Real-world example: callback chain — depth resets at every
			//     arrow boundary, but each inner arrow body still nests beyond
			//     the limit (max=1) ---
			{
				Code: `function load() {
  fetch('a').then((r) => {
    if (r.ok) {
      r.json().then((d) => {
        if (d.items) {
          d.items.forEach((it) => {
            if (it.active) {
              if (it.score > 0) {
              }
            }
          });
        }
      });
    }
  });
}`,
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					// Inside arrow 3 (forEach callback): if(it.active) depth=1,
					// if(it.score>0) depth=2 → reports at 2.
					{MessageId: "tooDeeply", Line: 8, Column: 15},
				},
			},
		},
	)
}
