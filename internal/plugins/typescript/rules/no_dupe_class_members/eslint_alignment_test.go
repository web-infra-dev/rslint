package no_dupe_class_members

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestESLintCoreAlignment verifies 1:1 parity with every ESLint core
// no-dupe-class-members test case (v10.x).
func TestESLintCoreAlignment(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDupeClassMembersRule,

		// ── ESLint core VALID (29 cases) ──
		[]rule_tester.ValidTestCase{
			{Code: `class A { foo() {} bar() {} }`},
			{Code: `class A { static foo() {} foo() {} }`},
			{Code: `class A { get foo() {} set foo(value) {} }`},
			{Code: `class A { static foo() {} get foo() {} set foo(value) {} }`},
			{Code: "class A { foo() {} }\nclass B { foo() {} }"},
			{Code: `class A { [foo]() {} foo() {} }`},
			{Code: `class A { 'foo'() {} 'bar'() {} baz() {} }`},
			{Code: `class A { *'foo'() {} *'bar'() {} *baz() {} }`},
			{Code: `class A { get 'foo'() {} get 'bar'() {} get baz() {} }`},
			{Code: `class A { 1() {} 2() {} }`},
			// computed bracket
			{Code: `class A { ['foo']() {} ['bar']() {} }`},
			{Code: "class A { [`foo`]() {} [`bar`]() {} }"},
			{Code: `class A { [12]() {} [123]() {} }`},
			// numeric format differences
			{Code: `class A { [1.0]() {} ['1.0']() {} }`},
			{Code: "class A { [0x1]() {} [`0x1`]() {} }"},
			{Code: `class A { [null]() {} ['']() {} }`},
			// computed getter+setter
			{Code: `class A { get ['foo']() {} set ['foo'](value) {} }`},
			// computed vs static
			{Code: `class A { ['foo']() {} static ['foo']() {} }`},
			// constructor keyword interactions
			{Code: `class A { ['constructor']() {} constructor() {} }`},
			{Code: "class A { 'constructor'() {} [`constructor`]() {} }"},
			{Code: "class A { constructor() {} get [`constructor`]() {} }"},
			{Code: `class A { 'constructor'() {} set ['constructor'](value) {} }`},
			// non-statically-known computed
			{Code: `class A { ['foo' + '']() {} ['foo']() {} }`},
			{Code: "class A { [`foo${''}`]() {} [`foo`]() {} }"},
			{Code: `class A { [-1]() {} ['-1']() {} }`},
			// dynamic computed
			{Code: `class A { [foo]() {} [foo]() {} }`},
			// private / static fields
			{Code: `class A { foo; static foo; }`},
			{Code: `class A { foo; #foo; }`},
			{Code: `class A { '#foo'; #foo; }`},
		},

		// ── ESLint core INVALID (28 cases) ──
		[]rule_tester.InvalidTestCase{
			// 1. basic
			{
				Code:   `class A { foo() {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 2. class expression
			{
				Code:   `!class A { foo() {} foo() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 3. string literal
			{
				Code:   `class A { 'foo'() {} 'foo'() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 4. numeric equivalence
			{
				Code:   `class A { 10() {} 1e1() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 5. computed bracket
			{
				Code:   `class A { ['foo']() {} ['foo']() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 6. static computed + static literal
			{
				Code:   `class A { static ['foo']() {} static foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 7. set with different syntax
			{
				Code:   `class A { set 'foo'(value) {} set ['foo'](val) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 8. empty string
			{
				Code:   `class A { ''() {} ['']() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 9. template literal
			{
				Code:   "class A { [`foo`]() {} [`foo`]() {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 10. static get template + bracket
			{
				Code:   "class A { static get [`foo`]() {} static get ['foo']() {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 11. identifier + template
			{
				Code:   "class A { foo() {} [`foo`]() {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 12. get template + string literal
			{
				Code:   "class A { get [`foo`]() {} 'foo'() {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 13. static string + static template
			{
				Code:   "class A { static 'foo'() {} static [`foo`]() {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 14. computed 'constructor' duplicate
			{
				Code:   `class A { ['constructor']() {} ['constructor']() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 15. static template constructor + static constructor
			{
				Code:   "class A { static [`constructor`]() {} static constructor() {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 16. static constructor + static string 'constructor'
			{
				Code:   `class A { static constructor() {} static 'constructor'() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 17. computed numeric
			{
				Code:   `class A { [123]() {} [123]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 18. hex + decimal
			{
				Code:   `class A { [0x10]() {} 16() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 19. bracket numeric + scientific
			{
				Code:   `class A { [100]() {} [1e2]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 20. float + template
			{
				Code:   "class A { [123.00]() {} [`123`]() {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 21. static string + static octal
			{
				Code:   `class A { static '65'() {} static [0o101]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 22. BigInt + numeric
			{
				Code:   `class A { [123n]() {} 123() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 23. null keyword + string 'null'
			{
				Code:   `class A { [null]() {} 'null'() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 24. triple duplicate
			{
				Code: `class A { foo() {} foo() {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
					{MessageId: "unexpected"},
				},
			},
			// 25. static duplicate
			{
				Code:   `class A { static foo() {} static foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 26. method then getter
			{
				Code:   `class A { foo() {} get foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 27. setter then method
			{
				Code:   `class A { set foo(value) {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// 28. duplicate properties
			{
				Code:   `class A { foo; foo; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
		},
	)
}

// TestTypeScriptESLintAlignment verifies 1:1 parity with every
// typescript-eslint no-dupe-class-members test case.
func TestTypeScriptESLintAlignment(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDupeClassMembersRule,

		// ── typescript-eslint VALID (11 cases) ──
		[]rule_tester.ValidTestCase{
			{Code: `class A { foo() {} bar() {} }`},
			{Code: `class A { static foo() {} foo() {} }`},
			{Code: `class A { get foo() {} set foo(value) {} }`},
			{Code: `class A { static foo() {} get foo() {} set foo(value) {} }`},
			{Code: "class A { foo() {} }\nclass B { foo() {} }"},
			{Code: `class A { [foo]() {} foo() {} }`},
			{Code: `class A { foo() {} bar() {} baz() {} }`},
			{Code: `class A { *foo() {} *bar() {} *baz() {} }`},
			{Code: `class A { get foo() {} get bar() {} get baz() {} }`},
			{Code: `class A { 1() {} 2() {} }`},
			// TypeScript method overloads (the key TS-eslint addition)
			{Code: "class Foo {\n  foo(a: string): string;\n  foo(a: number): number;\n  foo(a: any): any {}\n}"},
		},

		// ── typescript-eslint INVALID (10 cases) ──
		[]rule_tester.InvalidTestCase{
			{
				Code:   `class A { foo() {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `!class A { foo() {} foo() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `class A { 'foo'() {} 'foo'() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `class A { 10() {} 1e1() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code: `class A { foo() {} foo() {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
					{MessageId: "unexpected"},
				},
			},
			{
				Code:   `class A { static foo() {} static foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `class A { foo() {} get foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `class A { set foo(value) {} foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `class A { foo; foo = 42; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `class A { foo; foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
		},
	)
}
