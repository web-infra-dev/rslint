// TestNoUnneededAsyncExpectFunctionUpstream adapts the complete
// @vitest/eslint-plugin@v1.6.27 no-unneeded-async-expect-function suite to
// Rstest. Rstest-specific sources and tsgo edge shapes live in the extras suite.
package no_unneeded_async_expect_function

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnneededAsyncExpectFunctionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnneededAsyncExpectFunctionRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect.hasAssertions()`},
			{Code: `it('pass', async () => { expect(); })`},
			{Code: `it('pass', async () => { await expect(doSomethingAsync()).rejects.toThrow(); })`},
			{Code: `it('pass', async () => { await expect(doSomethingAsync(1, 2)).resolves.toBe(1); })`},
			{Code: `it('pass', async () => {
  await expect(async () => {
    await doSomethingAsync();
    await doSomethingTwiceAsync(1, 2);
  }).rejects.toThrow();
})`},
			{Code: `import { expect as pleaseExpect } from '@rstest/core';
it('pass', async () => { await pleaseExpect(doSomethingAsync()).rejects.toThrow(); })`},
			{Code: `it('pass', async () => {
  await expect(async () => { doSomethingAsync(); }).rejects.toThrow();
})`},
			{Code: `it('pass', async () => {
  await expect(async () => {
    const value = 1;
    await doSomethingAsync(value);
  }).rejects.toThrow();
})`},
			{Code: `it('pass for non-async expect', async () => {
  await expect(() => { doSomethingSync(); }).rejects.toThrow();
})`},
			{Code: `it('pass for await in expect', async () => {
  await expect(await doSomethingAsync()).rejects.toThrow();
})`},
			{Code: `it('pass for different matchers', async () => {
  await expect(await doSomething()).not.toThrow();
  await expect(await doSomething()).toHaveLength(2);
  await expect(await doSomething()).toHaveReturned();
  await expect(await doSomething()).not.toHaveBeenCalled();
  await expect(await doSomething()).not.toBeDefined();
  await expect(await doSomething()).toEqual(2);
})`},
			{Code: `it('pass for using await within for-loop', async () => {
  const callbacks = [async () => Promise.resolve(1), async () => Promise.reject(2)];
  await expect(async () => {
    for (const callback of callbacks) { await callback(); }
  }).rejects.toThrow();
})`},
			{Code: `it('pass for using await within array', async () => {
  await expect(async () => [await Promise.reject(2)]).rejects.toThrow(2);
})`},

			// Rstest permits functions with .rejects and invokes them before
			// observing the returned Promise. These upstream invalid cases are
			// therefore valid Rstest assertions rather than redundant wrappers.
			{Code: `it('keeps an async arrow', async () => {
  await expect(async () => {
    await doSomethingAsync();
  }).rejects.toThrow();
})`},
			{Code: `it('keeps a concise arrow', async () => {
  await expect(async () => await doSomethingAsync()).rejects.toThrow();
})`},
			{Code: `it('keeps an async function', async () => {
  await expect(async function () {
    await doSomethingAsync();
  }).rejects.toThrow();
})`},
			{Code: `it('keeps call arguments', async () => {
  await expect(async () => {
    await doSomethingAsync(1, 2);
  }).rejects.toThrow();
})`},
			{Code: `it('keeps call arguments in an async function', async () => {
  await expect(async function () {
    await doSomethingAsync(1, 2);
  }).rejects.toThrow();
})`},
			{Code: `it('keeps Promise.all', async () => {
  await expect(async function () {
    await Promise.all([doSomethingAsync(1, 2), doSomethingAsync()]);
  }).rejects.toThrow();
})`},
			{Code: `it('keeps an async reference', async () => {
  const operation = async () => { await doSomethingAsync() };
  await expect(async () => {
    await operation();
  }).rejects.toThrow();
})`},
		},
		nil,
	)
}
