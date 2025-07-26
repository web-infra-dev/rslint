package default_param_last

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
	"github.com/typescript-eslint/rslint/internal/rule_tester"
)

func TestDefaultParamLastRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DefaultParamLastRule,
		[]rule_tester.ValidTestCase{
			{Code: `function foo() {}`},
			{Code: `function foo(a: number) {}`},
			{Code: `function foo(a = 1) {}`},
			{Code: `function foo(a?: number) {}`},
			{Code: `function foo(a: number, b: number) {}`},
			{Code: `function foo(a: number, b: number, c?: number) {}`},
			{Code: `function foo(a: number, b = 1) {}`},
			{Code: `function foo(a: number, b = 1, c = 1) {}`},
			{Code: `function foo(a: number, b = 1, c?: number) {}`},
			{Code: `function foo(a: number, b?: number, c = 1) {}`},
			{Code: `function foo(a: number, b = 1, ...c) {}`},

			{Code: `const foo = function () {};`},
			{Code: `const foo = function (a: number) {};`},
			{Code: `const foo = function (a = 1) {};`},
			{Code: `const foo = function (a?: number) {};`},
			{Code: `const foo = function (a: number, b: number) {};`},
			{Code: `const foo = function (a: number, b: number, c?: number) {};`},
			{Code: `const foo = function (a: number, b = 1) {};`},
			{Code: `const foo = function (a: number, b = 1, c = 1) {};`},
			{Code: `const foo = function (a: number, b = 1, c?: number) {};`},
			{Code: `const foo = function (a: number, b?: number, c = 1) {};`},
			{Code: `const foo = function (a: number, b = 1, ...c) {};`},

			{Code: `const foo = () => {};`},
			{Code: `const foo = (a: number) => {};`},
			{Code: `const foo = (a = 1) => {};`},
			{Code: `const foo = (a?: number) => {};`},
			{Code: `const foo = (a: number, b: number) => {};`},
			{Code: `const foo = (a: number, b: number, c?: number) => {};`},
			{Code: `const foo = (a: number, b = 1) => {};`},
			{Code: `const foo = (a: number, b = 1, c = 1) => {};`},
			{Code: `const foo = (a: number, b = 1, c?: number) => {};`},
			{Code: `const foo = (a: number, b?: number, c = 1) => {};`},
			{Code: `const foo = (a: number, b = 1, ...c) => {};`},

			{Code: `
class Foo {
  constructor(a: number, b: number, c: number) {}
}
            `},
			{Code: `
class Foo {
  constructor(a: number, b?: number, c = 1) {}
}
            `},
			{Code: `
class Foo {
  constructor(a: number, b = 1, c?: number) {}
}
            `},
			{Code: `
class Foo {
  constructor(
    public a: number,
    protected b: number,
    private c: number,
  ) {}
}
            `},
			{Code: `
class Foo {
  constructor(
    public a: number,
    protected b?: number,
    private c = 10,
  ) {}
}
            `},
			{Code: `
class Foo {
  constructor(
    public a: number,
    protected b = 10,
    private c?: number,
  ) {}
}
            `},
			{Code: `
class Foo {
  constructor(
    a: number,
    protected b?: number,
    private c = 0,
  ) {}
}
            `},
			{Code: `
class Foo {
  constructor(
    a: number,
    b?: number,
    private c = 0,
  ) {}
}
            `},
			{Code: `
class Foo {
  constructor(
    a: number,
    private b?: number,
    c = 0,
  ) {}
}
            `},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `function foo(a = 1, b: number) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			{
				Code: `function foo(a = 1, b = 2, c: number) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 26,
					},
				},
			},
			{
				Code: `function foo(a = 1, b: number, c = 2, d: number) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    32,
						EndLine:   1,
						EndColumn: 37,
					},
				},
			},
			{
				Code: `function foo(a = 1, b: number, c = 2) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			{
				Code: `function foo(a = 1, b: number, ...c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			{
				Code: `function foo(a?: number, b: number) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 24,
					},
				},
			},
			{
				Code: `function foo(a: number, b?: number, c: number) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    25,
						EndLine:   1,
						EndColumn: 35,
					},
				},
			},
			{
				Code: `function foo(a = 1, b?: number, c: number) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 31,
					},
				},
			},
			{
				Code: `function foo(a = 1, { b }) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			{
				Code: `function foo({ a } = {}, b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 24,
					},
				},
			},
			{
				Code: `function foo({ a, b } = { a: 1, b: 2 }, c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 39,
					},
				},
			},
			{
				Code: `function foo([a] = [], b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 22,
					},
				},
			},
			{
				Code: `function foo([a, b] = [1, 2], c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 29,
					},
				},
			},
			{
				Code: `const foo = function (a = 1, b: number) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 28,
					},
				},
			},
			{
				Code: `const foo = function (a = 1, b = 2, c: number) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 28,
					},
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    30,
						EndLine:   1,
						EndColumn: 35,
					},
				},
			},
			{
				Code: `const foo = function (a = 1, b: number, c = 2, d: number) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 28,
					},
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    41,
						EndLine:   1,
						EndColumn: 46,
					},
				},
			},
			{
				Code: `const foo = function (a = 1, b: number, c = 2) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 28,
					},
				},
			},
			{
				Code: `const foo = function (a = 1, b: number, ...c) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 28,
					},
				},
			},
			{
				Code: `const foo = function (a?: number, b: number) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 33,
					},
				},
			},
			{
				Code: `const foo = function (a: number, b?: number, c: number) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    34,
						EndLine:   1,
						EndColumn: 44,
					},
				},
			},
			{
				Code: `const foo = function (a = 1, b?: number, c: number) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 28,
					},
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    30,
						EndLine:   1,
						EndColumn: 40,
					},
				},
			},
			{
				Code: `const foo = function (a = 1, { b }) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 28,
					},
				},
			},
			{
				Code: `const foo = function ({ a } = {}, b) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 33,
					},
				},
			},
			{
				Code: `const foo = function ({ a, b } = { a: 1, b: 2 }, c) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 48,
					},
				},
			},
			{
				Code: `const foo = function ([a] = [], b) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 31,
					},
				},
			},
			{
				Code: `const foo = function ([a, b] = [1, 2], c) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 38,
					},
				},
			},
			{
				Code: `const foo = (a = 1, b: number) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			{
				Code: `const foo = (a = 1, b = 2, c: number) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 26,
					},
				},
			},
			{
				Code: `const foo = (a = 1, b: number, c = 2, d: number) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    32,
						EndLine:   1,
						EndColumn: 37,
					},
				},
			},
			{
				Code: `const foo = (a = 1, b: number, c = 2) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			{
				Code: `const foo = (a = 1, b: number, ...c) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			{
				Code: `const foo = (a?: number, b: number) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 24,
					},
				},
			},
			{
				Code: `const foo = (a: number, b?: number, c: number) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    25,
						EndLine:   1,
						EndColumn: 35,
					},
				},
			},
			{
				Code: `const foo = (a = 1, b?: number, c: number) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 31,
					},
				},
			},
			{
				Code: `const foo = (a = 1, { b }) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			{
				Code: `const foo = ({ a } = {}, b) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 24,
					},
				},
			},
			{
				Code: `const foo = ({ a, b } = { a: 1, b: 2 }, c) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 39,
					},
				},
			},
			{
				Code: `const foo = ([a] = [], b) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 22,
					},
				},
			},
			{
				Code: `const foo = ([a, b] = [1, 2], c) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 29,
					},
				},
			},
			{
				Code: `
class Foo {
  constructor(
    public a: number,
    protected b?: number,
    private c: number,
  ) {}
}
                `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      5,
						Column:    5,
						EndLine:   5,
						EndColumn: 25,
					},
				},
			},
			{
				Code: `
class Foo {
  constructor(
    public a: number,
    protected b = 0,
    private c: number,
  ) {}
}
                `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      5,
						Column:    5,
						EndLine:   5,
						EndColumn: 20,
					},
				},
			},
			{
				Code: `
class Foo {
  constructor(
    public a?: number,
    private b: number,
  ) {}
}
                `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      4,
						Column:    5,
						EndLine:   4,
						EndColumn: 22,
					},
				},
			},
			{
				Code: `
class Foo {
  constructor(
    public a = 0,
    private b: number,
  ) {}
}
                `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      4,
						Column:    5,
						EndLine:   4,
						EndColumn: 17,
					},
				},
			},
			{
				Code: `
class Foo {
  constructor(a = 0, b: number) {}
}
                `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      3,
						Column:    15,
						EndLine:   3,
						EndColumn: 20,
					},
				},
			},
			{
				Code: `
class Foo {
  constructor(a?: number, b: number) {}
}
                `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "shouldBeLast",
						Line:      3,
						Column:    15,
						EndLine:   3,
						EndColumn: 25,
					},
				},
			},
		},
	)
}