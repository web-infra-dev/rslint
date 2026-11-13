package consistent_generic_constructors

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
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
	rule_tester.RunRuleTester(consistentGenericConstructorsRoot(), "tsconfig.json", t, &ConsistentGenericConstructorsRule, []rule_tester.ValidTestCase{
		{Code: "const x: Foo<string> = new Foo();", Options: "constructor", TSConfig: "tsconfig.isolated-declarations.json"},
		{Code: "class C { x: Foo<string> = new Foo(); }", Options: "constructor", TSConfig: "tsconfig.isolated-declarations.json"},
		{Code: "function f(x: Foo<string> = new Foo()) {}", Options: "constructor", TSConfig: "tsconfig.isolated-declarations.json"},
		{Code: "const x: Foo = new Foo<string>();", Options: "type-annotation"},
		{Code: "class C { x: Foo = new Foo<string>(); }", Options: "type-annotation"},
		{Code: "function f(x: Foo = new Foo<string>()) {}", Options: "type-annotation"},
		{Code: "const f = (x: Foo<string> = new Foo()) => x;", Options: "constructor"},
		{Code: "const f = (x = new Foo<string>()) => x;", Options: "type-annotation"},
		{Code: "const { x = new Foo<string>() } = {};", Options: "type-annotation"},
		{Code: "function f({ x = new Foo<string>() }) {}", Options: "type-annotation"},
		{Code: "class C { constructor(public readonly x: Foo<string> = new Foo()) {} }", Options: "constructor"},
		{Code: "class C { constructor(public readonly x = new Foo<string>()) {} }", Options: "type-annotation"},
		{Code: "const x: Foo<string> = new Foo;", Options: "constructor", TSConfig: "tsconfig.isolated-declarations.json"},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const x: Foo<string> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 33}},
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
			Code:     "const x = new Foo<string>();",
			Options:  "type-annotation",
			TSConfig: "tsconfig.isolated-declarations.json",
			Output:   []string{"const x: Foo<string> = new Foo();"},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 28}},
		},
	})
}

