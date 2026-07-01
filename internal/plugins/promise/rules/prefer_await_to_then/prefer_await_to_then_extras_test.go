package prefer_await_to_then_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/prefer_await_to_then"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferAwaitToThenExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_await_to_then.PreferAwaitToThenRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: AST node types ----

			// Element access with a string/template literal key is not an identifier
			// property, so it stays unmatched (mirrors upstream's node.property.name check).
			{Code: `function f() { obj['then']() }`},
			{Code: "function f() { obj[`then`]() }"},
			// Element access with an identifier key whose name isn't then/catch/finally.
			{Code: `function f() { obj[sym]() }`},

			// Member access without call (typeof check) — not a CallExpression
			{Code: `function f() { return typeof obj.then === 'function' }`},

			// ---- Dimension 2: Scoping & nesting ----

			// Inside an async generator — yield skips by default
			{Code: `async function * gen() { yield thing().then() }`},

			// Inside a class constructor (class expression form)
			{Code: `const C = class { constructor() { doSomething.then(cb) } }`},

			// Constructor nested inside another constructor (inner class)
			{Code: "class Outer {\n  constructor() {\n    class Inner {\n      constructor() { thing.then(cb) }\n    }\n  }\n}"},

			// Deeply nested inside constructor — all levels exempt (non-strict)
			{Code: "class Foo {\n  constructor() {\n    function inner() { thing.then(cb) }\n  }\n}"},

			// ---- Dimension 3: Autofix ----
			// N/A: this rule has no autofix.

			// ---- Dimension 4: Universal edge shapes ----

			// Parenthesized entire callee: (hey.then)(arg) — inside await → valid
			{Code: `async function f() { await (hey.then)() }`},

			// Non-null assertion on receiver: hey!.then() — inside await → valid
			{Code: `async function f() { await hey!.then() }`},

			// Type assertion on receiver: (hey as any).then() — inside await → valid
			{Code: `async function f() { await (hey as any).then() }`},

			// Optional chain on method: hey?.then() — inside await → valid
			{Code: `async function f() { await hey?.then() }`},

			// Optional call: hey.then?.() — inside await → valid
			{Code: `async function f() { await hey.then?.() }`},

			// Top-level: all forms are exempt from reporting
			{Code: `thing.then(cb)`},
			{Code: `thing.catch(cb)`},
			{Code: `thing.finally(cb)`},
			{Code: `(thing).then(cb)`},
			{Code: `thing?.then(cb)`},

			// ---- Branch lock-ins (Layer 3) ----

			// Locks in isTopLevelScoped branch: bare call at file top level → skip
			{Code: `somePromise.then(() => {})`},

			// Locks in isCypress direct-object branch: cy.then() (base case)
			{Code: `function f() { cy.then(cb) }`},

			// Locks in isCypress deeply-nested branch: 3-level cy chain
			{Code: `function f() { cy.get("button").click().then() }`},

			// Locks in isCypress element-access branch: cy['get']() chain
			{Code: `function f() { cy['get']('x').then(go) }`},

			// Locks in propName guard: non-then/catch/finally method call
			{Code: `function f() { p.resolve() }`},
			{Code: `function f() { p.thenSomething() }`},
			{Code: `function f() { p.catchError() }`},

			// ---- For-loop scoping: var / no-decl heads create no scope ----
			// eslint-scope does not open a scope unless the head is let/const/using.
			// Without a block body there is no KindBlock ancestor either, so these stay top-level.
			{Code: `for(;;) thing.then(cb)`},
			{Code: `for(var i=0;;) thing.then(cb)`},
			{Code: `for(var a of xs) thing.then(cb)`},
			{Code: `for(a of xs) thing.then(cb)`},
			{Code: `for(var k in o) thing.then(cb)`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 1: AST node types ----

			// Optional chain on method: hey?.then() inside a regular function IS reported.
			// '?' shifts the start of 'then' one column right vs plain hey.then.
			{
				Code: `function f() { hey?.then() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 21, EndLine: 1, EndColumn: 25},
				},
			},
			// .catch() on optional chain
			{
				Code: `function f() { hey?.catch(e => {}) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 21, EndLine: 1, EndColumn: 26},
				},
			},

			// ---- Dimension 2: Scoping & nesting ----

			// Inside arrow function (not top-level)
			{
				Code: `const f = () => { hey.then(x => {}) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 23, EndLine: 1, EndColumn: 27},
				},
			},
			// Top-level bare block: still a Program-level scope per eslint-scope, so reported
			{
				Code: `{ thing.then(cb) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 9, EndLine: 1, EndColumn: 13},
				},
			},
			// Top-level if-block
			{
				Code: `if (x) { thing.then(cb) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 16, EndLine: 1, EndColumn: 20},
				},
			},
			// Top-level for-of loop (braced body — KindBlock catches it regardless)
			{
				Code: `for (const a of xs) { thing.then(cb) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 29, EndLine: 1, EndColumn: 33},
				},
			},
			// Top-level for-loops with let/const but no braces — scope comes from the for-head
			{
				Code: `for(let i=0;;) thing.then(cb)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: `for(const a of xs) thing.then(cb)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 26, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code: `for(const k in o) thing.then(cb)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 25, EndLine: 1, EndColumn: 29},
				},
			},
			// Top-level switch/case
			{
				Code: `switch (x) { case 1: thing.then(cb) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 28, EndLine: 1, EndColumn: 32},
				},
			},
			// Top-level try/catch
			{
				Code: `try { x() } catch (e) { thing.then(cb) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 31, EndLine: 1, EndColumn: 35},
				},
			},
			// Class field initializer at top level
			{
				Code: `class C { x = p.then() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 17, EndLine: 1, EndColumn: 21},
				},
			},
			// Class static block at top level
			{
				Code: `class C { static { p.then() } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
				},
			},
			// Top-level inside a TS namespace body
			{
				Code: `namespace N { Promise.resolve().then(() => {}) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 33, EndLine: 1, EndColumn: 37},
				},
			},
			// Top-level inside a legacy TS module body
			{
				Code: `module M { Promise.resolve().then(() => {}) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 30, EndLine: 1, EndColumn: 34},
				},
			},
			// Inside class method (not constructor)
			{
				Code: `class C { method() { hey.then(x => {}) } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 26, EndLine: 1, EndColumn: 30},
				},
			},
			// Inside static class method
			{
				Code: `class C { static method() { hey.then(x => {}) } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 33, EndLine: 1, EndColumn: 37},
				},
			},
			// Inside getter — 'return ' prefix shifts column further right
			{
				Code: `class C { get val() { return hey.then(x => x) } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 34, EndLine: 1, EndColumn: 38},
				},
			},

			// ---- Dimension 4: Universal edge shapes ----

			// Parenthesized entire callee: (hey.then)(arg) outside await.
			// The opening '(' shifts the reported property by one column.
			{
				Code: `function f() { (hey.then)(x => {}) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 21, EndLine: 1, EndColumn: 25},
				},
			},
			// Non-null assertion on receiver: hey!.then()
			// The '!' shifts the column of 'then' one position right.
			{
				Code: `function f() { hey!.then(x => {}) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 21, EndLine: 1, EndColumn: 25},
				},
			},
			// Type assertion on receiver: (hey as any).then()
			{
				Code: `function f() { (hey as any).then(x => {}) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 29, EndLine: 1, EndColumn: 33},
				},
			},

			// ---- Branch lock-ins (Layer 3) ----

			// Locks in strict-mode yield branch: yield with strict → reported
			{
				Code:    `function * f() { yield thing().then() }`,
				Options: map[string]interface{}{"strict": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 32, EndLine: 1, EndColumn: 36},
				},
			},
			// Locks in strict-mode constructor branch: constructor with strict → reported
			{
				Code:    `class C { constructor() { thing.then(cb) } }`,
				Options: map[string]interface{}{"strict": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 33, EndLine: 1, EndColumn: 37},
				},
			},
			// Locks in strict-mode await branch: await-wrapped .catch() with strict → reported
			{
				Code:    `async function f() { await thing().catch(e => {}) }`,
				Options: map[string]interface{}{"strict": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 36, EndLine: 1, EndColumn: 41},
				},
			},

			// ---- Real-user shapes ----

			// Real-user: chained .then().catch() in a non-async function.
			// Errors are reported outer-to-inner: catch first, then second.
			{
				Code: `function fetch(url) { return http.get(url).then(r => r.json()).catch(e => null) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 64, EndLine: 1, EndColumn: 69},
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 44, EndLine: 1, EndColumn: 48},
				},
			},
			// Real-user: .then() nested as callback argument
			{
				Code: `function f() { Promise.all(items.map(x => x.then(cb))) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 45, EndLine: 1, EndColumn: 49},
				},
			},

			// ---- Element access with an identifier key ----
			// Upstream matches any MemberExpression.callee, including computed
			// identifier access like obj[then](), not just dot notation.
			{
				Code: `function f() { obj[then]() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: `function f() { obj[(then)]() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 21, EndLine: 1, EndColumn: 25},
				},
			},
		},
	)
}
