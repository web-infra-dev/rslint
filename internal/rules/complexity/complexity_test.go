package complexity

import (
	"fmt"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestComplexity covers two corpora:
//
//  1. ESLint parity — every case from eslint/tests/lib/rules/complexity.js
//     migrated 1:1, with line / column added on invalid cases that pin the
//     exact report position (upstream omits them on most cases — locking them
//     in here protects against future refactors silently shifting).
//  2. Additional edge cases — tsgo AST shape variants (option-shape JSON
//     round-trip, computed-key-vs-initializer boundaries on PropertyDeclaration,
//     async / generator function-likes with class-field arrow initializers,
//     destructuring defaults in regular `let`/`const`, etc.) that ESLint's
//     tests don't enumerate because ESTree collapses these forms into the
//     same Property / AssignmentPattern / FunctionExpression nodes.
func TestComplexity(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ComplexityRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// 1. ESLint parity — mirrors tests/lib/rules/complexity.js
			// ============================================================
			{Code: "function a(x) {}"},
			{Code: "function b(x) {}", Options: 1},
			{Code: "function a(x) {if (true) {return x;}}", Options: 2},
			{Code: "function a(x) {if (true) {return x;} else {return x+1;}}", Options: 2},
			{Code: "function a(x) {if (true) {return x;} else if (false) {return x+1;} else {return 4;}}", Options: 3},
			{Code: "function a(x) {for(var i = 0; i < 5; i ++) {x ++;} return x;}", Options: 2},
			{Code: "function a(obj) {for(var i in obj) {obj[i] = 3;}}", Options: 2},
			{Code: "function a(x) {for(var i = 0; i < 5; i ++) {if(i % 2 === 0) {x ++;}} return x;}", Options: 3},
			{Code: "function a(obj) {if(obj){ for(var x in obj) {try {x.getThis();} catch (e) {x.getThat();}}} else {return false;}}", Options: 4},
			{Code: "function a(x) {try {x.getThis();} catch (e) {x.getThat();}}", Options: 2},
			{Code: "function a(x) {return x === 4 ? 3 : 5;}", Options: 2},
			{Code: "function a(x) {return x === 4 ? 3 : (x === 3 ? 2 : 1);}", Options: 3},
			{Code: "function a(x) {return x || 4;}", Options: 2},
			{Code: "function a(x) {x && 4;}", Options: 2},
			{Code: "function a(x) {x ?? 4;}", Options: 2},
			{Code: "function a(x) {x ||= 4;}", Options: 2},
			{Code: "function a(x) {x &&= 4;}", Options: 2},
			{Code: "function a(x) {x ??= 4;}", Options: 2},
			{Code: "function a(x) {x = 4;}", Options: 1},
			{Code: "function a(x) {x |= 4;}", Options: 1},
			{Code: "function a(x) {x &= 4;}", Options: 1},
			{Code: "function a(x) {x += 4;}", Options: 1},
			{Code: "function a(x) {x >>= 4;}", Options: 1},
			{Code: "function a(x) {x >>>= 4;}", Options: 1},
			{Code: "function a(x) {x == 4;}", Options: 1},
			{Code: "function a(x) {x === 4;}", Options: 1},
			{Code: "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: 3;}}", Options: 3},
			{Code: "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: if(x == 'foo') {5;};}}", Options: 4},
			{Code: "function a(x) {while(true) {'foo';}}", Options: 2},
			{Code: "function a(x) {do {'foo';} while (true)}", Options: 2},
			{Code: "if (foo) { bar(); }", Options: 3},
			{Code: "var a = (x) => {do {'foo';} while (true)}", Options: 2},

			// ---- Modified complexity ----
			{
				Code:    "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: 3;}}",
				Options: map[string]interface{}{"max": 2, "variant": "modified"},
			},
			{
				Code:    "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: if(x == 'foo') {5;};}}",
				Options: map[string]interface{}{"max": 3, "variant": "modified"},
			},

			// ---- Class fields ----
			{Code: "function foo() { class C { x = a || b; y = c || d; } }", Options: 2},
			{Code: "function foo() { class C { static x = a || b; static y = c || d; } }", Options: 2},
			{Code: "function foo() { class C { x = a || b; y = c || d; } e || f; }", Options: 2},
			{Code: "function foo() { a || b; class C { x = c || d; y = e || f; } }", Options: 2},
			{Code: "function foo() { class C { [x || y] = a || b; } }", Options: 2},
			{Code: "class C { x = a || b; y() { c || d; } z = e || f; }", Options: 2},
			{Code: "class C { x() { a || b; } y = c || d; z() { e || f; } }", Options: 2},
			{Code: "class C { x = (() => { a || b }) || (() => { c || d }) }", Options: 2},
			{Code: "class C { x = () => { a || b }; y = () => { c || d } }", Options: 2},
			{Code: "class C { x = a || (() => { b || c }); }", Options: 2},
			{Code: "class C { x = class { y = a || b; z = c || d; }; }", Options: 2},
			{Code: "class C { x = a || class { y = b || c; z = d || e; }; }", Options: 2},
			{Code: "class C { x; y = a; static z; static q = b; }", Options: 1},

			// ---- Class static blocks ----
			{Code: "function foo() { class C { static { a || b; } static { c || d; } } }", Options: 2},
			{Code: "function foo() { a || b; class C { static { c || d; } } }", Options: 2},
			{Code: "function foo() { class C { static { a || b; } } c || d; }", Options: 2},
			{Code: "function foo() { class C { static { a || b; } } class D { static { c || d; } } }", Options: 2},
			{Code: "class C { static { a || b; } static { c || d; } }", Options: 2},
			{Code: "class C { static { a || b; } static { c || d; } static { e || f; } }", Options: 2},
			{Code: "class C { static { () => a || b; c || d; } }", Options: 2},
			{Code: "class C { static { a || b; () => c || d; } static { c || d; } }", Options: 2},
			{Code: "class C { static { a } }", Options: 1},
			{Code: "class C { static { a } static { b } }", Options: 1},
			{Code: "class C { static { a || b; } } class D { static { c || d; } }", Options: 2},
			{Code: "class C { static { a || b; } static c = d || e; }", Options: 2},
			{Code: "class C { static a = b || c; static { c || d; } }", Options: 2},
			{Code: "class C { static { a || b; } c = d || e; }", Options: 2},
			{Code: "class C { a = b || c; static { d || e; } }", Options: 2},
			{Code: "class C { static { a || b; c || d; } }", Options: 3},
			{Code: "class C { static { if (a || b) c = d || e; } }", Options: 4},

			// ---- Object property options ----
			{Code: "function b(x) {}", Options: map[string]interface{}{"max": 1}},

			// ---- Optional chaining ----
			{Code: "function a(b) { b?.c; }", Options: map[string]interface{}{"max": 2}},

			// ---- Default function parameter values ----
			{Code: "function a(b = '') {}", Options: map[string]interface{}{"max": 2}},

			// ---- Default destructuring values ----
			{Code: "function a(b) { const { c = '' } = b; }", Options: map[string]interface{}{"max": 2}},
			{Code: "function a(b) { const [ c = '' ] = b; }", Options: map[string]interface{}{"max": 2}},

			// ============================================================
			// 2. Additional edge cases
			// ============================================================

			// ---- Empty / degenerate forms ----
			{Code: ""},
			{Code: "function f() {}"},
			{Code: "var f = () => {};"},
			{Code: "class C {}"},
			{Code: "class C { static {} }"},

			// ---- Options shapes — exercise the JSON path ----
			// Bare number directly (single-option CLI).
			{Code: "function f() {}", Options: 5},
			// Bare object (single-option CLI).
			{Code: "function f() {}", Options: map[string]interface{}{"max": 5}},
			// Array-wrapped object (multi-element / rule_tester shape).
			{Code: "function f() {}", Options: []interface{}{map[string]interface{}{"max": 5}}},
			// Empty options array → defaults to 20.
			{Code: "function f() {}", Options: []interface{}{}},
			// Legacy `maximum` key.
			{Code: "function f() { if (a) {} }", Options: map[string]interface{}{"maximum": 2}},
			// `maximum` wins when both present and truthy
			// (mirrors ESLint's `option.maximum || option.max` coercion).
			{Code: "function f() { if (a) {} }", Options: map[string]interface{}{"maximum": 5, "max": 1}},

			// ---- Nested function-likes — each has its own counter ----
			// outer counts a||b (= 2) and inner counts c||d (= 2); neither bleeds
			// into the other.
			{Code: "function outer() { a || b; function inner() { c || d; } }", Options: 2},
			{Code: "var outer = () => { a || b; var inner = () => c || d; }", Options: 2},

			// ---- async / generator / async generator boundary ----
			{Code: "async function f() { a || b; }", Options: 2},
			{Code: "function* g() { a || b; }", Options: 2},
			{Code: "async function* g() { a || b; }", Options: 2},

			// ---- `for await (...of ...)` is the same KindForOfStatement, counted ----
			{Code: "async function f() { for await (const x of y) {} }", Options: 2},

			// ---- Optional-chain forms ----
			{Code: "function a(b) { b?.['c']; }", Options: map[string]interface{}{"max": 2}},
			{Code: "function a(b) { b?.(); }", Options: map[string]interface{}{"max": 2}},

			// ---- TypeScript-only: type annotations / generics don't count ----
			{Code: "function f<T extends { x: number }>(a: T): boolean { return a.x > 0; }", Options: 2},
			{Code: "function f(a: unknown): a is string { return typeof a === 'string'; }", Options: 1},
			// `as` / `satisfies` are expression wrappers, not branches.
			{Code: "function f(a: unknown) { return a as string; }", Options: 1},
			{Code: "function f() { return ({ a: 1 } satisfies { a: number }); }", Options: 1},

			// ---- Type-only declarations have no executable branches ----
			{Code: "interface I { x: { y: { z: number } } } function f() { return 1; }", Options: 1},
			{Code: "type T = { a: { b: number } }; function f() { return 1; }", Options: 1},
			{Code: "type T<X> = X extends Array<infer Y> ? Y : never; function f() { return 1; }", Options: 1},

			// ---- TS overload signatures / abstract / declare members — body-absent forms ----
			{Code: "abstract class A { abstract method(): void; }", Options: 1},
			{Code: "function f(x: number): number; function f(x: string): string; function f(x: any): any { return x; }", Options: 1},

			// ---- `else` block alternate (NOT chained `else if`) ----
			{Code: "function f() { if (a) {} else { if (b) {} } }", Options: 3},

			// ---- LabeledStatement is transparent — inner `while` still counts ----
			{Code: "function f() { outer: while (a) {} }", Options: 2},

			// ---- Statement bodies without braces ----
			{Code: "function f() { while (true) if (a) for (;;) {} }", Options: 4},

			// ---- Class expression vs declaration ----
			{Code: "var C = class { method() { a || b; } }", Options: 2},

			// ---- Method parameter default with arrow — arrow has its own scope ----
			{Code: "class C { method(cb = () => { a || b; }) {} }", Options: 2},

			// ---- TS Enum / namespace — no branches ----
			{Code: "enum E { A, B, C } function f() { return 1; }", Options: 1},
			{Code: "namespace N { export function f() { return 1; } }", Options: 1},

			// ---- Computed key in PropertyDeclaration: counts toward enclosing scope, NOT the field-init ----
			// Top-level: no enclosing function, so `a || b` in the computed key has no owner.
			{Code: "class C { [a || b] = c; }", Options: 1},
			// Top-level: `a || b` in the computed key has no owner; field-init complexity is just 1.
			{Code: "class C { [a || b]; }", Options: 1},

			// ---- Class-field initializer that IS a function takes over as the function code path ----
			// Without our handling, the field-init AND the arrow would both count `a||b||c`.
			// With it, only the arrow counts → complexity 3 inside the arrow.
			{Code: "class C { x = () => a || b || c; }", Options: 3},

			// ---- Optional-chain test cases ----
			// `b?.c.d` — only the inner `?.c` counts (outer `.d` has no `?.`).
			{Code: "function a(b) { b?.c.d; }", Options: map[string]interface{}{"max": 2}},

			// ---- Deeply nested function-likes ----
			// Each function-like is an independent counter; deep nesting must
			// not bleed into outer counters.
			{Code: "function a() { function b() { function c() { function d() { e || f; } } } }", Options: 2},
			// Mix of arrow / function expression / method, all reset the counter.
			{Code: "class C { m() { return (() => { var x = function () { a || b; }; return x; })(); } }", Options: 2},

			// ---- Parameter / destructuring defaults nested inside parameters ----
			// `[ , a = 1 ]` — Array hole then BindingElement with default.
			{Code: "function f([ , a = 1 ]) {}", Options: 2},
			// Nested destructuring defaults — each BindingElement with init counts once.
			{Code: "function f({ a: { b = 1 } = {} }) {}", Options: 3},
			// Rest pattern inside destructuring — RestElement has no Initializer.
			{Code: "function f({ a, ...rest }) {}", Options: 1},
			// SpreadAssignment in object literal — must not crash and counts no branches.
			{Code: "function f() { return { ...x }; }", Options: 1},

			// ---- Constructor parameter properties (TS-only): public/private/protected `x = default` ----
			// In tsgo, constructor parameter properties carry an Initializer
			// like normal parameters.
			{Code: "class C { constructor(public a = 1, private b = 2) {} }", Options: 3},

			// ---- Decorator with branch — decorator expression counts toward enclosing scope ----
			// `@(a || b)` is an expression in a position that has no enclosing
			// function/field-init owner (the class declaration is top-level),
			// so the `||` walks past the class and finds no owner → no diagnostic.
			{Code: "function dec() { return (...a) => {}; } @(dec() || dec()) class C {}", Options: 1},
			// But when the class is inside a function, the decorator branch
			// counts toward the enclosing function.
			{Code: "function dec() { return (...a) => {}; } function outer() { @(dec() || dec()) class C {} }", Options: 2},

			// ---- Object literal getter / setter — `KindGetAccessor`/`KindSetAccessor` form their own scopes ----
			{Code: "var o = { get x() { return a || b; } }", Options: 2},
			{Code: "var o = { set x(v) { a || b; } }", Options: 2},

			// ---- Optional chain mixed with logical ops ----
			// `(a?.b) || c` — `?.` counts on the inner node, `||` on the outer.
			{Code: "function f(a) { return (a?.b) || c; }", Options: 3},

			// ---- LogicalExpression in TS-only contexts ----
			// `as` wrapper around a logical expression — `||` still counts.
			{Code: "function f(a: any) { return (a || b) as string; }", Options: 2},
			// Non-null assertion wrapping a logical expression.
			{Code: "function f(a) { return (a || b)!; }", Options: 2},
			// Satisfies wrapper.
			{Code: "function f(a) { return (a || b) satisfies unknown; }", Options: 2},

			// ---- Real-world: simple guard-clause functions stay valid at default ----
			{Code: "function isEmpty(s) { return s == null || s === ''; }"},
			{Code: "const max = (a, b) => a > b ? a : b;"},
			{Code: "function clamp(v, lo, hi) { return Math.min(Math.max(v, lo), hi); }"},

			// ---- Real-world: pure factory function — small ----
			{
				Code: `function createCounter(start = 0) {
  let count = start;
  return {
    inc() { return ++count; },
    dec() { return --count; },
    get() { return count; },
  };
}`,
				Options: 2,
			},

			// ---- Just-at-the-limit: complexity exactly equals max ----
			// `if + ||` gives complexity 3, max 3 → valid.
			{Code: "function f(a, b) { if (a || b) return 1; return 0; }", Options: 3},
			// Modified switch + if = 1 (start) + 1 (switch) + 1 (if) = 3 ≤ 3.
			{
				Code:    "function f(x) { switch (x) { case 1: case 2: if (a) {} break; } }",
				Options: map[string]interface{}{"max": 3, "variant": "modified"},
			},

			// ---- Module-level branches: no diagnostic ----
			// Top-level branches do NOT report (program code path is excluded).
			{Code: "if (foo) { bar(); } else if (baz) { qux(); } else { boom(); }", Options: 1},
			{Code: "for (const x of [1, 2, 3]) { if (x > 0) {} }", Options: 1},
			{Code: "switch (x) { case 1: case 2: case 3: case 4: case 5: break; }", Options: 1},

			// ---- Optional chain on call expression with arguments ----
			{Code: "function f(g) { return g?.(1, 2); }", Options: 2},

			// ---- Class with only static fields, no methods ----
			{Code: "class C { static a = 1; static b = 2; static c = 3; }", Options: 1},

			// ---- Decorator factory call (non-branching) on class member ----
			{
				Code:    "function dec() { return (...a) => {}; } class C { @dec() method() {} }",
				Options: 1,
			},

			// ---- ESLint-style default threshold (20) — modest functions pass ----
			{Code: "function f(x) { if (x) return 1; if (!x) return 2; return 0; }"},

			// ---- TS overloads with empty implementation ----
			{
				Code: `class C {
  to(x: number): string;
  to(x: string): number;
  to(x: any): any { return x; }
}`,
				Options: 1,
			},

			// ---- Type predicate return-type annotation ----
			{Code: "function isString(v: unknown): v is string { return typeof v === 'string'; }", Options: 1},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// 1. ESLint parity — mirrors tests/lib/rules/complexity.js
			// ============================================================
			{
				Code:    "function a(x) {}",
				Options: 0,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 1. Maximum allowed is 0.",
					},
				},
			},
			{
				Code:    "function foo(x) {if (x > 10) {return 'x is greater than 10';} else if (x > 5) {return 'x is greater than 5';} else {return 'x is less than 5';}}",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
						Column:    1,
						EndColumn: 13,
					},
				},
			},
			{
				Code:    "var func = function () {}",
				Options: 0,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function has a complexity of 1. Maximum allowed is 0.",
					},
				},
			},
			{
				Code:    "var obj = { a(x) {} }",
				Options: 0,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'a' has a complexity of 1. Maximum allowed is 0.",
					},
				},
			},
			{
				Code:    "class Test { a(x) {} }",
				Options: 0,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'a' has a complexity of 1. Maximum allowed is 0.",
					},
				},
			},
			{
				Code:    "var a = (x) => {if (true) {return x;}}",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Arrow function has a complexity of 2. Maximum allowed is 1.",
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			{
				Code:    "function a(x) {if (true) {return x;}}",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 2. Maximum allowed is 1.",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 11,
					},
				},
			},
			{
				Code:    "function a(x) {if (true) {return x;} else {return x+1;}}",
				Options: 1,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(x) {if (true) {return x;} else if (false) {return x+1;} else {return 4;}}",
				Options: 2,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(x) {for(var i = 0; i < 5; i ++) {x ++;} return x;}",
				Options: 1,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(obj) {for(var i in obj) {obj[i] = 3;}}",
				Options: 1,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(obj) {for(var i of obj) {obj[i] = 3;}}",
				Options: 1,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(x) {for(var i = 0; i < 5; i ++) {if(i % 2 === 0) {x ++;}} return x;}",
				Options: 2,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(obj) {if(obj){ for(var x in obj) {try {x.getThis();} catch (e) {x.getThat();}}} else {return false;}}",
				Options: 3,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(x) {try {x.getThis();} catch (e) {x.getThat();}}",
				Options: 1,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(x) {return x === 4 ? 3 : 5;}",
				Options: 1,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(x) {return x === 4 ? 3 : (x === 3 ? 2 : 1);}",
				Options: 2,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{Code: "function a(x) {return x || 4;}", Options: 1, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "complex"}}},
			{Code: "function a(x) {x && 4;}", Options: 1, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "complex"}}},
			{Code: "function a(x) {x ?? 4;}", Options: 1, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "complex"}}},
			{Code: "function a(x) {x ||= 4;}", Options: 1, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "complex"}}},
			{Code: "function a(x) {x &&= 4;}", Options: 1, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "complex"}}},
			{Code: "function a(x) {x ??= 4;}", Options: 1, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "complex"}}},
			{
				Code:    "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: 3;}}",
				Options: 2,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: if(x == 'foo') {5;};}}",
				Options: 3,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(x) {while(true) {'foo';}}",
				Options: 1,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(x) {do {'foo';} while (true)}",
				Options: 1,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "function a(x) {(function() {while(true){'foo';}})(); (function() {while(true){'bar';}})();}",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "complex"},
					{MessageId: "complex"},
				},
			},
			{
				Code:    "function a(x) {(function() {while(true){'foo';}})(); (function() {'bar';})();}",
				Options: 1,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "complex"}},
			},
			{
				Code:    "var obj = { a(x) { return x ? 0 : 1; } };",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'a' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "var obj = { a: function b(x) { return x ? 0 : 1; } };",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'a' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code: createComplexity(21),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'test' has a complexity of 21. Maximum allowed is 20.",
					},
				},
			},
			{
				Code:    createComplexity(21),
				Options: map[string]interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'test' has a complexity of 21. Maximum allowed is 20.",
					},
				},
			},

			// ---- Modified complexity ----
			{
				Code:    "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: 3;}}",
				Options: map[string]interface{}{"max": 1, "variant": "modified"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: if(x == 'foo') {5;};}}",
				Options: map[string]interface{}{"max": 2, "variant": "modified"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},

			// ---- Class fields ----
			{
				Code:    "function foo () { a || b; class C { x; } c || d; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function foo () { a || b; class C { x = c; } d || e; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function foo () { a || b; class C { [x || y]; } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function foo () { a || b; class C { [x || y] = c; } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function foo () { class C { [x || y]; } a || b; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function foo () { class C { [x || y] = a; } b || c; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function foo () { class C { [x || y]; [z || q]; } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function foo () { class C { [x || y] = a; [z || q] = b; } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function foo () { a || b; class C { x = c || d; } e || f; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { x(){ a || b; } y = c || d || e; z() { f || g; } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { x = a || b; y() { c || d || e; } z = f || g; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'y' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { x; y() { c || d || e; } z; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'y' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { x = a || b; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "(class { x = a || b; })",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "class C { static x = a || b; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "(class { x = a ? b : c; })",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "class C { x = a || b || c; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { x = a || b; y = b || c || d; z = e || f; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 3. Maximum allowed is 2.",
						Line:      1,
						Column:    27,
						EndLine:   1,
						EndColumn: 38,
					},
				},
			},
			{
				Code:    "class C { x = a || b || c; y = d || e; z = f || g || h; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 3. Maximum allowed is 2.",
						Line:      1,
						Column:    15,
						EndLine:   1,
						EndColumn: 26,
					},
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 3. Maximum allowed is 2.",
						Line:      1,
						Column:    44,
						EndLine:   1,
						EndColumn: 55,
					},
				},
			},
			{
				Code:    "class C { x = () => a || b || c; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'x' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { x = (() => a || b || c) || d; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Arrow function has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { x = () => a || b || c; y = d || e; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'x' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { x = () => a || b || c; y = d || e || f; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'x' has a complexity of 3. Maximum allowed is 2.",
					},
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 3. Maximum allowed is 2.",
						Line:      1,
						Column:    38,
						EndLine:   1,
						EndColumn: 49,
					},
				},
			},
			{
				Code:    "class C { x = function () { a || b }; y = function () { c || d }; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'x' has a complexity of 2. Maximum allowed is 1.",
					},
					{
						MessageId: "complex",
						Message:   "Method 'y' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "class C { x = class { [y || z]; }; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 2. Maximum allowed is 1.",
						Line:      1,
						Column:    15,
						EndLine:   1,
						EndColumn: 34,
					},
				},
			},
			{
				Code:    "class C { x = class { [y || z] = a; }; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 2. Maximum allowed is 1.",
						Line:      1,
						Column:    15,
						EndLine:   1,
						EndColumn: 38,
					},
				},
			},
			{
				Code:    "class C { x = class { y = a || b; }; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 2. Maximum allowed is 1.",
						Line:      1,
						Column:    27,
						EndLine:   1,
						EndColumn: 33,
					},
				},
			},

			// ---- Class static blocks ----
			{
				Code:    "function foo () { a || b; class C { static {} } c || d; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function foo () { a || b; class C { static { c || d; } } e || f; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { static { a || b; }  }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 2. Maximum allowed is 1.",
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 17,
					},
				},
			},
			{
				Code:    "class C { static { a || b || c; }  }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { static { a || b; c || d; }  }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { static { a || b; c || d; e || f; }  }",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 4. Maximum allowed is 3.",
					},
				},
			},
			{
				Code:    "class C { static { a || b; c || d; { e || f; } }  }",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 4. Maximum allowed is 3.",
					},
				},
			},
			{
				Code:    "class C { static { if (a || b) c = d || e; } }",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 4. Maximum allowed is 3.",
					},
				},
			},
			{
				Code:    "class C { static { if (a || b) c = (d => e || f)() || (g => h || i)(); } }",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 4. Maximum allowed is 3.",
					},
				},
			},
			{
				Code:    "class C { x(){ a || b; } static { c || d || e; } z() { f || g; } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { x = a || b; static { c || d || e; } y = f || g; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { static x = a || b; static { c || d || e; } static y = f || g; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { static { a || b; } static(){ c || d || e; } static { f || g; } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'static' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { static { a || b; } static static(){ c || d || e; } static { f || g; } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Static method 'static' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "class C { static { a || b; } static x = c || d || e; static { f || g; } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 3. Maximum allowed is 2.",
						Column:    41,
						EndColumn: 52,
					},
				},
			},
			{
				Code:    "class C { static { a || b || c || d; } static { e || f || g; } }",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 4. Maximum allowed is 3.",
						Column:    11,
						EndColumn: 17,
					},
				},
			},
			{
				Code:    "class C { static { a || b || c; } static { d || e || f || g; } }",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 4. Maximum allowed is 3.",
						Column:    35,
						EndColumn: 41,
					},
				},
			},
			{
				Code:    "class C { static { a || b || c || d; } static { e || f || g || h; } }",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 4. Maximum allowed is 3.",
						Column:    11,
						EndColumn: 17,
					},
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 4. Maximum allowed is 3.",
						Column:    40,
						EndColumn: 46,
					},
				},
			},
			{
				Code:    "class C { x = () => a || b || c; y = f || g || h; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'x' has a complexity of 3. Maximum allowed is 2.",
						Column:    11,
						EndColumn: 15,
					},
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 3. Maximum allowed is 2.",
						Column:    38,
						EndColumn: 49,
					},
				},
			},

			// ---- Object property options ----
			{
				Code:    "function a(x) {}",
				Options: map[string]interface{}{"max": 0},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 1. Maximum allowed is 0.",
					},
				},
			},
			{
				Code:    "const obj = { b: (a) => a?.b?.c, c: function (a) { return a?.b?.c; } };",
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'b' has a complexity of 3. Maximum allowed is 2.",
						Column:    15,
						EndColumn: 18,
					},
					{
						MessageId: "complex",
						Message:   "Method 'c' has a complexity of 3. Maximum allowed is 2.",
						Column:    34,
						EndColumn: 46,
					},
				},
			},

			// ---- Optional chaining ----
			{
				Code:    "function a(b) { b?.c; }",
				Options: map[string]interface{}{"max": 1},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "function a(b) { b?.['c']; }",
				Options: map[string]interface{}{"max": 1},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "function a(b) { b?.c; d || e; }",
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function a(b) { b?.c?.d; }",
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function a(b) { b?.['c']?.['d']; }",
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function a(b) { b?.c?.['d']; }",
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function a(b) { b?.c.d?.e; }",
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 3. Maximum allowed is 2.",
						Column:    1,
						EndColumn: 11,
					},
				},
			},
			{
				Code:    "function a(b) { b?.c?.(); }",
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},
			{
				Code:    "function a(b) { b?.c?.()?.(); }",
				Options: map[string]interface{}{"max": 3},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 4. Maximum allowed is 3.",
					},
				},
			},

			// ---- Default function parameter values ----
			{
				Code:    "function a(b = '') {}",
				Options: map[string]interface{}{"max": 1},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Default destructuring values ----
			{
				Code:    "function a(b) { const { c = '' } = b; }",
				Options: map[string]interface{}{"max": 1},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "function a(b) { const [ c = '' ] = b; }",
				Options: map[string]interface{}{"max": 1},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "function a(b) { const [ { c: d = '' } = {} ] = b; }",
				Options: map[string]interface{}{"max": 1},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 3. Maximum allowed is 1.",
					},
				},
			},

			// ============================================================
			// 2. Additional edge cases
			// ============================================================

			// ---- Array-wrapped options shape (rule_tester / multi-element CLI) ----
			{
				Code:    "function a(x) {}",
				Options: []interface{}{map[string]interface{}{"max": 0}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 1. Maximum allowed is 0.",
					},
				},
			},
			// `{ maximum: N }` is honored even when `max` is also present and
			// `maximum` is truthy (locks in `option.maximum || option.max`).
			{
				Code:    "function a(x) { if (x) {} }",
				Options: map[string]interface{}{"maximum": 1, "max": 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			// `maximum: 0` falls through to `max` (matches `0 || N` JS coercion).
			{
				Code:    "function a(x) { if (x) {} }",
				Options: map[string]interface{}{"maximum": 0, "max": 1},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Multi-line code: line/column precision ----
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
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 4. Maximum allowed is 2.",
						Line:      1,
						Column:    1,
					},
				},
			},

			// ---- Multiple separate violations in one source ----
			{
				Code: `function a() { if (1) { if (2) { if (3) {} } } }
function b() { while (1) { while (2) { while (3) {} } } }`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'a' has a complexity of 4. Maximum allowed is 2.",
						Line:      1,
					},
					{
						MessageId: "complex",
						Message:   "Function 'b' has a complexity of 4. Maximum allowed is 2.",
						Line:      2,
					},
				},
			},

			// ---- async / generator boundary ----
			{
				Code:    "async function f() { if (a) {} else if (b) {} }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Async function 'f' has a complexity of 3. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "function* g() { if (a) {} else if (b) {} }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Generator function 'g' has a complexity of 3. Maximum allowed is 1.",
					},
				},
			},

			// ---- Async / static method, getter, setter ----
			{
				Code:    "class C { async method() { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Async method 'method' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "class C { static method() { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Static method 'method' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "class C { get x() { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Getter 'x' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "class C { set x(v) { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Setter 'x' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "class C { constructor() { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Constructor has a complexity of 2. Maximum allowed is 1.",
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 22,
					},
				},
			},

			// ---- Private member name ----
			{
				Code:    "class C { #x() { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Private method '#x' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Top-level (no enclosing function) statements: NO report ----
			// ESLint's complexity rule does not report on the program code path.
			// The `||` here would otherwise be counted, but there's no owner.
			//
			// Locked in as a SEPARATE valid case above to prove no diagnostic;
			// here we lock in the inverse — a function in the same file IS reported.
			{
				Code:    "a || b; function f() { c || d || e; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'f' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},

			// ---- Computed key + initializer-with-branches in the SAME field ----
			// The computed key counts toward the enclosing function; the
			// initializer counts toward its own field-init code path. The
			// field-init exits BEFORE the enclosing function, so its diagnostic
			// is reported first.
			{
				Code:    "function foo() { class C { [x || y] = a || b || c; } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 3. Maximum allowed is 1.",
					},
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Deeply nested function-likes — only the deepest reports ----
			// outer(1) → inner(1) → deepest(1+3=4). With max=1, all three
			// trigger because each starts at 1 and outer/inner remain at 1.
			// Wait — outer/inner have complexity 1 which is NOT > max=1. Only
			// deepest reports.
			{
				Code:    "function outer() { function inner() { function deepest() { return a || b || c || d; } } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'deepest' has a complexity of 4. Maximum allowed is 1.",
					},
				},
			},

			// ---- Class field initializer containing a function-like with branches ----
			// `[arrow with branches]` — arrow takes over as code path; field-init
			// counter is not pushed (initializer IS a function).
			{
				Code:    "class C { x = (a, b) => a || b || c; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'x' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},

			// ---- Static block nested inside class field initializer (via inner class) ----
			// Outer field-init has 1 + (`||`)=2 (NOT exceeding max=2);
			// Inner static block has 1 + (`||` + `||`)=3 (EXCEEDS max=2).
			{
				Code:    "class C { x = a || class { static { b || c || d; } } }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},

			// ---- Optional-chain inside parameter default ----
			// Parameter default counts +1; `?.` counts +1 (both attribute to
			// function f). 1 + 1 + 1 = 3.
			{
				Code:    "function f(a = b?.c) {}",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'f' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},

			// ---- Logical operator inside default destructuring ----
			// BindingElement default + `||` = 1 + 1 = 2 increments. 1 + 2 = 3.
			{
				Code:    "function f({ a = b || c } = {}) {}",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'f' has a complexity of 4. Maximum allowed is 2.",
					},
				},
			},

			// ---- Real-world: validation function with many guards ----
			// 1 (start) + 5 ifs + 2 (&&) + 1 (||) = 9.
			{
				Code: `function validate(input) {
  if (!input) return false;
  if (typeof input !== 'object') return false;
  if (input.type === 'a' && input.value > 0) return true;
  if (input.type === 'b' && input.value < 0) return true;
  if (input.type === 'c' || input.type === 'd') return true;
  return false;
}`,
				Options: 5,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'validate' has a complexity of 9. Maximum allowed is 5.",
					},
				},
			},

			// ---- Real-world: React-like reducer ----
			{
				Code: `function reducer(state, action) {
  switch (action.type) {
    case 'A': return { ...state, a: 1 };
    case 'B': return { ...state, b: 2 };
    case 'C': return { ...state, c: 3 };
    case 'D': return { ...state, d: 4 };
    case 'E': return { ...state, e: 5 };
    default: return state;
  }
}`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'reducer' has a complexity of 6. Maximum allowed is 3.",
					},
				},
			},
			// Same reducer in modified variant: switch counts 1, cases don't.
			{
				Code: `function reducer(state, action) {
  switch (action.type) {
    case 'A': return { ...state, a: 1 };
    case 'B': return { ...state, b: 2 };
    case 'C': return { ...state, c: 3 };
    case 'D': return { ...state, d: 4 };
    case 'E': return { ...state, e: 5 };
    default: return state;
  }
}`,
				Options: map[string]interface{}{"max": 1, "variant": "modified"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'reducer' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Real-world: async API client with retries and fallback ----
			// 1 (start) + 1 (param default) + 1 (for) + 1 (if) + 1 (catch) + 1 (if) = 6.
			{
				Code: `async function fetchWithRetry(url, retries = 3) {
  for (let i = 0; i <= retries; i++) {
    try {
      const response = await fetch(url);
      if (response.ok) {
        return await response.json();
      }
    } catch (e) {
      if (i === retries) throw e;
    }
  }
}`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Async function 'fetchWithRetry' has a complexity of 6. Maximum allowed is 3.",
					},
				},
			},

			// ---- Real-world: TypeScript type-narrowing chain ----
			{
				Code: `function process(value: unknown): string {
  if (typeof value === 'string') return value;
  if (typeof value === 'number') return String(value);
  if (typeof value === 'boolean') return value ? 'true' : 'false';
  if (value === null) return 'null';
  if (value === undefined) return 'undefined';
  if (Array.isArray(value)) return value.join(',');
  return JSON.stringify(value);
}`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'process' has a complexity of 8. Maximum allowed is 3.",
					},
				},
			},

			// ---- Real-world: deeply chained optional access ----
			{
				Code:    "function getUserCity(user) { return user?.profile?.address?.city ?? 'unknown'; }",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'getUserCity' has a complexity of 5. Maximum allowed is 3.",
					},
				},
			},

			// ---- Real-world: React-like component with conditional rendering ----
			// 1 (start) + 3 (destructuring defaults: items, onClick, hidden) + 2 (ifs) + 1 (&&) = 7.
			// The `?.` and `??` inside `items.map(item => ...)` belong to the
			// inner arrow's code path, not Component's.
			{
				Code: `function Component(props) {
  const { kind, items = [], onClick = () => {}, hidden = false } = props;
  if (hidden) return null;
  if (kind === 'list' && items.length > 0) {
    return items.map(item => item?.name ?? 'anonymous');
  }
  return null;
}`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'Component' has a complexity of 7. Maximum allowed is 3.",
					},
				},
			},

			// ---- Generator with yield expressions and branches ----
			{
				Code: `function* gen(arr) {
  for (const x of arr) {
    if (x > 0) yield x;
    else if (x < 0) yield -x;
  }
}`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Generator function 'gen' has a complexity of 4. Maximum allowed is 2.",
					},
				},
			},

			// ---- Async generator ----
			{
				Code: `async function* asyncGen(stream) {
  for await (const chunk of stream) {
    if (chunk.error) throw chunk.error;
    if (chunk.data) yield chunk.data;
  }
}`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Async generator function 'asyncGen' has a complexity of 4. Maximum allowed is 2.",
					},
				},
			},

			// ---- Try/catch with complex matching ----
			{
				Code: `function safeParse(input) {
  try {
    return JSON.parse(input);
  } catch (e) {
    if (e instanceof SyntaxError) return null;
    if (e?.code === 'EACCES') throw e;
    return undefined;
  }
}`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'safeParse' has a complexity of 5. Maximum allowed is 2.",
					},
				},
			},

			// ---- Class with complex constructor ----
			// 1 (start) + 1 (param default) + 1 (??) + 1 (||) + 1 (??) + 2 (ifs) = 7.
			// The arrow `() => {}` is its own code path with complexity 1
			// (under max), so it doesn't trigger.
			{
				Code: `class Cache {
  constructor(opts = {}) {
    this.max = opts.max ?? 100;
    this.ttl = opts.ttl || 60000;
    this.onEvict = opts.onEvict ?? (() => {});
    if (this.max < 0) throw new Error('invalid max');
    if (this.ttl < 0) throw new Error('invalid ttl');
  }
}`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Constructor has a complexity of 7. Maximum allowed is 3.",
					},
				},
			},

			// ---- Class field with complex initializer ----
			// 1 (start) + 3 (?. each) + 2 (?? each) = 6.
			{
				Code: `class Config {
  options = {
    debug: process.env?.DEBUG === 'true',
    retries: parseInt(process.env?.RETRIES ?? '3', 10),
    timeout: parseInt(process.env?.TIMEOUT ?? '5000', 10),
  };
}`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 6. Maximum allowed is 3.",
					},
				},
			},

			// ---- Class static block with multiple branches ----
			{
				Code: `class Registry {
  static {
    if (typeof globalThis !== 'undefined') {
      const env = globalThis.process?.env ?? {};
      if (env.DEBUG === 'true' || env.VERBOSE === '1') {
        Registry.log = console.log;
      }
    }
  }
}`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class static block has a complexity of 6. Maximum allowed is 3.",
						Column:    3,
						EndColumn: 9,
					},
				},
			},

			// ---- Multiple sibling functions with branches ----
			{
				Code: `function a() { if (x) {} if (y) {} }
function b() { if (x) {} if (y) {} }
function c() { if (x) {} if (y) {} }`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "complex", Message: "Function 'a' has a complexity of 3. Maximum allowed is 2.", Line: 1},
					{MessageId: "complex", Message: "Function 'b' has a complexity of 3. Maximum allowed is 2.", Line: 2},
					{MessageId: "complex", Message: "Function 'c' has a complexity of 3. Maximum allowed is 2.", Line: 3},
				},
			},

			// ---- IIFE that wraps a complex function ----
			{
				Code:    "const v = (function () { if (a) { if (b) { return c; } } return d; })();",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function has a complexity of 3. Maximum allowed is 1.",
					},
				},
			},

			// ---- Deeply nested ternaries ----
			{
				Code:    "function classify(x) { return x > 100 ? 'huge' : x > 50 ? 'big' : x > 10 ? 'medium' : x > 0 ? 'small' : 'zero'; }",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'classify' has a complexity of 5. Maximum allowed is 3.",
					},
				},
			},

			// ---- Mixed logical operators (&&, ||, ??) ----
			{
				Code:    "function check(a, b, c, d) { return a && b || c ?? d; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'check' has a complexity of 4. Maximum allowed is 2.",
					},
				},
			},

			// ---- Compound logical assignments inside loops ----
			{
				Code: `function accumulate(items) {
  let result;
  for (const item of items) {
    result ??= item;
    result &&= transform(item);
    result ||= fallback();
  }
  return result;
}`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'accumulate' has a complexity of 5. Maximum allowed is 3.",
					},
				},
			},

			// ---- Default export anonymous function ----
			{
				Code: `export default function () {
  if (a) {
    if (b) {
      if (c) return 1;
    }
  }
  return 0;
}`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function has a complexity of 4. Maximum allowed is 2.",
					},
				},
			},
			// Default export named function — name preserved.
			{
				Code:    "export default function foo() { if (a) { if (b) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'foo' has a complexity of 3. Maximum allowed is 1.",
					},
				},
			},

			// ---- Class expression assigned to variable, with complex method ----
			{
				Code:    "var Foo = class { method() { if (a) { if (b) {} } } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'method' has a complexity of 3. Maximum allowed is 1.",
					},
				},
			},

			// ---- Anonymous arrow assigned to variable ----
			{
				Code:    "const f = (x) => x > 0 ? 1 : x < 0 ? -1 : 0;",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Arrow function has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},

			// ---- Method with TypeScript generics and return-type annotation ----
			{
				Code: `class Resolver {
  resolve<T extends object>(input: T | null | undefined): T {
    if (input == null) throw new Error('null input');
    if (Array.isArray(input)) throw new Error('arrays not supported');
    return input;
  }
}`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'resolve' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},

			// ---- Computed property name with branches in OUTER function ----
			{
				Code: `function outer() {
  return { [a || b || c]: 1 };
}`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'outer' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},

			// ---- Class with overload signatures (no body) and one implementation ----
			// Overload signatures have no body and don't start a code path; the
			// implementation does and counts normally.
			{
				Code: `class Convert {
  to(x: number): string;
  to(x: string): number;
  to(x: number | string): number | string {
    if (typeof x === 'number') return String(x);
    if (typeof x === 'string') return Number(x);
    throw new Error('unreachable');
  }
}`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'to' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},

			// ---- Logical assignment chained with optional chain on left ----
			// `a?.b ||= c` — `?.` and `||=` both count.
			// 1 (start) + 1 (?.) + 1 (||=) = 3.
			{
				Code:    "function f(a) { a?.b ||= c; }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'f' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},

			// ---- Switch with fall-through cases (still counts each case) ----
			{
				Code: `function f(x) {
  switch (x) {
    case 1:
    case 2:
    case 3: return 'low';
    case 4:
    case 5: return 'mid';
    default: return 'other';
  }
}`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'f' has a complexity of 6. Maximum allowed is 3.",
					},
				},
			},

			// ---- Array destructuring with sparse holes and defaults ----
			// `[ , a = 1, , b = 2 ]` — holes do not count, defaults do.
			{
				Code:    "function f([ , a = 1, , b = 2 ]) {}",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'f' has a complexity of 3. Maximum allowed is 2.",
					},
				},
			},

			// ---- Nested arrow inside a class field initializer's object literal ----
			// The outer field-init has 1 (start). The inner method (object method
			// shorthand) is its own code path with branches.
			{
				Code:    "class C { handler = { onClick() { if (a) { if (b) {} } } }; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'onClick' has a complexity of 3. Maximum allowed is 1.",
					},
				},
			},

			// ---- Catch without binding (TS 4.0+) ----
			{
				Code:    "function f() { try { g(); } catch { return 'err'; } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'f' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Async arrow class field — kind tokens stack: "async method 'x'" ----
			{
				Code:    "class C { x = async () => a || b; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Async method 'x' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Static private method ----
			{
				Code:    "class C { static #x() { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Static private method '#x' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- IIFE arrow — ParenthesizedExpression wrapper does not affect counter ----
			// `((a) => ...)` — the arrow's `=>` token is at col 6-7;
			// GetFunctionHeadLoc returns the `=>` range, EndColumn 8 (exclusive).
			{
				Code:    "((a) => { return a || b || c; })();",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Arrow function has a complexity of 3. Maximum allowed is 2.",
						Line:      1,
						Column:    6,
						EndLine:   1,
						EndColumn: 8,
					},
				},
			},

			// ---- UTF-16 column counting — CJK characters ----
			// Each CJK character is 1 UTF-16 code unit (within BMP), so
			// `中文 || ` shifts the function head's column by exactly 2 units
			// from the start. Identifier `验证` for the function name shifts
			// `(` to column 14 (`function 验证(` = 8 + 1 + 2 = 11 chars wide,
			// + leading whitespace from the comment block none → start column 1).
			//
			// Layout (column 1-based):
			//   `function 验证(x) {`
			//    1234567890_2345
			//             1
			// `function ` = 9 cols, `验证` = 2 cols, `(` at col 12.
			{
				Code:    "function 验证(x) { if (a || b) {} }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function '验证' has a complexity of 3. Maximum allowed is 1.",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 12,
					},
				},
			},
			// ---- UTF-16 column counting — astral / emoji surrogate pair ----
			// `🚀` (U+1F680) is outside the BMP and counts as 2 UTF-16 code
			// units (surrogate pair), matching ECMAScript string length.
			// Emoji is invalid in TS identifiers; place it inside a string
			// literal preceding the function so the function head's column
			// counts the surrogate pair correctly.
			//
			// Layout (column 1-based):
			//   `'🚀'; function f(`
			//    1__456789012345__
			//     23         1
			// `'` = 1, `🚀` = 2 surrogate units (cols 2-3), `'` at col 4,
			// `;` at 5, ` ` at 6, `function` cols 7-14, ` ` at 15, `f` at 16,
			// `(` at col 17.
			{
				Code:    "'🚀'; function f() { if (a || b) {} }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'f' has a complexity of 3. Maximum allowed is 1.",
						Line:      1,
						Column:    7,
						EndLine:   1,
						EndColumn: 17,
					},
				},
			},

			// ---- TS-only number / string property keys ----
			// Numeric key in object literal: `{ 0: function () {} }` — the
			// `function` here is classified as "method '0'" (parent is
			// PropertyAssignment with numeric name).
			{
				Code:    "var o = { 0: function () { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method '0' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			// String key with spaces.
			{
				Code:    "var o = { 'foo bar': function () { if (a) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'foo bar' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Zero / falsy default values count ----
			// `function f(x = 0) {}` — even falsy defaults trigger the
			// AssignmentPattern increment.
			{
				Code:    "function f(x = 0) {}",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'f' has a complexity of 2. Maximum allowed is 1.",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 11,
					},
				},
			},

			// ---- Renamed destructuring with default ----
			// `const { a: b = 1 } = obj` — the BindingElement has both a
			// PropertyName and Initializer. Counts once.
			{
				Code:    "function f(obj) { const { a: b = 1 } = obj; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Function 'f' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Class field with arrow having parameter default ----
			// Class C: arrow assigned as field is "method 'x'". Inside, the
			// parameter `b = 1` adds +1, so arrow complexity = 1 + 1 = 2.
			{
				Code:    "class C { x = (b = 1) => b; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Method 'x' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Static getter / setter with branches ----
			{
				Code:    "class C { static get x() { return a || b; } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Static getter 'x' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},
			{
				Code:    "class C { static set x(v) { a || b; } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Static setter 'x' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Class private getter / setter ----
			{
				Code:    "class C { get #x() { return a || b; } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Private getter '#x' has a complexity of 2. Maximum allowed is 1.",
					},
				},
			},

			// ---- Async arrow without binding name in class field ----
			{
				Code:    "class C { x = async (a) => a || b || c; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Async method 'x' has a complexity of 3. Maximum allowed is 1.",
					},
				},
			},

			// ---- Generator method ----
			{
				Code:    "class C { *gen() { if (a || b) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Generator method 'gen' has a complexity of 3. Maximum allowed is 1.",
					},
				},
			},

			// ---- Async generator method ----
			{
				Code:    "class C { async *gen() { if (a || b) {} } }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Async generator method 'gen' has a complexity of 3. Maximum allowed is 1.",
					},
				},
			},

			// ---- Class field initializer position assertion ----
			{
				Code:    "class C { x = a || b || c; }",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "complex",
						Message:   "Class field initializer has a complexity of 3. Maximum allowed is 1.",
						Line:      1,
						Column:    15,
						EndLine:   1,
						EndColumn: 26,
					},
				},
			},
		},
	)
}

// createComplexity generates a function whose cyclomatic complexity equals
// the parameter — used for the threshold-default tests (replicates the
// upstream helper of the same name in tests/lib/rules/complexity.js).
func createComplexity(complexity int) string {
	var b strings.Builder
	b.WriteString("function test (a) { if (a === 1) {")
	for i := 2; i < complexity; i++ {
		fmt.Fprintf(&b, "} else if (a === %d) {", i)
	}
	b.WriteString("} };")
	return b.String()
}