// Safety expectations are authored independently of the rule and upstream fixes.
func TestConsistentGenericConstructorsAutofixSafety(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentGenericConstructorsRule, nil, []rule_tester.InvalidTestCase{{
		Code:    "const x: // note< >\nFoo<string> = new Foo(sideEffect());\nafter();",
		Options: "constructor",
		Output:  []string{"const x = new Foo// note< >\n<string>(sideEffect());\nafter();"},
		Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 2, EndColumn: 36}},
	},
		{
			Code:    "const x: /* before */ // after\nFoo<string> = new Foo;",
			Options: "constructor",
			Output:  []string{"const x = new Foo/* before */// after\n<string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 2, EndColumn: 22}},
		},
		{
			Code:    "const x = new Foo<string // tail >\n>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string // tail >\n> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 2, EndColumn: 4}},
		},
		{
			Code:    "const x: // note< >\r\nFoo<string> = new Foo(sideEffect());\r\nafter();",
			Options: "constructor",
			Output:  []string{"const x = new Foo// note< >\r\n<string>(sideEffect());\r\nafter();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 2, EndColumn: 36}},
		},
		{
			Code:    "const x: /* before */ // after\r\nFoo<string> = new Foo;",
			Options: "constructor",
			Output:  []string{"const x = new Foo/* before */// after\r\n<string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 2, EndColumn: 22}},
		},
		{
			Code:    "const x = new Foo<string // tail >\r\n>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string // tail >\r\n> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 2, EndColumn: 4}},
		},
		{
			Code:    "const x: // note< >\rFoo<string> = new Foo(sideEffect());\rafter();",
			Options: "constructor",
			Output:  []string{"const x = new Foo// note< >\r<string>(sideEffect());\rafter();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 2, EndColumn: 36}},
		},
		{
			Code:    "const x: /* before */ // after\rFoo<string> = new Foo;",
			Options: "constructor",
			Output:  []string{"const x = new Foo/* before */// after\r<string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 2, EndColumn: 22}},
		},
		{
			Code:    "const x = new Foo<string // tail >\r>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string // tail >\r> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 2, EndColumn: 4}},
		},
		{
			Code:    "const x: // note< > Foo<string> = new Foo(sideEffect()); after();",
			Options: "constructor",
			Output:  []string{"const x = new Foo// note< > <string>(sideEffect()); after();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 2, EndColumn: 36}},
		},
		{
			Code:    "const x: /* before */ // after Foo<string> = new Foo;",
			Options: "constructor",
			Output:  []string{"const x = new Foo/* before */// after <string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 2, EndColumn: 22}},
		},
		{
			Code:    "const x = new Foo<string // tail > >();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string // tail > > = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 2, EndColumn: 4}},
		},
		{
			Code:    "const x: // note< > Foo<string> = new Foo(sideEffect()); after();",
			Options: "constructor",
			Output:  []string{"const x = new Foo// note< > <string>(sideEffect()); after();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 2, EndColumn: 36}},
		},
		{
			Code:    "const x: /* before */ // after Foo<string> = new Foo;",
			Options: "constructor",
			Output:  []string{"const x = new Foo/* before */// after <string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 2, EndColumn: 22}},
		},
		{
			Code:    "const x = new Foo<string // tail > >();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string // tail > > = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 2, EndColumn: 4}},
		},
		{
			Code:    "const x: Foo<string /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<string /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 41}},
		},
		{
			Code:    "const x = new Foo<string /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 36}},
		},
		{
			Code:    "const x: Foo<string /* < */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<string /* < */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 41}},
		},
		{
			Code:    "const x = new Foo<string /* < */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string /* < */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 36}},
		},
		{
			Code:    "const x: Foo<string /* < > >> */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<string /* < > >> */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 46}},
		},
		{
			Code:    "const x = new Foo<string /* < > >> */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string /* < > >> */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 41}},
		},
		{
			Code:    "const x: Foo<string /* first > */ /* second < */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<string /* first > */ /* second < */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 62}},
		},
		{
			Code:    "const x = new Foo<string /* first > */ /* second < */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string /* first > */ /* second < */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 57}},
		},
		{
			Code:    "const x: Foo<Map<string, Array<number>> /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<Map<string, Array<number>> /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 61}},
		},
		{
			Code:    "const x = new Foo<Map<string, Array<number>> /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<Map<string, Array<number>> /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 56}},
		},
		{
			Code:    "const x: Foo<Map<string /* > */, Array<number /* > */>> /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<Map<string /* > */, Array<number /* > */>> /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 77}},
		},
		{
			Code:    "const x = new Foo<Map<string /* > */, Array<number /* > */>> /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<Map<string /* > */, Array<number /* > */>> /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 72}},
		},
		{
			Code:    "const x: Foo<{ key: \">\"; other: \"<\" } /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<{ key: \">\"; other: \"<\" } /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 59}},
		},
		{
			Code:    "const x = new Foo<{ key: \">\"; other: \"<\" } /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<{ key: \">\"; other: \"<\" } /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 54}},
		},
		{
			Code:    "const x: Foo<`prefix>${string}<suffix` /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<`prefix>${string}<suffix` /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 60}},
		},
		{
			Code:    "const x = new Foo<`prefix>${string}<suffix` /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<`prefix>${string}<suffix` /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 55}},
		},
		{
			Code:    "const x: Foo<string, /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<string, /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 42}},
		},
		{
			Code:    "const x = new Foo<string, /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string, /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 37}},
		},
		{
			Code:    "const x: Foo<string /* > */, number /* < */, /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<string /* > */, number /* < */, /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 66}},
		},
		{
			Code:    "const x = new Foo<string /* > */, number /* < */, /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string /* > */, number /* < */, /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 61}},
		},
		{
			Code:    "const x: Foo<string\n// tail >\n> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<string\n// tail >\n>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 3, EndColumn: 14}},
		},
		{
			Code:    "const x = new Foo<string\n// tail >\n>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string\n// tail >\n> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 3, EndColumn: 4}},
		},
		{
			Code:    "class C { x? = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { x?: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 34}},
		},
		{
			Code:    "class C { x! = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { x!: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 34}},
		},
		{
			Code:    "class C { [key]? = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { [key]?: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 38}},
		},
		{
			Code:    "class C { [key]! = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { [key]!: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 38}},
		},
		{
			Code:    "class C { [keys[index]]? = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { [keys[index]]?: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 46}},
		},
		{
			Code:    "class C { [keys[index]]! = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { [keys[index]]!: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 46}},
		},
		{
			Code:    "class C { accessor x? = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { accessor x?: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 43}},
		},
		{
			Code:    "class C { accessor x! = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { accessor x!: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 43}},
		},
		{
			Code:    "class C { accessor [key]? = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { accessor [key]?: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 47}},
		},
		{
			Code:    "class C { accessor [key]! = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { accessor [key]!: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 47}},
		},
		{
			Code:    "class C { public readonly x? = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { public readonly x?: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 50}},
		},
		{
			Code:    "class C { public readonly x! = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { public readonly x!: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 50}},
		},
		{
			Code:    "class C { static x? = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { static x?: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 41}},
		},
		{
			Code:    "class C { static x! = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { static x!: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 41}},
		},
		{
			Code:    "class C { \"key\"? = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { \"key\"?: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 38}},
		},
		{
			Code:    "class C { \"key\"! = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { \"key\"!: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 38}},
		},
		{
			Code:    "class C { 123? = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { 123?: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 36}},
		},
		{
			Code:    "class C { 123! = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { 123!: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 36}},
		},
		{
			Code:    "class C { #x! = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { #x!: Foo<string> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 35}},
		},
		{
			Code:    "class C { [key] /* > */ ? = new Foo<string /* > */>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { [key] /* > */ ?: Foo<string /* > */> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 55}},
		},
		{
			Code:    "const x: // first\n// second\nFoo<string> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo// first\n// second\n<string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 3, EndColumn: 24}},
		},
		{
			Code:    "const x: // note\nFoo<string /* > */> = new (Foo)(sideEffect());",
			Options: "constructor",
			Output:  []string{"const x = new (Foo)// note\n<string /* > */>(sideEffect());"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 2, EndColumn: 46}},
		},
		{
			Code:    "const x: Foo<string /* > */> = new Foo(), y: // note\nFoo<number> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<string /* > */>(), y = new Foo// note\n<number>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 41}, {MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 43, EndLine: 2, EndColumn: 24}},
		},
		{
			Code:    "class C { x? = new Foo<string /* > */>(); [key]? = new Foo<number>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { x?: Foo<string /* > */> = new Foo(); [key]?: Foo<number> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 42}, {MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 43, EndLine: 1, EndColumn: 70}},
		},
		{
			Code:    "const x:Foo<string>=new Foo();",
			Options: "constructor",
			Output:  []string{"const x=new Foo<string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 30}},
		},
		{
			Code:    "const x=new Foo<string>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<string>=new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 26}},
		},
		{
			Code:    "const x: Foo<(value: string) => number /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<(value: string) => number /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 60}},
		},
		{
			Code:    "const x = new Foo<(value: string) => number /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<(value: string) => number /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 55}},
		},
	})
}

func TestTypeArgumentsRangeRejectsMissingDelimiters(t *testing.T) {
	for _, tt := range []struct{ code, want string }{
		{code: "type X = Foo<string /* > */;"},
		{code: "type X = Foo<>;"},
		{code: "type X = Foo;"},
		{code: "type X = Foo<Bar<string /* > */> /* > */>;", want: "<Bar<string /* > */> /* > */>"},
		{code: "type X = Foo<<T>() => T>;", want: "<<T>() => T>"},
		{code: "type X = Foo< <T>() => T>;", want: "< <T>() => T>"},
		{code: "type X = Foo<<T>() => T /* > */;"},
	} {
		t.Run(tt.code, func(t *testing.T) {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/range.ts", Path: "/range.ts"}, tt.code, core.ScriptKindTS)
			var args *ast.NodeList
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindTypeReference {
					args = node.AsTypeReferenceNode().TypeArguments
					return true
				}
				return node.ForEachChild(visit)
			}
			if !visit(file.AsNode()) {
				t.Fatal("fixture has no type reference")
			}
			span, ok := typeArgumentsRange(file, args)
			if tt.want == "" {
				if ok {
					t.Fatalf("unsafe delimiter range accepted: %q", tt.code[span.Pos():span.End()])
				}
			} else if !ok || tt.code[span.Pos():span.End()] != tt.want {
				t.Fatalf("range = %v, %v; want %q", span, ok, tt.want)
			}
		})
	}
}

func TestConsistentGenericConstructorsAdjacentDelimiters(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentGenericConstructorsRule, nil, []rule_tester.InvalidTestCase{{
		Code:    "const x: Foo<<T>() => T> = new Foo();",
		Options: "constructor",
		Output:  []string{"const x = new Foo<<T>() => T>();"},
		Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 37}},
	},
		{
			Code:    "const x = new Foo<<T>() => T>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<<T>() => T> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 32}},
		},
		{
			Code:    "const x: Foo< <T>() => T> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo< <T>() => T>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 38}},
		},
		{
			Code:    "const x = new Foo< <T>() => T>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo< <T>() => T> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 33}},
		},
		{
			Code:    "const x: Foo</* opening < */<T>() => T /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo</* opening < */<T>() => T /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 60}},
		},
		{
			Code:    "const x = new Foo</* opening < */<T>() => T /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo</* opening < */<T>() => T /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 55}},
		},
		{
			Code:    "const x: Foo<<T extends Map<string, Array<number>>>() => T> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<<T extends Map<string, Array<number>>>() => T>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 72}},
		},
		{
			Code:    "const x = new Foo<<T extends Map<string, Array<number>>>() => T>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<<T extends Map<string, Array<number>>>() => T> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 67}},
		},
		{
			Code:    "const x: Foo<<T = Array<string>>() => T /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<<T = Array<string>>() => T /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 61}},
		},
		{
			Code:    "const x = new Foo<<T = Array<string>>() => T /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<<T = Array<string>>() => T /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 56}},
		},
		{
			Code:    "const x: Foo<<T>() => <U>() => [T, U]> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<<T>() => <U>() => [T, U]>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 51}},
		},
		{
			Code:    "const x = new Foo<<T>() => <U>() => [T, U]>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<<T>() => <U>() => [T, U]> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 46}},
		},
		{
			Code:    "const x: Foo<<T>() => T, string> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<<T>() => T, string>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 45}},
		},
		{
			Code:    "const x = new Foo<<T>() => T, string>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<<T>() => T, string> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 40}},
		},
		{
			Code:    "const x: Foo<Map<string, <T>() => T> /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<Map<string, <T>() => T> /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 58}},
		},
		{
			Code:    "const x = new Foo<Map<string, <T>() => T> /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<Map<string, <T>() => T> /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 53}},
		},
		{
			Code:    "const x: Foo<(<T>() => T)> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<(<T>() => T)>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 39}},
		},
		{
			Code:    "const x = new Foo<(<T>() => T)>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<(<T>() => T)> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 34}},
		},
		{
			Code:    "const x: Foo<<T extends { value: \">\" }>() => T> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<<T extends { value: \">\" }>() => T>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 60}},
		},
		{
			Code:    "const x = new Foo<<T extends { value: \">\" }>() => T>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<<T extends { value: \">\" }>() => T> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 55}},
		},
		{
			Code:    "const x: Foo<<T>() => T, /* tail > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<<T>() => T, /* tail > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 51}},
		},
		{
			Code:    "const x = new Foo<<T>() => T, /* tail > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<<T>() => T, /* tail > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 46}},
		},
		{
			Code:    "const x: Foo<<T>() => Array<Array<T>> /* > */> = new Foo();",
			Options: "constructor",
			Output:  []string{"const x = new Foo<<T>() => Array<Array<T>> /* > */>();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 59}},
		},
		{
			Code:    "const x = new Foo<<T>() => Array<Array<T>> /* > */>();",
			Options: "type-annotation",
			Output:  []string{"const x: Foo<<T>() => Array<Array<T>> /* > */> = new Foo();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 54}},
		},
		{
			Code:    "class C { [key]? = new Foo<<T>() => T /* > */>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { [key]?: Foo<<T>() => T /* > */> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 50}},
		},
		{
			Code:    "class C { accessor x! = new Foo<<T>() => T>(); }",
			Options: "type-annotation",
			Output:  []string{"class C { accessor x!: Foo<<T>() => T> = new Foo(); }"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 11, EndLine: 1, EndColumn: 47}},
		},
		{
			Code:    "function f(x = new Foo<<T>() => T>()) {}",
			Options: "type-annotation",
			Output:  []string{"function f(x: Foo<<T>() => T> = new Foo()) {}"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 12, EndLine: 1, EndColumn: 37}},
		},
		{
			Code:    "function f(x: Foo<<T>() => T> = new Foo()) {}",
			Options: "constructor",
			Output:  []string{"function f(x = new Foo<<T>() => T>()) {}"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 12, EndLine: 1, EndColumn: 42}},
		},
	})
}

// Keep the compiler option local to this rule's fixtures rather than changing
// the shared TypeScript fixture configuration.
func consistentGenericConstructorsRoot() rule_tester.Root {
	root := fixtures.GetRootDir()
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "tsconfig.inherited-isolated.json"):    `{"extends":"./tsconfig.isolated-declarations.json"}`,
		tspath.ResolvePath(root.Dir, "tsconfig.overridden-isolated.json"):   `{"extends":"./tsconfig.isolated-declarations.json","compilerOptions":{"isolatedDeclarations":false}}`,
		tspath.ResolvePath(root.Dir, "tsconfig.not-isolated.json"):          `{"extends":"./tsconfig.json","compilerOptions":{"isolatedDeclarations":false}}`,
		tspath.ResolvePath(root.Dir, "tsconfig.isolated-declarations.json"): `{"extends":"./tsconfig.json","compilerOptions":{"isolatedDeclarations":true,"declaration":true}}`,
	})
	return root
}

