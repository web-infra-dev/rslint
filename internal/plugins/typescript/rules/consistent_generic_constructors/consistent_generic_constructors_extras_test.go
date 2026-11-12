package consistent_generic_constructors

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Rslint regressions beyond the exact upstream suite.
func TestConsistentGenericConstructorsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentGenericConstructorsRule, []rule_tester.ValidTestCase{
		{Code: "class Foo { a: Foo = new Foo<string>(); }"},
		{Code: "class Foo { a: Foo<string> = new Foo<string>(); }"},
		{Code: "class Foo { accessor a: Foo = new Foo<string>(); }"},
		{Code: "class Foo { accessor a: Foo<string> = new Foo<string>(); }"},
		{Code: "function foo(a: Foo<string> = new Foo<string>()) {}"},
		{Code: "function foo(a = new Foo<string>()) {}"},
		{Code: "function foo({ a }: Foo<string> = new Foo<string>()) {}"},
		{Code: "function foo([a]: Foo<string> = new Foo<string>()) {}"},
		{Code: "class A { constructor(a: Foo<string> = new Foo<string>()) {} }"},
		{Code: "const a = function (a: Foo<string> = new Foo<string>()) {};"},
		{Code: "class Foo { a: Foo<string> = new Foo<string>(); }", Options: "type-annotation"},
		{Code: "class Foo { accessor a: Foo<string> = new Foo<string>(); }", Options: "type-annotation"},
		{Code: "function foo(a: Foo<string> = new Foo<string>()) {}", Options: "type-annotation"},
	}, []rule_tester.InvalidTestCase{})
}

func TestConsistentGenericConstructorsRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentGenericConstructorsRule, []rule_tester.ValidTestCase{
		{Code: "const x: Foo<string> = new Foo();", Options: "constructor", LanguageOptions: rule.LanguageOptions{IsolatedDeclarations: true}},
		{Code: "class C { x: Foo<string> = new Foo(); }", Options: "constructor", LanguageOptions: rule.LanguageOptions{IsolatedDeclarations: true}},
		{Code: "function f(x: Foo<string> = new Foo()) {}", Options: "constructor", LanguageOptions: rule.LanguageOptions{IsolatedDeclarations: true}},
		{Code: "const x: Foo = new Foo<string>();", Options: "type-annotation"},
		{Code: "class C { x: Foo = new Foo<string>(); }", Options: "type-annotation"},
		{Code: "function f(x: Foo = new Foo<string>()) {}", Options: "type-annotation"},
		{Code: "const f = (x: Foo<string> = new Foo()) => x;", Options: "constructor"},
		{Code: "const f = (x = new Foo<string>()) => x;", Options: "type-annotation"},
		{Code: "const { x = new Foo<string>() } = {};", Options: "type-annotation"},
		{Code: "function f({ x = new Foo<string>() }) {}", Options: "type-annotation"},
		{Code: "class C { constructor(public readonly x: Foo<string> = new Foo()) {} }", Options: "constructor"},
		{Code: "class C { constructor(public readonly x = new Foo<string>()) {} }", Options: "type-annotation"},
		{Code: "const x: Foo<string> = new Foo;", Options: "constructor", LanguageOptions: rule.LanguageOptions{IsolatedDeclarations: true}},
	}, []rule_tester.InvalidTestCase{
		{
			Code:            "const x: Foo<string> = new Foo();",
			Options:         "constructor",
			LanguageOptions: rule.LanguageOptions{IsolatedDeclarations: false},
			Output:          []string{"const x = new Foo<string>();"},
			Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 33}},
		},
		{
			Code:    "const { x } = new Foo<string>();",
			Options: "type-annotation",
			Output:  []string{"const { x }: Foo<string> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 32}},
		},
		{
			Code:    "const [x] = new Foo<string>();",
			Options: "type-annotation",
			Output:  []string{"const [x]: Foo<string> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 30}},
		},
		{
			Code:    "const x: (Foo<string>) = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 35}},
		},
		{
			Code:    "const x: Foo<string> = (new Foo());",
			Options: "constructor",
			Output:  []string{"const x = (new Foo<string>());"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 35}},
		},
		{
			Code:    "const x = (new Foo<string>());",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string> = (new Foo());"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 30}},
		},
		{
			Code:    "const x: Foo<string> = new (Foo)();",
			Options: "constructor",
			Output:  []string{"const x = new (Foo)<string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 35}},
		},
		{
			Code:    "const x = new (Foo)<string>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string> = new (Foo)();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 30}},
		},
		{
			Code:    "const x: Foo<Map<string, Array<number>>> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<Map<string, Array<number>>>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 53}},
		},
		{
			Code:    "const x = new Foo<Map<string, Array<number>>>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<Map<string, Array<number>>> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 48}},
		},
		{
			Code:    "const x: Foo</* > */ string,> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo</* > */ string,>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 42}},
		},
		{
			Code:    "const x = new Foo</* > */ string,>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo</* > */ string,> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 37}},
		},
		{
			Code:    "class C { accessor [keys[index]] = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { accessor [keys[index]]: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 54}},
		},
		{
			Code:    "const x: /* 😀 */ Foo</* inner */ string> /* tail */ = new Foo;",
			Options: "constructor",
			Output:  []string{"const x /* tail */ = new Foo/* 😀 */</* inner */ string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 63}},
		},
		{
			Code:    "const x: Foo<string> = new Foo(), y: Foo<number> = new Foo;",
			Options: "constructor",
			Output:  []string{"const x = new Foo<string>(), y = new Foo<number>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 33}, {MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 35, EndLine: 1, EndColumn: 59}},
		},
		{
			Code:    "const x = new Foo<string>(), y = new Foo<number>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string> = new Foo(), y: Foo<number> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 28}, {MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 30, EndLine: 1, EndColumn: 51}},
		},
		{
			Code:            "const x = new Foo<string>();",
			Options:         "type-annotation",
			LanguageOptions: rule.LanguageOptions{IsolatedDeclarations: true},
			Output:          []string{"const x: Foo<string> = new Foo();"},
			Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 28}},
		},
	})
}
