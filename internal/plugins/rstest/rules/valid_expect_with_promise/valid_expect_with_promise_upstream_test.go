// cspell:ignore Thenish thenish
package valid_expect_with_promise

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const customPromiseDeclaration = `declare class CustomPromise<T> {
  then<R>(
    onFulfilled?: (value: T) => R,
    onRejected?: (error: unknown) => R,
  ): CustomPromise<R>;
}

declare const promised: CustomPromise<string>;`

// eslint-plugin-jest v29.16.0; its lazy TypeScript-loading test has no Go equivalent.
func TestValidExpectWithPromiseUpstream(t *testing.T) {
	thenables := map[string]any{"checkThenables": true}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ValidExpectWithPromiseRule,
		[]rule_tester.ValidTestCase{
			// The upstream invalid fixture passes a constructor, not a Promise instance.
			{Code: `class MyPromise extends Promise<unknown> {};

declare function build<T extends Promise<number>>(): T;

expect(build<typeof MyPromise>()).toEqual('hello world');`},
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
it('is true', async () => { expect(x).toEqual(null); });`},
			{Code: `<T extends Promise<unknown> = Promise<string>>(v: T) => expect(v).resolves.toThrow()`},
			{Code: `<T = string>(v: T) => expect(v).toBe(1)`},
			{Code: customPromiseDeclaration + `
it('works', async () => { await expect(promised).resolves.toBe('value'); });`, Options: thenables},
			{Code: customPromiseDeclaration + `
expect(promised).toEqual(1);`},
			{Code: `expect({ then: 1 }).toBe(1)`},
			{Code: `expect({ then: 1 }).toBe(1)`, Options: thenables},
			{Code: `expect(1).toBe(1)`, Options: thenables},
			{Code: `expect().toBe(1)`},
			{Code: `declare class Chainable { then(next: string): this; }
declare const chain: Chainable;
expect(chain).toEqual(chain);`},
			{Code: `declare class Chainable { then(next: string): this; }
declare const chain: Chainable;
expect(chain).toEqual(chain);`, Options: thenables},
			{Code: `declare class Thenish<T> { then<R>(onFulfilled?: (value: T) => R): Thenish<R>; }
declare const thenish: Thenish<string>;
expect(thenish).toEqual(1);`, Options: thenables},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `expect(Promise.resolve()).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "poorlyExpectedPromise", Message: "Subject is a promise so resolve or reject should be used", Line: 1, Column: 1, EndLine: 1, EndColumn: 34}}},
			{Code: `expect(new Promise(r => r())).toBe(1)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "poorlyExpectedPromise", Line: 1}}},
			{Code: `const x = 'hello world';

expect(x as Promise<number>).toEqual('hello world');
expect(x as Promise<string[]> & Array<string>).not.toStrictEqual('hello world');
expect(x as Promise<string[]> | Array<string>).not.toStrictEqual('hello world');`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "poorlyExpectedPromise", Line: 3}, {MessageId: "poorlyExpectedPromise", Line: 4}}},
			{Code: `declare function build<T>(): T;

expect(build<Promise<string>>()).toEqual('hello world');`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "poorlyExpectedPromise", Line: 3}}},
			{Code: `async function promiseValue(value: string): Promise<string> {
  return value;
}

expect(promiseValue('hello world')).toEqual('hello world');
expect(promiseValue()).not.toEqual([]);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "poorlyExpectedPromise", Line: 5}, {MessageId: "poorlyExpectedPromise", Line: 6}}},
			{Code: `class PromisedString extends Promise<string> {}

it('works', async () => {
  await expect(PromisedString.resolve("hello sunshine")).toEqual(1);
  await expect(new PromisedString(r => r("value"))).toEqual(1);
});`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "poorlyExpectedPromise", Line: 4}, {MessageId: "poorlyExpectedPromise", Line: 5}}},
			{Code: customPromiseDeclaration + `

expect(promised).toEqual(1);`, Options: thenables, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "poorlyExpectedPromise", Line: 10}}},
			{Code: customPromiseDeclaration + `

it('works', async () => {
  await expect(promised).resolves.toBe('value');
});`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unneededRejectResolve", Line: 11}}},
			{Code: `expect(Promise.resolve()).toBeInstanceOf(Promise);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "poorlyExpectedPromise", Line: 1}}},
			{Code: `expect("hello world").resolves.toContain(1)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unneededRejectResolve", Message: "Subject is not a promise so resolves is not needed", Line: 1, Column: 23, EndLine: 1, EndColumn: 31}}},
			{Code: `expect({}).rejects.not.toContain(1)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unneededRejectResolve", Message: "Subject is not a promise so rejects is not needed", Line: 1, Column: 12, EndLine: 1, EndColumn: 19}}},
			{Code: `it('works', async () => {
  await expect(0).resolves.toEqual(1);
});`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unneededRejectResolve", Line: 2, Column: 19}}},
			{Code: `class PromisedString extends Promise<string> {}

it('works', async () => {
  const value = await PromisedString.resolve("hello sunshine");

  await expect(value).resolves.toEqual(1);
});`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unneededRejectResolve", Line: 6, Column: 23}}},
		})
}