// The owning Program provides the compiler option without a type-checker query.
func TestConsistentGenericConstructorsProgramOptions(t *testing.T) {
	if ConsistentGenericConstructorsRule.RequiresTypeInfo {
		t.Fatal("reading compiler options must not require type information")
	}
	root := consistentGenericConstructorsRoot()
	helper := rule_tester.NewProgramHelper(root)
	for _, config := range []struct {
		name     string
		isolated bool
	}{
		{"tsconfig.json", false}, {"tsconfig.not-isolated.json", false}, {"tsconfig.isolated-declarations.json", true},
		{"tsconfig.inherited-isolated.json", true},
		{"tsconfig.overridden-isolated.json", false},
	} {
		for _, style := range []string{"constructor", "type-annotation"} {
			code := "export const x: Foo<string> = new Foo();"
			if style == "type-annotation" {
				code = "export const x = new Foo<string>();"
			}
			raw, file, err := helper.CreateTestProgram(code, "program-options.ts", config.name)
			if err != nil {
				t.Fatal(err)
			}
			if raw.Options().IsolatedDeclarations.IsTrue() != config.isolated {
				t.Fatal("fixture compiler option was not loaded")
			}
			fs := utils.NewOverlayVFS(root.FS, map[string]string{file.FileName(): code})
			host := utils.CreateCompilerHost(root.Dir, fs)
			sourceOnly, err := lintprogram.NewFromRoots(lintprogram.RootOptions{RootFileNames: []string{file.FileName()}, Host: host, CompilerOptions: raw.Options().Clone(), SingleThreaded: true})
			if err != nil {
				t.Fatal(err)
			}
			programs := []*lintprogram.Program{lintprogram.NewFromCompiler(raw), sourceOnly}
			for index, sourceProgram := range programs {
				t.Run(config.name+"/"+style+"/"+[]string{"compiler", "source-only"}[index], func(t *testing.T) {
					if index == 1 && sourceProgram.CanProvideTypeChecker(sourceProgram.GetSourceFile(file.FileName())) {
						t.Fatal("source-only fixture unexpectedly provides a checker")
					}
					var diagnostics []rule.RuleDiagnostic
					linter.LintSingleFile(linter.LintSingleFileOptions{Program: sourceProgram, File: file.FileName(), HasTypeInfo: false,
						GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
							return []rule.ConfiguredRule{{Name: ConsistentGenericConstructorsRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
								if index == 1 && ctx.TypeChecker != nil {
									t.Fatal("source-only rule received a type checker")
								}
								return ConsistentGenericConstructorsRule.Run(ctx, []any{style})
							}}}
						}, Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
					})
					if config.isolated && style == "constructor" {
						if len(diagnostics) != 0 {
							t.Fatalf("required annotation was reported: %v", diagnostics)
						}
						return
					}
					if len(diagnostics) != 1 {
						t.Fatalf("diagnostics=%v, want one", diagnostics)
					}
					wantID := "preferConstructor"
					wantOutput := "export const x = new Foo<string>();"
					if style == "type-annotation" {
						wantID = "preferTypeAnnotation"
						wantOutput = "export const x: Foo<string> = new Foo();"
					}
					if diagnostics[0].Message.Id != wantID {
						t.Fatalf("message=%q, want %q", diagnostics[0].Message.Id, wantID)
					}
					output, _, fixed := linter.ApplyRuleFixes(code, diagnostics)
					if !fixed || output != wantOutput {
						t.Fatalf("output=%q, fixed=%v, want %q", output, fixed, wantOutput)
					}
				})
			}
		}
	}
}

func TestConsistentGenericConstructorsWithoutProgram(t *testing.T) {
	code := "const x: Foo<string> = new Foo();"
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/without-program.ts", Path: "/without-program.ts"}, code, core.ScriptKindTS)
	var diagnostics []rule.RuleDiagnostic
	ctx := rule.RuleContext{SourceFile: file, Comments: rule.NewCommentStore(file)}.WithDiagnosticConsumer(ConsistentGenericConstructorsRule.Name, rule.SeverityError, rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }})
	listeners := ConsistentGenericConstructorsRule.Run(ctx, nil)
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		return node.ForEachChild(visit)
	}
	visit(file.AsNode())
	if len(diagnostics) != 1 || diagnostics[0].Message.Id != "preferConstructor" {
		t.Fatalf("diagnostics=%v", diagnostics)
	}
	output, _, fixed := linter.ApplyRuleFixes(code, diagnostics)
	if !fixed || output != "const x = new Foo<string>();" {
		t.Fatalf("output=%q, fixed=%v", output, fixed)
	}
}
