package symbol_description

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSymbolDescriptionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SymbolDescriptionRule,
		[]rule_tester.ValidTestCase{
			// ---- From ESLint upstream ----
			{Code: `Symbol("Foo");`},
			{Code: `var foo = "foo"; Symbol(foo);`},
			// Ignore if it's shadowed.
			{Code: `var Symbol = function () {}; Symbol();`},
			{Code: `Symbol(); var Symbol = function () {};`},
			{Code: `function bar() { var Symbol = function () {}; Symbol(); }`},
			// Ignore if it's an argument.
			{Code: `function bar(Symbol) { Symbol(); }`},

			// ---- Argument shapes (any non-empty arg is OK) ----
			{Code: `Symbol("");`},
			{Code: "Symbol(`foo`);"},
			{Code: "Symbol(`${x}`);"},
			{Code: `Symbol(undefined);`},
			{Code: `Symbol(null);`},
			{Code: `Symbol(42);`},
			{Code: `Symbol(cond ? "a" : "b");`},
			{Code: `Symbol(getName());`},
			{Code: `Symbol(...args);`},
			{Code: `Symbol("a", "b");`},

			// ---- Not the global `Symbol` identifier being called ----
			{Code: `foo.Symbol();`},
			{Code: `obj["Symbol"]();`},
			{Code: `new Symbol();`},
			{Code: `Symbol;`}, // referenced but not called

			// ---- Shadowing — declaration forms ----
			{Code: `let Symbol = 1; Symbol();`},
			{Code: `const Symbol = () => {}; Symbol();`},
			{Code: `function Symbol() {} Symbol();`},
			{Code: `class Symbol {} new Symbol();`},
			{Code: `const f = (Symbol) => { Symbol(); };`},
			{Code: `function f(...Symbol) { Symbol(); }`},
			{Code: `function f({ Symbol }) { Symbol(); }`},
			{Code: `var { Symbol } = obj; Symbol();`},
			{Code: `var [Symbol] = arr; Symbol();`},
			{Code: `try {} catch (Symbol) { Symbol(); }`},
			{Code: `for (let Symbol of arr) { Symbol(); }`},
			{Code: `for (var Symbol = 0;;) { Symbol(); }`},
			// Hoisting — the call precedes the declaration in the same scope.
			{Code: `Symbol(); function Symbol() {}`},

			// ---- Outer shadow propagates into inner scopes ----
			{Code: `var Symbol = 1; function f() { Symbol(); }`},
			{Code: `var Symbol = 1; const f = () => Symbol();`},
			{Code: `var Symbol = 1; class C { m() { Symbol(); } }`},

			// ---- Import forms shadow the global Symbol ----
			{Code: `import { Symbol } from "x"; Symbol();`},
			{Code: `import Symbol from "x"; Symbol();`},
			{Code: `import * as Symbol from "x"; Symbol();`},
			{Code: `import { Foo as Symbol } from "x"; Symbol();`},

			// ---- Tagged template is not a CallExpression ----
			{Code: "Symbol`foo`;"},

			// ---- Named expression self-reference shadows inside its own body ----
			{Code: `const f = function Symbol() { Symbol(); };`},
			{Code: `const c = class Symbol { m() { Symbol(); } };`},

			// ---- TypeScript enum/class shadows the global value ----
			{Code: `enum Symbol { A } Symbol();`},

			// ---- TypeScript namespace/module with identifier name shadows
			//      (both at file scope and inside a function block).
			{Code: `namespace Symbol {} Symbol();`},
			{Code: `module Symbol {} Symbol();`},
			{Code: `function f() { namespace Symbol {} Symbol(); }`},

			// ---- `declare` value declarations shadow like runtime ones.
			{Code: `declare var Symbol: any; Symbol();`},
			{Code: `declare function Symbol(): any; Symbol();`},
			{Code: `declare const Symbol: any; Symbol();`},

			// ---- Class body computed key & class name — inner class name wins.
			{Code: `class Symbol { [Symbol()]() {} }`},
			{Code: `class X { static E = class Symbol { m() { Symbol(); } }; }`},

			// ---- Declaration merging: interface (type) + const (value) = shadowed.
			{Code: `interface Symbol {} const Symbol = 1; Symbol();`},

			// ---- `import type` — treated as a binding by ts-eslint's scope manager.
			{Code: `import type { Symbol } from "x"; Symbol();`},
			{Code: `import type Symbol from "x"; Symbol();`},
			{Code: `import { type Symbol } from "x"; Symbol();`},

			// ---- `export declare` value forms shadow.
			{Code: `export declare var Symbol: any; Symbol();`},
			{Code: `export declare function Symbol(): any; Symbol();`},
			{Code: `export declare namespace Symbol {} Symbol();`},
			{Code: `declare namespace Symbol {} Symbol();`},
			{Code: `declare class Symbol {} new Symbol(); Symbol();`},

			// ---- Inside a namespace block, local `var`/`function` shadow.
			{Code: `namespace NS { var Symbol = 1; Symbol(); }`},
			{Code: `namespace NS { function Symbol() {} Symbol(); }`},

			// ---- Nested-scope shadowing.
			{Code: `for (const Symbol = foo; ;) { Symbol(); }`},
			{Code: `class C { m() { function Symbol() {} Symbol(); } }`},
			{Code: `const f = () => { class Symbol {} return Symbol(); };`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- From ESLint upstream ----
			{
				Code: `Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expected",
						Message:   "Expected Symbol to have a description.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				// Bare `Symbol = ...` is reassignment, not a declaration — still global.
				Code: `Symbol(); Symbol = function () {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 1},
				},
			},

			// ---- Position / nested-scope assertions ----
			{
				Code: `var foo = Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 11, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: `function f() { return Symbol(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 23, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code: "Symbol(\n);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},

			// ---- Parenthesized callee: ESTree drops parens so ESLint reports;
			//      tsgo keeps them as explicit nodes, so SkipParentheses is required.
			{
				Code: `(Symbol)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 1},
				},
			},
			{
				Code: `((Symbol))();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 1},
				},
			},

			// ---- Optional call — `Symbol?.()` still has zero arguments.
			{
				Code: `Symbol?.();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 1},
				},
			},

			// ---- Comments don't count as arguments.
			{
				Code: `Symbol(/* no description */);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 1},
				},
			},

			// ---- Block-scoped shadow doesn't leak out.
			{
				Code: `{ let Symbol = 1; } Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 21},
				},
			},
			// Nested shadow doesn't affect outer call.
			{
				Code: `function f(Symbol) { Symbol(); } Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 34},
				},
			},
			// IIFE parameter is scoped to the IIFE.
			{
				Code: `(function (Symbol) {})(1); Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 28},
				},
			},

			// ---- Call sites inside various containers, no shadow in scope.
			{
				Code: `const f = () => Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 17},
				},
			},
			{
				Code: `class C { m() { Symbol(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 17},
				},
			},
			{
				Code: `class C { static { Symbol(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 20},
				},
			},

			// ---- Multi-line: report spans from callee start to closing paren.
			{
				Code: "var s = Symbol(\n\n);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 9, EndLine: 3, EndColumn: 2},
				},
			},

			// ---- Call embedded in various expression positions.
			{
				Code: `throw Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 7},
				},
			},
			{
				// Chained MemberExpression around the bare Symbol() call.
				Code: `Symbol().toString();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code: `x || Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 6},
				},
			},
			{
				// Sequence expression via comma operator — inner Symbol() is still a CallExpression.
				Code: `(a, Symbol());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 5},
				},
			},

			// ---- Class bodies: field initializer and computed method key.
			{
				Code: `class C { x = Symbol(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 15},
				},
			},
			{
				Code: `class C { [Symbol()]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 12},
				},
			},
			// Object literal computed property key — different AST path than class.
			{
				Code: `const obj = { [Symbol()]: 1 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 16},
				},
			},

			// ---- Type-only declarations do NOT shadow the runtime value.
			{
				Code: `type Symbol = string; Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 23},
				},
			},
			{
				Code: `interface Symbol {} Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 21},
				},
			},
			// Ambient module with string-literal name does NOT bind `Symbol`.
			{
				Code: `declare module "x" {} Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 23},
				},
			},
			// Type alias followed by runtime use still reports — `type` is type-only.
			{
				Code: `type Symbol = string; var s: Symbol = "x"; Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 44},
				},
			},
			// Decorator position — the CallExpression is still detected.
			{
				Code: `class C { @Symbol() method() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 12},
				},
			},

			// ---- Class field named Symbol does NOT shadow the global.
			{
				Code: `class C { accessor Symbol = 1; m() { Symbol(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 38},
				},
			},

			// ---- Destructuring assignment is not a declaration — global stays.
			{
				Code: `({ Symbol } = obj); Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 21},
				},
			},
			{
				Code: `[Symbol] = arr; Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 17},
				},
			},

			// ---- `declare global { var Symbol }` augments the global type but
			//      the rule still reports — declaration isn't in file scope.
			{
				Code: `declare global { var Symbol: any; } Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 37},
				},
			},

			// ---- Type-position usage doesn't shadow the runtime value.
			{
				Code: `const x: Symbol = null as any; Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 32},
				},
			},
			{
				Code: `const x: { Symbol: any } = {} as any; Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 39},
				},
			},

			// ---- Each call evaluated independently.
			{
				Code: `Symbol(); Symbol("b");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 1},
				},
			},
			{
				Code: `Symbol("a"); Symbol();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected", Line: 1, Column: 14},
				},
			},
		},
	)
}
