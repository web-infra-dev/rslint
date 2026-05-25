// TestPreferFunctionTypeUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/tests/rules/prefer-function-type.test.ts 1:1.
// Position assertions cover line/column for every invalid case. Rslint-
// specific lock-in cases and tsgo edge-shape coverage live in
// prefer_function_type_extras_test.go.
package prefer_function_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferFunctionTypeUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferFunctionTypeRule, []rule_tester.ValidTestCase{
		{Code: `
interface Foo {
  (): void;
  bar: number;
}
    `},
		{Code: `
type Foo = {
  (): void;
  bar: number;
};
    `},
		{Code: `
function foo(bar: { (): string; baz: number }): string {
  return bar();
}
    `},
		{Code: `
interface Foo {
  bar: string;
}
interface Bar extends Foo {
  (): void;
}
    `},
		{Code: `
interface Foo {
  bar: string;
}
interface Bar extends Function, Foo {
  (): void;
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
interface Foo {
  (): string;
}
      `,
			Output: []string{`
type Foo = () => string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},
		// https://github.com/typescript-eslint/typescript-eslint/issues/3004
		{
			Code: `
export default interface Foo {
  /** comment */
  (): string;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
interface Foo {
  // comment
  (): string;
}
      `,
			Output: []string{`
// comment
type Foo = () => string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
export interface Foo {
  /** comment */
  (): string;
}
      `,
			Output: []string{`
/** comment */
export type Foo = () => string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
export interface Foo {
  // comment
  (): string;
}
      `,
			Output: []string{`
// comment
export type Foo = () => string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
function foo(bar: { /* comment */ (s: string): number } | undefined): number {
  return bar('hello');
}
      `,
			Output: []string{`
function foo(bar: /* comment */ ((s: string) => number) | undefined): number {
  return bar('hello');
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    35,
				},
			},
		},
		{
			Code: `
type Foo = {
  (): string;
};
      `,
			Output: []string{`
type Foo = () => string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
function foo(bar: { (s: string): number }): number {
  return bar('hello');
}
      `,
			Output: []string{`
function foo(bar: (s: string) => number): number {
  return bar('hello');
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    21,
				},
			},
		},
		{
			Code: `
function foo(bar: { (s: string): number } | undefined): number {
  return bar('hello');
}
      `,
			Output: []string{`
function foo(bar: ((s: string) => number) | undefined): number {
  return bar('hello');
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    21,
				},
			},
		},
		{
			Code: `
interface Foo extends Function {
  (): void;
}
      `,
			Output: []string{`
type Foo = () => void;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
interface Foo<T> {
  (bar: T): string;
}
      `,
			Output: []string{`
type Foo<T> = (bar: T) => string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
interface Foo<T> {
  (this: T): void;
}
      `,
			Output: []string{`
type Foo<T> = (this: T) => void;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
type Foo<T> = { (this: string): T };
      `,
			Output: []string{`
type Foo<T> = (this: string) => T;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    17,
				},
			},
		},
		{
			Code: `
interface Foo {
  (arg: this): void;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedThisOnFunctionOnlyInterface",
					Line:      3,
					Column:    9,
				},
			},
		},
		{
			Code: `
interface Foo {
  (arg: number): this | undefined;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedThisOnFunctionOnlyInterface",
					Line:      3,
					Column:    18,
				},
			},
		},
		{
			Code: `
// isn't actually valid ts but want to not give message saying it refers to Foo.
interface Foo {
  (): {
    a: {
      nested: this;
    };
    between: this;
    b: {
      nested: string;
    };
  };
}
      `,
			Output: []string{`
// isn't actually valid ts but want to not give message saying it refers to Foo.
type Foo = () => {
    a: {
      nested: this;
    };
    between: this;
    b: {
      nested: string;
    };
  };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
type X = {} | { (): void; }
      `,
			Output: []string{`
type X = {} | (() => void)
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    17,
				},
			},
		},
		{
			Code: `
type X = {} & { (): void; };
      `,
			Output: []string{`
type X = {} & (() => void);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    17,
				},
			},
		},
	})
}
