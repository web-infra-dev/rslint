package consistent_generic_constructors

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Legacy cases mapped to upstream v8.46.3; the complete suite is mirrored in JS.
func TestConsistentGenericConstructorsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentGenericConstructorsRule, []rule_tester.ValidTestCase{
		{Code: "const a = new Foo();"},
		{Code: "const a = new Foo<string>();"},
		{Code: "const a: Foo<string> = new Foo<string>();"},
		{Code: "const a: Foo = new Foo();"},
		{Code: "const a: Bar<string> = new Foo();"},
		{Code: "const a: Foo = new Foo<string>();"},
		{Code: "const a: Bar = new Foo<string>();"},
		{Code: "const a: Bar<string> = new Foo<string>();"},
		{Code: "const a: Foo<string> = Foo<string>();"},
		{Code: "const a: Foo<string> = Foo();"},
		{Code: "const a: Foo = Foo<string>();"},
		{Code: "class Foo { a = new Foo<string>(); }"},
		{Code: "class Foo { accessor a = new Foo<string>(); }"},
		{Code: "function foo(a: Foo = new Foo<string>()) {}"},
		{Code: "function foo({ a }: Foo = new Foo<string>()) {}"},
		{Code: "function foo([a]: Foo = new Foo<string>()) {}"},
		{Code: "class A { constructor(a: Foo = new Foo<string>()) {} }"},
		{Code: "const a = function (a: Foo = new Foo<string>()) {};"},
		{Code: "const a = new Foo();", Options: "type-annotation"},
		{Code: "const a: Foo<string> = new Foo();", Options: "type-annotation"},
		{Code: "const a: Foo<string> = new Foo<string>();", Options: "type-annotation"},
		{Code: "const a: Foo = new Foo();", Options: "type-annotation"},
		{Code: "const a: Bar = new Foo<string>();", Options: "type-annotation"},
		{Code: "const a: Bar<string> = new Foo<string>();", Options: "type-annotation"},
		{Code: "const a: Foo<string> = Foo<string>();", Options: "type-annotation"},
		{Code: "const a: Foo<string> = Foo();", Options: "type-annotation"},
		{Code: "const a: Foo = Foo<string>();", Options: "type-annotation"},
		{Code: "const a = new (class C<T> {})<string>();", Options: "type-annotation"},
		{Code: "class Foo { a: Foo<string> = new Foo(); }", Options: "type-annotation"},
		{Code: "class Foo { accessor a: Foo<string> = new Foo(); }", Options: "type-annotation"},
		{Code: "function foo(a: Foo<string> = new Foo()) {}", Options: "type-annotation"},
		{Code: "function foo({ a }: Foo<string> = new Foo()) {}", Options: "type-annotation"},
		{Code: "function foo([a]: Foo<string> = new Foo()) {}", Options: "type-annotation"},
		{Code: "class A { constructor(a: Foo<string> = new Foo()) {} }", Options: "type-annotation"},
		{Code: "const a = function (a: Foo<string> = new Foo()) {};", Options: "type-annotation"},
		{Code: "const [a = new Foo<string>()] = [];", Options: "type-annotation"},
		{Code: "function a([a = new Foo<string>()]) {}", Options: "type-annotation"},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const a: Foo<string> = new Foo();",
			Options: "constructor",
			Output:  []string{"const a = new Foo<string>();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 33},
			},
		},
		{
			Code:    "const a: Map<string, number> = new Map();",
			Options: "constructor",
			Output:  []string{"const a = new Map<string, number>();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 41},
			},
		},
		{
			Code:   "const a: Foo<string> = new Foo();",
			Output: []string{"const a = new Foo<string>();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 33},
			},
		},
		{
			Code:   "const a: /* comment */ Foo /* another */ <string> = new Foo();",
			Output: []string{"const a = new Foo/* comment *//* another */<string>();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 62},
			},
		},
		{
			Code:   "const a: Foo<number> = new Foo;",
			Output: []string{"const a = new Foo<number>();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 7, EndLine: 1, EndColumn: 31},
			},
		},
		{
			Code:   "class Foo { a: Foo<string> = new Foo(); }",
			Output: []string{"class Foo { a = new Foo<string>(); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 13, EndLine: 1, EndColumn: 40},
			},
		},
		{
			Code:   "class Foo { [a]: Foo<string> = new Foo(); }",
			Output: []string{"class Foo { [a] = new Foo<string>(); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 13, EndLine: 1, EndColumn: 42},
			},
		},
		{
			Code:   "class Foo { accessor a: Foo<string> = new Foo(); }",
			Output: []string{"class Foo { accessor a = new Foo<string>(); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 13, EndLine: 1, EndColumn: 49},
			},
		},
		{
			Code:   "function foo(a: Foo<string> = new Foo()) {}",
			Output: []string{"function foo(a = new Foo<string>()) {}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 14, EndLine: 1, EndColumn: 40},
			},
		},
		{
			Code:   "function foo({ a }: Foo<string> = new Foo()) {}",
			Output: []string{"function foo({ a } = new Foo<string>()) {}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 14, EndLine: 1, EndColumn: 44},
			},
		},
		{
			Code:   "function foo([a]: Foo<string> = new Foo()) {}",
			Output: []string{"function foo([a] = new Foo<string>()) {}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 14, EndLine: 1, EndColumn: 42},
			},
		},
		{
			Code:   "class A { constructor(a: Foo<string> = new Foo()) {} }",
			Output: []string{"class A { constructor(a = new Foo<string>()) {} }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 23, EndLine: 1, EndColumn: 49},
			},
		},
		{
			Code:   "const a = function (a: Foo<string> = new Foo()) {};",
			Output: []string{"const a = function (a = new Foo<string>()) {};"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor", Message: "The generic type arguments should be specified as part of the constructor type arguments.", Line: 1, Column: 21, EndLine: 1, EndColumn: 47},
			},
		},
		{
			Code:    "const a = new Foo<string>();",
			Options: "type-annotation",
			Output:  []string{"const a: Foo<string> = new Foo();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 28},
			},
		},
		{
			Code:    "const a = new Map<string, number>();",
			Options: "type-annotation",
			Output:  []string{"const a: Map<string, number> = new Map();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 36},
			},
		},
		{
			Code:    "const a = new Foo<string>();",
			Options: "type-annotation",
			Output:  []string{"const a: Foo<string> = new Foo();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 28},
			},
		},
		{
			Code:    "const a = new Foo  <  string  >();",
			Options: "type-annotation",
			Output:  []string{"const a: Foo<  string  > = new Foo  ();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 7, EndLine: 1, EndColumn: 34},
			},
		},
		{
			Code:    "class Foo { a = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class Foo { a: Foo<string> = new Foo(); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 13, EndLine: 1, EndColumn: 35},
			},
		},
		{
			Code:    "class Foo { [a] = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class Foo { [a]: Foo<string> = new Foo(); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 13, EndLine: 1, EndColumn: 37},
			},
		},
		{
			Code:    "class Foo { accessor a = new Foo<string>(); }",
			Options: "type-annotation",
			Output:  []string{"class Foo { accessor a: Foo<string> = new Foo(); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 13, EndLine: 1, EndColumn: 44},
			},
		},
		{
			Code:    "function foo(a = new Foo<string>()) {}",
			Options: "type-annotation",
			Output:  []string{"function foo(a: Foo<string> = new Foo()) {}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 14, EndLine: 1, EndColumn: 35},
			},
		},
		{
			Code:    "function foo({ a } = new Foo<string>()) {}",
			Options: "type-annotation",
			Output:  []string{"function foo({ a }: Foo<string> = new Foo()) {}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 14, EndLine: 1, EndColumn: 39},
			},
		},
		{
			Code:    "function foo([a] = new Foo<string>()) {}",
			Options: "type-annotation",
			Output:  []string{"function foo([a]: Foo<string> = new Foo()) {}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 14, EndLine: 1, EndColumn: 37},
			},
		},
		{
			Code:    "class A { constructor(a = new Foo<string>()) {} }",
			Options: "type-annotation",
			Output:  []string{"class A { constructor(a: Foo<string> = new Foo()) {} }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 23, EndLine: 1, EndColumn: 44},
			},
		},
		{
			Code:    "const a = function (a = new Foo<string>()) {};",
			Options: "type-annotation",
			Output:  []string{"const a = function (a: Foo<string> = new Foo()) {};"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation", Message: "The generic type arguments should be specified as part of the type annotation.", Line: 1, Column: 21, EndLine: 1, EndColumn: 42},
			},
		},
	})
}
