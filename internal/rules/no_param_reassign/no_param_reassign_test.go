package no_param_reassign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoParamReassignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoParamReassignRule,
		[]rule_tester.ValidTestCase{
			// === Ported from ESLint ===

			// No reassignment
			{Code: `function foo(a) { var b = a; }`},
			{Code: `function foo(a) { for (b in a); }`},
			{Code: `function foo(a) { for (b of a); }`},

			// Property modifications (props: false, default)
			{Code: `function foo(a) { a.prop = 'value'; }`},
			{Code: `function foo(a) { for (a.prop in obj); }`},
			{Code: `function foo(a) { for (a.prop of arr); }`},
			{Code: `function foo(a) { a.b = 0; }`},
			{Code: `function foo(a) { delete a.b; }`},
			{Code: `function foo(a) { ++a.b; }`},

			// Shadowing and globals
			{Code: `function foo(a) { (function() { var a = 12; a++; })(); }`},
			{Code: `function foo() { someGlobal = 13; }`},

			// With props: true - does not flag non-property reads
			{
				Code:    `function foo(a) { bar(a.b).c = 0; }`,
				Options: map[string]interface{}{"props": true},
			},
			{
				Code:    `function foo(a) { data[a.b] = 0; }`,
				Options: map[string]interface{}{"props": true},
			},
			{
				Code:    `function foo(a) { +a.b; }`,
				Options: map[string]interface{}{"props": true},
			},
			{
				Code:    `function foo(a) { (a ? [] : [])[0] = 1; }`,
				Options: map[string]interface{}{"props": true},
			},
			{
				Code:    `function foo(a) { (a.b ? [] : [])[0] = 1; }`,
				Options: map[string]interface{}{"props": true},
			},

			// ignorePropertyModificationsFor
			{
				Code: `function foo(a) { a.b = 0; }`,
				Options: map[string]interface{}{
					"props":                          true,
					"ignorePropertyModificationsFor": []interface{}{"a"},
				},
			},
			{
				Code: `function foo(a) { ++a.b; }`,
				Options: map[string]interface{}{
					"props":                          true,
					"ignorePropertyModificationsFor": []interface{}{"a"},
				},
			},
			{
				Code: `function foo(a) { delete a.b; }`,
				Options: map[string]interface{}{
					"props":                          true,
					"ignorePropertyModificationsFor": []interface{}{"a"},
				},
			},
			{
				Code: `function foo(a) { for (a.b in obj); }`,
				Options: map[string]interface{}{
					"props":                          true,
					"ignorePropertyModificationsFor": []interface{}{"a"},
				},
			},
			{
				Code: `function foo(a) { for (a.b of arr); }`,
				Options: map[string]interface{}{
					"props":                          true,
					"ignorePropertyModificationsFor": []interface{}{"a"},
				},
			},
			{
				Code: `function foo(a) { a.b.c = 0; }`,
				Options: map[string]interface{}{
					"props":                          true,
					"ignorePropertyModificationsFor": []interface{}{"a"},
				},
			},

			// ignorePropertyModificationsForRegex
			{
				Code: `function foo(aFoo) { aFoo.b = 0; }`,
				Options: map[string]interface{}{
					"props":                               true,
					"ignorePropertyModificationsForRegex": []interface{}{"^a.*$"},
				},
			},
			{
				Code: `function foo(aFoo) { ++aFoo.b; }`,
				Options: map[string]interface{}{
					"props":                               true,
					"ignorePropertyModificationsForRegex": []interface{}{"^a.*$"},
				},
			},
			{
				Code: `function foo(aFoo) { delete aFoo.b; }`,
				Options: map[string]interface{}{
					"props":                               true,
					"ignorePropertyModificationsForRegex": []interface{}{"^a.*$"},
				},
			},
			{
				Code: `function foo(aFoo) { aFoo.b.c = 0; }`,
				Options: map[string]interface{}{
					"props":                               true,
					"ignorePropertyModificationsForRegex": []interface{}{"^a.*$"},
				},
			},

			// Destructuring / loop patterns that do not reassign params
			{
				Code:    `function foo(a) { ({ [a]: variable } = value) }`,
				Options: map[string]interface{}{"props": true},
			},
			{Code: `function foo(a) { ([...a.b] = obj); }`},
			{Code: `function foo(a) { ({...a.b} = obj); }`},
			{
				Code:    `function foo(a) { for (obj[a.b] in obj); }`,
				Options: map[string]interface{}{"props": true},
			},
			{
				Code:    `function foo(a) { for (obj[a.b] of arr); }`,
				Options: map[string]interface{}{"props": true},
			},
			{
				Code:    `function foo(a) { for (bar in a.b); }`,
				Options: map[string]interface{}{"props": true},
			},
			{
				Code:    `function foo(a) { for (bar of a.b); }`,
				Options: map[string]interface{}{"props": true},
			},
			{
				Code:    `function foo(a) { for (bar in baz) a.b; }`,
				Options: map[string]interface{}{"props": true},
			},
			{
				Code:    `function foo(a) { for (bar of baz) a.b; }`,
				Options: map[string]interface{}{"props": true},
			},

			// === Edge cases: function-like kinds (all 7) reading their params ===

			{Code: `const f = (a) => a + 1;`},
			{Code: `function* gen(a) { yield a; }`},
			{Code: `async function asyncF(a) { return a; }`},
			{Code: `async function* ag(a) { yield a; }`},
			{Code: `class C { m(a: number) { return a; } }`},
			{Code: `class C { static m(a: number) { return a; } }`},
			{Code: `class C { get x() { return 1; } }`},
			{Code: `class C { set x(v: number) { this._x = v; } }`},
			{Code: `class C { constructor(a: number) { this.a = a; } }`},
			{Code: `const o = { m(a: number) { return a; } };`},
			{Code: `class C { async m(a: number) { return a; } }`},

			// === Edge cases: shadowing — outer param unaffected ===

			// Inner arrow's own param shadows; arrow just reads its own.
			{Code: `function foo(a) { ((a: number) => a + 1)(2); }`},
			// Inner function body reads outer closure.
			{Code: `function foo(a) { function inner() { return a; } }`},
			// IIFE with its own param reads its own param.
			{Code: `function foo(a) { (function(a) { return a; })(a); }`},
			// Class method's own param shadows; just reads it.
			{Code: `function foo(a) { class C { m(a: number) { return a; } } }`},
			// Inner let binding shadows — only reads after let.
			{Code: `function foo(a) { { let a = 1; console.log(a); } }`},
			// Catch clause param shadows; write is to catch binding.
			{Code: `function foo(a) { try {} catch (a) { a; } }`},
			// Catch clause destructured name shadows the param.
			{Code: `function foo(a) { try {} catch ({a}) { a; } }`},
			// for-let creates its own scope.
			{Code: `function foo(a) { for (let a = 0; a < 1; a++) {} }`},
			// Class declaration name shadows.
			{Code: `function foo(A) { class A {} }`},

			// === Edge cases: default values (reads are fine) ===

			{Code: `function foo(a, b = a) { return b; }`},
			{Code: `const x = 1; function foo({a = x}) { return a; }`},
			{Code: `function foo({a, b = a}) { return b; }`},
			{Code: `function foo({a: {b: [c]}}) { return c; }`},

			// === Edge cases: TypeScript syntax ===

			{Code: `function foo<T>(a: T) { return a; }`},
			{Code: `function foo(a?: number) { return a; }`},
			{Code: `function foo(...a: number[]) { return a; }`},
			{Code: `function foo(this: number, a: number) { return a; }`},
			{Code: `class C { constructor(public a: number) { console.log(this.a); } }`},

			// === Edge cases: loop bodies reading param ===

			{Code: `function foo(a) { for (const x of a) { console.log(x); } }`},
			{Code: `function foo(a) { for (const k in a) { console.log(k); } }`},

			// === Edge cases: props: true — non-modifying patterns ===

			{
				Code:    `function foo(a) { a.map(x => x); }`,
				Options: map[string]interface{}{"props": true},
			},
			{
				Code:    `function foo(a) { const o = { [a]: 1 }; }`,
				Options: map[string]interface{}{"props": true},
			},
			{
				Code:    `function foo(a) { return a?.x; }`,
				Options: map[string]interface{}{"props": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			// === Ported from ESLint ===

			// Direct reassignment
			{
				Code: `function foo(bar) { bar = 13; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 21},
				},
			},
			{
				Code: `function foo(bar) { bar += 13; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 21},
				},
			},
			{
				Code: `function foo(bar) { (function() { bar = 13; })(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 35},
				},
			},
			{
				Code: `function foo(bar) { ++bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 23},
				},
			},
			{
				Code: `function foo(bar) { bar++; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 21},
				},
			},
			{
				Code: `function foo(bar) { --bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 23},
				},
			},
			{
				Code: `function foo(bar) { bar--; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 21},
				},
			},

			// Destructured parameters
			{
				Code: `function foo({bar}) { bar = 13; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 23},
				},
			},
			{
				Code: `function foo([, {bar}]) { bar = 13; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 27},
				},
			},
			{
				Code: `function foo(bar) { ({bar} = {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 23},
				},
			},
			{
				Code: `function foo(bar) { ({x: [, bar = 0]} = {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 29},
				},
			},

			// Loop assignment
			{
				Code: `function foo(bar) { for (bar in baz); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 26},
				},
			},
			{
				Code: `function foo(bar) { for (bar of baz); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 26},
				},
			},

			// Property modification with props: true
			{
				Code:    `function foo(bar) { bar.a = 0; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 21},
				},
			},
			{
				Code:    `function foo(bar) { bar.get(0).a = 0; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 21},
				},
			},
			{
				Code:    `function foo(bar) { delete bar.a; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 28},
				},
			},
			{
				Code:    `function foo(bar) { ++bar.a; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 23},
				},
			},
			{
				Code:    `function foo(bar) { for (bar.a in {}); }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 26},
				},
			},
			{
				Code:    `function foo(bar) { for (bar.a of []); }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 26},
				},
			},
			{
				Code:    `function foo(bar) { (bar ? bar : [])[0] = 1; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 28},
				},
			},
			{
				Code:    `function foo(bar) { [bar.a] = []; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 22},
				},
			},

			// Parameter reassignment in destructuring
			{
				Code:    `function foo(a) { ({a} = obj); }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 21},
				},
			},
			{
				Code: `function foo(a) { ([...a] = obj); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 24},
				},
			},
			{
				Code: `function foo(a) { ({...a} = obj); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 24},
				},
			},

			// Spread/rest with property access
			{
				Code:    `function foo(a) { ([...a.b] = obj); }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 24},
				},
			},
			{
				Code:    `function foo(a) { ({...a.b} = obj); }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 24},
				},
			},
			{
				Code:    `function foo(a) { for ([a.b] of []); }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 25},
				},
			},

			// Logical assignment operators
			{
				Code: `function foo(a) { a &&= b; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 19},
				},
			},
			{
				Code: `function foo(a) { a ||= b; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 19},
				},
			},
			{
				Code: `function foo(a) { a ??= b; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 19},
				},
			},
			{
				Code:    `function foo(a) { a.b &&= c; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 19},
				},
			},
			{
				Code:    `function foo(a) { a.b.c ||= d; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 19},
				},
			},
			{
				Code:    `function foo(a) { a[b] ??= c; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 19},
				},
			},

			// ignorePropertyModificationsFor bypass (property not whitelisted)
			{
				Code: `function foo(bar) { [bar.a] = []; }`,
				Options: map[string]interface{}{
					"props":                          true,
					"ignorePropertyModificationsFor": []interface{}{"a"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 22},
				},
			},
			{
				Code: `function foo(bar) { [bar.a] = []; }`,
				Options: map[string]interface{}{
					"props":                               true,
					"ignorePropertyModificationsForRegex": []interface{}{"^B.*$"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 22},
				},
			},

			// === Edge cases: function-like kinds — each flags its own param ===

			// Arrow with expression body
			{
				Code: `const f = (a) => (a = 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 19},
				},
			},
			// Generator
			{
				Code: `function* gen(a) { a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 20},
				},
			},
			// Async function
			{
				Code: `async function af(a) { a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 24},
				},
			},
			// Class method
			{
				Code: `class C { m(a: number) { a = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 26},
				},
			},
			// Class setter
			{
				Code: `class C { set x(v: number) { v = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 30},
				},
			},
			// Constructor
			{
				Code: `class C { constructor(a: number) { a = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 36},
				},
			},
			// Object shorthand method
			{
				Code: `const o = { m(a: number) { a = 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 28},
				},
			},
			// TS parameter property — two symbols share the Parameter decl;
			// rule must still flag via decl comparison.
			{
				Code: `class C { constructor(public a: number) { a = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 43},
				},
			},
			// Private method
			{
				Code: `class C { #m(a: number) { a = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 27},
				},
			},

			// === Edge cases: nested scope — outer param written from inside
			// an inner scope that does NOT shadow ===

			// Arrow closure writes outer param
			{
				Code: `function foo(a) { const f = () => { a = 1; }; f(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 37},
				},
			},
			// Inner fn (no shadow) writes outer param
			{
				Code: `function foo(a) { function inner() { a = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 38},
				},
			},
			// Class method writes outer closure param
			{
				Code: `function foo(a) { class C { m() { a = 1; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 35},
				},
			},
			// Deeply nested, no shadowing intermediate
			{
				Code: `function foo(a) { (function() { (() => { a = 1; })(); })(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 42},
				},
			},

			// === Edge cases: shadowed inner params — reported once, by
			// the inner function's listener, never by the outer one ===

			// Nested function's own param written (bar listener reports)
			{
				Code: `function foo(a) { function bar(a) { a = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 37},
				},
			},
			// Arrow's own param written (arrow listener reports)
			{
				Code: `function foo(a) { ((a: number) => { a = 1; })(2); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 37},
				},
			},
			// Method's own param written (method listener reports)
			{
				Code: `function foo(a) { class C { m(a: number) { a = 1; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 44},
				},
			},
			// `var a` in same scope as param — same binding
			{
				Code: `function foo(a) { var a; a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 26},
				},
			},

			// === Edge cases: multiple reassignments reported independently ===

			{
				Code: `function foo(a) { a = 1; a = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 19},
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 26},
				},
			},
			{
				Code: `function foo(a, b) { ({a, b} = {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 24},
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 27},
				},
			},
			// Two separate inner scopes each write once
			{
				Code: `function foo(a) { (() => { a = 1; })(); (() => { a = 2; })(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 28},
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 50},
				},
			},

			// === Edge cases: parens / type assertions wrapping target ===

			{
				Code: `function foo(a) { (a) = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 20},
				},
			},
			{
				Code: `function foo(a: unknown) { (a as any) = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 29},
				},
			},
			{
				Code: `function foo(a: number | undefined) { a! = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 39},
				},
			},

			// === Edge cases: destructured param reassignment ===

			// Rename destructure: binding is `x`
			{
				Code: `function foo({a: x}) { x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 24},
				},
			},
			// Rest destructure
			{
				Code: `function foo(...rest) { rest = []; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 25},
				},
			},
			// Nested destructure write
			{
				Code: `function foo({a: {b}}) { b = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 26},
				},
			},

			// === Edge cases: default value containing reassignment ===

			// Reassign earlier param in later default
			{
				Code: `function foo(a, b = (a = 1)) { return b; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 22},
				},
			},
			// Inner function in a default value writes outer param
			{
				Code: `function foo(a, b = function() { a = 1; }) { return b; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 34},
				},
			},
			// Inner arrow in a later default writes an earlier param
			{
				Code: `function foo(a, b = () => (a = 1)) { return b; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 28},
				},
			},

			// === Edge cases: property modification (props: true) ===

			// Property modification in nested arrow
			{
				Code:    `function foo(a) { const f = () => { a.x = 1; }; f(); }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 37},
				},
			},
			// Compound on property in method
			{
				Code:    `class C { m(a: any) { a.x += 1; } }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 23},
				},
			},
			// Setter property modification
			{
				Code:    `class C { set x(v: any) { v.prop = 1; } }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 27},
				},
			},
			// Deep property write
			{
				Code:    `function foo(a: any) { a.b.c.d = 1; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 24},
				},
			},
			// Destructured binding — property write via destructuring target
			{
				Code:    `function foo({a}: any) { a.x = 1; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 26},
				},
			},
			// Rest param property write
			{
				Code:    `function foo(...rest: any[]) { rest[0] = 1; }`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 32},
				},
			},

			// === Edge cases: ignore lists combined ===

			// `bar` ignored, `baz` not — flag only baz.
			{
				Code: `function foo(bar, baz) { bar.a = 1; baz.b = 2; }`,
				Options: map[string]interface{}{
					"props":                               true,
					"ignorePropertyModificationsFor":      []interface{}{"bar"},
					"ignorePropertyModificationsForRegex": []interface{}{"^qux"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParamProp", Line: 1, Column: 37},
				},
			},

			// === Edge cases: expression hosts the walker must descend into ===

			// Template literal interpolation
			{
				Code: "function foo(a) { `${a = 1}`; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 22},
				},
			},
			// Tagged template interpolation
			{
				Code: "function foo(a) { String.raw`${a = 1}`; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 32},
				},
			},
			// Comma expression — left side writes
			{
				Code: `function foo(a) { (a = 1, a); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 20},
				},
			},
			// Return statement
			{
				Code: `function foo(a) { return (a = 1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 27},
				},
			},
			// Throw statement
			{
				Code: `function foo(a) { throw (a = 1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 26},
				},
			},
			// Switch case body
			{
				Code: `function foo(a, x) { switch (x) { case 1: a = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 43},
				},
			},
			// Labeled statement
			{
				Code: `function foo(a) { outer: a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 26},
				},
			},
			// Await operand
			{
				Code: `async function foo(a: any) { await (a = 1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 37},
				},
			},
			// Yielded assignment
			{
				Code: `function* foo(a) { yield (a = 1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 27},
				},
			},
			// JSX attribute handler
			{
				Code: `function Comp(a: any) { return <div onClick={() => { a = 1; }} />; }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 1, Column: 54},
				},
			},

			// === Edge cases: TS overloaded function declarations ===
			// Only the implementation signature (with the body) flags.
			{
				Code: `function foo(a: number): void;
function foo(a: string): void;
function foo(a: any) { a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToFunctionParam", Line: 3, Column: 24},
				},
			},
		},
	)
}
