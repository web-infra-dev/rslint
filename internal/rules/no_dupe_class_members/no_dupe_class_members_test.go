package no_dupe_class_members

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeClassMembersRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeClassMembersRule,
		[]rule_tester.ValidTestCase{
			// ---- basic: different names ----
			{Code: `class A { foo() {} bar() {} }`},
			{Code: `class A { static foo() {} foo() {} }`},
			{Code: `class A { get foo() {} set foo(value) {} }`},
			{Code: `class A { static foo() {} get foo() {} set foo(value) {} }`},
			{Code: `class A { foo() { } } class B { foo() { } }`},
			{Code: `class A { [foo]() {} foo() {} }`},
			{Code: `class A { 'foo'() {} 'bar'() {} baz() {} }`},
			{Code: `class A { *'foo'() {} *'bar'() {} *baz() {} }`},
			{Code: `class A { get 'foo'() {} get 'bar'() {} get baz() {} }`},
			{Code: `class A { 1() {} 2() {} }`},
			{Code: `class A { ['foo']() {} ['bar']() {} }`},
			{Code: "class A { [`foo`]() {} [`bar`]() {} }"},
			{Code: `class A { [12]() {} [123]() {} }`},
			{Code: `class A { [1.0]() {} ['1.0']() {} }`},
			{Code: `class A { [0x1]() {} [` + "`0x1`" + `]() {} }`},
			{Code: `class A { [null]() {} ['']() {} }`},
			{Code: `class A { get ['foo']() {} set ['foo'](value) {} }`},
			{Code: `class A { ['foo']() {} static ['foo']() {} }`},

			// ---- computed "constructor" key doesn't create constructor ----
			{Code: `class A { ['constructor']() {} constructor() {} }`},
			{Code: "class A { 'constructor'() {} [`constructor`]() {} }"},
			{Code: "class A { constructor() {} get [`constructor`]() {} }"},
			{Code: `class A { 'constructor'() {} set ['constructor'](value) {} }`},

			// ---- not assumed to be statically-known values ----
			{Code: `class A { ['foo' + '']() {} ['foo']() {} }`},
			{Code: "class A { [`foo${''}`]() {} [`foo`]() {} }"},
			{Code: `class A { [-1]() {} ['-1']() {} }`},

			// ---- computed with non-literal expression (ESLint: "not supported by this rule") ----
			{Code: `class A { [foo]() {} [foo]() {} }`},

			// ---- private and public ----
			{Code: `class A { foo; static foo; }`},
			{Code: `class A { foo; #foo; }`},
			{Code: `class A { '#foo'; #foo; }`},

			// ---- TypeScript method overloads (upstream ts parser suite) ----
			{Code: `class Foo { foo(a: string): string; foo(a: number): number; foo(a: any): any {} }`},
			{Code: `class A { foo(a: string): void; foo(a: number): void; foo(a: boolean): void; foo(a: any) {} }`},
			{Code: `class A { static foo(a: string): void; static foo(a: number): void; static foo(a: any) {} }`},
			{Code: `class A { constructor(a: string); constructor(a: number); constructor(a: any) {} }`},
			{Code: `class A { foo(a: string): void; get foo() { return ''; } }`},

			// ---- abstract methods (no body, like overloads) ----
			{Code: `abstract class A { abstract foo(): void; foo() {} }`},
			{Code: `abstract class A { abstract foo(): void; abstract bar(): void; }`},
			{Code: `abstract class A { abstract get foo(): string; foo = 1; }`},
			{Code: `abstract class A { abstract set foo(v: string); foo = 1; }`},
			{Code: `abstract class A { abstract get foo(): string; abstract set foo(v: string); }`},

			// ---- nested classes: inner members don't conflict with outer ----
			{Code: `class Outer { foo() {} bar() { class Inner { foo() {} } } }`},
			{Code: `class Outer { foo = class Inner { foo() {} }; }`},
			{Code: `class L1 { foo() {} bar() { class L2 { foo() {} baz() { class L3 { foo() {} } } } } }`},
			{Code: `class Outer { static foo() {} static { class Inner { foo() {} } } }`},

			// ---- extends does not create conflicts across classes ----
			{Code: `class Base { foo() {} } class Child extends Base { foo() {} }`},

			// ---- empty / single member ----
			{Code: `class A {}`},
			{Code: `class A { foo() {} }`},

			// ---- mixed async / generator / property types ----
			{Code: `class A { async foo() {} bar() {} }`},
			{Code: `class A { *foo() {} bar() {} }`},
			{Code: `class A { async *foo() {} bar() {} }`},
			{Code: `class A { get foo() {} set foo(v) {} bar() {} baz; }`},

			// ---- property with function-expression initializer (single, no dup) ----
			{Code: `class A { foo = () => {} }`},
			{Code: `class A { foo = function() {} }`},

			// ---- computed + identifier mixed accessor pair (allowed) ----
			{Code: `class A { get ['foo']() {} set foo(v) {} }`},

			// ---- index signature does not participate ----
			{Code: `class A { [key: string]: any; foo() {} bar() {} }`},

			// ---- declare property alone (no dup) ----
			{Code: `class A { declare foo: string; }`},

			// ---- non-static string-literal 'constructor' (parsed as constructor — skipped) ----
			{Code: `class A { 'constructor'() {} }`},
			{Code: `class A { 'constructor'() {} constructor() {} }`},
			{Code: `class A { 'constructor'() {} 'constructor'() {} }`},

			// ---- static + non-static constructor don't conflict ----
			{Code: `class A { static constructor() {} constructor() {} }`},

			// ---- Symbol as computed key: dynamic, not deduplicated ----
			{Code: `class A { [Symbol.iterator]() {} [Symbol.hasInstance]() {} }`},

			// ---- expressions with side effects: not statically analyzed ----
			{Code: `class A { [a++]() {} [a++]() {} }`},
			{Code: `class A { [foo + bar]() {} [foo + bar]() {} }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- basic duplicates ----
			{
				Code: `class A { foo() {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20, Message: "Duplicate name 'foo'."},
				},
			},
			{
				Code: `!class A { foo() {} foo() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A { 'foo'() {} 'foo'() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			{
				Code: `class A { 10() {} 1e1() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19, Message: "Duplicate name '10'."},
				},
			},
			{
				Code: `class A { ['foo']() {} ['foo']() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `class A { static ['foo']() {} static foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 38},
				},
			},
			{
				Code: `class A { set 'foo'(value) {} set ['foo'](val) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36},
				},
			},
			{
				Code: `class A { ''() {} ['']() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20, Message: "Duplicate name ''."},
				},
			},
			{
				Code: "class A { [`foo`]() {} [`foo`]() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: "class A { static get [`foo`]() {} static get ['foo']() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 47},
				},
			},
			{
				Code: "class A { foo() {} [`foo`]() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			{
				Code: "class A { get [`foo`]() {} 'foo'() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			{
				Code: "class A { static 'foo'() {} static [`foo`]() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 37},
				},
			},

			// ---- computed 'constructor' duplicates (not the keyword constructor) ----
			{
				Code: `class A { ['constructor']() {} ['constructor']() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 33, Message: "Duplicate name 'constructor'."},
				},
			},
			// Divergence: tsgo parses `static constructor()` and `static 'constructor'()`
			// as KindConstructor (no name node), so we fall back to reporting at the
			// member start (including `static`) rather than ESLint's name position.
			{
				Code: "class A { static [`constructor`]() {} static constructor() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 39},
				},
			},
			{
				Code: `class A { static constructor() {} static 'constructor'() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},

			// ---- numeric literal equivalence ----
			{
				Code: `class A { [123]() {} [123]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23, Message: "Duplicate name '123'."},
				},
			},
			{
				Code: `class A { [0x10]() {} 16() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23, Message: "Duplicate name '16'."},
				},
			},
			{
				Code: `class A { [100]() {} [1e2]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
			{
				Code: "class A { [123.00]() {} [`123`]() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 26},
				},
			},
			{
				Code: `class A { static '65'() {} static [0o101]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36},
				},
			},
			{
				Code: `class A { [123n]() {} 123() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23, Message: "Duplicate name '123'."},
				},
			},
			{
				Code: `class A { [null]() {} 'null'() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23, Message: "Duplicate name 'null'."},
				},
			},

			// ---- triple / multi-duplicates ----
			{
				Code: `class A { foo() {} foo() {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},

			// ---- static duplicates ----
			{
				Code: `class A { static foo() {} static foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 34},
				},
			},

			// ---- method / accessor cross-kind conflicts ----
			{
				Code: `class A { foo() {} get foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			{
				Code: `class A { set foo(value) {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},

			// ---- property declarations ----
			{
				Code: `class A { foo; foo; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},

			// ---- TypeScript overload edge case: two implementations after overloads ----
			{
				Code: `class A { foo(a: string): void; foo(a: any) {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 48},
				},
			},
			// Property + method (TS-specific scenario)
			{
				Code: `class A { foo; foo = 42; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			{
				Code: `class A { foo; foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			// Multi-line class body (position assertion with EndLine / EndColumn)
			{
				Code: "class A {\n  foo() {}\n  foo() {}\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 3, Column: 3, EndLine: 3, EndColumn: 6},
				},
			},

			// ---- async / generator duplicates ----
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
				Code: `class A { foo() {} async foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 26},
				},
			},

			// ---- class expression with duplicates ----
			{
				Code: `var x = class { foo() {} foo() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 26},
				},
			},

			// ---- nested classes: duplicates inside the inner class only ----
			{
				Code: `class Outer { bar() {} foo() { class Inner { baz() {} baz() {} } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 55},
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

			// ---- state isolation: independent names tracked separately ----
			{
				Code: `class A { foo() {} bar() {} foo() {} bar() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
					{MessageId: "unexpected", Line: 1, Column: 38},
				},
			},

			// ---- init-before-accessor-pair: init taints both get and set ----
			{
				Code: `class A { foo = 1; get foo() {} set foo(v) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
					{MessageId: "unexpected", Line: 1, Column: 37},
				},
			},

			// ---- dup getter then setter: only getter errors; setter still pairs ----
			{
				Code: `class A { get foo() {} get foo() {} set foo(v) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},

			// ---- method taints get; dup getter + setter after ----
			{
				Code: `class A { foo() {} get foo() {} get foo() {} set foo(v) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
					{MessageId: "unexpected", Line: 1, Column: 37},
					{MessageId: "unexpected", Line: 1, Column: 50},
				},
			},

			// ---- class with extends: child duplicates still detected ----
			{
				Code: `class Base { foo() {} } class Child extends Base { foo() {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 61},
				},
			},

			// ---- __proto__ as class method is a regular name, duplicates detected ----
			{
				Code: `class A { __proto__() {} __proto__() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 26},
				},
			},

			// ---- property with arrow function initializer duplicate ----
			{
				Code: `class A { foo = () => {}; foo = () => {}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 27},
				},
			},

			// ---- index signature doesn't participate but other duplicates still caught ----
			{
				Code: `class A { [key: string]: any; foo() {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 40},
				},
			},

			// ---- definite assignment (!) doesn't affect name resolution ----
			{
				Code: `class A { foo!: string; foo!: number; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},

			// ---- access modifiers (public/private/protected) don't affect name ----
			{
				Code: `class A { private foo() {} public foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			{
				Code: `class A { protected foo: string; foo: number; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 34},
				},
			},

			// ---- readonly doesn't affect name ----
			{
				Code: `class A { readonly foo: string; foo: number; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 33},
				},
			},

			// ---- optional property doesn't affect name ----
			{
				Code: `class A { foo?: string; foo: number; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},

			// ---- PropertyDeclaration with computed name ----
			{
				Code: `class A { ['foo'] = 1; foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},

			// ---- declare property + method (declare is still a PropertyDeclaration) ----
			{
				Code: `class A { declare foo: string; foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 32},
				},
			},

			// ---- static and non-static each independently duplicated ----
			{
				Code: `class A { foo() {} static foo() {} foo() {} static foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36},
					{MessageId: "unexpected", Line: 1, Column: 52},
				},
			},
		},
	)
}
