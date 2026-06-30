package no_unneeded_async_expect_function_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_unneeded_async_expect_function"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnneededAsyncExpectFunctionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_unneeded_async_expect_function.NoUnneededAsyncExpectFunctionRule,
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
        import { expect as pleaseExpect } from '@jest/globals';
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
			{Code: `
      it('does not unwrap awaited identifiers', async () => {
        const promise = doSomethingAsync();
        await expect(async () => {
          await promise;
        }).rejects.toThrow();
      })
    `},
			{Code: `
      it('does not unwrap async functions for sync matchers', async () => {
        expect(async () => {
          await doSomethingAsync();
        }).toThrow();
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
				Output: []string{`
        it('should be fixed', async () => {
          await expect(doSomethingAsync()).rejects.toThrow(); 
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncWrapperForExpectedPromise"},
				},
			},
			{
				Code: `
        it('should be fixed', async () => {
          await expect(async function () {
            await doSomethingAsync();
          }).rejects.toThrow(); 
        })
      `,
				Output: []string{`
        it('should be fixed', async () => {
          await expect(doSomethingAsync()).rejects.toThrow(); 
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncWrapperForExpectedPromise"},
				},
			},
			{
				Code: `
        it('should be fixed for async arrow function', async () => {
          await expect(async () => {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow(); 
        })
      `,
				Output: []string{`
        it('should be fixed for async arrow function', async () => {
          await expect(doSomethingAsync(1, 2)).rejects.toThrow(); 
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncWrapperForExpectedPromise"},
				},
			},
			{
				Code: `
        it('should be fixed for async normal function', async () => {
          await expect(async function () {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow(); 
        })
      `,
				Output: []string{`
        it('should be fixed for async normal function', async () => {
          await expect(doSomethingAsync(1, 2)).rejects.toThrow(); 
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncWrapperForExpectedPromise"},
				},
			},
			{
				Code: `
        it('should be fixed for Promise.all', async () => {
          await expect(async function () {
            await Promise.all([doSomethingAsync(1, 2), doSomethingAsync()]);
          }).rejects.toThrow(); 
        })
      `,
				Output: []string{`
        it('should be fixed for Promise.all', async () => {
          await expect(Promise.all([doSomethingAsync(1, 2), doSomethingAsync()])).rejects.toThrow(); 
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncWrapperForExpectedPromise"},
				},
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
				Output: []string{`
        it('should be fixed for async ref to expect', async () => {
          const a = async () => { await doSomethingAsync() };
          await expect(a()).rejects.toThrow();
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncWrapperForExpectedPromise"},
				},
			},
			{
				Code: `
        it('should be fixed for resolves', async () => {
          await expect(async () => {
            await doSomethingAsync();
          }).resolves.toBe(1);
        })
      `,
				Output: []string{`
        it('should be fixed for resolves', async () => {
          await expect(doSomethingAsync()).resolves.toBe(1);
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncWrapperForExpectedPromise"},
				},
			},
			{
				// rslint enhancement: tsgo exposes concise arrow bodies directly,
				// so the same safe unwrap can be applied without a block body.
				Code: `
        it('fixes concise async arrow functions', async () => {
          await expect(async () => await doSomethingAsync()).rejects.toThrow();
        })
      `,
				Output: []string{`
        it('fixes concise async arrow functions', async () => {
          await expect(doSomethingAsync()).rejects.toThrow();
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncWrapperForExpectedPromise"},
				},
			},
			{
				// rslint enhancement: parenthesized async function arguments are
				// matched even though ESTree-based upstream tests do not cover them.
				Code: `
        it('fixes parenthesized async functions', async () => {
          await expect((async () => {
            await doSomethingAsync();
          })).rejects.toThrow();
        })
      `,
				Output: []string{`
        it('fixes parenthesized async functions', async () => {
          await expect(doSomethingAsync()).rejects.toThrow();
        })
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncWrapperForExpectedPromise"},
				},
			},
		},
	)
}
