// TestNoErrorEqualUpstream migrates the full valid/invalid suite from
// eslint-plugin-jest@v29.16.6 src/rules/__tests__/no-error-equal.test.ts 1:1,
// plus the examples from docs/rules/no-error-equal.md. Every invalid case
// asserts the complete range of the reported matcher call.
package no_error_equal_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_error_equal"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const msg = "Avoid using equality matchers to check errors"

func TestNoErrorEqualUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_error_equal.NoErrorEqualRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect`},
			{Code: `expect.hasAssertions`},
			{Code: `expect.hasAssertions()`},
			{Code: `expect(a).toBe(b)`},
			{Code: `expect(a).toThrow(b)`},
			{Code: `expect(a).toThrowError(b)`},
			{Code: `class MyError {}

expect(new MyError()).toBeInstanceOf(Error);

expect(new MyError()).toBe(new Error());
expect(new MyError()).toEqual(new Error());
expect(new MyError()).toStrictEqual(new Error());`},
			{Code: `class MyError {}

const myError = new MyError();

expect(myError).toBeInstanceOf(Error);

expect(myError).toBe(new Error());
expect(myError).toEqual(new Error());
expect(myError).toStrictEqual(new Error());`},
			{Code: `interface WithCode { code: string };

class MyError implements WithCode {}

const myError = new MyError();

expect(myError).toBeInstanceOf(Error);

expect(myError).toBe(new Error());
expect(myError).toEqual(new Error());
expect(myError).toStrictEqual(new Error());`},
			{Code: `function buildError(): Error | { code: string } | null {
  return new Error('oh noes');
}

const x: Array<number> = [1, 2, 3];

expect(buildError()).toEqual(null);
expect(x).toEqual(null);
expect(buildError()).toEqual(new Error());`},
			{Code: `type Mx = Array<number>;

const x: Mx = [1, 2, 3];

