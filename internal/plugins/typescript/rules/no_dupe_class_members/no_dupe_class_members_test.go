package no_dupe_class_members

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeClassMembersRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDupeClassMembersRule, []rule_tester.ValidTestCase{
		// ─── basic: different names ───
		{Code: `class A { foo() {} bar() {} }`},
		{Code: `class A { foo() {} bar() {} baz() {} }`},
		{Code: `class A { *foo() {} *bar() {} *baz() {} }`},
		{Code: `class A { get foo() {} get bar() {} get baz() {} }`},
		{Code: `class A { 1() {} 2() {} }`},
		{Code: `class A { foo; bar; }`},

		// ─── static vs non-static (same name allowed) ───
		{Code: `class A { static foo() {} foo() {} }`},
		{Code: `class A { foo; static foo; }`},
		{Code: `class A { static foo() {} get foo() {} set foo(value) {} }`},
		{Code: `class A { ['foo']() {} static ['foo']() {} }`},
		{Code: `class A { static get foo() {} get foo() {} }`},
		{Code: `class A { static set foo(v) {} set foo(v) {} }`},
		{Code: `class A { static foo = 1; foo() {} }`},

		// ─── getter + setter pair (allowed) ───
		{Code: `class A { get foo() {} set foo(value) {} }`},
		{Code: `class A { set foo(value) {} get foo() {} }`},
		{Code: `class A { get ['foo']() {} set ['foo'](value) {} }`},
		{Code: `class A { static get foo() {} static set foo(value) {} }`},

		// ─── computed properties ───
		{Code: `class A { [foo]() {} foo() {} }`},
		{Code: `class A { [foo]() {} [foo]() {} }`},
		{Code: `class A { ['foo']() {} ['bar']() {} }`},
		{Code: "class A { [`foo`]() {} [`bar`]() {} }"},
		{Code: `class A { [Symbol.iterator]() {} [Symbol.hasInstance]() {} }`},
		{Code: `class A { [foo + bar]() {} [foo + bar]() {} }`},
		// non-expression computed properties share syntax but resolve to different strings
		{Code: `class A { [1.0]() {} ['1.0']() {} }`},
		{Code: `class A { [0x1]() {} ['0x1']() {} }`},
		// expression with side effects — not statically analyzed
		{Code: `class A { [a++]() {} [a++]() {} }`},

		// ─── constructor keyword ───
		{Code: `class A { ['constructor']() {} constructor() {} }`},
		{Code: `class A { constructor() {} ['constructor']() {} }`},

		// ─── private fields vs public ───
		{Code: `class A { foo; #foo; }`},
		{Code: `class A { '#foo'; #foo; }`},

		// ─── same name in different classes ───
		{Code: "class A { foo() {} }\nclass B { foo() {} }"},

		// ─── TypeScript method overloads ───
		{Code: "class Foo {\n  foo(a: string): string;\n  foo(a: number): number;\n  foo(a: any): any {}\n}"},
		// many overloads before implementation
		{Code: `class A { foo(a: string): void; foo(a: number): void; foo(a: boolean): void; foo(a: any) {} }`},
		// static overloads
		{Code: `class A { static foo(a: string): void; static foo(a: number): void; static foo(a: any) {} }`},
		// overloaded constructor
		{Code: `class A { constructor(a: string); constructor(a: number); constructor(a: any) {} }`},
		// overload signature does not conflict with getter
		{Code: `class A { foo(a: string): void; get foo() { return ''; } }`},

		// ─── abstract methods (no body, like overloads) ───
		{Code: `abstract class A { abstract foo(): void; foo() {} }`},
		{Code: `abstract class A { abstract foo(): void; abstract bar(): void; }`},

		// ─── nested classes: inner members don't conflict with outer ───
		{Code: `class Outer { foo() {} bar() { class Inner { foo() {} } } }`},
		{Code: `class Outer { foo = class Inner { foo() {} }; }`},
		{Code: `class Outer { foo() {} method() { return class { foo() {} }; } }`},
		// deeply nested (3 levels)
		{Code: `class L1 { foo() {} bar() { class L2 { foo() {} baz() { class L3 { foo() {} } } } } }`},
		// class expression in static block
		{Code: `class Outer { static foo() {} static { class Inner { foo() {} } } }`},

		// ─── class with extends (child duplicates only matter within the child) ───
		{Code: `class Base { foo() {} } class Child extends Base { foo() {} }`},

		// ─── empty / single member ───
		{Code: `class A {}`},
		{Code: `class A { foo() {} }`},

		// ─── mixed member types (no conflict) ───
		{Code: `class A { get foo() {} set foo(v) {} bar() {} baz; }`},

		// ─── semicolon class elements are silently ignored ───
		{Code: `class A { ; foo() {} ; bar() {} ; }`},

		// ─── async / generator are still methods ───
		{Code: `class A { async foo() {} bar() {} }`},
		{Code: `class A { *foo() {} bar() {} }`},
		{Code: `class A { async *foo() {} bar() {} }`},

		// ─── __proto__ as single member (no collision) ───
		{Code: `class A { __proto__() {} }`},

		// ─── property with function-expression initializer (single, no dup) ───
		{Code: `class A { foo = () => {} }`},
		{Code: `class A { foo = function() {} }`},

		// ─── computed + identifier mixed accessor pair (allowed) ───
		{Code: `class A { get ['foo']() {} set foo(v) {} }`},
		{Code: `class A { get foo() {} set ['foo'](v) {} }`},

		// ─── index signature does not interfere ───
		{Code: `class A { [key: string]: any; foo() {} bar() {} }`},

		// ─── dup getter then setter is OK (dup getter already reported, setter still pairs) ───
		{Code: `class A { get foo() {} set foo(v) {} }`},

		// ─── declare property alone (no dup) ───
		{Code: `class A { declare foo: string; }`},

		// ─── abstract getter/setter: no body → must be skipped like overloads ───
		{Code: `abstract class A { abstract get foo(): string; foo = 1; }`},
		{Code: `abstract class A { abstract set foo(v: string); foo = 1; }`},
		{Code: `abstract class A { abstract get foo(): string; abstract set foo(v: string); }`},

		// ─── constructor: non-static KindConstructor skipped (matches ESLint) ───
		// In ESLint, both constructor() and 'constructor'() have kind="constructor"
		// and are skipped. TypeScript-Go parses them as the same KindConstructor.
		{Code: `class A { 'constructor'() {} }`},
		{Code: `class A { 'constructor'() {} constructor() {} }`},
		{Code: `class A { 'constructor'() {} 'constructor'() {} }`},
		// static + non-static don't conflict
		{Code: `class A { static constructor() {} constructor() {} }`},
	}, []rule_tester.InvalidTestCase{
		// ─── basic duplicates ───
		{
			Code: `class A { foo() {} foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 20},
			},
		},
		{
			Code: `!class A { foo() {} foo() {} };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 21},
			},
		},
		// triple duplicate
		{
			Code: `class A { foo() {} foo() {} foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 20},
				{MessageId: "unexpected", Line: 1, Column: 29},
			},
		},

		// ─── string / numeric literal names ───
		{
			Code: `class A { 'foo'() {} 'foo'() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 22},
			},
		},
		// empty string duplicate
		{
			Code: `class A { ''() {} ['']() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 19},
			},
		},
		// 10 == 1e1
		{
			Code: `class A { 10() {} 1e1() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 19},
			},
		},
		// 0x10 == 16
		{
			Code: `class A { [0x10]() {} 16() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 23},
			},
		},
		// [100] == [1e2]
		{
			Code: `class A { [100]() {} [1e2]() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 22},
			},
		},
		// BigInt [123n] == 123
		{
			Code: `class A { [123n]() {} 123() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 23},
			},
		},

		// ─── computed (static expression) duplicates ───
		{
			Code: `class A { ['foo']() {} ['foo']() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
			},
		},
		{
			Code: "class A { [`foo`]() {} [`foo`]() {} }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
			},
		},
		// identifier + template literal
		{
			Code: "class A { foo() {} [`foo`]() {} }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 20},
			},
		},
		// computed 'constructor' duplicates (not the keyword constructor)
		{
			Code: `class A { ['constructor']() {} ['constructor']() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 32},
			},
		},
		// three different syntaxes, same resolved name
		{
			Code: "class A { foo() {} ['foo']() {} [`foo`]() {} }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 20},
				{MessageId: "unexpected", Line: 1, Column: 33},
			},
		},

		// ─── static duplicates ───
		{
			Code: `class A { static foo() {} static foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 34},
			},
		},
		{
			Code: `class A { static ['foo']() {} static foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 38},
			},
		},

		// ─── method / accessor cross-kind conflicts ───
		// method then getter
		{
			Code: `class A { foo() {} get foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
			},
		},
		// setter then method
		{
			Code: `class A { set foo(value) {} foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 29},
			},
		},
		// getter then method
		{
			Code: `class A { get foo() {} foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
			},
		},
		// method then setter
		{
			Code: `class A { foo() {} set foo(v) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
			},
		},
		// duplicate getters
		{
			Code: `class A { get foo() {} get foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 28},
			},
		},
		// duplicate setters
		{
			Code: `class A { set foo(v) {} set foo(v) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 29},
			},
		},
		// getter+setter pair then method (method conflicts with both)
		{
			Code: `class A { get foo() {} set foo(v) {} foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 38},
			},
		},
		// static getter then static property (cross-kind within static)
		{
			Code: `class A { static get foo() {} static foo = 1; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 38},
			},
		},

		// ─── property conflicts ───
		{
			Code: `class A { foo; foo = 42; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 16},
			},
		},
		// property then method
		{
			Code: `class A { foo; foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 16},
			},
		},
		// method then property
		{
			Code: `class A { foo() {} foo = 1; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 20},
			},
		},
		// property then getter
		{
			Code: `class A { foo = 1; get foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
			},
		},
		// property then setter
		{
			Code: `class A { foo = 1; set foo(v) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
			},
		},
		// property with arrow function initializer duplicate
		{
			Code: `class A { foo = () => {}; foo = () => {}; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 27},
			},
		},

		// ─── nested classes: inner duplicates still reported ───
		{
			Code: `class Outer { bar() {} foo() { class Inner { baz() {} baz() {} } } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 55},
			},
		},
		// class expression with duplicates
		{
			Code: `var x = class { foo() {} foo() {} };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 26},
			},
		},
		// duplicates in both outer and inner class
		{
			Code: `class Outer { foo() {} foo() {} bar() { class Inner { baz() {} baz() {} } } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
				{MessageId: "unexpected", Line: 1, Column: 64},
			},
		},
		// deeply nested: duplicate at innermost level
		{
			Code: `class L1 { a() { class L2 { b() { class L3 { foo() {} foo() {} } } } } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 55},
			},
		},
		// class expression as property initializer with duplicates
		{
			Code: `class Outer { inner = class { foo() {} foo() {} }; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 40},
			},
		},

		// ─── overload edge cases ───
		// overload + implementation + extra implementation = duplicate
		{
			Code: `class A { foo(a: string): void; foo(a: any) {} foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 48},
			},
		},

		// ─── async / generator duplicates ───
		{
			Code: `class A { async foo() {} async foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 32},
			},
		},
		{
			Code: `class A { *foo() {} *foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 22},
			},
		},
		{
			Code: `class A { async *foo() {} async *foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 34},
			},
		},
		// async and non-async with same name
		{
			Code: `class A { foo() {} async foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 26},
			},
		},

		// ─── __proto__ duplicate ───
		{
			Code: `class A { __proto__() {} __proto__() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 26},
			},
		},

		// ─── class with extends: duplicates in child class still detected ───
		{
			Code: `class Base { foo() {} } class Child extends Base { foo() {} foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 61},
			},
		},

		// ─── init before accessor pair → two errors (init taints both get and set) ───
		{
			Code: `class A { foo = 1; get foo() {} set foo(v) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
				{MessageId: "unexpected", Line: 1, Column: 37},
			},
		},

		// ─── multiple independent duplicates (state isolation per name) ───
		{
			Code: `class A { foo() {} bar() {} foo() {} bar() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 29},
				{MessageId: "unexpected", Line: 1, Column: 38},
			},
		},

		// ─── dup getter, then setter: only getter errors, setter still pairs ───
		{
			Code: `class A { get foo() {} get foo() {} set foo(v) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 28},
			},
		},

		// ─── method + dup getter + setter: method taints get, dup getter also errors ───
		{
			Code: `class A { foo() {} get foo() {} get foo() {} set foo(v) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
				{MessageId: "unexpected", Line: 1, Column: 37},
				{MessageId: "unexpected", Line: 1, Column: 50},
			},
		},

		// ─── computed property declaration + method ───
		{
			Code: `class A { ['foo'] = 1; foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
			},
		},

		// ─── declare property + method (declare is still a PropertyDeclaration) ───
		{
			Code: `class A { declare foo: string; foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 32},
			},
		},

		// ─── index signature skipped but duplicates still caught ───
		{
			Code: `class A { [key: string]: any; foo() {} foo() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 40},
			},
		},

		// ─── definite assignment (!) doesn't affect name ───
		{
			Code: `class A { foo!: string; foo!: number; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 25},
			},
		},

		// ─── [null] resolves to "null" ───
		{
			Code: `class A { [null]() {} 'null'() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 23},
			},
		},

		// ─── octal [0o101] resolves to "65" ───
		{
			Code: `class A { [0o101]() {} 65() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 24},
			},
		},

		// ─── binary [0b1010] resolves to "10" ───
		{
			Code: `class A { [0b1010]() {} 10() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 25},
			},
		},

		// ─── static constructor duplicates (static constructor is a method, not the class constructor) ───
		{
			Code: `class A { static constructor() {} static constructor() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 35},
			},
		},
		// static template constructor + static constructor
		{
			Code: "class A { static [`constructor`]() {} static constructor() {} }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 39},
			},
		},
		// static constructor + static string 'constructor'
		{
			Code: `class A { static constructor() {} static 'constructor'() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected", Line: 1, Column: 35},
			},
		},
	})
}
