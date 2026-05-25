package no_loop_func

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoLoopFuncRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLoopFuncRule,
		[]rule_tester.ValidTestCase{
			// ---- typescript-eslint specific tests ----
			{Code: `
for (let i = 0; i < 10; i++) {
  function foo() {
    console.log('A');
  }
}`},
			{Code: `
let someArray: MyType[] = [];
for (let i = 0; i < 10; i += 1) {
  someArray = someArray.filter((item: MyType) => !!item);
}`},
			{Code: `
type MyType = 1;
let someArray: MyType[] = [];
for (let i = 0; i < 10; i += 1) {
  someArray = someArray.filter((item: MyType) => !!item);
}`},

			// ---- Forked ESLint tests ----

			// Not inside a loop.
			{Code: `string = 'function a() {}';`},
			{Code: `for (var i=0; i<l; i++) { } var a = function() { i; };`},

			// Function declared in for-init — not considered "inside the loop".
			{Code: `for (var i=0, a=function() { i; }; i<l; i++) { }`},

			// Function declared in the for-in/for-of iterable — not "inside the loop".
			{Code: `for (var x in xs.filter(function(x) { return x != upper; })) { }`},
			{Code: `for (var x of xs.filter(function(x) { return x != upper; })) { }`},

			// No reference to loop-modified variables.
			{Code: `for (var i=0; i<l; i++) { (function() {}) }`},
			{Code: `for (var i in {}) { (function() {}) }`},
			{Code: `for (var i of {}) { (function() {}) }`},

			// Functions using unmodified `let` / `const` are OK (fresh per iteration).
			{Code: `for (let i=0; i<l; i++) { (function() { i; }) }`},
			{Code: `for (let i in {}) { i = 7; (function() { i; }) }`},
			{Code: `for (const i of {}) { (function() { i; }) }`},

			// Nested loops with `let` iteration variables.
			{Code: `for (let i = 0; i < 10; ++i) { for (let x in xs.filter(x => x != i)) {  } }`},

			// Closures over outer-scope `let` / `var` that aren't modified are OK.
			{Code: `let a = 0; for (let i=0; i<l; i++) { (function() { a; }); }`},
			{Code: `let a = 0; for (let i in {}) { (function() { a; }); }`},
			{Code: `let a = 0; for (let i of {}) { (function() { a; }); }`},
			{Code: `let a = 0; for (let i=0; i<l; i++) { (function() { (function() { a; }); }); }`},
			{Code: `let a = 0; for (let i in {}) { function foo() { (function() { a; }); } }`},
			{Code: `let a = 0; for (let i of {}) { (() => { (function() { a; }); }); }`},
			{Code: `var a = 0; for (let i=0; i<l; i++) { (function() { a; }); }`},
			{Code: `var a = 0; for (let i in {}) { (function() { a; }); }`},
			{Code: `var a = 0; for (let i of {}) { (function() { a; }); }`},

			// Closure over outer const — safe even though reassigned later (runtime error).
			{Code: `
let result = {};
for (const score in scores) {
  const letters = scores[score];
  letters.split('').forEach(letter => {
    result[letter] = score;
  });
}
result.__default = 6;`},

			// Variable declared after the loop — no modification visible.
			{Code: `
while (true) {
  (function() { a; });
}
let a;`},

			// Undeclared variables in the loop condition are picked up by no-undef.
			{Code: `while(i) { (function() { i; }) }`},
			{Code: `do { (function() { i; }) } while (i)`},

			// Variables declared outside the loop and not updated.
			{Code: `var i; while(i) { (function() { i; }) }`},
			{Code: `var i; do { (function() { i; }) } while (i)`},

			// Undeclared references — handled by no-undef, not here.
			{Code: `for (var i=0; i<l; i++) { (function() { undeclared; }) }`},
			{Code: `for (let i=0; i<l; i++) { (function() { undeclared; }) }`},
			{Code: `for (var i in {}) { i = 7; (function() { undeclared; }) }`},
			{Code: `for (let i in {}) { i = 7; (function() { undeclared; }) }`},
			{Code: `for (const i of {}) { (function() { undeclared; }) }`},
			{Code: `for (let i = 0; i < 10; ++i) { for (let x in xs.filter(x => x != undeclared)) {  } }`},

			// IIFE — immediately invoked, not saved off.
			{Code: `
let current = getStart();
while (current) {
  (() => {
    current;
    current.a;
    current.b;
    current.c;
    current.d;
  })();

  current = current.upper;
}`},
			{Code: `for (var i=0; (function() { i; })(), i<l; i++) { }`},
			{Code: `for (var i=0; i<l; (function() { i; })(), i++) { }`},
			{Code: `for (var i = 0; i < 10; ++i) { (()=>{ i;})() }`},
			{Code: `for (var i = 0; i < 10; ++i) { (function a(){i;})() }`},
			{Code: `
var arr = [];
for (var i = 0; i < 5; i++) {
  arr.push((f => f)((() => i)()));
}`},
			{Code: `
var arr = [];
for (var i = 0; i < 5; i++) {
  arr.push((() => {
    return (() => i)();
  })());
}`},

			// ---- ts-eslint specific: const captures are still safe ----
			// Parallel valid cases for the using/await-using invalid cases below — only
			// the writable form must be unsafe. `const` outside the loop is
			// safe because it has no writes anywhere.
			{Code: `const k = 10; for (var i = 0; i < l; i++) { (function () { k; }); }`},
			// Class declaration / Enum const-like binding.
			{Code: `class K {} for (var i = 0; i < l; i++) { (function () { K; }); }`},
			{Code: `enum E { A } for (var i = 0; i < l; i++) { (function () { E.A; }); }`},
			// `const` block-scoped inside the loop — fresh per iteration.
			{Code: `for (var i = 0; i < l; i++) { const k = i; (function () { k; }); }`},

			// ---- Real-world / edge cases (verified against ts-eslint upstream) ----

			// Closure references only an unmodified outer `let`.
			{Code: `let unchanged = 1; for (var i = 0; i < 5; i++) { (function () { unchanged; })(); }`},
			// Optional chain IIFE — `(arrow)?.()` still IIFE-shaped.
			{Code: `for (var i = 0; i < 5; i++) { ((() => i)?.()); }`},
			// Tagged template with arrow capturing fresh `let i`.
			{Code: "declare function tag(s: TemplateStringsArray, x: any): any; for (let i = 0; i < l; i++) { tag`${() => i}`; }"},
			// Spread arg with arrow capturing fresh `let i`.
			{Code: `declare function foo(...args: any[]): void; for (let i = 0; i < l; i++) { foo(...[() => i]); }`},
			// Labeled loop, closure references only its own locals.
			{Code: `outer: for (var i = 0; i < l; i++) { (function () { var inner = 1; inner; }); }`},
			// Function declared in for-init, references init's `let i`.
			{Code: `for (let i = 0, f = function() { i; }; i < 5; i++) { f(); }`},
			// Closure references only globals (no through refs to flag).
			{Code: `for (var i = 0; i < 5; i++) { (function () { console.log("hi"); }); }`},
			// Imported binding is read-only.
			{Code: `import { foo } from "./mod"; for (var i = 0; i < l; i++) { (function () { foo; }); }`},
			// `let i` block-scoped, plus `let f = () => i` block-scoped — both fresh.
			{Code: `for (let i = 0; i < l; i++) { let f = () => i; f(); }`},
			// Inner `let shadow = i` block-scoped per iteration; closure reads it, not the outer var.
			{Code: `var shadow = 1; for (var i = 0; i < 5; i++) { let shadow = i; (function () { shadow; }); }`},
			// Non-IIFE class field initializer is an expression, not a function — not flagged here
			// (only the constructor / method below get flagged in the invalid suite).
			{Code: `for (var i = 0; i < l; i++) { class C { f = i; } }`},

			// `using` declared OUTSIDE the loop with no write inside — safe.
			{Code: `async function f(bar: any) { using foo = bar(); for (var i = 0; i < 10; ++i) { (function () { foo; }); } }`},
			// Closure body references only `this`, no through refs.
			{Code: `class A { m(items: any[]) { for (var i = 0; i < items.length; i++) { (function (this: any) { console.log(this); }); } } }`},
			// Closure body is empty.
			{Code: `for (var i = 0; i < 5; i++) { (function () {}); }`},
			// Closure references global function (not a through-write target).
			{Code: `function helper() { return 1; } for (var i = 0; i < 5; i++) { (function () { helper(); }); }`},
			// `eval("i")` — `i` is inside a string literal, not a real reference.
			{Code: `for (var i = 0; i < 5; i++) { (function () { eval("i"); }); }`},
			// `new` IIFE with `this` only — no through refs.
			{Code: `for (var i = 0; i < 5; i++) { new (function () { this; })(); }`},
			// Parameter consumed (no through ref to flag).
			{Code: `for (var i = 0; i < 5; i++) { (function (x) { x; })(i); }`},
		},
		[]rule_tester.InvalidTestCase{
			// Each of these asserts complete position info to lock the report site
			// for the canonical container shapes (FE, arrow, FD, var = FE, FD + call).
			{
				Code: `for (var i=0; i<l; i++) { (function() { i; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
					Line:      1,
					Column:    28,
				}},
			},
			{
				Code: `for (var i=0; i<l; i++) { for (var j=0; j<m; j++) { (function() { i+j; }) } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i', 'j'.",
					Line:      1,
					Column:    54,
				}},
			},
			{
				Code: `for (var i in {}) { (function() { i; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
					Line:      1,
					Column:    22,
				}},
			},
			{
				Code: `for (var i of {}) { (function() { i; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
					Line:      1,
					Column:    22,
				}},
			},
			{
				Code: `for (var i=0; i < l; i++) { (() => { i; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      1,
					Column:    30,
				}},
			},
			{
				Code: `for (var i=0; i < l; i++) { var a = function() { i; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      1,
					Column:    37,
				}},
			},
			{
				Code: `for (var i=0; i < l; i++) { function a() { i; }; a(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      1,
					Column:    29,
				}},
			},

			// Closure captures a variable written inside the loop.
			{
				Code: `let a; for (let i=0; i<l; i++) { a = 1; (function() { a; });}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'a'.",
				}},
			},
			{
				Code: `let a; for (let i in {}) { (function() { a; }); a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `let a; for (let i of {}) { (function() { a; }); } a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `let a; for (let i=0; i<l; i++) { (function() { (function() { a; }); }); a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `let a; for (let i in {}) { a = 1; function foo() { (function() { a; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `let a; for (let i of {}) { (() => { (function() { a; }); }); } a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// Closure in nested loop captures the outer `var`.
			{
				Code: `for (var i = 0; i < 10; ++i) { for (let x in xs.filter(x => x != i)) {  } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `for (let x of xs) { let a; for (let y of ys) { a = 1; (function() { a; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `for (var x of xs) { for (let y of ys) { (function() { x; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `for (var x of xs) { (function() { x; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// Write inside the loop occurs before/after the closure declaration.
			{
				Code: `var a; for (let x of xs) { a = 1; (function() { a; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `var a; for (let x of xs) { (function() { a; }); a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `let a; function foo() { a = 10; } for (let x of xs) { (function() { a; }); } foo();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `let a; function foo() { a = 10; for (let x of xs) { (function() { a; }); } } foo();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- TypeScript-specific: loop var captured by a typed function ----
			{
				Code: `
for (var i = 0; i < 10; i++) {
  function foo() {
    console.log(i);
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `
for (var i = 0; i < 10; i++) {
  const handler = (event: Event) => {
    console.log(i);
  };
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- ts-eslint vs ESLint-core divergence lock-ins ----
			//
			// The upstream @typescript-eslint/no-loop-func plugin forks an OLDER
			// snapshot of ESLint's no-loop-func that pre-dates two later upstream
			// changes:
			//   (1) `using` / `await using` were not yet treated as constant
			//       bindings (they go through the unsafe-write check).
			//   (2) Repeated unsafe variable names were not deduplicated.
			// rslint's core no-loop-func tracks the *current* ESLint behavior
			// (so this plugin must NOT inherit it). The next 7 cases lock the
			// ts-eslint behavior in.

			// (1a) `using` declared inside a `var`-iteration loop body and
			// captured by a closure: ts-eslint reports it; ESLint core no longer.
			{
				Code: `
async function f(bar: any) {
  for (var i = 0; i < 10; ++i) {
    using foo = bar(i);
    (function () { foo; });
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'foo'.",
				}},
			},
			// (1b) Same shape with `await using`.
			{
				Code: `
async function g(bar: any) {
  for (var i = 0; i < 10; ++i) {
    await using x = bar(i);
    (function () { x; });
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'x'.",
				}},
			},
			// (1c) `for (using i of ...)` — iteration writes the binding each
			// step; ts-eslint reports.
			{
				Code: `async function h(foo: any) { for (using i of foo) { (function () { i; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},
			// (1d) `for (await using i of ...)` — same as above for the async form.
			{
				Code: `async function h(foo: any) { for (await using i of foo) { (function () { i; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},

			// (2a) Same name appearing N times must appear N times in the
			// message. ESLint core de-duplicates; ts-eslint does not.
			{
				Code: `for (var i = 0; i < 10; i++) { (function () { i; i; i; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i', 'i', 'i'.",
				}},
			},
			// (2b) Mixed names also preserve repetition order.
			{
				Code: `var a, b; for (var i = 0; i < l; i++) { a = i; b = i; (function () { a + b; a + b; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'a', 'b', 'a', 'b'.",
				}},
			},
			// (2c) Single-occurrence reference still appears exactly once.
			{
				Code: `for (var i = 0; i < 10; i++) { (function () { i; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},

			// ---- Additional tsgo edge-shape coverage (Dimensions 1–4) ----

			// Class-field arrow re-creates the closure per iteration.
			{
				Code: `for (var i = 0; i < l; i++) { class C { f = () => i; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},
			// Object-literal getter — tsgo splits accessors out of FE.
			{
				Code: `for (var i = 0; i < l; i++) { const o = { get x() { return i; } }; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Class static block — transparent boundary; the inner FE leaks `i`.
			{
				Code: `
for (var i = 0; i < l; i++) {
  class C {
    static {
      var f = function () { i; };
    }
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      5,
				}},
			},
			// Async IIFE — `async` flag bypasses the IIFE skip optimization.
			{
				Code: `
let current: any = null;
const arr: any[] = [];
while (current) {
  const p = (async () => {
    await Promise.resolve();
    current;
  })();
  arr.push(p);
  current = current.upper;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'current'.",
				}},
			},
			// Generator IIFE — same: `generator` flag bypasses the IIFE skip.
			{
				Code: `let a; for (var i = 0; i < l; i++) { (function* () { i; })(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Self-referencing named IIFE — referenced inside its own body, so
			// it is not skipped.
			{
				Code: `
let current: any = null;
const arr: any[] = [];
while (current) {
  (function f() {
    current;
    arr.push(f);
  })();
  current = current.upper;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Catch-clause binding captured by a closure inside a loop — `e`
			// rebinds per exception.
			{
				Code: `for (var i = 0; i < l; i++) { try { throw 0; } catch (e) { (function () { e; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'e'.",
				}},
			},
			// Destructuring iteration writes the binding each step.
			{
				Code: `for (var {a} of arr) { (function () { a; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'a'.",
				}},
			},
			// Computed key + function value both reference the loop var.
			{
				Code: `for (var i = 0; i < l; i++) { var o = { [i]: function () { return i; } }; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Async FunctionDeclaration inside a loop.
			{
				Code: `for (var i = 0; i < l; i++) { async function fn() { i; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- Real-world / edge cases (verified against ts-eslint upstream) ----

			// `Array#forEach` callback inside a `var`-iter loop.
			{
				Code: `declare const items: any[]; for (var i = 0; i < items.length; i++) { items.forEach(function (it) { console.log(i, it); }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},
			// Classic `setTimeout` capture-of-var pitfall.
			{
				Code: `for (var i = 0; i < 5; i++) { setTimeout(function () { console.log(i); }, 100); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},
			// Promise.then arrow capture.
			{
				Code: `declare const xs: Promise<any>[]; for (var i = 0; i < xs.length; i++) { xs[i].then(v => console.log(i, v)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Default parameter is itself an arrow capturing the loop var.
			{
				Code: `for (var i = 0; i < l; i++) { function m(x = (() => i)) { return x(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Self-reassigning FunctionDeclaration: `foo` becomes unsafe AFTER reassignment.
			// Both the FD `foo` and the FE `g` get flagged (with different vars).
			{
				Code: `for (var i = 0; i < l; i++) { function foo() { return i; } foo = null as any; var g = function () { foo; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeRefs"},
					{MessageId: "unsafeRefs"},
				},
			},
			// Conditional expression branching to an arrow.
			{
				Code: `declare const cond: boolean; for (var i = 0; i < l; i++) { const f = cond ? (() => i) : null; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Switch case closure.
			{
				Code: `declare const x: number; for (var i = 0; i < l; i++) { switch (x) { case 1: (function () { i; }); break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Deep curried arrow chain — outer arrow leaks `i` through 4 nestings.
			{
				Code: `for (var i = 0; i < 5; i++) { const f = () => () => () => () => i; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Destructuring assignment inside loop body counts as a write.
			{
				Code: `var a; for (var i = 0; i < l; i++) { ({a} = {a: i}); (function () { a; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'a'.",
				}},
			},
			// Same-name nested var loops merge declarations; closure still unsafe.
			{
				Code: `for (var i = 0; i < 5; i++) { for (var i = 0; i < 5; i++) { (function () { i; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},
			// Generator FunctionDeclaration.
			{
				Code: `for (var i = 0; i < l; i++) { function* gen() { yield i; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Async arrow IIFE inside a do-while.
			{
				Code: `let cur: any = null; const ps: Promise<any>[] = []; do { ps.push((async () => { await Promise.resolve(); cur; })()); cur = cur?.next; } while (cur);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'cur'.",
				}},
			},
			// Class with constructor + method both referencing loop var.
			{
				Code: `for (var i = 0; i < l; i++) { class C { f = i; constructor() { this.f = i; } m() { return i; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeRefs"},
					{MessageId: "unsafeRefs"},
				},
			},
			// `let` declared OUTSIDE the loop, modified AFTER the loop.
			{
				Code: `let modified = 0; const callbacks: Array<() => void> = []; for (let i = 0; i < 5; i++) { callbacks.push(() => modified); } modified = 99;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'modified'.",
				}},
			},
			// Read-modify-write counter (++) is itself a write.
			{
				Code: `var counter = 0; for (var i = 0; i < 5; i++) { counter++; (function () { counter; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'counter'.",
				}},
			},
			// Try/finally — closure in finally block.
			{
				Code: `for (var i = 0; i < l; i++) { try { } finally { (function () { i; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Array.from second-arg callback.
			{
				Code: `for (var i = 0; i < 5; i++) { Array.from({length: 3}, function (_, j) { return i + j; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- Position assertions (multi-line, exact range) ----

			// Multi-line FE — verifies report range spans full function expression.
			{
				Code: `
for (var i = 0; i < l; i++) {
  (function() {
    i;
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
					Line:      3,
					Column:    4,
					EndLine:   5,
					EndColumn: 4,
				}},
			},
			// Mixed safe / unsafe closures in same loop body — only the second flags.
			{
				Code: `let safe = 1; for (var i = 0; i < l; i++) { (function () { safe; }); (function () { i; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},
			// Two closures referring to the same loop var — both flagged independently.
			{
				Code: `for (var i = 0; i < l; i++) { (function () { i; }); (function () { i; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeRefs"},
					{MessageId: "unsafeRefs"},
				},
			},

			// ---- Real-user / additional edge shapes ----

			// Decorator on a method — body still references loop var.
			{
				Code: `declare function dec(...args: any[]): any; for (var i = 0; i < l; i++) { class C { @dec m() { return i; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Arrow with rest param.
			{
				Code: `for (var i = 0; i < 5; i++) { const f = (...args: any[]) => i + args.length; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Object destructuring with rename.
			{
				Code: `declare const arr: any[]; for (var {a: x} of arr) { (function () { x; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'x'.",
				}},
			},
			// Array destructuring with default.
			{
				Code: `declare const arr: any[]; for (var [x = 1] of arr) { (function () { x; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// 3-level nested destructuring.
			{
				Code: `declare const arr: any[]; for (var {a: {b: {c}}} of arr) { (function () { c; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Function inside method body — method scope is fine; the FE-in-loop is still inside the loop.
			{
				Code: `class A { m() { for (var i = 0; i < 5; i++) { const f = function () { return i; }; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// `typeof i` is still a reference to `i`.
			{
				Code: `for (var i = 0; i < 5; i++) { (function () { typeof i; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Assignment in `while` condition — `item` is rewritten each iteration.
			{
				Code: `declare const next: () => any; let item: any; while ((item = next())) { (function () { item; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'item'.",
				}},
			},
			// `var local = i;` initializer is itself a write each iteration.
			{
				Code: `for (var i = 0; i < 5; i++) { var local = i; (function () { local; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'local'.",
				}},
			},
			// Computed key with string concat + closure value.
			{
				Code: `for (var i = 0; i < 5; i++) { var o = { ["k" + i]: function () { return i; } }; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Class extends + method referencing loop var.
			{
				Code: `declare const Base: any; for (var i = 0; i < 5; i++) { class C extends Base { m() { return i; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Function pushed into outer array (default-export pattern).
			{
				Code: `const fns: any[] = []; for (var i = 0; i < 5; i++) { fns.push(function () { return i; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Closure that throws the captured var.
			{
				Code: `for (var i = 0; i < 5; i++) { (function () { throw i; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// `yield*` delegation in generator FunctionDeclaration.
			{
				Code: `for (var i = 0; i < 5; i++) { function* g() { yield* [i]; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// `??=` logical-assignment is a write to `x`.
			{
				Code: `let x: any = null; for (var i = 0; i < 5; i++) { x ??= i; (function () { x; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'x'.",
				}},
			},
			// `||=` logical-assignment is a write to `x`.
			{
				Code: `let x: any = null; for (var i = 0; i < 5; i++) { x ||= i; (function () { x; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// FE wrapped in `as any` type assertion still detected.
			{
				Code: `for (var i = 0; i < 5; i++) { const f = (function () { return i; }) as any; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			// Shorthand property `{port}` in object literal value position.
			// The identifier is a real read of the variable, not a property
			// name — `port` must be picked up as a through reference.
			{
				Code: `declare function setup(opts: any, cb: () => void): void; declare const tryLimits: number; let port = 3000; let found = false; let attempts = 0; const host = 'localhost'; while (!found && attempts <= tryLimits) { try { new Promise((resolve, reject) => { setup({ port, host }, () => { found = true; resolve(0); }); }); } catch { port++; } attempts++; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'port', 'found'.",
				}},
			},
			// Shorthand property in destructuring assignment write — closure
			// after reads the same variable, which was overwritten.
			{
				Code: `let port = 0; for (var i = 0; i < 5; i++) { ({port} = {port: i}); (function () { port; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'port'.",
				}},
			},
			// Closure body itself contains a destructuring shorthand write.
			{
				Code: `let port = 0; for (var i = 0; i < 5; i++) { (function () { ({port} = {port: i}); }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'port', 'i'.",
				}},
			},
			// Private class field arrow capturing loop var. Only the field
			// arrow is a function-in-loop; `g()` only calls `this.#f()`
			// which contains no through reference to the loop var.
			{
				Code: `for (var i = 0; i < 5; i++) { class C { #f = () => i; g() { return this.#f(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},
			// JSX expression container with arrow capturing loop var.
			{
				Code:     `declare const React: any; declare function Foo(p: any): any; function App() { const items: any[] = []; for (var i = 0; i < 5; i++) { items.push(<Foo onClick={() => i} />); } return items; }`,
				FileName: "react.tsx",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},
		},
	)
}