expect(buildError()).toEqual(null);
expect(x).toEqual(null);
expect(buildError()).toEqual(new Error());`},
			{Code: `<T = Error>(v) => expect(v as T).toEqual(new Error())`},
			{Code: `expect(new AggregateError()).toBe(0)`},
			// docs/rules/no-error-equal.md: correct examples. The documented arrow
			// bodies `() => throw ...` are not valid syntax, so they use blocks here.
			{Code: "expect(() => { throw new AggregateError([], expect.any(String)); }).toThrow(new Error(expect.any(String)));"},
			{Code: "expect(() => { throw new Error('hello world'); }).toThrow('hello sunshine');"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(new Error("hello world")).toEqual(new Error("hello world"))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 1, Column: 1, EndLine: 1, EndColumn: 67},
				},
			},
			{
				Code: `expect(Error("hello world")).toEqual(new Error("hello world"))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 1, Column: 1, EndLine: 1, EndColumn: 63},
				},
			},
			{
				Code: `expect(new Error("hello world")).toStrictEqual(new Error("hello world"))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 1, Column: 1, EndLine: 1, EndColumn: 73},
				},
			},
			{
				Code: `class MyError extends Error {}

expect(new MyError()).toBeInstanceOf(Error);

expect(new MyError()).toBe(new Error());
expect(new MyError()).toEqual(new Error());
expect(new MyError()).toStrictEqual(new Error());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 6, Column: 1, EndLine: 6, EndColumn: 43},
					{MessageId: "equalError", Message: msg, Line: 7, Column: 1, EndLine: 7, EndColumn: 49},
				},
			},
			{
				Code: `class MyError extends Error {}

const myError = new MyError();

expect(myError).toBeInstanceOf(Error);

expect(myError).toBe(new Error());
expect(myError).toEqual(new Error());
expect(myError).toStrictEqual(new Error());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 8, Column: 1, EndLine: 8, EndColumn: 37},
					{MessageId: "equalError", Message: msg, Line: 9, Column: 1, EndLine: 9, EndColumn: 43},
				},
			},
			{
				Code: `expect(new AggregateError([], 'hello world')).toBe(new Error());
expect(new AggregateError([], 'hello world')).toEqual(new Error());
expect(new AggregateError([], 'hello world')).toStrictEqual(new Error());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 67},
					{MessageId: "equalError", Message: msg, Line: 3, Column: 1, EndLine: 3, EndColumn: 73},
				},
			},
			{
				Code: `const x = 'hello world';

expect(x as Error).toEqual('hello world');
expect(x as Error & {}).not.toStrictEqual('hello world');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 3, Column: 1, EndLine: 3, EndColumn: 42},
					{MessageId: "equalError", Message: msg, Line: 4, Column: 1, EndLine: 4, EndColumn: 57},
				},
			},
			{
				Code: `declare function buildError<T>(): T;

expect(buildError<Error>()).toEqual('hello world');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 3, Column: 1, EndLine: 3, EndColumn: 51},
				},
			},
			{
				Code: `declare function addCode<T extends Error>(err: T, code: string): T & { code: string };

expect(addCode(new Error('hello world'), 'MODULE_NOT_FOUND')).toEqual('hello world');
expect(addCode(new AggregateError([], 'hello world'), 'MODULE_NOT_FOUND')).toEqual('hello world');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 3, Column: 1, EndLine: 3, EndColumn: 85},
					{MessageId: "equalError", Message: msg, Line: 4, Column: 1, EndLine: 4, EndColumn: 98},
				},
			},
			{
				Code: `<T extends Error>(v: any) => expect(v as T).toStrictEqual(new Error())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 1, Column: 30, EndLine: 1, EndColumn: 71},
				},
			},
			{
				Code: `function buildError(msg: string) {
  return new Error(msg);
}

expect(buildError('hello world')).toEqual('hello world');
expect(buildError('hello world')).toEqual(new Error('hello world'));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 5, Column: 1, EndLine: 5, EndColumn: 57},
					{MessageId: "equalError", Message: msg, Line: 6, Column: 1, EndLine: 6, EndColumn: 68},
				},
			},
			{
				Code: `function buildError(msg: string): Error {
  return new Error(msg);
}

expect(buildError('hello world')).toEqual('hello world');
expect(buildError('hello world')).toEqual(new Error('hello world'));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 5, Column: 1, EndLine: 5, EndColumn: 57},
					{MessageId: "equalError", Message: msg, Line: 6, Column: 1, EndLine: 6, EndColumn: 68},
				},
			},
			// Upstream leaves line 5 unmatched with a todo: calling a class without new has
			// an error type, so only line 4 is reported.
			{
				Code: `class MyGenericError extends Error {}
class MySpecificError extends MyGenericError {}

expect(new MySpecificError('hello world')).toEqual('hello world');
expect(MySpecificError('hello world')).toStrictEqual(new Error('hello world'));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 4, Column: 1, EndLine: 4, EndColumn: 66},
				},
			},
			{
				Code: `interface Mx { code: string }

class MyError extends Error implements Mx {}

expect(new MyError('hello world')).toEqual('hello world');
expect(new MyError('hello world')).not.toStrictEqual(new Error('hello world'));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 5, Column: 1, EndLine: 5, EndColumn: 58},
					{MessageId: "equalError", Message: msg, Line: 6, Column: 1, EndLine: 6, EndColumn: 79},
				},
			},
			{
				Code: `it('works', async () => {
  const err = await Promise.resolve(new Error('oh noes'));

  expect(err).toEqual(new Error('oh noes'));
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 4, Column: 3, EndLine: 4, EndColumn: 44},
				},
			},
			// docs/rules/no-error-equal.md: incorrect examples.
			{
				Code: `expect(new AggregateError([], expect.any(String))).toEqual(
  new Error(expect.any(String)),
);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},
			{
				Code: `expect(new Error('hello world')).toStrictEqual('hello sunshine');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "equalError", Message: msg, Line: 1, Column: 1, EndLine: 1, EndColumn: 65},
				},
			},
		},
	)
}
