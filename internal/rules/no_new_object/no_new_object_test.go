package no_new_object

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewObjectRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewObjectRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Dimension 1: Non-matching AST shapes (not NewExpression + Identifier("Object"))
			// ============================================================

			// Object literal — not a NewExpression
			{Code: `var myObject = {};`},

			// Custom constructor — callee name is not "Object"
			{Code: `var myObject = new CustomObject();`},

			// Object() without new — CallExpression, not NewExpression
			{Code: `var foo = Object("foo");`},
			{Code: `Object();`},

			// Callee is a property access, not a bare Identifier
			{Code: `var foo = new foo.Object()`},
			{Code: `var foo = new foo.bar.Object()`},
			{Code: `var foo = new foo['Object']()`},

			// Callee is not "Object" even if it references Object indirectly
			{Code: `var x = something ? MyClass : Object; var y = new x();`},
			{Code: `var Obj = Object; new Obj();`},

			// Shadowed via rest destructuring
			{Code: `var { ...Object } = foo; new Object();`},

			// Shadowed via ambient (declare) declarations
			{Code: `declare class Object {} new Object();`},
			{Code: `declare function Object(): void; new Object();`},

			// ============================================================
			// Dimension 2: Shadowing — all declaration forms
			// ============================================================

			// var / let / const
			{Code: `var Object = function() {}; new Object();`},
			{Code: `var Object = function Object() {}; new Object();`},
			{Code: `var Object; new Object();`},
			{Code: `let Object = 1; new Object();`},
			{Code: `const Object = null; new Object();`},

			// function / class
			{Code: `function Object() {} new Object();`},
			{Code: `class Object { constructor(){} } new Object();`},

			// TypeScript enum / namespace
			{Code: `enum Object { A, B } new Object();`},
			{Code: `namespace Object { export const x = 1; } new Object();`},

			// ============================================================
			// Dimension 3: Shadowing — parameter variants
			// ============================================================

			{Code: `function bar(Object) { var baz = new Object(); }`},
			{Code: `const f = (Object) => { new Object(); };`},
			{Code: `function f(...Object) { new Object(); }`},
			{Code: `function f({ Object }) { new Object(); }`},
			{Code: `function f([Object]) { new Object(); }`},
			{Code: `function f(Object = 1) { new Object(); }`},
			{Code: `function f({ a: Object }) { new Object(); }`},
			{Code: `function f({ a: { Object } }) { new Object(); }`},

			// ============================================================
			// Dimension 4: Shadowing — destructuring in declarations
			// ============================================================

			{Code: `var { Object } = obj; new Object();`},
			{Code: `var [Object] = arr; new Object();`},
			{Code: `var { a: { Object } } = obj; new Object();`},
			{Code: `var { ["Object"]: Object } = obj; new Object();`},
			{Code: `let { Object } = obj; new Object();`},
			{Code: `const [, Object] = arr; new Object();`},

			// ============================================================
			// Dimension 5: Shadowing — loop variables
			// ============================================================

			{Code: `for (var Object = 0;;) { new Object(); }`},
			{Code: `for (var Object in obj) { new Object(); }`},
			{Code: `for (let Object of arr) { new Object(); }`},
			{Code: `for (const Object of arr) { new Object(); }`},

			// ============================================================
			// Dimension 6: Shadowing — catch clause
			// ============================================================

			{Code: `try {} catch(Object) { new Object(); }`},

			// ============================================================
			// Dimension 7: Shadowing — imports
			// ============================================================

			{Code: `import { Object } from './mod'; new Object();`},
			{Code: `import { X as Object } from './mod'; new Object();`},
			{Code: `import Object from './mod'; new Object();`},
			{Code: `import * as Object from './mod'; new Object();`},

			// ============================================================
			// Dimension 8: Shadowing scope propagation (top-level shadow
			// reaches into nested scopes)
			// ============================================================

			{Code: `function Object() {} function f() { new Object(); }`},
			{Code: `var Object = 1; function f() { new Object(); }`},
			{Code: `var Object = 1; const f = () => new Object();`},
			{Code: `var Object = 1; class C { m() { new Object(); } }`},
			{Code: `var Object = 1; function a() { function b() { function c() { new Object(); } } }`},
			{Code: `var Object = 1; (() => { (() => { new Object(); })(); })();`},
			{Code: `class Object {} class C extends Object { m() { new Object(); } }`},

			// ============================================================
			// Dimension 9: Hoisting — var/function declarations hoist to scope top
			// ============================================================

			{Code: `new Object(); var Object = 1;`},
			{Code: `new Object(); function Object() {}`},
			{Code: `function f() { new Object(); var Object = 1; }`},

			// ============================================================
			// Dimension 10: Shadow inside inner scope, new Object() also inside
			// that same scope (both shadowed — valid)
			// ============================================================

			{Code: `function f(Object) { new Object(); }`},
			{Code: `{ let Object = 1; new Object(); }`},
			{Code: `for (let Object of arr) { new Object(); }`},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Dimension 1: Basic forms
			// ============================================================

			{
				Code:   `var foo = new Object()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 11}},
			},
			{
				Code:   `new Object();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 1}},
			},
			{
				Code:   `const a = new Object()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 11}},
			},
			{
				Code:   `let x = new Object()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 9}},
			},

			// With arguments — ESLint still reports regardless of arguments
			{
				Code:   `var foo = new Object("foo")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 11}},
			},
			{
				Code:   `new Object(1, 2, 3)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 1}},
			},

			// Without parentheses — still a valid NewExpression
			{
				Code:   `new Object`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 1}},
			},

			// Parenthesized callee — SkipOuterExpressions unwraps parens
			{
				Code:   `new (Object)()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 1}},
			},
			{
				Code:   `new ((Object))()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 1}},
			},

			// TS type assertions on callee — SkipOuterExpressions unwraps assertions
			{
				Code:   `new (Object as any)()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 1}},
			},
			{
				Code:   `new (<any>Object)()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 1}},
			},

			// TypeScript generic type argument — callee is still Identifier
			{
				Code:   `new Object<any>()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 1}},
			},

			// ============================================================
			// Dimension 2: Multiple errors in the same file
			// ============================================================

			{
				Code: `new Object(); new Object();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferLiteral", Line: 1, Column: 1},
					{MessageId: "preferLiteral", Line: 1, Column: 15},
				},
			},
			{
				Code: `var x = new Object(), y = new Object();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferLiteral", Line: 1, Column: 9},
					{MessageId: "preferLiteral", Line: 1, Column: 27},
				},
			},

			// ============================================================
			// Dimension 3: Multiline — verify line/column tracking
			// ============================================================

			{
				Code: "var a = 1;\nnew Object();\nvar b = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferLiteral", Line: 2, Column: 1},
				},
			},
			{
				Code: "function f() {\n  return new Object();\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferLiteral", Line: 2, Column: 10},
				},
			},

			// ============================================================
			// Dimension 4: Non-shadowing scope boundaries
			// (declaration inside a scope does NOT shadow outside that scope)
			// ============================================================

			// Nested function expression name doesn't shadow global
			{
				Code:   "function bar() { return function Object() {}; } var baz = new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral", Line: 1, Column: 59}},
			},

			// Block-scoped declarations don't shadow outside the block
			{
				Code:   "{ function Object() {} } new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "{ let Object = 1; } new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "{ const Object = 1; } new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "if (true) { let Object = 1; } new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "try { let Object = 1; } catch(e) {} new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Function/arrow scoped var doesn't shadow outside
			{
				Code:   "function foo() { var Object = 1; } new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "const f = () => { var Object = 1; }; new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "function a() { function b() { var Object = 1; } } new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// IIFE parameters don't shadow outside
			{
				Code:   "(function(Object) {})(1); new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "((Object) => {})(1); new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Class static block scoped var doesn't shadow outside
			{
				Code:   "class C { static { var Object = 1; } } new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Catch clause var doesn't shadow outside
			{
				Code:   "try {} catch(Object) {} new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Loop variable doesn't shadow outside the loop
			{
				Code:   "for (let Object of arr) {} new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "for (const Object in obj) {} new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// ============================================================
			// Dimension 5: new Object() inside various scopes (no shadow anywhere)
			// ============================================================

			{
				Code:   "function f() { new Object(); }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "const f = () => new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "class C { m() { new Object(); } }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "if (true) { if (true) { new Object(); } }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "class C { static { new Object(); } }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Deep nesting without shadow — 3+ levels
			{
				Code:   "function a() { function b() { function c() { new Object(); } } }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "(() => { (() => { (() => { new Object(); })(); })(); })();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// ============================================================
			// Dimension 6: Expression contexts — new Object() as part of
			// larger expressions
			// ============================================================

			// Object literal method
			{
				Code:   "({ m() { new Object() } })",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Getter / setter
			{
				Code:   "({ get x() { return new Object(); } })",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "({ set x(v) { this.v = new Object(); } })",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Computed property key
			{
				Code:   "({ [new Object()]: 1 })",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Default parameter value
			{
				Code:   "function f(x = new Object()) {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "const f = (x = new Object()) => x;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Template literal
			{
				Code:   "`${new Object()}`",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Conditional / logical
			{
				Code:   "condition ? new Object() : {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "var x = foo || new Object()",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "var x = foo ?? new Object()",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Array / spread
			{
				Code:   "[new Object()]",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "[...new Object()]",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// As function argument
			{
				Code:   "foo(new Object())",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "foo(1, new Object(), 'a')",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Destructuring from new Object()
			{
				Code:   "const { a } = new Object()",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Assignment
			{
				Code:   "x = new Object()",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Return / yield / await
			{
				Code:   "function f() { return new Object(); }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "function* g() { yield new Object(); }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},
			{
				Code:   "async function f() { await new Object(); }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Export
			{
				Code:   "export default new Object()",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// ============================================================
			// Dimension 7: Class body contexts
			// ============================================================

			// Instance property initializer
			{
				Code:   "class C { x = new Object() }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Static property initializer
			{
				Code:   "class C { static x = new Object() }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Constructor
			{
				Code:   "class C { constructor() { this.x = new Object(); } }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Extends clause — callee is not "Object", but new Object() in body
			{
				Code:   "class C extends Base { m() { new Object(); } }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// ============================================================
			// Dimension 8: Other statement contexts
			// ============================================================

			// Switch
			{
				Code:   "switch(x) { case 1: new Object(); }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// try / catch / finally — all three positions report
			{
				Code: "try { new Object() } catch(e) { new Object() } finally { new Object() }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferLiteral"},
					{MessageId: "preferLiteral"},
					{MessageId: "preferLiteral"},
				},
			},

			// Labeled statement
			{
				Code:   "label: new Object()",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Comma operator
			{
				Code:   "(0, new Object())",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// TS type assertion on result (not callee) — still a NewExpression
			{
				Code:   "var x = new Object() as any",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Non-null assertion on result
			{
				Code:   "var x = new Object()!",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// Re-export does NOT create a local binding
			{
				Code:   `export { Object } from './mod'; new Object();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// ============================================================
			// Dimension 9: TypeScript — type-level declarations do NOT shadow
			// ============================================================

			// type alias — does not create a value binding
			{
				Code:   "type Object = string; new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// interface — does not create a value binding
			{
				Code:   "interface Object { x: number } new Object();",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferLiteral"}},
			},

			// ============================================================
			// Dimension 9: Mixed scopes — shadow in one, global in another
			// ============================================================

			{
				Code: "function f(Object) { new Object(); } new Object();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferLiteral", Line: 1, Column: 38},
				},
			},
			{
				Code: "{ let Object = 1; new Object(); } new Object();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferLiteral", Line: 1, Column: 35},
				},
			},
			{
				Code: "function f() { let Object = 1; new Object(); } function g() { new Object(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferLiteral", Line: 1, Column: 63},
				},
			},
			// Shadow inside class method, global outside
			{
				Code: "class C { m(Object) { new Object(); } } new Object();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferLiteral", Line: 1, Column: 41},
				},
			},
		},
	)
}
