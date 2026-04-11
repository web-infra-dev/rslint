package no_eval

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEvalRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEvalRule,
		[]rule_tester.ValidTestCase{
			// ================================================================
			// Basic: not eval
			// ================================================================
			{Code: `Eval(foo)`},
			{Code: `setTimeout('foo')`},
			{Code: `setInterval('foo')`},
			{Code: `window.noeval('foo')`},
			// global is not declared in test env (no @types/node) → treated as unknown
			{Code: `global.eval('foo')`},
			{Code: `global.noeval('foo')`},
			{Code: `globalThis.noeval('foo')`},
			{Code: `this.noeval('foo');`},

			// ================================================================
			// Property / member names — NOT references to global eval
			// ================================================================
			// Object literal property name
			{Code: `var obj = { eval: 42 }`},
			{Code: `var obj = { eval: function() {} }`},
			// Object method shorthand
			{Code: `var obj = { eval() { return 1; } }`},
			// Class member names
			{Code: `class A { eval = 42 }`},
			{Code: `class A { eval() { return 1; } }`},
			{Code: `class A { get eval() { return 1; } }`},
			{Code: `class A { set eval(v) {} }`},
			{Code: `class A { static eval() {} }`},
			{Code: `class A { static eval = 42 }`},
			// Destructuring property name (not a reference)
			{Code: `var { eval: e } = obj`},
			{Code: `var { eval: e } = obj; e()`},
			{Code: `function foo({ eval: e }) { e() }`},
			// TypeScript interface / type property name
			{Code: `interface I { eval: string }`},
			{Code: `type T = { eval: string }`},
			// Enum member name
			{Code: `enum E { eval }`},
			// Label name
			{Code: `eval: while(true) { break eval; }`},
			// Import property name: import { eval as foo } — eval is just a source name
			{Code: `import { eval as foo } from 'mod'; foo()`},
			// Export rename target: export { foo as eval } — eval is just the exported alias
			{Code: `var foo = 1; export { foo as eval }`},
			// Re-export: export { eval } from 'mod' — eval is source module name, not local ref
			{Code: `export { eval } from 'mod'`},
			{Code: `export { eval as foo } from 'mod'`},
			{Code: `export { foo as eval } from 'mod'`},

			// ================================================================
			// this.eval — safe contexts (this is NOT global)
			// ================================================================
			// Strict mode: function body directive
			{Code: `function foo() { 'use strict'; this.eval('foo'); }`},
			// Strict mode: source file directive
			{Code: `'use strict'; function foo() { this.eval('foo'); }`},
			// Module: always strict
			{Code: `import x from 'y'; this.eval('foo');`},
			{Code: `import x from 'y'; function foo() { this.eval('foo'); }`},
			{Code: `export {}; () => { this.eval('foo') }`},
			// Object method: this is the object
			{Code: `var obj = {foo: function() { this.eval('foo'); }}`},
			{Code: `var obj = {}; obj.foo = function() { this.eval('foo'); }`},
			// Object getter/setter: this is the object
			{Code: `var obj = { get foo() { return this.eval(); } }`},
			{Code: `var obj = { set foo(v) { this.eval(v); } }`},
			// Arrow inside strict function: inherits strict this
			{Code: `function f() { 'use strict'; () => { this.eval('foo') } }`},
			{Code: `(function f() { 'use strict'; () => { this.eval('foo') } })`},
			{Code: `function f() { 'use strict'; () => () => this.eval('foo') }`},
			// Class methods: this is the instance
			{Code: `class A { foo() { this.eval(); } }`},
			{Code: `class A { static foo() { this.eval(); } }`},
			{Code: `class A { constructor() { this.eval(); } }`},
			// Class method + nested arrow: still instance
			{Code: `class A { foo() { () => this.eval(); } }`},
			{Code: `class A { foo() { () => () => this.eval(); } }`},
			// Class field initializer: this is the instance
			{Code: `class A { field = this.eval(); }`},
			{Code: `class A { field = () => this.eval(); }`},
			// Static block: this is the class
			{Code: `class A { static { this.eval(); } }`},
			// Class extends: function is in strict context due to class
			{Code: `class C extends function () { this.eval('foo'); } {}`},
			// Uppercase variable name → constructor convention, this is instance
			{Code: `var Foo = function() { this.eval('foo') }`},
			{Code: `var MyClass = function() { this.eval('bar') }`},
			// Lowercase → NOT constructor, should still flag (tested in invalid)

			// IIFE — callee of call, this is caller-determined
			{Code: `(function() { this.eval('foo') })()`},
			// .call()/.apply()/.bind() with non-null first arg
			{Code: `(function() { this.eval('foo') }).call(obj)`},
			{Code: `(function() { this.eval('foo') }).apply(obj, [])`},
			{Code: `var f = function() { this.eval('foo') }.bind(obj); f()`},
			// Transparent expressions reaching member-assignment
			{Code: `obj.method = true ? function() { this.eval('foo') } : null`},
			{Code: `obj.method = null || function() { this.eval('foo') }`},
			// IIFE with return — return inside IIFE is transparent
			{Code: `obj.foo = (function() { return function() { this.eval('foo') } })()`},
			// Callback with thisArg: arr.method(callback, thisArg)
			{Code: `arr.forEach(function() { this.eval('foo') }, obj)`},
			{Code: `arr.map(function(x) { return this.eval(x) }, obj)`},
			{Code: `arr.filter(function(x) { return this.eval(x) }, obj)`},
			{Code: `arr.find(function(x) { return this.eval(x) }, obj)`},
			{Code: `arr.findIndex(function(x) { return this.eval(x) }, obj)`},
			{Code: `arr.findLast(function(x) { return this.eval(x) }, obj)`},
			{Code: `arr.findLastIndex(function(x) { return this.eval(x) }, obj)`},
			{Code: `arr.flatMap(function(x) { return this.eval(x) }, obj)`},
			{Code: `arr.some(function(x) { return this.eval(x) }, obj)`},
			{Code: `arr.every(function(x) { return this.eval(x) }, obj)`},
			// Reflect.apply(callback, thisArg, args)
			{Code: `Reflect.apply(function() { this.eval('foo') }, obj, [])`},
			// Array.from(iter, callback, thisArg)
			{Code: `Array.from(iter, function(x) { return this.eval(x) }, obj)`},
			// TS type assertion wrappers with IIFE
			{Code: `(function() { this.eval('foo') } as any)()`},
			{Code: `(<any>function() { this.eval('foo') })()`},

			// ================================================================
			// Shadowing — eval is a local variable, not the global
			// ================================================================
			{Code: `function foo() { var eval = 'foo'; window[eval]('foo') }`},
			{Code: `function foo(eval) { var x = eval }`},
			{Code: `function foo() { let eval = 1; eval }`},
			{Code: `function foo() { const eval = 1; eval }`},
			// Catch clause shadowing (non-call reference — call always flags)
			{Code: `try {} catch(eval) { var x = eval }`},
			// Function declaration shadowing
			{Code: `function eval() {} var x = eval`},
			// Destructuring binding creates local eval
			{Code: `var { a: eval } = obj; var x = eval`},
			// Import creates local eval binding
			{Code: `import { eval } from 'mod'; var x = eval`},

			// ================================================================
			// allowIndirect: true — all indirect forms allowed
			// ================================================================
			{Code: `(0, eval)('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `(0, window.eval)('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `(0, window['eval'])('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `var EVAL = eval; EVAL('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `var EVAL = this.eval; EVAL('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `(function(exe){ exe('foo') })(eval);`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `window.eval('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `window.window.eval('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `window.window['eval']('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `global.eval('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `global.global.eval('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `this.eval('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `function foo() { this.eval('foo') }`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `(0, globalThis.eval)('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `(0, globalThis['eval'])('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `var EVAL = globalThis.eval; EVAL('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `function foo() { globalThis.eval('foo') }`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `globalThis.globalThis.eval('foo');`, Options: map[string]interface{}{"allowIndirect": true}},
			// Optional call is not direct eval
			{Code: `eval?.('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `window?.eval('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
			{Code: `(window?.eval)('foo')`, Options: map[string]interface{}{"allowIndirect": true}},
		},
		[]rule_tester.InvalidTestCase{
			// ================================================================
			// Direct eval calls — always flagged regardless of scope
			// ================================================================
			{
				Code: `eval(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// eval as parameter name — still flagged as direct call
			{
				Code: `function foo(eval) { eval('foo') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			// eval in nested function
			{
				Code: `function foo() { function bar() { eval('x') } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			// eval in arrow function
			{
				Code: `var fn = () => eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			// eval in IIFE
			{
				Code: `(function() { eval('foo') })()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			// eval in async function
			{
				Code: `async function foo() { eval('bar') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			// eval in generator function
			{
				Code: `function* gen() { eval('bar') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},
			// eval in class method
			{
				Code: `class A { foo() { eval('bar') } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},
			// eval in class field initializer
			{
				Code: `class A { field = eval('bar') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},
			// eval in class static block
			{
				Code: `class A { static { eval('bar') } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			// eval in default parameter value
			{
				Code: `function foo(x = eval('bar')) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},

			// ================================================================
			// Direct eval with allowIndirect: true
			// ================================================================
			{
				Code:    `eval(foo)`,
				Options: map[string]interface{}{"allowIndirect": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `eval('foo')`,
				Options: map[string]interface{}{"allowIndirect": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `function foo(eval) { eval('foo') }`,
				Options: map[string]interface{}{"allowIndirect": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},

			// ================================================================
			// Indirect eval — flagged in default mode
			// ================================================================
			{
				Code: `(0, eval)('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5},
				},
			},

			// ================================================================
			// Global object property access: dot notation
			// ================================================================
			{
				Code: `window.eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
				},
			},
			{
				Code: `window.window.eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			// global.eval — global is not declared in test env (no @types/node)
			// so these are valid. Tested via window/globalThis which ARE declared.
			{
				Code: `globalThis.eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			{
				Code: `globalThis.globalThis.eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},

			// ================================================================
			// Global object property access: bracket notation
			// ================================================================
			{
				Code: `window.window['eval']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			{
				Code: `globalThis.globalThis['eval']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
			// Template literal in element access (use window since global is undeclared in test env)
			{
				Code: "window.window[`eval`]('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			// Bracket notation chain: window['window']['eval']
			{
				Code: `window['window']['eval']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			// Mixed chain: bracket then dot
			{
				Code: `window['window'].eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},

			// ================================================================
			// Indirect eval via global object (in comma expression)
			// ================================================================
			{
				Code: `(0, window.eval)('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			{
				Code: `(0, window['eval'])('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			{
				Code: `(0, globalThis.eval)('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			{
				Code: `(0, globalThis['eval'])('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},

			// ================================================================
			// Non-call eval references (various expression positions)
			// ================================================================
			// Variable assignment
			{
				Code: `var EVAL = eval; EVAL('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			// Passed as argument
			{
				Code: `(function(exe){ exe('foo') })(eval);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			// Global object property reference
			{
				Code: `var EVAL = globalThis.eval; EVAL('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
			// Array literal
			{
				Code: `var arr = [eval]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			// Logical expression
			{
				Code: `var fn = eval || function() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			// Ternary expression
			{
				Code: `var fn = true ? eval : null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			// Shorthand property — eval IS a reference to the variable
			{
				Code: `var obj = { eval }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			// Spread element
			{
				Code: `var arr = [...eval]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			// Template expression
			{
				Code: "var s = `${eval}`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			// Computed property key — eval IS a reference
			{
				Code: `var obj = { [eval]: 42 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
				},
			},
			// Tagged template — eval IS a reference
			{
				Code: "eval`template`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// typeof eval — eval IS a reference
			{
				Code: `typeof eval`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
				},
			},
			// Assignment target
			{
				Code: `eval = function() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// for-in target — eval IS a reference
			{
				Code: `for (eval in obj) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			// export { eval } without rename — eval IS a reference
			{
				Code: `export { eval }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			// export { eval as foo } — eval is the local reference
			{
				Code: `export { eval as foo }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},

			// ================================================================
			// this.eval — top level of script (this IS global)
			// ================================================================
			// Dot notation
			{
				Code: `this.eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			// Top-level strict — this is still global in scripts
			{
				Code: `'use strict'; this.eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			// Bracket notation (was missing!)
			{
				Code: `this['eval']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			// Template literal bracket notation
			{
				Code: "this[`eval`]('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},

			// ================================================================
			// this.eval — non-strict function (this defaults to global)
			// ================================================================
			{
				Code: `function foo() { this.eval('foo') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
			// Expression statement is not a directive
			{
				Code: `function foo() { ('use strict'); this.eval; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 39},
				},
			},
			// Lowercase variable name → NOT constructor, this IS default
			{
				Code: `var foo = function() { this.eval('bar') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},
			// Named function → NOT anonymous, constructor heuristic doesn't apply
			{
				Code: `var Foo = function foo() { this.eval('bar') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 33},
				},
			},
			// .call()/.apply() with null/undefined/void/no args → this IS default
			{
				Code: `(function() { this.eval('foo') }).call(null)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				Code: `(function() { this.eval('foo') }).call(undefined)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			{
				Code: `(function() { this.eval('foo') }).call()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			// Returned function from non-IIFE → this IS default
			{
				Code: `function foo() { return function() { this.eval('bar') } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 43},
				},
			},
			// Callback without thisArg → this IS default
			{
				Code: `arr.forEach(function() { this.eval('foo') })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			{
				Code: `arr.flatMap(function(x) { return this.eval(x) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 39},
				},
			},
			// Callback with null thisArg → this IS default
			{
				Code: `arr.forEach(function() { this.eval('foo') }, null)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			// reduce is NOT a thisArg method — second arg is initialValue
			{
				Code: `arr.reduce(function(a, b) { return this.eval(a) ? a : b }, '0')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 41},
				},
			},
			// Async function — same as regular function
			{
				Code: `async function foo() { this.eval('bar') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},
			// Generator function — same as regular function
			{
				Code: `function* gen() { this.eval('bar') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},

			// ================================================================
			// Arrow functions — this inherits from enclosing scope
			// ================================================================
			// Arrow at top level (script) — this is global
			{
				Code: `() => { this.eval('foo'); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
				},
			},
			{
				Code: `() => { 'use strict'; this.eval('foo'); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			{
				Code: `'use strict'; () => { this.eval('foo'); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			// Nested arrows at top level
			{
				Code: `() => { 'use strict'; () => { this.eval('foo'); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36},
				},
			},
			// Arrow inside non-strict function — this inherits function's global this
			{
				Code: `function foo() { () => this.eval('bar') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},
			// Deeply nested arrows in non-strict function
			{
				Code: `function foo() { () => () => this.eval('bar') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			// Arrow in object literal — arrow has NO own this, inherits from outer scope
			{
				Code: `var obj = { foo: () => this.eval('bar') }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},

			// ================================================================
			// this.eval in reference context (not as call)
			// ================================================================
			{
				Code: `var EVAL = this.eval; EVAL('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			{
				Code: `'use strict'; var EVAL = this.eval; EVAL('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},

			// ================================================================
			// Optional chaining — flagged in default mode
			// ================================================================
			{
				Code: `window?.eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			{
				Code: `(window?.eval)('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			{
				Code: `(window?.window).eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},

			// ================================================================
			// Computed property name in class — this is outer scope
			// ================================================================
			{
				Code: `class C { [this.eval('foo')] = 0 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			{
				Code: `'use strict'; class C { [this.eval('foo')] = 0 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},

			// ================================================================
			// Optional chaining on this
			// ================================================================
			{
				Code: `this?.eval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 7},
				},
			},

			// ================================================================
			// new eval() — NewExpression, caught by KindIdentifier
			// ================================================================
			{
				Code: `new eval()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5},
				},
			},

			// ================================================================
			// eval as expression in various non-call contexts
			// ================================================================
			// export default
			{
				Code: `export default eval`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			// Property value (non-shorthand) — eval IS a reference
			{
				Code: `var obj = { key: eval }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			// eval as element access expression (eval['foo'])
			{
				Code: `eval['foo']`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// eval as property access expression (eval.foo)
			{
				Code: `eval.foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Multiple violations
			// ================================================================
			{
				Code: `eval('foo'); window.eval('bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			{
				Code: `eval('a'); eval('b'); eval('c')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 12},
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
		},
	)
}
