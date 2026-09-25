package no_unneeded_async_expect_function

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnneededAsyncExpectFunctionUpstream(t *testing.T) {
	// @vitest/eslint-plugin v1.6.27 and eslint-plugin-jest v29.16.1 share this
	// suite; the vitest copy adds the concise arrow body case, which is kept
	// here. Every upstream invalid case asserts through `rejects`, so all of
	// them stay invalid, reported without an automatic fix.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnneededAsyncExpectFunctionRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect.hasAssertions()`},
			{Code: `
it('pass', async () => {
  expect();
})
`},
			{Code: `
it('pass', async () => {
  await expect(doSomethingAsync()).rejects.toThrow();
})
`},
			{Code: `
it('pass', async () => {
  await expect(doSomethingAsync(1, 2)).resolves.toBe(1);
})
`},
			{Code: `
it('pass', async () => {
  await expect(async () => {
    await doSomethingAsync();
    await doSomethingTwiceAsync(1, 2);
  }).rejects.toThrow();
})
`},
			{Code: `
import { expect as pleaseExpect } from '@rstest/core';
it('pass', async () => {
  await pleaseExpect(doSomethingAsync()).rejects.toThrow();
})
`},
			{Code: `
it('pass', async () => {
  await expect(async () => {
    doSomethingAsync();
  }).rejects.toThrow();
})
`},
			{Code: `
it('pass', async () => {
  await expect(async () => {
    const a = 1;
    await doSomethingAsync(a);
  }).rejects.toThrow();
})
`},
			{Code: `
it('pass for non-async expect', async () => {
  await expect(() => {
    doSomethingSync(a);
  }).rejects.toThrow();
})
`},
			{Code: `
it('pass for await in expect', async () => {
  await expect(await doSomethingAsync()).rejects.toThrow();
})
`},
			{Code: `
it('pass for different matchers', async () => {
  await expect(await doSomething()).not.toThrow();
  await expect(await doSomething()).toHaveLength(2);
  await expect(await doSomething()).toHaveReturned();
  await expect(await doSomething()).not.toHaveBeenCalled();
  await expect(await doSomething()).not.toBeDefined();
  await expect(await doSomething()).toEqual(2);
})
`},
			{Code: `
it('pass for using await within for-loop', async () => {
  const b = [async () => Promise.resolve(1), async () => Promise.reject(2)];
  await expect(async() => {
    for (const a of b) {
      await b();
    }
  }).rejects.toThrow();
})
`},
			{Code: `
it('pass for using await within array', async () => {
  await expect(async() => [await Promise.reject(2)]).rejects.toThrow(2);
})
`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
it('should be fixed', async () => {
  await expect(async () => {
    await doSomethingAsync();
  }).rejects.toThrow();
})
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noAsyncWrapperForExpectedPromise",
					Line:      3, Column: 16, EndLine: 5, EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsyncWrapper",
						Output: `
it('should be fixed', async () => {
  await expect(doSomethingAsync()).rejects.toThrow();
})
`,
					}},
				}},
			},
			{
				Code: `
it('should be fixed', async () => {
  await expect(async () => await doSomethingAsync()).rejects.toThrow();
})
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noAsyncWrapperForExpectedPromise",
					Line:      3, Column: 16, EndLine: 3, EndColumn: 52,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsyncWrapper",
						Output: `
it('should be fixed', async () => {
  await expect(doSomethingAsync()).rejects.toThrow();
})
`,
					}},
				}},
			},
			{
				Code: `
it('should be fixed', async () => {
  await expect(async function () {
    await doSomethingAsync();
  }).rejects.toThrow();
})
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noAsyncWrapperForExpectedPromise",
					Line:      3, Column: 16, EndLine: 5, EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsyncWrapper",
						Output: `
it('should be fixed', async () => {
  await expect(doSomethingAsync()).rejects.toThrow();
})
`,
					}},
				}},
			},
			{
				Code: `
it('should be fixed for async arrow function', async () => {
  await expect(async () => {
    await doSomethingAsync(1, 2);
  }).rejects.toThrow();
})
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noAsyncWrapperForExpectedPromise",
					Line:      3, Column: 16, EndLine: 5, EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsyncWrapper",
						Output: `
it('should be fixed for async arrow function', async () => {
  await expect(doSomethingAsync(1, 2)).rejects.toThrow();
})
`,
					}},
				}},
			},
			{
				Code: `
it('should be fixed for async normal function', async () => {
  await expect(async function () {
    await doSomethingAsync(1, 2);
  }).rejects.toThrow();
})
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noAsyncWrapperForExpectedPromise",
					Line:      3, Column: 16, EndLine: 5, EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsyncWrapper",
						Output: `
it('should be fixed for async normal function', async () => {
  await expect(doSomethingAsync(1, 2)).rejects.toThrow();
})
`,
					}},
				}},
			},
			{
				Code: `
it('should be fixed for Promise.all', async () => {
  await expect(async function () {
    await Promise.all([doSomethingAsync(1, 2), doSomethingAsync()]);
  }).rejects.toThrow();
})
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noAsyncWrapperForExpectedPromise",
					Line:      3, Column: 16, EndLine: 5, EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsyncWrapper",
						Output: `
it('should be fixed for Promise.all', async () => {
  await expect(Promise.all([doSomethingAsync(1, 2), doSomethingAsync()])).rejects.toThrow();
})
`,
					}},
				}},
			},
			{
				Code: `
it('should be fixed for async ref to expect', async () => {
  const a = async () => { await doSomethingAsync() };
  await expect(async () => {
    await a();
  }).rejects.toThrow();
})
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noAsyncWrapperForExpectedPromise",
					Line:      4, Column: 16, EndLine: 6, EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveAsyncWrapper",
						Output: `
it('should be fixed for async ref to expect', async () => {
  const a = async () => { await doSomethingAsync() };
  await expect(a()).rejects.toThrow();
})
`,
					}},
				}},
			},
		},
	)
}
