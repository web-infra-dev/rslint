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

			// Functions closing over unmodified `let` / `const` / `using` are OK.
			{Code: `for (let i=0; i<l; i++) { (function() { i; }) }`},
			{Code: `for (let i in {}) { i = 7; (function() { i; }) }`},
			{Code: `for (const i of {}) { (function() { i; }) }`},
			{Code: `for (using i of foo) { (function() { i; }) }`},
			{Code: `for (await using i of foo) { (function() { i; }) }`},
			{Code: `for (var i = 0; i < 10; ++i) { using foo = bar(i); (function() { foo; }) }`},
			{Code: `for (var i = 0; i < 10; ++i) { await using foo = bar(i); (function() { foo; }) }`},

			// Functions that never reference the enclosing loop variable.
			{Code: `for (let i = 0; i < 10; ++i) { for (let x in xs.filter(x => x != i)) {  } }`},
			{Code: `let a = 0; for (let i=0; i<l; i++) { (function() { a; }); }`},
			{Code: `let a = 0; for (let i in {}) { (function() { a; }); }`},
			{Code: `let a = 0; for (let i of {}) { (function() { a; }); }`},
			{Code: `let a = 0; for (let i=0; i<l; i++) { (function() { (function() { a; }); }); }`},
			{Code: `let a = 0; for (let i in {}) { function foo() { (function() { a; }); } }`},
			{Code: `let a = 0; for (let i of {}) { (() => { (function() { a; }); }); }`},
			{Code: `var a = 0; for (let i=0; i<l; i++) { (function() { a; }); }`},
			{Code: `var a = 0; for (let i in {}) { (function() { a; }); }`},
			{Code: `var a = 0; for (let i of {}) { (function() { a; }); }`},

			// Closure over outer const — safe.
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

			// Variables declared outside the loop and not updated in or after it.
			{Code: `var i; while(i) { (function() { i; }) }`},
			{Code: `var i; do { (function() { i; }) } while (i)`},

			// Undeclared references — handled by no-undef, not here.
			{Code: `for (var i=0; i<l; i++) { (function() { undeclared; }) }`},
			{Code: `for (let i=0; i<l; i++) { (function() { undeclared; }) }`},
			{Code: `for (var i in {}) { i = 7; (function() { undeclared; }) }`},
			{Code: `for (let i in {}) { i = 7; (function() { undeclared; }) }`},
			{Code: `for (const i of {}) { (function() { undeclared; }) }`},
			{Code: `for (let i = 0; i < 10; ++i) { for (let x in xs.filter(x => x != undeclared)) {  } }`},

			// IIFE — immediately invoked, not saved off, so closure semantics don't matter.
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

			// Outer binding is const — function is safe even though reassigned later (runtime error).
			{Code: `
const foo = bar;

for (var i = 0; i < 5; i++) {
    arr.push(() => foo);
}

foo = baz;`},
			{Code: `
using foo = bar;

for (var i = 0; i < 5; i++) {
    arr.push(() => foo);
}

foo = baz;`},
			{Code: `
await using foo = bar;

for (var i = 0; i < 5; i++) {
    arr.push(() => foo);
}

foo = baz;`},

			// TypeScript: plain type references inside closures never count as value references.
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
			// Unconfigured global type inside a closure: arrow only references the type, not the loop var.
			{Code: `
for (var i = 0; i < 10; i++) {
  const process = (item: UnconfiguredGlobalType) => {
    return item.id;
  };
}`},
			{Code: `
for (var i = 0; i < 10; i++) {
  const process = (configItem: ConfiguredType, unconfigItem: UnconfiguredType) => {
    return {
      config: configItem.value,
      unconfig: unconfigItem.value
    };
  };
}`},

			// ---- Destructuring: iteration binds to fresh const — safe ----
			{Code: `for (const {a} of arr) { (function() { a; }) }`},
			{Code: `for (const [a] of arr) { (function() { a; }) }`},
			{Code: `for (const [[a]] of arr) { (function() { a; }) }`},
			{Code: `for (const {a: {b}} of arr) { (function() { b; }) }`},
			{Code: `for (const {a = 10} of arr) { (function() { a; }) }`},

			// ---- Method-bound this / labeled statements — no through refs matter ----
			{Code: `outer: for (var i = 0; i < l; i++) { (function() { break outer; }); }`},

			// ---- Nested loops where outer-loop-declared let is read but not modified ----
			{Code: `for (let i = 0; i < l; i++) { for (let j = 0; j < i; j++) { (function() { j; }) } }`},

			// ---- Methods/getters inside class/object boundary shields inner closure ----
			// A plain FunctionExpression inside a method inside a loop is "inside the method",
			// not inside the loop — methods/accessors create their own scope boundary, just
			// like plain FunctionExpressions. Matches ESLint (where class methods map to
			// FunctionExpression in ESTree).
			{Code: `for (var i = 0; i < l; i++) { const o = { m() { var f = function() { x; }; } }; }`},
			{Code: `for (var i = 0; i < l; i++) { class C { m() { var f = function() { x; }; } } }`},

			// ---- Methods referencing only local state or const outer bindings — safe ----
			{Code: `const k = 10; for (var i = 0; i < l; i++) { class C { m() { return k; } } }`},

			// ---- Computed property key captures loop var but the function body doesn't ----
			// The key `[i]` evaluates in the outer scope each iteration; the function
			// body has no through refs, so no report.
			{Code: `for (var i = 0; i < l; i++) { var o = { [i]: function() {} }; }`},

			// ---- ClassExpression name as through ref of a method inside the class ----
			// `Foo` lives in the class's name scope; it has no writes, so referencing
			// it from a method is safe on its own — only a write to a through ref
			// would flag.
			{Code: `for (var i = 0; i < l; i++) { const C = class Foo { m() { return Foo; } }; }`},

			// ---- for-await-of with fresh const binding — safe ----
			{Code: `async function f(xs) { for await (const x of xs) { (function() { x; }) } }`},

			// ---- Parameter default value that references a let declared inside the loop ----
			// `j` is fresh per iteration, so capturing it via a default is safe.
			{Code: `for (let i = 0; i < l; i++) { let j = i; (function(x = j) { x; }); }`},

			// ---- ForStatement with no init: loop var declared outside, no modification in/after ----
			{Code: `var i; for (; i < l; ) { (function() { i; }) }`},

			// ---- Import binding is read-only — safe through ref ----
			{Code: `import { foo } from "./mod"; for (var i = 0; i < l; i++) { (function() { foo; }) }`},

			// ---- Enum reference (TS value+type) is a const-like binding ----
			{Code: `enum Color { Red } for (var i = 0; i < l; i++) { (function() { Color.Red; }) }`},

			// ---- Through refs collected from nested functions inside a non-loop
			// outer function (the inner function itself IS the closure at risk) ----
			{Code: `function outer() { for (var i = 0; i < l; i++) { let j = i; (function() { j; }); } }`},

			// ---- Skip-IIFE propagates: inner FE assigned to a var reads loop var
			// only through an outer IIFE; only the inner FE should be flagged, not the outer ----
			// (covered by the "outer IIFE skipped, inner FE reported" invariant)
			{Code: `for (var i = 0; i < l; i++) { (function() { /* no through refs */ var local = 1; local; })(); }`},
		},
		[]rule_tester.InvalidTestCase{
			// Each of these four asserts a complete position range so future
			// refactors can't silently shift the report site across the three
			// kinds of function-like containers we care about (FE, arrow, FD)
			// and the two body-containment shapes (direct, via VariableDecl).
			{
				Code: `for (var i=0; i<l; i++) { (function() { i; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
					Line:      1,
					Column:    28,
					EndLine:   1,
					EndColumn: 45,
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

			// Closure in a nested loop captures the outer `var`.
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

			// Write inside the loop occurs before or after the closure declaration.
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

			// Generator / async IIFE — still checked.
			{
				Code: `let a; for (var i=0; i<l; i++) { (function* (){i;})() }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `let a; for (var i=0; i<l; i++) { (async function (){i;})() }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// Self-referencing named IIFE — skip-IIFE optimization doesn't apply.
			{
				Code: `
let current = getStart();
const arr = [];
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
			{
				Code: `
var arr = [];

for (var i = 0; i < 5; i++) {
    (function fun () {
        if (arr.includes(fun)) return i;
        else arr.push(fun);
    })();
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// Async arrow IIFE — the async flag means we don't skip it.
			{
				Code: `
let current = getStart();
const arr = [];
while (current) {
    const p = (async () => {
        await someDelay();
        current;
    })();

    arr.push(p);
    current = current.upper;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// Nested arrow produced by an IIFE leaks the loop variable.
			{
				Code: `
var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((f => f)(
        () => i
    ));
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      6,
				}},
			},
			{
				Code: `
var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((() => {
        return () => i;
    })());
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      6,
				}},
			},
			{
				Code: `
var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((() => {
        return () => { return i };
    })());
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      6,
				}},
			},
			{
				Code: `
var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((() => {
        return () => {
            return () => i
        };
    })());
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      6,
				}},
			},
			{
				Code: `
var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((() => {
        return () =>
            (() => i)();
    })());
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      6,
				}},
			},
			{
				Code: `
var arr = [];

for (var i = 0; i < 5; i ++) {
    (() => {
        arr.push((async () => {
            await 1;
            return i;
        })());
    })();
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      6,
				}},
			},
			{
				Code: `
var arr = [];

for (var i = 0; i < 5; i ++) {
    (() => {
        (function f() {
            if (!arr.includes(f)) {
                arr.push(f);
            }
            return i;
        })();
    })();

}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `
var arr1 = [], arr2 = [];

for (var [i, j] of ["a", "b", "c"].entries()) {
    (() => {
        arr1.push((() => i)());
        arr2.push(() => j);
    })();
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'j'.",
					Line:      7,
				}},
			},
			{
				Code: `
var arr = [];

for (var i = 0; i < 5; i ++) {
    ((f) => {
        arr.push(f);
    })(() => {
        return (() => i)();
    });

}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      7,
				}},
			},
			{
				Code: `
for (var i = 0; i < 5; i++) {
    (async () => {
        () => i;
    })();
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      3,
				}},
			},

			// Closure declared after other loop-body statements captures the loop var.
			{
				Code: `
for (var i = 0; i < 10; i++) {
    items.push({
        id: i,
        name: "Item " + i
    });

    const process = function (callback){
        callback({ id: i, name: "Item " + i });
    };
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      8,
				}},
			},

			// TypeScript: loop var is captured by a function whose only explicit
			// annotations are types; `i` still leaks through.
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
			{
				Code: `
interface Item {
  id: number;
  name: string;
}

const items: Item[] = [];
for (var i = 0; i < 10; i++) {
  items.push({
    id: i,
    name: "Item " + i
  });

  const process = function(callback: (item: Item) => void): void {
    callback({ id: i, name: "Item " + i });
  };
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `
type Processor<T> = (item: T) => void;

for (var i = 0; i < 10; i++) {
  const processor: Processor<number> = (item) => {
    return item + i;
  };
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `
for (var i = 0; i < 10; i++) {
  const process = (item: UnconfiguredGlobalType) => {
    console.log(i, item.value);
  };
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- Destructuring bindings in for-in/of iterate (write each step) ----
			{
				Code: `for (var {a} of arr) { (function() { a; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'a'.",
				}},
			},
			{
				Code: `for (var [[a]] of arr) { (function() { a; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `for (var {a = 10} of arr) { (function() { a; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- Destructuring assignment in the body counts as a write ----
			{
				Code: `var a; for (var i = 0; i < l; i++) { [a] = [i]; (function() { a; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `var a; for (var i = 0; i < l; i++) { ({a} = {a: i}); (function() { a; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- Async function declaration (not expression) inside a loop ----
			{
				Code: `for (var i = 0; i < l; i++) { async function f() { i; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- Nested functions inherit the outer's through references ----
			{
				Code: `for (var i = 0; i < l; i++) { function foo() { function bar() { i; } bar(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- Class/object methods inside a loop that reference loop vars ----
			// ESLint's ESTree treats these as FunctionExpression values so it reports
			// them; we register explicit listeners for Method/Accessor/Constructor.
			{
				Code: `for (var i = 0; i < l; i++) { const o = { m() { return i; } }; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      1,
					Column:    43,
				}},
			},
			{
				Code: `for (var i = 0; i < l; i++) { class C { m() { return i; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      1,
					Column:    41,
				}},
			},
			{
				Code: `for (var i = 0; i < l; i++) { class C { get x() { return i; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `for (var i = 0; i < l; i++) { class C { set x(v) { this.a = i + v; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `for (var i = 0; i < l; i++) { class C { constructor() { this.a = i; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `for (var i = 0; i < l; i++) { const o = { async m() { return i; } }; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},
			{
				Code: `for (var i = 0; i < l; i++) { const o = { *m() { yield i; } }; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- Functions inside a static block inside a loop — static block is transparent ----
			{
				Code: `
for (var i = 0; i < l; i++) {
    class C {
        static {
            var f = function() { i; };
        }
    }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Line:      5,
				}},
			},

			// ---- Parameter default value captures loop var (the function still
			// closes over `i` across iterations) ----
			{
				Code: `for (var i = 0; i < l; i++) { (function(x = i) { x; }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},

			// ---- Computed key + function value both reference the loop var ----
			{
				Code: `for (var i = 0; i < l; i++) { var o = { [i]: function() { return i; } }; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- ClassExpression name that IS reassigned inside the class — through ref unsafe ----
			// (Not truly possible via `Foo = ...` at runtime — the class name is read-only
			// within the class body — but a reassignment of an outer same-named binding is.
			// The interesting shape is: the method references an OUTER var that's modified.)
			{
				Code: `var Foo; for (var i = 0; i < l; i++) { Foo = i; const C = class { m() { return Foo; } }; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'Foo'.",
				}},
			},

			// ---- for-await-of with `var` iterator — iteration writes the same binding ----
			{
				Code: `async function f(xs) { for await (var x of xs) { (function() { x; }) } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
				}},
			},

			// ---- Through refs leak up through nested functions: outer FE is
			// flagged when a nested FE references a loop-modified outer-scope var ----
			{
				Code: `for (var i = 0; i < l; i++) { function outer() { (function() { i; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					// Only the outer FunctionDeclaration is in the loop; the
					// inner FE is inside a non-loop function, so only the
					// outer one reports.
				}},
			},

			// ---- Reassigned FunctionDeclaration name captured by a non-IIFE function ----
			// - `function foo()` itself references `i` (unsafe).
			// - `var g = function() { foo }` references `foo`, which is reassigned
			//   inside the loop (`foo = null`), so `foo` is unsafe too.
			{
				Code: `for (var i = 0; i < l; i++) { function foo() { return i; } foo = null; var g = function() { foo; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeRefs"},
					{MessageId: "unsafeRefs"},
				},
			},

			// ---- FunctionDeclaration used directly in loop body (its own through refs) ----
			{
				Code: `for (var i = 0; i < l; i++) { function foo() { return i; } foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'i'.",
				}},
			},

			// ---- Multiple distinct unsafe refs appear in stable source order ----
			{
				Code: `var a, b; for (var i = 0; i < l; i++) { a = i; b = i; (function() { a + b; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeRefs",
					Message:   "Function declared in a loop contains unsafe references to variable(s) 'a', 'b'.",
				}},
			},
		},
	)
}
