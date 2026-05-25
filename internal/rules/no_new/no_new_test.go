package no_new

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewRule,
		// Valid cases — ESLint-ported + extended edge cases
		[]rule_tester.ValidTestCase{
			// ---- Original ESLint cases ----
			{Code: `var a = new Date()`},
			{Code: `var a; if (a === new Date()) { a = false; }`},

			// ---- Used as a value (assignment / return / argument) ----
			{Code: `const thing = new Thing();`},
			{Code: `let x = new Foo();`},
			{Code: `foo(new Bar());`},
			{Code: `function f() { return new Baz(); }`},
			{Code: `var x = new A() || new B();`},
			{Code: `var x = true ? new A() : new B();`},
			{Code: `var x = [new A(), new B()];`},
			{Code: `var x = { a: new A() };`},
			{Code: `class C { m() { return new A(); } }`},
			{Code: `(function () { return new Foo(); })();`},

			// ---- Result is further dereferenced / called (not a standalone new) ----
			{Code: `(new Foo).bar;`},
			{Code: `(new Foo()).bar();`},
			{Code: `new Foo().bar;`},
			{Code: `new Foo()?.bar;`},
			{Code: `new Foo()[0];`},
			{Code: `new Foo().bar();`},

			// ---- Wrapped by another operator so the statement expr is not a NewExpression ----
			{Code: `void new Foo();`},
			{Code: `!new Foo();`},
			{Code: `typeof new Foo();`},
			{Code: `delete new Foo().x;`},
			{Code: `new Foo(), new Bar();`},
			{Code: `async function f() { await new Promise(r => r(1)); }`},
			{Code: `function* g() { yield new Foo(); }`},

			// ---- Plain call (not new) ----
			{Code: `Foo();`},
			{Code: `foo.bar();`},

			// ---- Arrow with expression body (not a statement) ----
			{Code: `var f = () => new Foo();`},

			// ---- Wrapping operators make the statement expression non-NewExpression ----
			{Code: `true && new Foo();`},
			{Code: `a ? new Foo() : new Bar();`},

			// ---- TypeScript type casts / satisfies around the new ----
			{Code: `new Foo() as Bar;`},
			{Code: `(new Foo() as Bar);`},
			{Code: `<Foo>new Bar();`},
			{Code: `new Foo() satisfies Bar;`},

			// ---- Not an ExpressionStatement: export default / decorator argument ----
			{Code: `export default new Foo();`},
			{Code: `@new Dec() class C {}`},

			// ---- Class field / default parameter initializers (not ExpressionStatement) ----
			{Code: `class C { x = new Foo(); }`},
			{Code: `class C { static x = new Foo(); }`},
			{Code: `function f(x = new Foo()) {}`},

			// ---- Chained type casts ----
			{Code: `new Foo() as any as Bar;`},

			// ---- Destructuring default values / computed keys (not ExpressionStatement) ----
			{Code: `var { a = new Foo() } = {};`},
			{Code: `var [a = new Foo()] = [];`},
			{Code: `var obj = { [new Foo()]: 1 };`},

			// ---- Template literal expression slot ----
			{Code: "`${new Foo()}`;"},

			// ---- Calling the result of new ----
			{Code: `(new Foo())();`},

			// ---- TC39 accessor keyword (class field init, not ExpressionStatement) ----
			{Code: `class C { accessor x = new Foo(); }`},

			// ---- Decorator factory call (new is argument to decorator) ----
			{Code: `@factory(new Foo()) class C {}`},

			// ---- import.meta call is not NewExpression ----
			{Code: `import.meta.foo();`},

			// ---- `using` / `await using` declarations (Stage 3) ----
			{Code: `function f() { using x = new Foo(); }`},
			{Code: `async function f() { await using x = new Foo(); }`},

			// ---- Chained / combined TS type operators on new (not ExpressionStatement's direct kind) ----
			{Code: `new Foo() as const;`},
			{Code: `<const>new Foo();`},
			{Code: `(new Foo())!;`},

			// ---- Non-null or type-assertion on callee, but used as a value ----
			{Code: `var x = new Foo!();`},
			{Code: `var x = new (Foo as any)();`},

			// ---- Decorators as values / on methods ----
			{Code: `@a @b() @c(new X()) class C {}`},
			{Code: `class C { @dec method() { return new Foo(); } }`},

			// ---- Class with implements / extending mixin, used as value ----
			{Code: `class C implements I { m() { return new Foo(); } }`},
			{Code: `class C extends A implements I { m() { return new Foo(); } }`},

			// ---- Generator + yield-new / yield* new ----
			{Code: `function* g() { yield* new Foo(); }`},
			{Code: `async function* g() { yield new Foo(); }`},

			// ---- Multi-generic / member-access generic ----
			{Code: `var x = new Foo<number, string>();`},
			{Code: `var x = new ns.Foo<number>();`},

			// ---- JSX: new as a value inside JsxExpression / attribute slot / fragment ----
			{Code: `const j = <Foo />;`, Tsx: true},
			{Code: `const j = <div>{new Foo()}</div>;`, Tsx: true},
			{Code: `const j = <Foo attr={new Bar()} />;`, Tsx: true},
			{Code: `function render() { return <Foo>{new Bar()}</Foo>; }`, Tsx: true},
			{Code: `const j = <>{new Foo()}</>;`, Tsx: true},
		},
		// Invalid cases — ESLint-ported + extended edge cases
		[]rule_tester.InvalidTestCase{
			// ---- Original ESLint case ----
			{
				Code: `new Date()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Basic forms ----
			{
				Code: `new Foo;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Foo('a', 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},
			{
				Code: `new foo.Bar();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},
			{
				Code: `new foo.bar.Baz();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Parenthesized: ESLint still reports (paren-transparent); tsgo requires SkipParentheses ----
			{
				Code: `(new Foo());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},
			{
				Code: `((new Foo()));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- new of new (outer is still a NewExpression directly under the statement) ----
			{
				Code: `new new Foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Labeled statement: inner ExpressionStatement still matches ----
			{
				Code: `label: new Foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 8},
				},
			},

			// ---- Nested inside various constructs ----
			{
				Code: `function f() { new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 16},
				},
			},
			{
				Code: `var f = () => { new Foo(); };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 17},
				},
			},
			{
				Code: `if (true) { new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 13},
				},
			},
			{
				Code: `class C { m() { new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 17},
				},
			},
			{
				Code: `for (;;) { new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 12},
				},
			},
			{
				Code: `try { new Foo(); } catch {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 7},
				},
			},

			// ---- Multiple / multi-line ----
			{
				Code: `new Foo(); new Bar();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
					{MessageId: "noNewStatement", Line: 1, Column: 12},
				},
			},
			{
				Code: "new\n  Foo();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- TypeScript generics on the callee ----
			{
				Code: `new Foo<number>();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Class-body static block (ES2022) ----
			{
				Code: `class C { static { new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 20},
				},
			},

			// ---- Anonymous class as constructor ----
			{
				Code: `new class {}();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Tagged-template as new callee ----
			{
				Code: "new Foo``;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- More container scopes ----
			{
				Code: `do { new Foo(); } while (false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 6},
				},
			},
			{
				Code: `while (true) { new Foo(); break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 16},
				},
			},
			{
				Code: `switch (x) { case 1: new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 22},
				},
			},
			{
				Code: `if (a) {} else { new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 18},
				},
			},
			{
				Code: `try {} finally { new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 18},
				},
			},
			{
				Code: `{ new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 3},
				},
			},

			// ---- Comments around / inside the expression ----
			{
				Code: `/* a */ new Foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 9},
				},
			},
			{
				Code: `new /* a */ Foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Single-statement branches (no block) ----
			{
				Code: `if (a) new Foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 8},
				},
			},
			{
				Code: `if (a) new Foo(); else new Bar();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 8},
					{MessageId: "noNewStatement", Line: 1, Column: 24},
				},
			},
			{
				Code: `for (var i = 0; i < 1; i++) new Foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 29},
				},
			},
			{
				Code: `while (a) new Foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 11},
				},
			},

			// ---- IIFE-style: callee is a parenthesized function expression ----
			{
				Code: `new (function(){})();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Spread argument ----
			{
				Code: `new Foo(...args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Computed-access callee (no call parens) ----
			{
				Code: `new Foo[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- ASI splits into two statements ----
			{
				Code: "var a = b\nnew Foo()",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 2, Column: 1},
				},
			},

			// ---- Trailing extra semicolon is an EmptyStatement, not a second report ----
			{
				Code: `new Foo();;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- TypeScript namespace body hosts ExpressionStatement ----
			{
				Code: `namespace N { new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 15},
				},
			},

			// ---- Class expression as constructor ----
			{
				Code: `new (class {})();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},
			{
				Code: `new (class extends Base {})();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Async IIFE ----
			{
				Code: `new (async function(){})();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Computed-member callee ----
			{
				Code: `new obj[key]();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Deep member chain ----
			{
				Code: `new a.b.c.d();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Tagged template with member-access tag ----
			{
				Code: "new foo.Bar`x`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Double report: outer new-statement AND inner new inside its callback ----
			{
				Code: `new Promise(r => { new Foo(); r(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
					{MessageId: "noNewStatement", Line: 1, Column: 20},
				},
			},

			// ---- After a directive prologue ----
			{
				Code: `"use strict"; new Foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 15},
				},
			},

			// ---- Interleaved between other statements ----
			{
				Code: `foo; new Bar(); baz;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 6},
				},
			},

			// ---- Class body: constructor / derived constructor / method / static / getter / setter / private method ----
			{
				Code: `class C { constructor() { new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 27},
				},
			},
			{
				Code: `class C extends B { constructor() { super(); new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 46},
				},
			},
			{
				Code: `class C { static m() { new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 24},
				},
			},
			{
				Code: `class C { get x() { new Foo(); return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 21},
				},
			},
			{
				Code: `class C { set x(v) { new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 22},
				},
			},
			{
				Code: `class C { #m() { new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 18},
				},
			},

			// ---- Object literal method / accessor bodies ----
			{
				Code: `var obj = { m() { new Foo(); } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 19},
				},
			},
			{
				Code: `var obj = { get x() { new Foo(); return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 23},
				},
			},
			{
				Code: `var obj = { set x(v) { new Foo(); } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 24},
				},
			},

			// ---- Catch block / for-in / for-of / for-await-of single-statement bodies ----
			{
				Code: `try {} catch (e) { new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 20},
				},
			},
			{
				Code: `for (var x in arr) new Foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 20},
				},
			},
			{
				Code: `for (var x of arr) new Foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 20},
				},
			},
			{
				Code: `async function f() { for await (const x of arr) new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 49},
				},
			},

			// ---- Async IIFE with inner new statement ----
			{
				Code: `(async () => { new Foo(); })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 16},
				},
			},

			// ---- new used as a statement AND one of its arguments contains another new statement ----
			{
				Code: `new Foo(function() { new Bar(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
					{MessageId: "noNewStatement", Line: 1, Column: 22},
				},
			},

			// ---- Nested namespace body ----
			{
				Code: `namespace A.B { new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 17},
				},
			},

			// ---- Legacy `module` keyword (TS) ----
			{
				Code: `module N { new Foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 12},
				},
			},

			// ---- Abstract class constructor ----
			{
				Code: `abstract class A { constructor() { new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 36},
				},
			},

			// ---- `override` method body ----
			{
				Code: `class B extends A { override m() { new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 36},
				},
			},

			// ---- Decorated method body ----
			{
				Code: `class C { @dec m() { new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 22},
				},
			},

			// ---- Class with implements clause ----
			{
				Code: `class C implements I { m() { new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 30},
				},
			},

			// ---- Mixin-style extends ----
			{
				Code: `class C extends Mix(Base) { m() { new Foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 35},
				},
			},

			// ---- Static-initialization block with mixed declaration + new statement ----
			{
				Code: `class C { static { let x = new Foo(); new Bar(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 39},
				},
			},

			// ---- Multi-generic / member-access generic as statement ----
			{
				Code: `new Foo<number, string>();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},
			{
				Code: `new ns.Foo<number>();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Non-null assertion on callee, used as statement ----
			{
				Code: `new Foo!();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Type assertion on callee ----
			{
				Code: `new (Foo as any)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 1},
				},
			},

			// ---- Dynamic import awaited then new'd ----
			{
				Code: `async function f() { new (await import("x")).default(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 22},
				},
			},

			// ---- JSX: new-as-statement inside an arrow body nested in a JSX handler / child ----
			{
				Code: `const j = <Foo onClick={() => { new Bar(); }} />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 33},
				},
			},
			{
				Code: `const j = <div>{() => { new Foo(); return 1; }}</div>;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewStatement", Line: 1, Column: 25},
				},
			},
		},
	)
}
