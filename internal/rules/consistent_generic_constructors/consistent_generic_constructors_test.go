package consistent_generic_constructors

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestConsistentGenericConstructors(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentGenericConstructorsRule,
		[]rule_tester.ValidTestCase{
			// default: constructor
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
			{Code: `
class Foo {
  a = new Foo<string>();
}
			`},
			{Code: `
class Foo {
  accessor a = new Foo<string>();
}
			`},
			{Code: `
function foo(a: Foo = new Foo<string>()) {}
			`},
			{Code: `
function foo({ a }: Foo = new Foo<string>()) {}
			`},
			{Code: `
function foo([a]: Foo = new Foo<string>()) {}
			`},
			{Code: `
class A {
  constructor(a: Foo = new Foo<string>()) {}
}
			`},
			{Code: `
const a = function (a: Foo = new Foo<string>()) {};
			`},
			// type-annotation mode
			{
				Code:    "const a = new Foo();",
				Options: []interface{}{"type-annotation"},
			},
			{
				Code:    "const a: Foo<string> = new Foo();",
				Options: []interface{}{"type-annotation"},
			},
			{
				Code:    "const a: Foo<string> = new Foo<string>();",
				Options: []interface{}{"type-annotation"},
			},
			{
				Code:    "const a: Foo = new Foo();",
				Options: []interface{}{"type-annotation"},
			},
			{
				Code:    "const a: Bar = new Foo<string>();",
				Options: []interface{}{"type-annotation"},
			},
			{
				Code:    "const a: Bar<string> = new Foo<string>();",
				Options: []interface{}{"type-annotation"},
			},
			{
				Code:    "const a: Foo<string> = Foo<string>();",
				Options: []interface{}{"type-annotation"},
			},
			{
				Code:    "const a: Foo<string> = Foo();",
				Options: []interface{}{"type-annotation"},
			},
			{
				Code:    "const a: Foo = Foo<string>();",
				Options: []interface{}{"type-annotation"},
			},
			{
				Code:    "const a = new (class C<T> {})<string>();",
				Options: []interface{}{"type-annotation"},
			},
			{
				Code: `
class Foo {
  a: Foo<string> = new Foo();
}
				`,
				Options: []interface{}{"type-annotation"},
			},
			{
				Code: `
class Foo {
  accessor a: Foo<string> = new Foo();
}
				`,
				Options: []interface{}{"type-annotation"},
			},
			{
				Code: `
function foo(a: Foo<string> = new Foo()) {}
				`,
				Options: []interface{}{"type-annotation"},
			},
			{
				Code: `
function foo({ a }: Foo<string> = new Foo()) {}
				`,
				Options: []interface{}{"type-annotation"},
			},
			{
				Code: `
function foo([a]: Foo<string> = new Foo()) {}
				`,
				Options: []interface{}{"type-annotation"},
			},
			{
				Code: `
class A {
  constructor(a: Foo<string> = new Foo()) {}
}
				`,
				Options: []interface{}{"type-annotation"},
			},
			{
				Code: `
const a = function (a: Foo<string> = new Foo()) {};
				`,
				Options: []interface{}{"type-annotation"},
			},
			{
				Code: `
const [a = new Foo<string>()] = [];
				`,
				Options: []interface{}{"type-annotation"},
			},
			{
				Code: `
function a([a = new Foo<string>()]) {}
				`,
				Options: []interface{}{"type-annotation"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "const a: Foo<string> = new Foo();",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const a = new Foo<string>();"},
			},
			{
				Code: "const a: Map<string, number> = new Map();",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const a = new Map<string, number>();"},
			},
			{
				Code: "const a: Map <string, number> = new Map();",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const a = new Map<string, number>();"},
			},
			{
				Code: "const a: Map< string, number > = new Map();",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const a = new Map< string, number >();"},
			},
			{
				Code: "const a: Map<string, number> = new Map ();",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const a = new Map<string, number> ();"},
			},
			{
				Code: "const a: Foo<number> = new Foo;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const a = new Foo<number>();"},
			},
			{
				Code: `
class Foo {
  a: Foo<string> = new Foo();
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      3,
						Column:    3,
					},
				},
				Output: []string{`
class Foo {
  a = new Foo<string>();
}
				`},
			},
			{
				Code: `
class Foo {
  [a]: Foo<string> = new Foo();
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      3,
						Column:    3,
					},
				},
				Output: []string{`
class Foo {
  [a] = new Foo<string>();
}
				`},
			},
			{
				Code: `
class Foo {
  accessor a: Foo<string> = new Foo();
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      3,
						Column:    3,
					},
				},
				Output: []string{`
class Foo {
  accessor a = new Foo<string>();
}
				`},
			},
			{
				Code: `
class Foo {
  accessor a = new Foo<string>();
}
				`,
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      3,
						Column:    3,
					},
				},
				Output: []string{`
class Foo {
  accessor a: Foo<string> = new Foo();
}
				`},
			},
			{
				Code: `
class Foo {
  accessor [a]: Foo<string> = new Foo();
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      3,
						Column:    3,
					},
				},
				Output: []string{`
class Foo {
  accessor [a] = new Foo<string>();
}
				`},
			},
			{
				Code: `
function foo(a: Foo<string> = new Foo()) {}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      2,
						Column:    14,
					},
				},
				Output: []string{`
function foo(a = new Foo<string>()) {}
				`},
			},
			{
				Code: `
function foo({ a }: Foo<string> = new Foo()) {}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      2,
						Column:    14,
					},
				},
				Output: []string{`
function foo({ a } = new Foo<string>()) {}
				`},
			},
			{
				Code: `
function foo([a]: Foo<string> = new Foo()) {}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      2,
						Column:    14,
					},
				},
				Output: []string{`
function foo([a] = new Foo<string>()) {}
				`},
			},
			{
				Code: `
class A {
  constructor(a: Foo<string> = new Foo()) {}
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      3,
						Column:    15,
					},
				},
				Output: []string{`
class A {
  constructor(a = new Foo<string>()) {}
}
				`},
			},
			{
				Code: `
const a = function (a: Foo<string> = new Foo()) {};
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferConstructor",
						Line:      2,
						Column:    21,
					},
				},
				Output: []string{`
const a = function (a = new Foo<string>()) {};
				`},
			},
			{
				Code: "const a = new Foo<string>();",
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const a: Foo<string> = new Foo();"},
			},
			{
				Code: "const a = new Map<string, number>();",
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const a: Map<string, number> = new Map();"},
			},
			{
				Code: "const a = new Map <string, number> ();",
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const a: Map<string, number> = new Map  ();"},
			},
			{
				Code: "const a = new Map< string, number >();",
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"const a: Map< string, number > = new Map();"},
			},
			{
				Code: `
class Foo {
  a = new Foo<string>();
}
				`,
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      3,
						Column:    3,
					},
				},
				Output: []string{`
class Foo {
  a: Foo<string> = new Foo();
}
				`},
			},
			{
				Code: `
class Foo {
  [a] = new Foo<string>();
}
				`,
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      3,
						Column:    3,
					},
				},
				Output: []string{`
class Foo {
  [a]: Foo<string> = new Foo();
}
				`},
			},
			{
				Code: `
class Foo {
  [a + b] = new Foo<string>();
}
				`,
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      3,
						Column:    3,
					},
				},
				Output: []string{`
class Foo {
  [a + b]: Foo<string> = new Foo();
}
				`},
			},
			{
				Code: `
function foo(a = new Foo<string>()) {}
				`,
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      2,
						Column:    14,
					},
				},
				Output: []string{`
function foo(a: Foo<string> = new Foo()) {}
				`},
			},
			{
				Code: `
function foo({ a } = new Foo<string>()) {}
				`,
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      2,
						Column:    14,
					},
				},
				Output: []string{`
function foo({ a }: Foo<string> = new Foo()) {}
				`},
			},
			{
				Code: `
function foo([a] = new Foo<string>()) {}
				`,
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      2,
						Column:    14,
					},
				},
				Output: []string{`
function foo([a]: Foo<string> = new Foo()) {}
				`},
			},
			{
				Code: `
class A {
  constructor(a = new Foo<string>()) {}
}
				`,
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      3,
						Column:    15,
					},
				},
				Output: []string{`
class A {
  constructor(a: Foo<string> = new Foo()) {}
}
				`},
			},
			{
				Code: `
const a = function (a = new Foo<string>()) {};
				`,
				Options: []interface{}{"type-annotation"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferTypeAnnotation",
						Line:      2,
						Column:    21,
					},
				},
				Output: []string{`
const a = function (a: Foo<string> = new Foo()) {};
				`},
			},
		},
	)
}