// cspell:ignore Thenish thenish
package valid_expect_with_promise

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const customPromiseDeclaration = `declare class CustomPromise<T> {
  then<R>(
    onFulfilled?: (value: T) => R,
    onRejected?: (error: unknown) => R,
  ): CustomPromise<R>;
}

declare const promised: CustomPromise<string>;`

var checkThenables = map[string]any{"checkThenables": true}

func poorly(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: "poorlyExpectedPromise", Line: line, Column: column, EndLine: endLine, EndColumn: endColumn}
}

func unneeded(modifier string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unneededRejectResolve", Message: "Subject is not a promise so " + modifier + " is not needed",
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// TestValidExpectWithPromiseUpstream migrates the full valid/invalid suite from
// upstream src/rules/__tests__/valid-expect-with-promise.test.ts 1:1. Position
// assertions cover line/column for every invalid case. The lazy TypeScript
// loading test has no Go equivalent. rslint-specific lock-in cases live in the
// valid_expect_with_promise_extras_test.go file.
func TestValidExpectWithPromiseUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ValidExpectWithPromiseRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect`},
			{Code: `expect.hasAssertions`},
			{Code: `expect.hasAssertions()`},
			{Code: `expect(a).toBe(b)`},
			{Code: `expect(Promise.resolve()).resolves.toBe(1)`},
			{Code: `expect(Promise.resolve()).rejects.toBe(1)`},
			{Code: `it('is correct', async () => {
  await expect(Promise.resolve()).rejects.toEqual(1);
  await expect(Promise.reject()).resolves.not.toStrictEqual(1);

  expect(await Promise.resolve()).toEqual(1);
});`},
			{Code: `interface WithCode { code: string };

class MyError implements WithCode {}

const myError = new MyError();

expect(myError).toBeInstanceOf(Error);

expect(myError).toBe(new Error());
expect(myError).toEqual(new Error());
expect(myError).toStrictEqual(new Error());`},
			{Code: `const x: Array<number> = [1, 2, 3];

it('is true', async () => {
  expect(x).toEqual(null);
});`},
			{Code: `<T extends Promise<unknown> = Promise<string>>(v: T) => expect(v).resolves.toThrow()`},
			{Code: `<T = string>(v: T) => expect(v).toBe(1)`},
			{Code: customPromiseDeclaration + `

it('works', async () => {
  await expect(promised).resolves.toBe('value');
});`, Options: checkThenables},
			// thenables are not treated as promises unless checkThenables is enabled
			{Code: customPromiseDeclaration + `

expect(promised).toEqual(1);`},
			{Code: `expect({ then: 1 }).toBe(1)`},
			{Code: `expect({ then: 1 }).toBe(1)`, Options: checkThenables},
			{Code: `expect(1).toBe(1)`, Options: checkThenables},
			{Code: `expect().toBe(1)`},
			{Code: `declare class Chainable {
  then(next: string): this;
}

declare const chain: Chainable;

expect(chain).toEqual(chain);`},
			{Code: `declare class Chainable {
  then(next: string): this;
}

declare const chain: Chainable;

expect(chain).toEqual(chain);`, Options: checkThenables},
			// a `then` accepting only a fulfillment callback (like Cypress) is not
			// enough to be considered thenable - a rejection callback is required too
			{Code: `declare class Thenish<T> {
  then<R>(onFulfilled?: (value: T) => R): Thenish<R>;
}

declare const thenish: Thenish<string>;

expect(thenish).toEqual(1);`, Options: checkThenables},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(Promise.resolve()).toBe(1)`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "poorlyExpectedPromise", Message: "Subject is a promise so resolve or reject should be used",
					Line: 1, Column: 1, EndLine: 1, EndColumn: 34,
				}},
			},
			{
				Code:   `expect(new Promise(r => r())).toBe(1)`,
				Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 38)},
			},
			{
				Code: `const x = 'hello world';

expect(x as Promise<number>).toEqual('hello world');
expect(x as Promise<string[]> & Array<string>).not.toStrictEqual('hello world');
expect(x as Promise<string[]> | Array<string>).not.toStrictEqual('hello world');`,
				Errors: []rule_tester.InvalidTestCaseError{poorly(3, 1, 3, 52), poorly(4, 1, 4, 80)},
			},
			{
				Code: `declare function build<T>(): T;

expect(build<Promise<string>>()).toEqual('hello world');`,
				Errors: []rule_tester.InvalidTestCaseError{poorly(3, 1, 3, 56)},
			},
			{
				Code: `class MyPromise extends Promise<unknown> {};

declare function build<T extends Promise<number>>(): T;

expect(build<typeof MyPromise>()).toEqual('hello world');`,
				Errors: []rule_tester.InvalidTestCaseError{poorly(5, 1, 5, 57)},
			},
			{
				Code: `async function promiseValue(value: string): Promise<string> {
  return value;
}

expect(promiseValue('hello world')).toEqual('hello world');
expect(promiseValue()).not.toEqual([]);`,
				Errors: []rule_tester.InvalidTestCaseError{poorly(5, 1, 5, 59), poorly(6, 1, 6, 39)},
			},
			{
				Code: `class PromisedString extends Promise<string> {}

it('works', async () => {
  await expect(PromisedString.resolve("hello sunshine")).toEqual(1);
  await expect(new PromisedString(r => r("value"))).toEqual(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{poorly(4, 9, 4, 68), poorly(5, 9, 5, 63)},
			},
			{
				Code: customPromiseDeclaration + `

expect(promised).toEqual(1);`,
				Options: checkThenables,
				Errors:  []rule_tester.InvalidTestCaseError{poorly(10, 1, 10, 28)},
			},
			// without checkThenables, a thenable is not considered a promise, so
			// using resolves or rejects on one is reported as unneeded
			{
				Code: customPromiseDeclaration + `

it('works', async () => {
  await expect(promised).resolves.toBe('value');
});`,
				Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 11, 26, 11, 34)},
			},
			// technically this is valid, but we choose not to give it special treatment
			// as it should be very rare and doing so could give a lot of false negatives
			{
				Code:   `expect(Promise.resolve()).toBeInstanceOf(Promise);`,
				Errors: []rule_tester.InvalidTestCaseError{poorly(1, 1, 1, 50)},
			},
			{
				Code:   `expect("hello world").resolves.toContain(1)`,
				Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 1, 23, 1, 31)},
			},
			{
				Code:   `expect({}).rejects.not.toContain(1)`,
				Errors: []rule_tester.InvalidTestCaseError{unneeded("rejects", 1, 12, 1, 19)},
			},
			{
				Code: `it('works', async () => {
  await expect(0).resolves.toEqual(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 2, 19, 2, 27)},
			},
			{
				Code: `class PromisedString extends Promise<string> {}

it('works', async () => {
  const value = await PromisedString.resolve("hello sunshine");

  await expect(value).resolves.toEqual(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{unneeded("resolves", 6, 23, 6, 31)},
			},
		},
	)
}
