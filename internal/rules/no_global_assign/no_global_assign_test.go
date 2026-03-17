package no_global_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoGlobalAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoGlobalAssignRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Lowercase identifier is not a builtin
			{Code: `string = 'hello world';`},

			// Shadowed by var declaration
			{Code: `var String; String = 'test';`},

			// Shadowed by let declaration
			{Code: `let Array; Array = 1;`},

			// Shadowed by function parameter
			{Code: `function foo(String) { String = 'test'; }`},

			// Shadowed by function declaration
			{Code: `function Object() {} Object = 'test';`},

			// Exception option
			{Code: `Object = 0;`, Options: map[string]interface{}{"exceptions": []interface{}{"Object"}}},

			// Read-only usage (not a write reference)
			{Code: `var x = String(123);`},

			// Property access (not an identifier assignment)
			{Code: `var x = Math.PI;`},

			// Not a builtin name
			{Code: `foo = 'bar';`},

			// Shadowed by class declaration
			{Code: `class Array {} Array = 1;`},

			// Var shadows global inside function
			{Code: `function foo() { var Object: any; Object = 1; }`},

			// Block-scoped shadow with let
			{Code: `{ let String: any; String = 'test'; }`},

			// For-loop let variable shadows
			{Code: `for (let Object = 0; Object < 10; Object++) {}`},

			// For-in let variable shadows
			{Code: `for (let String in {}) {}`},

			// For-of let variable shadows
			{Code: `for (let Array of []) {}`},

			// Enum declaration shadows (TS)
			{Code: `enum Promise { A, B } Promise = 1;`},

			// Function expression name shadows inside body
			{Code: `const x = function Number() { Number = 1; };`},

			// Arrow function parameter shadows
			{Code: `const f = (Object: any) => { Object = 1; };`},

			// Catch variable shadows
			{Code: `try {} catch (String) { String = 'x'; }`},

			// Import default shadows
			{Code: `import Array from 'some-module'; Array = 1;`},

			// Var hoisting shadows (assignment before declaration)
			{Code: `Number = 1; var Number: any;`},

			// Shadowed by parameter in constructor
			{Code: `class C { constructor(Object: any) { Object = 1; } }`},

			// Shadowed by parameter in setter
			{Code: `class C { set prop(Array: any) { Array = 1; } }`},

			// Var hoists from for-loop to enclosing function scope
			{Code: `function f() { for (var Object = 0; Object < 1; Object++) {} Object = 99; }`},

			// Var hoists from nested if block to function scope
			{Code: `function f() { Object = 1; if (true) { var Object: any; } }`},

			// Class expression name shadows inside class body
			{Code: `const c = class Object { method() { Object = 1; } };`},

			// Function expression name shadows in nested function within body
			{Code: `const x = function Object() { function inner() { Object = 1; } };`},

			// Var hoists from switch case
			{Code: `function f() { switch(1) { case 1: var Object: any; } Object = 1; }`},

			// Var hoists from try block
			{Code: `function f() { try { var Object: any; } catch(e) {} Object = 1; }`},

			// Var hoists from labeled statement
			{Code: `function f() { label: { var Object: any; } Object = 1; }`},

			// Var hoists from while body
			{Code: `function f() { while(false) { var Object: any; } Object = 1; }`},

			// Var hoists from do-while body
			{Code: `function f() { do { var Object: any; } while(false); Object = 1; }`},

			// Var hoists from for-in body
			{Code: `function f() { for (var x in {}) { var Object: any; } Object = 1; }`},

			// Var hoists from deeply nested blocks
			{Code: `function f() { if (true) { for (;;) { { var Object: any; break; } } } Object = 1; }`},

			// Namespace declaration shadows global
			{Code: `namespace Map { export const x = 1; } Map = 1;`},

			// Const enum shadows global
			{Code: `const enum Set { A, B } Set = 1;`},

			// Declare var shadows global
			{Code: `declare var WeakMap: any; WeakMap = 1;`},

			// Declare function shadows global
			{Code: `declare function Symbol(): any; Symbol = 1;`},

			// Declare class shadows global
			{Code: `declare class Promise {} Promise = 1;`},

			// Let destructuring in for-of is a declaration (not global write)
			{Code: `for (let {Object} of [{}]) {}`},

			// Type assertion write is NOT detected by ESLint scope analysis
			{Code: `((Object as any) as any) = 1;`},

			// Satisfies expression write is NOT detected by ESLint scope analysis
			{Code: `(Object satisfies any) = 1;`},

			// Object destructuring parameter shadows
			{Code: `function f({Object}: any) { Object = 1; }`},

			// Array destructuring parameter shadows
			{Code: `function f([Object]: any) { Object = 1; }`},

			// Renamed destructuring parameter shadows
			{Code: `function f({a: Object}: any) { Object = 1; }`},

			// Destructuring parameter with default shadows
			{Code: `function f({Object = 0}: any) { Object = 1; }`},

			// Nested destructuring parameter shadows
			{Code: `function f({a: {b: Object}}: any) { Object = 1; }`},

			// Arrow with destructuring parameter shadows
			{Code: `const f = ({Object}: any) => { Object = 1; };`},

			// Method with destructuring parameter shadows
			{Code: `class C { method({Object}: any) { Object = 1; } }`},

			// Destructured catch variable shadows
			{Code: `try {} catch ({Object}: any) { Object = 1; }`},

			// Var with object destructuring shadows
			{Code: `var {Object}: any = {}; Object = 1;`},

			// Let with array destructuring shadows
			{Code: `let [Array]: any = []; Array = 1;`},

			// Var with renamed destructuring shadows
			{Code: `var {a: Map}: any = {}; Map = 1;`},

			// Var with nested destructuring shadows
			{Code: `var {a: {b: Set}}: any = {}; Set = 1;`},

			// Hoisted var destructuring from nested block
			{Code: `function f() { if (true) { var {Object}: any = {}; } Object = 1; }`},

			// Hoisted var array destructuring from nested block
			{Code: `function f() { if (true) { var [Array]: any = []; } Array = 1; }`},

			// Hoisted var destructuring from switch
			{Code: `function f() { switch(1) { case 1: var {Object}: any = {}; } Object = 1; }`},

			// Import equals declaration shadows
			{Code: `import Object = require('some-module'); Object = 1;`},

			// Hoisted var destructuring in for-of
			{Code: `function f() { for (var {Object}: any of [{}]) {} Object = 1; }`},

			// Multiple var declarations in one statement shadow both
			{Code: `var Object: any, Array: any; Object = 1; Array = 1;`},

			// Closure accessing outer var
			{Code: `function f() { var Object: any; function inner() { Object = 1; } }`},

			// Var after inner function still shadows (hoisting)
			{Code: `function f() { function inner() { Object = 1; } var Object: any; }`},

			// Constructor parameter property shadows
			{Code: `class C { constructor(public Object: any) { Object = 1; } }`},

			// Declare enum shadows
			{Code: `declare enum Promise { A, B } Promise = 1;`},

			// Nested satisfies + as is not a real write
			{Code: `((Object satisfies any) as any) = 1;`},

			// Nested as + satisfies is not a real write
			{Code: `((Object as any) satisfies any) = 1;`},

			// Double satisfies is not a real write
			{Code: `((Object satisfies any) satisfies any) = 1;`},

			// Non-null wrapped in satisfies is not a real write
			{Code: `((Object!) satisfies any) = 1;`},

			// Satisfies wrapped in non-null is not a real write
			{Code: `((Object satisfies any)!) = 1;`},

			// Type-only import shadows in ESLint scope
			{Code: `import type Object from "test"; Object = 1;`},

			// Named import with alias shadows
			{Code: `import { foo as Object } from "test"; Object = 1;`},

			// Namespace import shadows
			{Code: `import * as Object from "test"; Object = 1;`},

			// Named import shadows
			{Code: `import { Object } from "test"; Object = 1;`},

			// Overloaded function declaration shadows
			{Code: "function Object(x: string): string;\nfunction Object(x: number): number;\nfunction Object(x: any): any { return x; }\nObject = 1;"},

			// Using declaration shadows in same scope
			{Code: "function f() { using Object = { [Symbol.dispose]() {} }; Object = 1; }"},

			// Await using declaration shadows in same scope
			{Code: "async function f() { await using Object = { async [Symbol.asyncDispose]() {} }; Object = 1; }"},

			// Var with computed property destructuring shadows
			{Code: `const k = "x"; var {[k]: Object}: any = {}; Object = 1;`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Direct assignment to builtin
			{
				Code: `String = 'hello world';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Postfix increment on builtin
			{
				Code: `String++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Assignment to Array
			{
				Code: `Array = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Assignment to Number
			{
				Code: `Number = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Compound assignment
			{
				Code: `Math += 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Prefix decrement
			{
				Code: `--Object;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 3},
				},
			},

			// Assignment to undefined
			{
				Code: `undefined = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Assignment to NaN
			{
				Code: `NaN = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Assignment to Infinity
			{
				Code: `Infinity = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Multiple globals assigned
			{
				Code: `String = 1; Array = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 13},
				},
			},

			// Destructuring assignment of global
			{
				Code: `({Object} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 3},
				},
			},

			// Destructuring with default value
			{
				Code: `({Object = 0, String = 0} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 3},
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 15},
				},
			},

			// Array destructuring
			{
				Code: `[String] = ['x'];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 2},
				},
			},

			// Object destructuring with rename
			{
				Code: `({a: Object} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 6},
				},
			},

			// Rest element in destructuring
			{
				Code: `[...Array] = [1, 2];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 5},
				},
			},

			// For-in with global as target
			{
				Code: `for (Object in {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 6},
				},
			},

			// For-of with global as target
			{
				Code: `for (Array of []) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 6},
				},
			},

			// Logical assignment
			{
				Code: `Object ??= {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Inner var does NOT shadow outer scope
			{
				Code: `function f() { Object = 1; function inner() { var Object: any; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 16},
				},
			},

			// Interface does NOT shadow value binding (TS)
			{
				Code: `interface JSON { x: number } JSON = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 30},
				},
			},

			// Type alias does NOT shadow value binding (TS)
			{
				Code: "type RegExp = string; RegExp = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 23},
				},
			},

			// Assignment inside nested function
			{
				Code: `function outer() { Object = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 20},
				},
			},

			// Assignment inside arrow function
			{
				Code: `const f = () => { String = 'x'; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 19},
				},
			},

			// Assignment inside class method
			{
				Code: `class C { method() { Object = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 22},
				},
			},

			// Deeply nested array destructuring
			{
				Code: `[[[String]]] = [[['x']]];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 4},
				},
			},

			// Chained assignment
			{
				Code: `Object = Array = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 10},
				},
			},

			// Function expression name only shadows inside, not outside
			{
				Code: `const x = function Number() {}; Number = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 33},
				},
			},

			// Object destructuring in for-of
			{
				Code: `for ({Object} of [{}]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 7},
				},
			},

			// Array destructuring in for-of
			{
				Code: `for ([Object] of [[1]]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 7},
				},
			},

			// Var in arrow function body does NOT hoist to outer
			{
				Code: `function f() { const fn = () => { var Object: any; }; Object = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 55},
				},
			},

			// Var in class method does NOT hoist to outer
			{
				Code: `function f() { class C { m() { var Object: any; } } Object = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 53},
				},
			},

			// Let does NOT shadow outside for-loop
			{
				Code: `function f() { for (let Object = 0;;) { break; } Object = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 50},
				},
			},

			// Const does NOT shadow outside for-of
			{
				Code: `function f() { for (const Object of []) {} Object = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 44},
				},
			},

			// Nested destructuring in for-of
			{
				Code: `for ({a: {b: Object}} of [{a: {b: 1}}]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 14},
				},
			},

			// Rest element in for-of destructuring
			{
				Code: `for ([...Object] of [[1,2]]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 10},
				},
			},

			// Destructuring with default in for-of
			{
				Code: `for ({Object = 0} of [{}]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 7},
				},
			},

			// Non-null assertion write
			{
				Code: `(Object!) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 2},
				},
			},

			// Nested parenthesized write
			{
				Code: `(((((Object))))) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 6},
				},
			},

			// Var in static block does NOT hoist to enclosing scope
			{
				Code: `class C { static { var Object: any; } } Object = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 41},
				},
			},

			// Assignment in class field initializer
			{
				Code: `class C { field = (Object = 1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 20},
				},
			},

			// Assignment inside comma expression
			{
				Code: `(0, Object = 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 5},
				},
			},

			// Assignment in default parameter value
			{
				Code: `function f(x = (Object = 1)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 17},
				},
			},

			// Assignment in generator function
			{
				Code: `function* gen() { Object = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 19},
				},
			},

			// Assignment in async function
			{
				Code: `async function af() { Object = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 23},
				},
			},

			// IIFE var does NOT hoist out
			{
				Code: `(function() { var Object: any; })(); Object = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 38},
				},
			},

			// Assignment in while condition
			{
				Code: `while (Object = 1) { break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 8},
				},
			},

			// Assignment in for-loop update
			{
				Code: `for (let i = 0; i < 1; Object++) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 24},
				},
			},

			// Module augmentation interface does NOT shadow
			{
				Code: "declare module \"test\" {\n  interface Object { custom: string; }\n}\nObject = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 4, Column: 1},
				},
			},

			// Declare module var does NOT shadow outside
			{
				Code: "declare module \"foo\" {\n  var Object: any;\n}\nObject = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 4, Column: 1},
				},
			},

			// Var inside namespace does NOT hoist to file scope
			{
				Code: "namespace Foo {\n  var Object: any;\n}\nObject = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 4, Column: 1},
				},
			},

			// Declare global var does NOT shadow
			{
				Code: "declare global {\n  var Object: any;\n}\nObject = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 4, Column: 1},
				},
			},

			// Assignment in template expression
			{
				Code: "const s = `${Object = 1}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 14},
				},
			},

			// Assignment in array literal
			{
				Code: `const a = [Object = 1];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 12},
				},
			},

			// Using declaration in nested block does NOT hoist
			{
				Code: "function f() {\n  if (true) {\n    using Object = { [Symbol.dispose]() {} };\n  }\n  Object = 1;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 5, Column: 3},
				},
			},

			// Object rest in destructuring assignment
			{
				Code: `({...Object} = {a: 1});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 6},
				},
			},

			// Mixed object+array destructuring
			{
				Code: `({a: [Object]} = {a: [1]});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 7},
				},
			},

			// Mixed array+object destructuring
			{
				Code: `([{a: Object}] = [{a: 1}]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 7},
				},
			},

			// Skip + rest in array destructuring
			{
				Code: `[, ...Object] = [1, 2, 3];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 7},
				},
			},

			// For-await-of with global target
			{
				Code: `async function f() { for await (Object of []) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 33},
				},
			},

			// Export default with assignment
			{
				Code: `export default Object = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 16},
				},
			},

			// Computed property destructuring
			{
				Code: `const key = "x"; ({[key]: Object} = {x: 1});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 27},
				},
			},

			// Getter body var does NOT hoist to outer function
			{
				Code: `function f() { const o = { get x() { var Object: any; return 1; } }; Object = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 70},
				},
			},

			// Destructuring in for-in: object shorthand
			{
				Code: `for ({Object} in {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 7},
				},
			},

			// Destructuring in for-in: array
			{
				Code: `for ([Object] in {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 7},
				},
			},

			// Destructuring in for-in: rename
			{
				Code: `for ({a: Object} in {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 10},
				},
			},

			// Destructuring in for-in: rest
			{
				Code: `for ([...Object] in {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 10},
				},
			},

			// Object method shorthand body
			{
				Code: `const o = { m() { Object = 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 19},
				},
			},

			// Static method body
			{
				Code: `class C { static m() { Object = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 24},
				},
			},

			// Optional catch binding (no catch parameter)
			{
				Code: `try {} catch { Object = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 16},
				},
			},

			// Arrow in class field
			{
				Code: `class C { prop = () => { Object = 1; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 26},
				},
			},

			// Setter body var does NOT hoist outside
			{
				Code: `function f() { const o = { set x(v: any) { var Object: any; } }; Object = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 66},
				},
			},

			// Await using in nested block does NOT hoist
			{
				Code: "async function f() {\n  if (true) {\n    await using Object = { async [Symbol.asyncDispose]() {} };\n  }\n  Object = 1;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 5, Column: 3},
				},
			},

			// AggregateError is a read-only global
			{
				Code: `AggregateError = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// FinalizationRegistry is a read-only global
			{
				Code: `FinalizationRegistry = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Intl is a read-only global
			{
				Code: `Intl = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Object rest in for-in destructuring
			{
				Code: `for ({...Object} in {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 10},
				},
			},

			// Default value in for-in destructuring
			{
				Code: `for ({Object = 0} in {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 7},
				},
			},

			// Nested destructuring in for-in
			{
				Code: `for ({a: {b: Object}} in {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 14},
				},
			},

			// Ternary with assignments to different globals
			{
				Code: `true ? (String = 1) : (Array = 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 9},
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 24},
				},
			},

			// Inner catch param does NOT shadow outer scope
			{
				Code: `try { try {} catch (Object) {} Object = 1; } catch(e) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 32},
				},
			},

			// Multiple globals in one array destructuring
			{
				Code: `[Object, Array, ...String] = [1, 2, 3, 4];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 2},
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 10},
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 20},
				},
			},

			// Default value in destructuring triggers write to both globals
			{
				Code: `({x: Object = Array = 1} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 6},
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 15},
				},
			},

			// Generator yield assignment
			{
				Code: `function* g() { Object = yield 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 17},
				},
			},

			// Async await assignment
			{
				Code: `async function f() { Object = await Promise.resolve(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 22},
				},
			},

			// Inner class does NOT shadow outer scope write
			{
				Code: `function f() { Object = 1; class Inner { m() { class Object {} } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 16},
				},
			},

			// Iterator is a read-only global
			{
				Code: `Iterator = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// SuppressedError is a read-only global
			{
				Code: `SuppressedError = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// DisposableStack is a read-only global
			{
				Code: `DisposableStack = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},
		},
	)
}
