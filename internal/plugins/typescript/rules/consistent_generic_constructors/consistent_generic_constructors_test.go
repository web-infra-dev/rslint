package consistent_generic_constructors

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentGenericConstructorsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentGenericConstructorsRule, []rule_tester.ValidTestCase{
		// Default mode (constructor style)
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

		// Class properties
		{Code: "class Foo { a = new Foo<string>(); }"},
		{Code: "class Foo { a: Foo = new Foo<string>(); }"},
		{Code: "class Foo { a: Foo<string> = new Foo<string>(); }"},

		// Accessor properties
		{Code: "class Foo { accessor a = new Foo<string>(); }"},
		{Code: "class Foo { accessor a: Foo = new Foo<string>(); }"},
		{Code: "class Foo { accessor a: Foo<string> = new Foo<string>(); }"},

		// Function parameters
		{Code: "function foo(a: Foo = new Foo<string>()) {}"},
		{Code: "function foo(a: Foo<string> = new Foo<string>()) {}"},
		{Code: "function foo(a = new Foo<string>()) {}"},

		// Destructuring patterns
		{Code: "function foo({ a }: Foo = new Foo<string>()) {}"},
		{Code: "function foo({ a }: Foo<string> = new Foo<string>()) {}"},
		{Code: "function foo([a]: Foo = new Foo<string>()) {}"},
		{Code: "function foo([a]: Foo<string> = new Foo<string>()) {}"},

		// Constructor parameters
		{Code: "class A { constructor(a: Foo = new Foo<string>()) {} }"},
		{Code: "class A { constructor(a: Foo<string> = new Foo<string>()) {} }"},

		// Arrow functions
		{Code: "const a = function (a: Foo = new Foo<string>()) {};"},
		{Code: "const a = function (a: Foo<string> = new Foo<string>()) {};"},

		// Type-annotation mode
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

		// Class properties (type-annotation mode)
		{Code: "class Foo { a: Foo<string> = new Foo(); }", Options: "type-annotation"},
		{Code: "class Foo { a: Foo<string> = new Foo<string>(); }", Options: "type-annotation"},

		// Accessor properties (type-annotation mode)
		{Code: "class Foo { accessor a: Foo<string> = new Foo(); }", Options: "type-annotation"},
		{Code: "class Foo { accessor a: Foo<string> = new Foo<string>(); }", Options: "type-annotation"},

		// Function parameters (type-annotation mode)
		{Code: "function foo(a: Foo<string> = new Foo()) {}", Options: "type-annotation"},
		{Code: "function foo(a: Foo<string> = new Foo<string>()) {}", Options: "type-annotation"},

		// Destructuring patterns (type-annotation mode)
		{Code: "function foo({ a }: Foo<string> = new Foo()) {}", Options: "type-annotation"},
		{Code: "function foo([a]: Foo<string> = new Foo()) {}", Options: "type-annotation"},

		// Constructor parameters (type-annotation mode)
		{Code: "class A { constructor(a: Foo<string> = new Foo()) {} }", Options: "type-annotation"},

		// Arrow functions (type-annotation mode)
		{Code: "const a = function (a: Foo<string> = new Foo()) {};", Options: "type-annotation"},

		// Variable destructuring
		{Code: "const [a = new Foo<string>()] = [];", Options: "type-annotation"},
		{Code: "function a([a = new Foo<string>()]) {}", Options: "type-annotation"},
	}, []rule_tester.InvalidTestCase{
		// Default mode (prefer constructor)
		{
			Code:    "const a: Foo<string> = new Foo();",
			Options: "constructor",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},
		{
			Code:    "const a: Map<string, number> = new Map();",
			Options: "constructor",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},
		{
			Code: "const a: Foo<string> = new Foo();",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},
		{
			Code: "const a: /* comment */ Foo /* another */ <string> = new Foo();",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},
		{
			Code: "const a: Foo<number> = new Foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},

		// Class properties (prefer constructor)
		{
			Code: "class Foo { a: Foo<string> = new Foo(); }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},
		{
			Code: "class Foo { [a]: Foo<string> = new Foo(); }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},

		// Accessor properties (prefer constructor)
		{
			Code: "class Foo { accessor a: Foo<string> = new Foo(); }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},

		// Function parameters (prefer constructor)
		{
			Code: "function foo(a: Foo<string> = new Foo()) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},
		{
			Code: "function foo({ a }: Foo<string> = new Foo()) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},
		{
			Code: "function foo([a]: Foo<string> = new Foo()) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},

		// Constructor parameters (prefer constructor)
		{
			Code: "class A { constructor(a: Foo<string> = new Foo()) {} }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},

		// Arrow functions (prefer constructor)
		{
			Code: "const a = function (a: Foo<string> = new Foo()) {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferConstructor"},
			},
		},

		// Type-annotation mode (prefer type-annotation)
		{
			Code:    "const a = new Foo<string>();",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},
		{
			Code:    "const a = new Map<string, number>();",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},
		{
			Code:    "const a = new Foo<string>();",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},
		{
			Code:    "const a = new Foo  <  string  >();",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},

		// Class properties (prefer type-annotation)
		{
			Code:    "class Foo { a = new Foo<string>(); }",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},
		{
			Code:    "class Foo { [a] = new Foo<string>(); }",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},

		// Accessor properties (prefer type-annotation)
		{
			Code:    "class Foo { accessor a = new Foo<string>(); }",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},

		// Function parameters (prefer type-annotation)
		{
			Code:    "function foo(a = new Foo<string>()) {}",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},
		{
			Code:    "function foo({ a } = new Foo<string>()) {}",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},
		{
			Code:    "function foo([a] = new Foo<string>()) {}",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},

		// Constructor parameters (prefer type-annotation)
		{
			Code:    "class A { constructor(a = new Foo<string>()) {} }",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},

		// Arrow functions (prefer type-annotation)
		{
			Code:    "const a = function (a = new Foo<string>()) {};",
			Options: "type-annotation",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferTypeAnnotation"},
			},
		},
	})
}
