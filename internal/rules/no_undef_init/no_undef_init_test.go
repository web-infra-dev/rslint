package no_undef_init

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUndefInit(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUndefInitRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Dimension 1: Declaration kinds that should NOT be reported
			// ============================================================

			// No initializer
			{Code: `var a;`},
			{Code: `let b;`},
			{Code: `const c = 1;`},

			// const = undefined is allowed (const requires initializer, can't omit)
			{Code: `const foo = undefined`},

			// using / await using are allowed (resource management semantics)
			{Code: `using foo = undefined`},
			{Code: `async function f() { await using foo = undefined }`},
			{Code: `using a = condition ? getDisposableResource() : undefined;`},

			// ============================================================
			// Dimension 2: Initializer is NOT bare `undefined` identifier
			// ============================================================

			{Code: `var a = null;`},
			{Code: `var a = 0;`},
			{Code: `var a = '';`},
			{Code: `var a = false;`},
			{Code: `let a = void 0;`},
			{Code: `let a = void undefined;`},
			{Code: `let a = typeof undefined;`},
			// Property access to `undefined` — not the bare identifier
			{Code: `let a = window.undefined;`},
			{Code: `let a = globalThis.undefined;`},
			// Conditional/ternary with undefined
			{Code: `let a = true ? undefined : 1;`},
			// Template literal that looks like undefined — not an Identifier
			{Code: "let a = `undefined`;"},
			// String literal "undefined"
			{Code: `let a = "undefined";`},
			// TypeScript `as` expression — init is AsExpression, not Identifier
			{Code: `let a = undefined as any;`},
			// TypeScript type assertion — init is TypeAssertion, not Identifier
			{Code: `let a = <any>undefined;`},
			// TypeScript `satisfies` — init is SatisfiesExpression, not Identifier
			{Code: `let a = undefined satisfies any;`},
			// Non-null assertion on undefined — init is NonNullExpression
			{Code: `let a = undefined!;`},

			// ============================================================
			// Dimension 3: Class fields (PropertyDeclaration, not VariableDeclaration)
			// ============================================================

			{Code: `class C { field = undefined; }`},
			{Code: `class C { static field = undefined; }`},
			{Code: `class C { readonly field: string | undefined = undefined; }`},
			{Code: `class C { #private = undefined; }`},

			// ============================================================
			// Dimension 4: Shadowed `undefined`
			// ============================================================

			// var shadows undefined at top level
			{Code: `var undefined = 5; var foo = undefined;`},
			{Code: `var undefined = 5; let foo = undefined;`},
			// function parameter shadows undefined
			{Code: `function f(undefined: any) { var a = undefined; }`},
			{Code: `function f(undefined: any) { let a = undefined; }`},
			{Code: `(function(undefined: any) { var a = undefined; })()`},
			// function name shadows undefined
			{Code: `function undefined() {} var a = undefined;`},
			// let/const in block scope shadows undefined
			{Code: `{ let undefined = 1; var a = undefined; }`},
			{Code: `{ const undefined = 1; let a = undefined; }`},
			// catch variable shadows undefined
			{Code: `try {} catch(undefined) { var a = undefined; }`},
			// class name shadows undefined
			{Code: `class undefined {} let a = undefined;`},
			// enum name shadows undefined
			{Code: `enum undefined { A } let a = undefined;`},
			// hoisted var shadows undefined inside a function
			{Code: `function f() { let a = undefined; var undefined = 5; }`},
			// nested function parameter shadow
			{Code: `function outer() { function inner(undefined: any) { let a = undefined; } }`},
			// Inner function inherits outer scope's shadowed `undefined` (lexical scoping)
			{Code: `function f(undefined: any) { function g() { let a = undefined; } }`},
			// Arrow function parameter shadows undefined
			{Code: `const f = (undefined: any) => { let a = undefined; }`},
			// Destructuring parameter shadows undefined
			{Code: `function f({undefined}: any) { let a = undefined; }`},
			// for-in variable shadows undefined
			{Code: `for (var undefined in obj) { let a = undefined; }`},
			// for-of variable shadows undefined
			{Code: `for (let undefined of [1]) { let a = undefined; }`},
			// import shadows undefined (import name binding)
			{Code: `import undefined from 'module'; let a = undefined;`},
			// bare `let undefined;` (no initializer) shadows via declaration
			{Code: `{ let undefined; let a = undefined; }`},
			// arrow param shadows, nested arrow inherits (lexical scoping chain)
			{Code: `const f = (undefined: any) => { const g = () => { let a = undefined; }; }`},
			// for-initializer shadows
			{Code: `for (let undefined = 0; undefined < 10; undefined++) { let a = undefined; }`},
			// Parenthesized + shadowed
			{Code: `var undefined = 5; let a = (undefined);`},
			// var at top-level shadows inside class static block (scoping chain)
			{Code: "var undefined = 5;\nclass C { static { let a = undefined; } }"},
			// let inside static block shadows within the block
			{Code: `class C { static { let undefined = 1; let a = undefined; } }`},

			// ============================================================
			// Dimension 5a: `= undefined` that is NOT a VariableDeclaration initializer
			// ============================================================

			// Function parameter defaults
			{Code: `function f(a = undefined) {}`},
			{Code: `const f = (a = undefined) => a;`},
			{Code: `class C { method(a = undefined) {} }`},
			// Destructuring defaults inside binding patterns
			{Code: `let {a = undefined} = {};`},
			{Code: `let [a = undefined] = [];`},
			{Code: `let {a: {b = undefined}} = {a: {}};`},
			// Assignment expression (not declaration)
			{Code: `var a; a = undefined;`},

			// ============================================================
			// Dimension 5b: TypeScript-specific valid patterns
			// ============================================================

			{Code: `declare let a: string;`},
			{Code: `declare var a: number | undefined;`},
			{Code: `interface I { a: undefined; }`},
			{Code: `type T = { a: undefined };`},

			// ============================================================
			// Dimension 6: Various nesting contexts (valid — no `= undefined`)
			// ============================================================

			{Code: `function f() { var a; }`},
			{Code: `const f = () => { let a; }`},
			{Code: `if (true) { let a; }`},
			{Code: `for (let a;;) {}`},
			{Code: `switch (x) { case 1: let a; break; }`},

			// ============================================================
			// Dimension 7: Multi-byte characters
			// ============================================================

			{Code: `let 变量 = 1;`},
			{Code: `const café = undefined;`},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Dimension 1: var — always reported, never autofixed
			// ============================================================

			{
				Code:   `var a = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   `var a = undefined, b = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   `var a = 1, b = undefined, c = 5;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 12}},
			},
			{
				Code:   `var a = 1, b = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 12}},
			},
			// var destructuring
			{
				Code:   `var [a] = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   `var {a} = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Multiple var = undefined — each reported independently, no autofix
			{
				Code: `var a = undefined, b = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5},
					{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 20},
				},
			},

			// ============================================================
			// Dimension 2: let — reported and autofixed
			// ============================================================

			{
				Code:   `let a = undefined;`,
				Output: []string{`let a;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   `let a = undefined, b = 1;`,
				Output: []string{`let a, b = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   `let a = 1, b = undefined, c = 5;`,
				Output: []string{`let a = 1, b, c = 5;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 12}},
			},
			{
				Code:   `let a = 1, b = undefined;`,
				Output: []string{`let a = 1, b;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 12}},
			},
			// let destructuring — reported but no autofix
			{
				Code:   `let [a] = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   `let {a} = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   `let [[a, b], c] = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   `let {a: {b}} = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// No semicolon
			{
				Code:   `let a = undefined`,
				Output: []string{`let a`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},

			// ============================================================
			// Dimension 3: Nesting contexts — var (no autofix)
			// ============================================================

			{
				Code:   `for(var i in [1,2,3]){var a = undefined; for(var j in [1,2,3]){}}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 27}},
			},
			{
				Code:   `function f() { var a = undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 20}},
			},
			{
				Code:   `const f = () => { var a = undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 23}},
			},
			{
				Code:   `async function f() { var a = undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 26}},
			},
			{
				Code:   `function* g() { var a = undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 21}},
			},
			// try / catch / finally
			{
				Code:   `try { var a = undefined; } catch(e) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 11}},
			},
			{
				Code:   `try {} catch(e) { var a = undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 23}},
			},
			{
				Code:   `try {} finally { var a = undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 22}},
			},
			// switch
			{
				Code:   `switch(x) { case 1: var a = undefined; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 25}},
			},
			// if/else — multiple errors
			{
				Code: `if (true) { var a = undefined; } else { var b = undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 17},
					{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 45},
				},
			},

			// ============================================================
			// Dimension 4: Nesting contexts — let (with autofix)
			// ============================================================

			{
				Code:   `for(var i in [1,2,3]){let a = undefined; for(var j in [1,2,3]){}}`,
				Output: []string{`for(var i in [1,2,3]){let a; for(var j in [1,2,3]){}}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 27}},
			},
			{
				Code:   `function f() { let a = undefined; }`,
				Output: []string{`function f() { let a; }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 20}},
			},
			{
				Code:   `const f = () => { let a = undefined; }`,
				Output: []string{`const f = () => { let a; }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 23}},
			},
			{
				Code:   `class C { method() { let a = undefined; } }`,
				Output: []string{`class C { method() { let a; } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 26}},
			},
			{
				Code:   `class C { constructor() { let a = undefined; } }`,
				Output: []string{`class C { constructor() { let a; } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 31}},
			},
			{
				Code:   `class C { static init() { let a = undefined; } }`,
				Output: []string{`class C { static init() { let a; } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 31}},
			},
			{
				Code:   `const f = async () => { let a = undefined; }`,
				Output: []string{`const f = async () => { let a; }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 29}},
			},
			{
				Code:   `for (let i = 0; i < 10; i++) { let a = undefined; }`,
				Output: []string{`for (let i = 0; i < 10; i++) { let a; }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 36}},
			},
			{
				Code:   `while(true) { let a = undefined; }`,
				Output: []string{`while(true) { let a; }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 19}},
			},
			{
				Code:   `do { let a = undefined; } while(true);`,
				Output: []string{`do { let a; } while(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 10}},
			},

			// ============================================================
			// Dimension 5: Deeply nested structures
			// ============================================================

			{
				Code:   `function a() { if (true) { { let x = undefined; } } }`,
				Output: []string{`function a() { if (true) { { let x; } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 34}},
			},
			{
				Code:   `const f = () => { class C { m() { let x = undefined; } } }`,
				Output: []string{`const f = () => { class C { m() { let x; } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 39}},
			},

			// ============================================================
			// Dimension 6: TypeScript type annotations — autofix preserves them
			// ============================================================

			{
				Code:   `let a: string | undefined = undefined;`,
				Output: []string{`let a: string | undefined;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   `let a: any = undefined;`,
				Output: []string{`let a: any;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// var with type annotation — no autofix
			{
				Code:   `var a: number | undefined = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Generic type annotation
			{
				Code:   `let a: Array<string | undefined> = undefined;`,
				Output: []string{`let a: Array<string | undefined>;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},

			// ============================================================
			// Dimension 7: Multi-line code
			// ============================================================

			{
				Code:   "var a\n  = undefined;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   "let a\n  = undefined;",
				Output: []string{"let a;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   "let a: string\n  = undefined;",
				Output: []string{"let a: string;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},

			// ============================================================
			// Dimension 8: Comment placement (autofix suppression)
			// ============================================================

			// Comment between name and `=` — no autofix
			{
				Code:   "let a/**/ = undefined;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   "let a /**/ = undefined;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Comment between `=` and `undefined` — no autofix
			{
				Code:   "let a = /**/undefined;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Line comment between tokens — no autofix
			{
				Code:   "let a//\n= undefined;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   "let a = //\nundefined;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Comment BEFORE name — autofix still works
			{
				Code:   "let /* comment */a = undefined;",
				Output: []string{"let /* comment */a;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 18}},
			},
			// Comment AFTER `undefined` (trailing) — autofix works, comment preserved
			{
				Code:   "let a = undefined/* comment */;",
				Output: []string{"let a/* comment */;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   "let a = undefined/* comment */, b;",
				Output: []string{"let a/* comment */, b;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			{
				Code:   "let a = undefined//comment\n, b;",
				Output: []string{"let a//comment\n, b;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Comment between type annotation and `=` — no autofix
			{
				Code:   "let a: string/**/ = undefined;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},

			// ============================================================
			// Dimension 9: Additional nesting contexts
			// ============================================================

			// export let
			{
				Code:   `export let a = undefined;`,
				Output: []string{`export let a;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 12}},
			},
			// export var
			{
				Code:   `export var a = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 12}},
			},
			// class static block
			{
				Code:   `class C { static { let a = undefined; } }`,
				Output: []string{`class C { static { let a; } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 24}},
			},
			// namespace
			{
				Code:   `namespace N { let a = undefined; }`,
				Output: []string{`namespace N { let a; }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 19}},
			},

			// ============================================================
			// Dimension 10: Additional destructuring patterns
			// ============================================================

			// Empty array destructuring
			{
				Code:   `let [] = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Empty object destructuring
			{
				Code:   `let {} = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Destructuring with rest
			{
				Code:   `let [...a] = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Destructuring with defaults
			{
				Code:   `let {a = 1} = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Destructuring with rename
			{
				Code:   `let {a: b} = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},

			// ============================================================
			// Dimension 11: Whitespace and formatting edge cases
			// ============================================================

			// No space around `=`
			{
				Code:   `let a=undefined;`,
				Output: []string{`let a;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Multiple spaces
			{
				Code:   `let a  =  undefined;`,
				Output: []string{`let a;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Tab indentation
			{
				Code:   "\tlet a = undefined;",
				Output: []string{"\tlet a;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 6}},
			},

			// ============================================================
			// Dimension 12: TypeScript type annotation + autofix edge cases
			// ============================================================

			// Type annotation is literally `undefined`
			{
				Code:   `let a: undefined = undefined;`,
				Output: []string{`let a: undefined;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Type annotation + trailing comment after undefined — autofix preserves both
			{
				Code:   "let a: string = undefined/* comment */;",
				Output: []string{"let a: string/* comment */;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Tuple type annotation
			{
				Code:   `let a: [string, number] = undefined;`,
				Output: []string{`let a: [string, number];`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Function type annotation
			{
				Code:   `let a: () => void = undefined;`,
				Output: []string{`let a: () => void;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},

			// ============================================================
			// Dimension 15: Parenthesized `undefined` (ESTree strips parens)
			// ============================================================

			{
				Code:   `let a = (undefined);`,
				Output: []string{`let a;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Multi-byte: CJK characters
			{
				Code:   `let 变量 = undefined;`,
				Output: []string{`let 变量;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Multi-byte: CJK with type annotation
			{
				Code:   `let 数据: string = undefined;`,
				Output: []string{`let 数据: string;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Multiple levels of parentheses
			{
				Code:   `let a = ((undefined));`,
				Output: []string{`let a;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// var + parenthesized — no autofix
			{
				Code:   `var a = (undefined);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// Parenthesized + shadowed — should NOT report
			// (already valid via shadowing test, no need to add here)
			// Conditional type annotation
			{
				Code:   `let a: string extends number ? never : string = undefined;`,
				Output: []string{`let a: string extends number ? never : string;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},
			// ExclamationToken preserved (let a!: T = undefined → let a!: T)
			{
				Code:   `let a!: string = undefined;`,
				Output: []string{`let a!: string;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5}},
			},

			// ============================================================
			// Dimension 13: for-initializer (VariableDeclarationList inside ForStatement)
			// ============================================================

			// var in for-initializer — no autofix
			{
				Code:   `for (var a = undefined;;) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 10}},
			},
			// let in for-initializer — autofix
			{
				Code:   `for (let a = undefined;;) {}`,
				Output: []string{`for (let a;;) {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 10}},
			},
			// let with type in for-initializer
			{
				Code:   `for (let a: any = undefined;;) {}`,
				Output: []string{`for (let a: any;;) {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 10}},
			},

			// ============================================================
			// Dimension 14: Multiple let = undefined with autofix
			// ============================================================

			// Both declarators fixed
			{
				Code:   `let a = undefined, b = undefined;`,
				Output: []string{`let a, b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5},
					{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 20},
				},
			},
			// Mixed: one with type annotation, one without
			{
				Code:   `let a: string = undefined, b = undefined;`,
				Output: []string{`let a: string, b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 5},
					{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 28},
				},
			},
			// Three declarators, only middle is = undefined
			{
				Code:   `let a = 1, b = undefined, c = 2;`,
				Output: []string{`let a = 1, b, c = 2;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryUndefinedInit", Line: 1, Column: 12}},
			},
		},
	)
}
