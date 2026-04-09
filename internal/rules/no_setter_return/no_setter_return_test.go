package no_setter_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSetterReturnRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoSetterReturnRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			//--------------------------------------------------------------
			// General: not a setter
			//--------------------------------------------------------------
			{Code: `function foo() { return 1; }`},
			{Code: `function set(val: any) { return 1; }`},
			{Code: `var foo = function() { return 1; };`},
			{Code: `var foo = function set() { return 1; };`},
			{Code: `var set = function() { return 1; };`},
			{Code: `var set = function set(val: any) { return 1; };`},
			{Code: `var set = (val: any) => { return 1; };`},
			{Code: `var set = (val: any) => 1;`},

			// setters do not affect other functions
			{Code: `({ set a(val) { }}); function foo() { return 1; }`},
			{Code: `({ set a(val) { }}); (function () { return 1; });`},
			{Code: `({ set a(val) { }}); (() => { return 1; });`},
			{Code: `({ set a(val) { }}); (() => 1);`},

			//--------------------------------------------------------------
			// Object literals and classes
			//--------------------------------------------------------------

			// return without a value is allowed
			{Code: `({ set foo(val) { return; } })`},
			{Code: `({ set foo(val) { if (val) { return; } } })`},
			{Code: `class A { set foo(val) { return; } }`},
			{Code: `(class { set foo(val) { if (val) { return; } else { return; } return; } })`},
			{Code: `class A { set foo(val) { try {} catch(e) { return; } } }`},

			// not a setter
			{Code: `({ get foo() { return 1; } })`},
			{Code: `({ get set() { return 1; } })`},
			{Code: `({ set(val: any) { return 1; } })`},
			{Code: `({ set: function(val: any) { return 1; } })`},
			{Code: `({ foo: function set(val: any) { return 1; } })`},
			{Code: `({ set: function set(val: any) { return 1; } })`},
			{Code: `({ set: (val: any) => { return 1; } })`},
			{Code: `({ set: (val: any) => 1 })`},
			{Code: `var set = { foo(val: any) { return 1; } };`},
			{Code: `class A { constructor(val: any) { return; } }`},
			{Code: `class A { get foo() { return 1; } }`},
			{Code: `class A { get set() { return 1; } }`},
			{Code: `class A { set(val: any) { return 1; } }`},
			{Code: `class A { static set(val: any) { return 1; } }`},

			// not returning from the setter
			{Code: `({ set foo(val) { function foo(val: any) { return 1; } } })`},
			{Code: `({ set foo(val) { var foo = function(val: any) { return 1; } } })`},
			{Code: `({ set foo(val) { var foo = (val: any) => { return 1; } } })`},
			{Code: `({ set foo(val) { var foo = (val: any) => 1; } })`},
			{Code: `({ set foo(val = function() { return 1; }) {} })`},
			{Code: `({ set foo(val = (v: any) => 1) {} })`},
			{Code: `(class { set foo(val) { function foo(val: any) { return 1; } } })`},
			{Code: `(class { set foo(val) { var foo = function(val: any) { return 1; } } })`},
			{Code: `(class { set foo(val) { var foo = (val: any) => { return 1; } } })`},
			{Code: `(class { set foo(val) { var foo = (val: any) => 1; } })`},
			{Code: `(class { set foo(val = function() { return 1; }) {} })`},
			{Code: `(class { set foo(val = (v: any) => 1) {} })`},

			// computed property key containing return (not in setter body)
			{Code: `({ set [function() { return 1; } as any](val) {} })`},
			{Code: `(class { set [function() { return 1; } as any](val) {} })`},

			//--------------------------------------------------------------
			// Property descriptors
			//--------------------------------------------------------------

			// return without a value is allowed
			{Code: `Object.defineProperty(foo, 'bar', { set(val) { return; } })`},
			{Code: `Reflect.defineProperty(foo, 'bar', { set(val) { if (val) { return; } } })`},
			{Code: `Object.defineProperties(foo, { bar: { set(val) { try { return; } catch(e){} } } })`},
			{Code: `Object.create(foo, { bar: { set: function(val: any) { return; } } })`},

			// not a setter
			{Code: `var x = { set(val: any) { return 1; } }`},
			{Code: `var x = { foo: { set(val: any) { return 1; } } }`},
			{Code: `Object.defineProperty(foo, 'bar', { value(val: any) { return 1; } })`},
			{Code: `Reflect.defineProperty(foo, 'bar', { value: function set(val: any) { return 1; } })`},
			{Code: `Object.create(foo, { bar: { 'set ': function(val: any) { return 1; } } })`},
			{Code: `Reflect.defineProperty(foo, 'bar', { Set(val: any) { return 1; } })`},
			{Code: `Object.defineProperties(foo, { bar: { value: (val: any) => 1 } })`},
			{Code: `Object.create(foo, { set: { value: function(val: any) { return 1; } } })`},
			{Code: `Object.defineProperty(foo, 'bar', { baz(val: any) { return 1; } })`},
			{Code: `Reflect.defineProperty(foo, 'bar', { get(val: any) { return 1; } })`},
			{Code: `Object.create(foo, { set: function(val: any) { return 1; } } as any)`},
			{Code: `Object.defineProperty(foo, { set: (val: any) => 1 } as any)`},

			// computed property name is a variable, not string "set"
			{Code: `declare var set: any; Object.defineProperties(foo, { bar: { [set](val: any) { return 1; } } })`},

			// not returning from the setter
			{Code: `Object.defineProperty(foo, 'bar', { set(val) { function foo() { return 1; } } })`},
			{Code: `Reflect.defineProperty(foo, 'bar', { set(val) { var foo = function() { return 1; } } })`},
			{Code: `Object.defineProperties(foo, { bar: { set(val) { () => { return 1 }; } } })`},
			{Code: `Object.create(foo, { bar: { set: (val: any) => { (val: any) => 1; } } })`},

			// invalid argument index
			{Code: `Object.defineProperty(foo, 'bar', 'baz' as any, { set(val) { return 1; } })`},
			{Code: `Object.defineProperty(foo, { set(val) { return 1; } } as any, 'bar')`},
			{Code: `Object.defineProperty({ set(val) { return 1; } } as any, foo, 'bar')`},
			{Code: `Reflect.defineProperty(foo, 'bar', 'baz' as any, { set(val) { return 1; } })`},
			{Code: `Reflect.defineProperty(foo, { set(val) { return 1; } } as any, 'bar')`},
			{Code: `Reflect.defineProperty({ set(val) { return 1; } } as any, foo, 'bar')`},
			{Code: `Object.defineProperties(foo, bar, { baz: { set(val) { return 1; } } })`},
			{Code: `Object.defineProperties({ bar: { set(val) { return 1; } } }, foo)`},
			{Code: `Object.create(foo, bar, { baz: { set(val) { return 1; } } })`},
			{Code: `Object.create({ bar: { set(val) { return 1; } } }, foo)`},

			// wrong method name
			{Code: `(Object as any).DefineProperty(foo, 'bar', { set(val) { return 1; } })`},
			{Code: `(Object as any).Create(foo, { bar: { set: function(val: any) { return 1; } } })`},

			// wrong object name
			{Code: `declare var object: any; object.defineProperty(foo, 'bar', { set(val) { return 1; } })`},
			{Code: `declare var reflect: any; reflect.defineProperty(foo, 'bar', { set(val) { if (val) { return 1; } } })`},
			{Code: `declare var object: any; object.create(foo, { bar: { set: function(val: any) { return 1; } } })`},
			// Reflect.defineProperties does not exist as a standard API
			{Code: `(Reflect as any).defineProperties(foo, { bar: { set(val) { try { return 1; } catch(e){} } } })`},

			// global object is shadowed
			{Code: `function f(Object: any) { Object.defineProperties(foo, { bar: { set(val) { try { return 1; } catch(e){} } } }) }`},
			{Code: `function f() { Reflect.defineProperty(foo, 'bar', { set(val) { if (val) { return 1; } } }); var Reflect: any; }`},
			{Code: `function f(Reflect: any) { Reflect.defineProperty(foo, 'bar', { set(val) { return 1; } }) }`},

			//--------------------------------------------------------------
			// Nesting / composition edge cases
			//--------------------------------------------------------------

			// class inside setter — inner method return is NOT the setter's return
			{Code: `({ set a(val) { class Inner { method() { return 1; } } } })`},
			{Code: `class A { set a(val) { const B = class { get x() { return 1; } }; } }`},

			// generator / async function inside setter — inner return NOT flagged
			{Code: `({ set a(val) { function* gen() { return 1; } } })`},
			{Code: `({ set a(val) { async function f() { return 1; } } })`},

			// constructor inside class in setter — inner return NOT flagged
			{Code: `({ set a(val) { class Inner { constructor() { return; } } } })`},

			// getter inside same object — getter return is fine
			{Code: `({ get a() { return 1; }, set a(val) { } })`},

			// nested object with "set" method — NOT in a descriptor context
			{Code: `var x = { outer: { set(val: any) { return 1; } } }`},

			// Object.defineProperty inside setter — inner non-setter return is fine
			{Code: `({ set a(val) { Object.defineProperty(foo, 'bar', { get() { return 1; } }); } })`},

			// map/reduce callback inside setter
			{Code: `({ set a(val) { [1,2,3].map(function(x: any) { return x * 2; }); } })`},
			{Code: `({ set a(val) { [1,2,3].forEach((x) => { return; }); } })`},

			// setter with arrow expression returning non-value (descriptor context only)
			{Code: `({ set: (val: any) => { return 1; } })`},

			// deeply nested non-setter: method "set" in regular object chain
			{Code: `var x = { a: { b: { c: { set(val: any) { return 1; } } } } }`},

			// property descriptor setter with spread — spread doesn't change the detection
			{Code: `Object.defineProperty(foo, 'bar', { ...defaults, value(val: any) { return 1; } } as any)`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			//--------------------------------------------------------------
			// Object literals and classes
			//--------------------------------------------------------------

			// basic tests
			{
				Code: `({ set a(val) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 17},
				},
			},
			{
				Code: `class A { set a(val) { return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 24},
				},
			},
			{
				Code: `class A { static set a(val) { return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 31},
				},
			},
			{
				Code: `(class { set a(val) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 23},
				},
			},

			// any value
			{
				Code: `({ set a(val) { return val; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 17},
				},
			},
			{
				Code: `class A { set a(val) { return undefined; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 24},
				},
			},
			{
				Code: `(class { set a(val) { return null; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 23},
				},
			},
			{
				Code: `({ set a(val) { return this.a; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 17},
				},
			},

			// any location
			{
				Code: `({ set a(val) { if (foo) { return 1; }; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 28},
				},
			},
			{
				Code: `class A { set a(val) { try { return 1; } catch(e) {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 30},
				},
			},
			{
				Code: `(class { set a(val) { while (foo){ if (bar) break; else return 1; } } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 57},
				},
			},

			// multiple invalid in same object/class
			{
				Code: `({ set a(val) { return 1; }, set b(val) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 17},
					{MessageId: "returnsValue", Line: 1, Column: 43},
				},
			},
			{
				Code: `class A { set a(val) { return 1; } set b(val) { return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 24},
					{MessageId: "returnsValue", Line: 1, Column: 49},
				},
			},
			{
				Code: `(class { set a(val) { return 1; } static set b(val) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 23},
					{MessageId: "returnsValue", Line: 1, Column: 55},
				},
			},

			// multiple invalid in same setter
			{
				Code: `({ set a(val) { if(val) { return 1; } else { return 2 }; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 27},
					{MessageId: "returnsValue", Line: 1, Column: 46},
				},
			},
			{
				Code: `class A { set a(val) { switch(val) { case 1: return x; case 2: return y; default: return z } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 46},
					{MessageId: "returnsValue", Line: 1, Column: 64},
					{MessageId: "returnsValue", Line: 1, Column: 83},
				},
			},

			// static setter with multiple returns
			{
				Code: `(class { static set a(val) { if (val > 0) { (this as any)._val = val; return val; } return false; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 71},
					{MessageId: "returnsValue", Line: 1, Column: 85},
				},
			},

			// valid and invalid in the same setter
			{
				Code: `({ set a(val) { if(val) { return 1; } else { return; }; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 27},
				},
			},
			{
				Code: `class A { set a(val) { switch(val) { case 1: return x; case 2: return; default: return z } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 46},
					{MessageId: "returnsValue", Line: 1, Column: 81},
				},
			},
			{
				Code: `(class { static set a(val) { if (val > 0) { (this as any)._val = val; return; } return false; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 81},
				},
			},

			// inner functions do not affect
			{
				Code: `({ set a(val) { function b(){} return b(); } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 32},
				},
			},
			{
				Code: `class A { set a(val) { return () => {}; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 24},
				},
			},
			{
				Code: `(class { set a(val) { function b(){ return 1; } return 2; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 49},
				},
			},
			{
				Code: `({ set a(val) { function b(){ return; } return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 41},
				},
			},
			{
				Code: `class A { set a(val) { var x = function() { return 1; }; return 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 58},
				},
			},
			{
				Code: `(class { set a(val) { var x = () => { return; }; return 2; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 50},
				},
			},

			// other functions do not affect
			{
				Code: `function f(){}; ({ set a(val) { return 1; } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 33},
				},
			},
			{
				Code: `var x = function f(){}; class A { set a(val) { return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 48},
				},
			},
			{
				Code: `var x = () => {}; var A = class { set a(val) { return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 48},
				},
			},

			//--------------------------------------------------------------
			// Property descriptors
			//--------------------------------------------------------------

			// basic tests
			{
				Code: `Object.defineProperty(foo, 'bar', { set(val) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 48},
				},
			},
			{
				Code: `Reflect.defineProperty(foo, 'bar', { set(val) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 49},
				},
			},
			{
				Code: `Object.defineProperties(foo, { baz: { set(val) { return 1; } } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 50},
				},
			},
			{
				Code: `Object.create(null, { baz: { set(val) { return 1; } } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 41},
				},
			},

			// arrow implicit return
			{
				Code: `Object.defineProperty(foo, 'bar', { set: (val: any) => val })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 56},
				},
			},
			{
				Code: `Reflect.defineProperty(foo, 'bar', { set: (val: any) => f(val) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 57},
				},
			},
			{
				Code: `Object.defineProperties(foo, { baz: { set: (val: any) => a + b } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 58},
				},
			},
			{
				Code: `Object.create({}, { baz: { set: (val: any) => (this as any)._val } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 47},
				},
			},

			// mixed valid/invalid in descriptor
			{
				Code: `Object.defineProperty(foo, 'bar', { set(val) { if (val) { return; } return false; }, get(val) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 69},
				},
			},
			{
				Code: `Reflect.defineProperty(foo, 'bar', { set(val) { try { return f(val) } catch (e) { return e }; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 55},
					{MessageId: "returnsValue", Line: 1, Column: 83},
				},
			},
			{
				Code: `Object.defineProperties(foo, { bar: { get(){ return null; }, set(val) { return null; } } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 73},
				},
			},

			// Object.create with 3 returns, 2 errors (bare return in the middle is allowed)
			{
				Code: `Object.create(null, { baz: { set(val) { return (this as any)._val; return; return undefined; } } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 41},
					{MessageId: "returnsValue", Line: 1, Column: 76},
				},
			},

			// multiple invalid in same descriptors object
			{
				Code: `Object.defineProperties(foo, { baz: { set(val) { return 1; } }, bar: { set(val) { return 1; } } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 50},
					{MessageId: "returnsValue", Line: 1, Column: 83},
				},
			},
			{
				Code: `Object.create({}, { baz: { set(val) { return 1; } }, bar: { set: (val: any) => 1 } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 39},
					{MessageId: "returnsValue", Line: 1, Column: 80},
				},
			},

			// bracket notation
			{
				Code: `Object['defineProperty'](foo, 'bar', { set: function bar(val: any) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 70},
				},
			},

			// string literal property name
			{
				Code: "Reflect.defineProperty(foo, 'bar', { 'set'(val) { return 1; } })",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 51},
				},
			},

			// computed property name with string
			{
				Code: "Object.defineProperties(foo, { baz: { ['set'](val) { return 1; } } })",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 54},
				},
			},

			// element access method name with template literal
			{
				Code: "Object[`defineProperties`](foo, { baz: { ['set'](val) { return 1; } } })",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 57},
				},
			},

			// template literal property name
			{
				Code: "Object.create({}, { baz: { [`set`]: (val: any) => { return 1; } } })",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 53},
				},
			},

			// edge cases for global objects - function name doesn't shadow
			{
				Code: `Object.defineProperty(foo, 'bar', { set: function Object(val: any) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 70},
				},
			},
			{
				Code: `Object.defineProperty(foo, 'bar', { set: function(Object: any) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 66},
				},
			},

			// optional chaining
			{
				Code: `Object?.defineProperty(foo, 'bar', { set(val) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 49},
				},
			},
			{
				Code: `(Object?.defineProperty)(foo, 'bar', { set(val) { return 1; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 51},
				},
			},

			//--------------------------------------------------------------
			// Nesting / composition edge cases
			//--------------------------------------------------------------

			// setter inside setter — both returns flagged
			{
				Code: `({ set a(val) { return { set b(v) { return 1; } }; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 17},
					{MessageId: "returnsValue", Line: 1, Column: 37},
				},
			},

			// deeply nested return in setter — still flagged
			{
				Code: `({ set a(val) { if (x) { if (y) { if (z) { return 1; } } } } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 44},
				},
			},

			// try/catch/finally with returns in setter — all value returns flagged
			{
				Code: `class A { set a(val) { try { return 1; } catch(e) { return 2; } finally { } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 30},
					{MessageId: "returnsValue", Line: 1, Column: 53},
				},
			},

			// setter returning type assertion — still a value return
			{
				Code: `class A { set a(val) { return val as any; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 24},
				},
			},

			// setter returning non-null assertion — still a value return
			{
				Code: `class A { set a(val) { return val!; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 24},
				},
			},

			// arrow setter with ternary expression body in descriptor
			{
				Code: `Object.defineProperty(foo, 'bar', { set: (val: any) => val > 0 ? val : -val })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 56},
				},
			},

			// nested Object.defineProperty inside setter — both setter returns flagged
			{
				Code: `Object.defineProperty(foo, 'a', { set(val) { Object.defineProperty(bar, 'b', { set(v) { return 1; } }); return 2; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 89},
					{MessageId: "returnsValue", Line: 1, Column: 105},
				},
			},

			// setter returning result of map (outer return flagged, inner callback NOT)
			{
				Code: `({ set a(val) { return [1,2,3].map(function(x: any) { return x * 2; }); } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 17},
				},
			},

			// class setter inside object literal setter (both flagged)
			{
				Code: `({ set a(val) { return class { static set b(v) { return 1; } }; } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 17},
					{MessageId: "returnsValue", Line: 1, Column: 50},
				},
			},

			// Object.defineProperties with getter + setter — only setter flagged
			{
				Code: `Object.defineProperties(foo, { bar: { get() { return 1; }, set(val) { return 2; } } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 71},
				},
			},

			// setter in anonymous class expression
			{
				Code: `var C = class { set x(val) { return val; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 30},
				},
			},

			// for loop inside setter with return
			{
				Code: `({ set a(val) { for (var i = 0; i < 10; i++) { return i; } } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 48},
				},
			},

			// labeled statement inside setter
			{
				Code: `({ set a(val) { label: { return 1; } } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 26},
				},
			},
		},
	)
}
