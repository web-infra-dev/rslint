package max_nested_callbacks

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// nestFunctions returns `foo(function(){...})` nested `times` deep, matching
// the helper of the same name in eslint/tests/lib/rules/max-nested-callbacks.js.
func nestFunctions(times int) string {
	const opening = "foo(function() {"
	const closing = "});"
	return strings.Repeat(opening, times) + strings.Repeat(closing, times)
}

// TestMaxNestedCallbacks covers two corpora:
//
//  1. ESLint parity — every case from eslint/tests/lib/rules/max-nested-callbacks.js
//     migrated 1:1. Line and column are added to every invalid case (upstream
//     omits them) so subsequent refactors can't silently shift report position.
//  2. Additional edge cases — option-shape coverage, paren-walked parents,
//     tsgo-specific syntax forms (TS expressions, optional chains), and
//     lock-in tests for the asymmetric push/pop branch in the rule source.
func TestMaxNestedCallbacks(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxNestedCallbacksRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// 1. ESLint parity — mirrors tests/lib/rules/max-nested-callbacks.js
			// ============================================================
			{
				Code:    "foo(function() { bar(thing, function(data) {}); });",
				Options: 3,
			},
			{
				Code:    "var foo = function() {}; bar(function(){ baz(function() { qux(foo); }) });",
				Options: 2,
			},
			{
				Code:    "fn(function(){}, function(){}, function(){});",
				Options: 2,
			},
			{
				Code:    "fn(() => {}, function(){}, function(){});",
				Options: 2,
			},
			// 10 deep is exactly at the default threshold of 10 — valid.
			{Code: nestFunctions(10)},

			// object property options
			{
				Code:    "foo(function() { bar(thing, function(data) {}); });",
				Options: map[string]interface{}{"max": 3},
			},

			// ============================================================
			// 2. Additional edge cases
			// ============================================================

			// --- Default option (max=10) ---
			// One-deep callback with default — well below the limit.
			{
				Code: "foo(function() { console.log('hi'); });",
			},

			// --- Options shapes — exercise the JSON path explicitly ---
			// Number directly (not wrapped in array — bare CLI shape).
			{Code: "foo(function() {});", Options: 1},
			// Bare object (single-option CLI shape).
			{
				Code:    "foo(function() { bar(function() {}); });",
				Options: map[string]interface{}{"max": 2},
			},
			// Array-wrapped object (multi-element / rule_tester shape).
			{
				Code:    "foo(function() { bar(function() {}); });",
				Options: []interface{}{map[string]interface{}{"max": 2}},
			},
			// Empty options array → default 10.
			{Code: nestFunctions(10), Options: []interface{}{}},
			// Legacy `maximum` key.
			{
				Code:    "foo(function() { bar(function() { baz(function() {}); }); });",
				Options: map[string]interface{}{"maximum": 3},
			},

			// --- Function-likes that are NOT direct CallExpression children
			//     do not push, even if deeply nested ---
			// Variable declaration value — never pushes.
			{
				Code:    "var a = function() { var b = function() { var c = function() { var d = function() {}; }; }; };",
				Options: 0,
			},
			// Property value — never pushes.
			{
				Code:    "var x = { a: function() { return { b: function() { return { c: function() {} }; } }; } };",
				Options: 0,
			},
			// Array element — never pushes.
			{
				Code:    "var arr = [function() {}, function() {}];",
				Options: 0,
			},
			// `new Foo(fn)` — NewExpression, not CallExpression, no push.
			{
				Code:    "new Foo(function() { new Bar(function() {}); });",
				Options: 0,
			},
			// Tagged template — TemplateExpression child, not CallExpression.
			{
				Code:    "tag`${function() {}}`;",
				Options: 0,
			},
			// JSX expression container — not CallExpression.
			{
				Code:    "var el = <Foo cb={function() { return <Bar cb={function() {}} />; }} />;",
				Options: 0,
				Tsx:     true,
			},
			// Function declaration — listener never fires.
			{
				Code:    "function a() { function b() { function c() { function d() {} } } }",
				Options: 0,
			},

			// --- Class / object method bodies are NOT FunctionExpressions in
			//     tsgo; methods, getters, setters, constructors, static blocks
			//     don't fire the listener at all ---
			{
				Code:    "class C { m() { class D { m() { class E { m() {} } } } } }",
				Options: 0,
			},
			{
				Code:    "var o = { m() { return { m() { return { m() {} }; } }; } };",
				Options: 0,
			},
			{
				Code:    "class C { get x() { return 1; } set x(v) {} constructor() {} static {} }",
				Options: 0,
			},

			// --- Class / object-method exit pop mirrors ESLint's exit-pop bug:
			//     a class or object method sitting between two callback frames
			//     pops the outer frame on its exit, so siblings after the
			//     method see a smaller depth count. Without this, rslint would
			//     over-report compared to ESLint.
			//   `outer(function(){ class C { m(){} } bar(function(){}); })`:
			//     ESLint sees bar at depth 1 (m's exit-pop cleared outer);
			//     rslint matches via the KindMethodDeclaration exit listener.
			{
				Code:    "outer(function() { class C { m() {} } bar(function() {}); });",
				Options: 1,
			},
			// Same with object shorthand method.
			{
				Code:    "outer(function() { ({ m() {} }); bar(function() {}); });",
				Options: 1,
			},
			// Same with getter / setter / constructor.
			{
				Code:    "outer(function() { class C { get x() {} } bar(function() {}); });",
				Options: 1,
			},
			{
				Code:    "outer(function() { class C { set x(v) {} } bar(function() {}); });",
				Options: 1,
			},
			{
				Code:    "outer(function() { class C { constructor() {} } bar(function() {}); });",
				Options: 1,
			},

			// --- Paren-walked parent: `(function(){})()` IIFE callee position ---
			// Single paren wrap.
			{
				Code:    "(function() {})();",
				Options: 1,
			},
			// Multiple paren wraps.
			{
				Code:    "(((function() {})))();",
				Options: 1,
			},
			// Paren-wrapped argument.
			{
				Code:    "foo((function() {}));",
				Options: 1,
			},

			// --- Optional call chains push like normal calls ---
			{
				Code:    "foo?.(function() {});",
				Options: 1,
			},
			{
				Code:    "obj?.method(function() {});",
				Options: 1,
			},

			// --- TS expression wrappers around the function do NOT push ---
			// `as` wrapper around the callback argument — function's parent is
			// AsExpression (not CallExpression), so no push, even though the
			// outer call IS a CallExpression. Mirrors ESLint+TS-parser.
			{
				Code:    "foo(function() {} as any);",
				Options: 0,
			},
			// `satisfies` wrapper — same.
			{
				Code:    "foo(function() {} satisfies any);",
				Options: 0,
			},
			// Type assertion `<any>fn` — same.
			{
				Code:    "foo(<any>function() {});",
				Options: 0,
			},

			// --- TS wrappers on the *callee* don't change push behavior ---
			// Function in arguments still pushes.
			{
				Code:    "(foo as any)(function() {});",
				Options: 1,
			},
			{
				Code:    "foo!(function() {});",
				Options: 1,
			},
			{
				Code:    "foo<T>(function() {});",
				Options: 1,
			},

			// --- async / generator / async generator function expressions ---
			{
				Code:    "foo(async function() {});",
				Options: 1,
			},
			{
				Code:    "foo(function*() {});",
				Options: 1,
			},
			{
				Code:    "foo(async function*() {});",
				Options: 1,
			},
			{
				Code:    "foo(async () => {});",
				Options: 1,
			},

			// --- Mixed argument list: only function-likes count, others ignored ---
			{
				Code:    "foo(1, 'two', null, function() {});",
				Options: 1,
			},

			// --- Spread argument with function ---
			// `foo(...args, function(){})` — function is a regular CallExpression
			// argument, parent=CallExpression, pushed.
			{
				Code:    "foo(...args, function() {});",
				Options: 1,
			},
			// SpreadElement with function inside — `foo(...[function(){}])` — the
			// function is inside an ArrayLiteralExpression spread, parent=Array,
			// no push.
			{
				Code:    "foo(...[function() {}, function() {}]);",
				Options: 0,
			},

			// --- Default parameter with function-like value — never pushes ---
			{
				Code:    "function f(cb = function() {}) {}",
				Options: 0,
			},
			{
				Code:    "function f(cb = () => {}) {}",
				Options: 0,
			},

			// --- Class field arrow / default member ---
			{
				Code:    "class C { handler = function() {}; arrow = () => {}; }",
				Options: 0,
			},

			// --- Conditional expression branches — neither pushes ---
			{
				Code:    "var x = cond ? function() {} : function() {};",
				Options: 0,
			},

			// --- Logical / comma expression branches — neither pushes ---
			{
				Code:    "var x = a || function() {};",
				Options: 0,
			},
			{
				Code:    "var x = (1, function() {});",
				Options: 0,
			},

			// --- Decorator with function-like argument: parent IS CallExpression ---
			// `@dec(function(){})` — decorator wraps CallExpression `dec(function(){})`.
			// Function's parent = CallExpression. Pushed.
			{
				Code:    "function dec(cb) { return () => {}; } @dec(function() {}) class C {}",
				Options: 1,
			},

			// --- Chained method calls — each call's argument pushes & pops in sequence ---
			{
				Code:    "promise.then(function(a){}).catch(function(b){}).finally(function(c){});",
				Options: 1,
			},

			// --- super() call with function arg (inside a constructor) ---
			{
				Code:    "class B { constructor(cb) {} } class D extends B { constructor() { super(function() {}); } }",
				Options: 1,
			},

			// --- TaggedTemplateExpression — function in template substitution
			//     is not a CallExpression child, no push ---
			{
				Code:    "tag`x ${function() {}} y`",
				Options: 0,
			},

			// --- Nested calls but with intermediate non-callback nesting ---
			// `foo(function(){})` then unrelated `bar(function(){})` at same level.
			{
				Code:    "foo(function() {}); bar(function() {}); baz(function() {});",
				Options: 1,
			},

			// --- Empty / degenerate forms ---
			{Code: ""},
			{Code: "foo();"},
			{Code: ";"},

			// --- Real-world: shallow callback chain (typical promise then) ---
			{
				Code: `fetch('/api')
  .then(function(r) { return r.json(); })
  .then(function(d) { console.log(d); });`,
				Options: 2,
			},

			// ============================================================
			// 3. Container/container nesting — function-likes as containers
			//    around CallExpressions (every depth-bearing JS form)
			// ============================================================

			// --- Callback inside if / else / else if branches ---
			{
				Code:    "foo(function() { if (a) { bar(function() {}); } else { baz(function() {}); } });",
				Options: 2,
			},
			{
				Code:    "foo(function() { if (a) {} else if (b) { bar(function() {}); } });",
				Options: 2,
			},

			// --- Callback inside switch case / default ---
			{
				Code:    "foo(function() { switch (k) { case 1: bar(function() {}); break; default: baz(function() {}); } });",
				Options: 2,
			},

			// --- Callback inside try / catch / finally ---
			{
				Code:    "foo(function() { try { bar(function() {}); } catch (e) { baz(function() {}); } finally { qux(function() {}); } });",
				Options: 2,
			},

			// --- Callback inside while / do-while / for / for-in / for-of / for-await ---
			{
				Code:    "foo(function() { while (a) { bar(function() {}); } });",
				Options: 2,
			},
			{
				Code:    "foo(function() { do { bar(function() {}); } while (a); });",
				Options: 2,
			},
			{
				Code:    "foo(function() { for (let i = 0; i < 10; i++) { bar(function() {}); } });",
				Options: 2,
			},
			{
				Code:    "foo(function() { for (const k in o) { bar(function() {}); } });",
				Options: 2,
			},
			{
				Code:    "foo(function() { for (const v of a) { bar(function() {}); } });",
				Options: 2,
			},
			{
				Code:    "foo(async function() { for await (const v of a) { bar(function() {}); } });",
				Options: 2,
			},

			// --- Callback after `await` / `yield` in async / generator ---
			{
				Code:    "foo(async function() { await bar(function() {}); });",
				Options: 2,
			},
			{
				Code:    "foo(function*() { yield bar(function() {}); });",
				Options: 2,
			},

			// --- Arrow with concise-body invocation (no block) ---
			// `() => fn(function(){})` — arrow body IS the call expression.
			{
				Code:    "var f = () => bar(function() {});",
				Options: 2,
			},
			// Arrow returning arrow returning callback — each arrow pushes.
			{
				Code:    "foo(() => () => bar(function() {}));",
				Options: 3,
			},

			// --- Function in array spread — element of ArrayLiteralExpression
			//     (parent ≠ CallExpression even after walking spread).
			{
				Code:    "foo(...[1, function() {}, 2]);",
				Options: 0,
			},

			// --- Object/array destructuring binding with function default ---
			{
				Code:    "function f({ cb = function() {} } = {}) {}",
				Options: 0,
			},
			{
				Code:    "function f([cb = function() {}]) {}",
				Options: 0,
			},

			// --- Class with computed / private / static method names — none push ---
			{
				Code:    "class C { ['m']() {} #priv() {} static st() {} }",
				Options: 0,
			},
			// Async / async-generator class methods.
			{
				Code:    "class C { async m() {} async *g() {} }",
				Options: 0,
			},
			// Setter with rest pattern (TS-only) and complex parameter form.
			{
				Code:    "class C { set v(value) {} get v() { return 1; } }",
				Options: 0,
			},

			// --- Class extends with super call passing callback ---
			{
				Code:    "class B { constructor(cb) {} } class D extends B { constructor() { super(function() {}); } }",
				Options: 1,
			},

			// --- yield* delegation with callback arg ---
			{
				Code:    "foo(function*() { yield* bar(function() {}); });",
				Options: 2,
			},

			// --- await import() then ---
			{
				Code: `async function load() {
  const m = await import('mod');
  return m.run(function() {});
}`,
				Options: 1,
			},

			// --- Dynamic import with then chain ---
			{
				Code:    "import('mod').then(function(m) { m.run(); });",
				Options: 1,
			},

			// --- Multi-line nested IIFE pattern ---
			{
				Code: `(function() {
  (function() {
    (function() {})();
  })();
})();`,
				Options: 3,
			},

			// --- Sibling chains separated by class declaration with method ---
			// Validates the method-exit pop fully clears between siblings.
			{
				Code: `outer(function() {
  class A { m() {} }
  class B { m() {} }
  bar(function() {});
});`,
				Options: 1,
			},

			// --- Real-world: jQuery-like AJAX shape ---
			{
				Code: `$(function() {
  $.ajax({
    success: function(data) { handle(data); },
  });
});`,
				Options: 2,
			},

			// --- Real-world: Express middleware chain ---
			{
				Code: `app.get('/x', function(req, res, next) {
  next();
});`,
				Options: 1,
			},

			// --- Real-world: Node.js style fs.readFile ---
			{
				Code: `fs.readFile('a', function(err, data) {
  if (err) return cb(err);
  cb(null, data);
});`,
				Options: 1,
			},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// 1. ESLint parity — mirrors tests/lib/rules/max-nested-callbacks.js
			// ============================================================
			{
				Code:    "foo(function() { bar(thing, function(data) { baz(function() {}); }); });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    50,
					},
				},
			},
			{
				Code:    "foo(function() { bar(thing, (data) => { baz(function() {}); }); });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    45,
					},
				},
			},
			{
				Code:    "foo(() => { bar(thing, (data) => { baz( () => {}); }); });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    41,
					},
				},
			},
			{
				Code:    "foo(function() { if (isTrue) { bar(function(data) { baz(function() {}); }); } });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    57,
					},
				},
			},
			{
				Code: nestFunctions(11),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (11). Maximum allowed is 10.",
						Line:      1,
						Column:    165,
					},
				},
			},
			{
				Code:    nestFunctions(11),
				Options: map[string]interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (11). Maximum allowed is 10.",
						Line:      1,
						Column:    165,
					},
				},
			},
			{
				Code:    "foo(function() {})",
				Options: map[string]interface{}{"max": 0},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (1). Maximum allowed is 0.",
						Line:      1,
						Column:    5,
					},
				},
			},

			// object property options
			{
				Code:    "foo(function() { bar(thing, function(data) { baz(function() {}); }); });",
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    50,
					},
				},
			},

			// ============================================================
			// 2. Additional edge cases
			// ============================================================

			// --- Multi-line code: line/column precision ---
			{
				Code: `foo(function() {
    bar(function() {
        baz(function() {
            qux(function() {});
        });
    });
});`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      3,
						Column:    13,
					},
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (4). Maximum allowed is 2.",
						Line:      4,
						Column:    17,
					},
				},
			},

			// --- IIFE callee + nested callbacks ---
			{
				Code:    "(function() { foo(function() { bar(function() {}); }); })();",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    36,
					},
				},
			},
			// IIFE wrapped in extra parens — paren walking lets the outer FE push.
			{
				Code:    "((function() { foo(function() { bar(function() {}); }); }))();",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    37,
					},
				},
			},
			// Paren-wrapped argument — function's parent is ParenthesizedExpression,
			// but paren walking treats CallExpression as the effective parent.
			{
				Code:    "foo((function() { bar((function() { baz((function() {})); })); }));",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    42,
					},
				},
			},

			// --- Mixed FunctionExpression / ArrowFunction nesting ---
			{
				Code:    "foo(() => bar(function() { baz(() => {}); }));",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    32,
					},
				},
			},

			// --- async / generator function-likes still push ---
			{
				Code:    "foo(async function() { bar(async function() { baz(async function() {}); }); });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    51,
					},
				},
			},
			{
				Code:    "foo(function*() { bar(function*() { baz(function*() {}); }); });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    41,
					},
				},
			},
			{
				Code:    "foo(async () => { bar(async () => { baz(async () => {}); }); });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    41,
					},
				},
			},

			// --- Optional chain calls push the same as plain calls ---
			{
				Code:    "foo?.(function() { bar?.(function() { baz?.(function() {}); }); });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    45,
					},
				},
			},

			// --- max=0 reports every direct callback ---
			// All three are direct CallExpression children → all pushed and
			// reported (length goes 1, 1, 1 each time, all > 0).
			{
				Code:    "foo(function() {}); bar(function() {}); baz(function() {});",
				Options: 0,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (1). Maximum allowed is 0.", Line: 1, Column: 5},
					{MessageId: "exceed", Message: "Too many nested callbacks (1). Maximum allowed is 0.", Line: 1, Column: 25},
					{MessageId: "exceed", Message: "Too many nested callbacks (1). Maximum allowed is 0.", Line: 1, Column: 45},
				},
			},

			// --- Sibling callback chains under a shared outer callback ---
			{
				Code:    "foo(function() { bar(function() { baz(function() {}); }); qux(function() { quux(function() {}); }); });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    39,
					},
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    81,
					},
				},
			},

			// --- Legacy `maximum` key honored ---
			{
				Code:    "foo(function() { bar(function() { baz(function() {}); }); });",
				Options: map[string]interface{}{"maximum": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    39,
					},
				},
			},
			// `maximum` wins when both keys are present and `maximum` is truthy
			// (matches ESLint's `option.maximum || option.max`).
			{
				Code:    "foo(function() { bar(function() { baz(function() {}); }); });",
				Options: map[string]interface{}{"maximum": 2, "max": 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    39,
					},
				},
			},
			// `{ maximum: 0 }` short-circuits to undefined in ESLint (disabling
			// the check). Lock in the falsy-but-valid path with `{ max: 0 }`
			// (which DOES report).
			{
				Code:    "foo(function() {});",
				Options: []interface{}{map[string]interface{}{"max": 0}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (1). Maximum allowed is 0.",
						Line:      1,
						Column:    5,
					},
				},
			},

			// --- Lock-in: asymmetric exit pop. A non-pushed FunctionExpression
			//     (var-declared, not a callback) still pops the stack on exit,
			//     mirroring ESLint's `popStack` ALWAYS calling Array#pop. The
			//     practical effect: code AFTER `var x = function(){}` inside a
			//     callback sees a stack length one less than its true depth. ---
			// Without the asymmetric pop, the inner `bar(...)` would see depth=2
			// (outer + bar). With the pop, it sees depth=1 (outer was popped by
			// `var x`'s exit). With max=1, no report. We assert this by using
			// max=0 instead — bar's callback still reports as length=1.
			{
				Code:    "foo(function() { var x = function() {}; bar(function() {}); });",
				Options: 0,
				Errors: []rule_tester.InvalidTestCaseError{
					// Outer callback is direct call argument → pushed → reported.
					{MessageId: "exceed", Message: "Too many nested callbacks (1). Maximum allowed is 0.", Line: 1, Column: 5},
					// `var x = function(){}` doesn't push. But ESLint also runs
					// the length check after the (skipped) push: stack=[outer]
					// length=1 > 0 → REPORT.
					{MessageId: "exceed", Message: "Too many nested callbacks (1). Maximum allowed is 0.", Line: 1, Column: 26},
					// `var x`'s exit pops → stack=[]. `bar(function(){})` pushes.
					// Length=1 > 0 → REPORT.
					{MessageId: "exceed", Message: "Too many nested callbacks (1). Maximum allowed is 0.", Line: 1, Column: 45},
				},
			},

			// --- Function-likes inside conditional / loop bodies still nest ---
			{
				Code:    "foo(function() { while (true) { bar(function() { for (;;) { baz(function() {}); } }); } });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    65,
					},
				},
			},

			// --- TS-only: function-likes inside CallExpression with type args ---
			{
				Code:    "foo<T>(function() { bar<U>(function() { baz<V>(function() {}); }); });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (3). Maximum allowed is 2.",
						Line:      1,
						Column:    48,
					},
				},
			},

			// --- Real-world: callback hell (the canonical anti-pattern) ---
			{
				Code: `getData('a', function (a) {
  getData(a, function (b) {
    getData(b, function (c) {
      getData(c, function (d) {
        process(d);
      });
    });
  });
});`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Too many nested callbacks (4). Maximum allowed is 3.",
						Line:      4,
						Column:    18,
					},
				},
			},

			// --- Lock-in: callbacks INSIDE a class method body still see the
			//     outer callback frame, because the method's exit-pop runs only
			//     after we've finished traversing its body. ESLint and rslint
			//     agree: bar inside m sees depth=2, baz inside bar sees depth=3.
			{
				Code:    "outer(function() { class C { m() { bar(function() { baz(function() {}); }); } } });",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (2). Maximum allowed is 1.", Line: 1, Column: 40},
					{MessageId: "exceed", Message: "Too many nested callbacks (3). Maximum allowed is 1.", Line: 1, Column: 57},
				},
			},

			// --- Chained method calls combined with nesting reports correctly ---
			{
				Code: `promise
  .then(function(a) {
    return a.next(function(b) {
      return b.next(function(c) {});
    });
  });`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (3). Maximum allowed is 2.", Line: 4, Column: 21},
				},
			},

			// ============================================================
			// 4. Container nesting — callbacks deep inside control flow
			// ============================================================

			// --- Inside if branch ---
			{
				Code:    "foo(function() { if (a) { bar(function() { baz(function() {}); }); } });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (3). Maximum allowed is 2.", Line: 1, Column: 48},
				},
			},
			// --- Inside for-of body ---
			{
				Code:    "foo(function() { for (const v of a) { bar(function() { baz(function() {}); }); } });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (3). Maximum allowed is 2.", Line: 1, Column: 60},
				},
			},
			// --- Inside try / catch ---
			{
				Code:    "foo(function() { try { bar(function() { baz(function() {}); }); } catch (e) {} });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (3). Maximum allowed is 2.", Line: 1, Column: 45},
				},
			},

			// --- await on a call whose argument is the deeper callback ---
			{
				Code:    "foo(async function() { await bar(async function() { await baz(async function() {}); }); });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (3). Maximum allowed is 2.", Line: 1, Column: 63},
				},
			},

			// --- Concise-body arrow chain reaches threshold ---
			{
				Code:    "foo(() => bar(() => baz(() => qux(() => {}))));",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (4). Maximum allowed is 3.", Line: 1, Column: 35},
				},
			},

			// ============================================================
			// 5. Real-world patterns — anti-patterns we'd actually flag
			// ============================================================

			// --- Express middleware with deeply nested callbacks ---
			{
				Code: `app.get('/x', function(req, res) {
  db.connect(function(err, client) {
    client.query('SELECT', function(err, rows) {
      rows.forEach(function(row) {
        process(row);
      });
    });
  });
});`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (4). Maximum allowed is 3.", Line: 4, Column: 20},
				},
			},

			// --- Node.js fs callback hell ---
			{
				Code: `fs.readFile('a', function(err, a) {
  fs.readFile('b', function(err, b) {
    fs.readFile('c', function(err, c) {
      fs.writeFile('out', a + b + c, function(err) {});
    });
  });
});`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (3). Maximum allowed is 2.", Line: 3, Column: 22},
					{MessageId: "exceed", Message: "Too many nested callbacks (4). Maximum allowed is 2.", Line: 4, Column: 38},
				},
			},

			// --- jQuery DOM-ready + AJAX + iteration. ESLint counts only
			//     direct CallExpression-children: `$(fn)`, `forEach(fn)`,
			//     `each(fn)` — three callbacks. The `success: function(){}`
			//     inside the `$.ajax({...})` object literal is a Property
			//     value, NOT a CallExpression child, so it does not push. ---
			{
				Code: `$(function() {
  $.ajax({
    success: function(data) {
      data.items.forEach(function(item) {
        $(item).each(function() {});
      });
    },
  });
});`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (3). Maximum allowed is 2.", Line: 5, Column: 22},
				},
			},

			// --- setTimeout inside setTimeout (timer pyramid) ---
			{
				Code: `setTimeout(function() {
  setTimeout(function() {
    setTimeout(function() {
      setTimeout(function() {});
    }, 10);
  }, 10);
}, 10);`,
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (4). Maximum allowed is 3.", Line: 4, Column: 18},
				},
			},

			// ============================================================
			// 6. Pure-syntax boundaries — paren / type-wrapper interactions
			// ============================================================

			// --- Optional call combined with paren-wrapped function arg ---
			{
				Code:    "foo?.(function() { bar?.((function() { baz?.((function() {})); })); });",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (3). Maximum allowed is 2.", Line: 1, Column: 47},
				},
			},

			// --- Mixed paren wraps in IIFE callee + nested ---
			{
				Code:    "(((function() { (((function() { (((function() {})))(); })))(); })))();",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: "Too many nested callbacks (3). Maximum allowed is 2.", Line: 1, Column: 36},
				},
			},

			// --- Function inside object spread (not a CallExpression child)
			//     yet OUTER call still pushes because the function in args
			//     directly is. Tests that we don't mistakenly attribute push
			//     to the inner one. ---
			{
				Code:    "foo({...x, cb: function(){}}, function(){});",
				Options: 0,
				Errors: []rule_tester.InvalidTestCaseError{
					// Only the trailing function is a direct CallExpression argument.
					{MessageId: "exceed", Message: "Too many nested callbacks (1). Maximum allowed is 0.", Line: 1, Column: 31},
				},
			},
		},
	)
}
