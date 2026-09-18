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
			{Code: `expect(async () => { /* preserve why */ await doSomethingAsync(); }).rejects.toThrow();`},
			{Code: `expect(async () => { await doSomethingAsync(); }, makeMessage()).rejects.toThrow();`},
			{Code: `expect(async function () { await client.arguments(); }).rejects.toThrow();`},
			{Code: `function returnsPlainValue() { return 1; }
expect(async () => { await returnsPlainValue(); }).resolves.toBe(1);`},
			{Code: `function throwsSynchronously() { throw new Error('sync failure'); }
expect(async () => { await throwsSynchronously(); }).rejects.toThrow('sync failure');`},
			{Code: `expect(async () => { await doSomethingAsync(1, 2); }).rejects.toThrow();`},
			{Code: `const operation = async (value) => value;
expect(async () => { await operation(1); }).resolves.toBe(1);`},
			{Code: `let operation = async () => 1;
expect(async () => { await operation(); }).resolves.toBe(1);`},
			{Code: `async function operation() { return 1; }
operation = () => 1;
expect(async () => { await operation(); }).resolves.toBe(1);`},
			{Code: `async function operation() { return 1; }
expect(async (value = 1) => { await operation(); }).resolves.toBe(1);`},
			{Code: `async function operation() { return 1; }
expect(async () => { await operation(); }).resolves.toBeUndefined();`},
			{Code: `async function operation() { return 1; }
expect(async () => { await operation(); }).rejects.toThrow();`},
			{Code: `async function operation() { return 1; }
expect(async () => await operation()).resolves.toBe(1);`},
			{Code: `const operation = async () => 1;
expect(operation).resolves.toBe(1);`},
			{Code: `const operation = async () => { throw new Error('failure'); };
expect(operation).rejects.toThrow('failure');`},
			{Code: `expect(async () => await operation()).resolves.toBe(1);
const operation = async () => 1;`},
			{Code: `const operation = async <Value,>() => undefined as Value;
expect(async () => await operation<number>()).resolves.toBe(1);`},
			{Code: `
        it('keeps calls with arguments', async () => {
          await expect(async () => {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow();
        })
      `},
			{Code: `
        it('keeps function calls with arguments', async () => {
          await expect(async function () {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow();
        })
      `},
			{Code: `
        it('keeps member calls', async () => {
          await expect(async function () {
            await Promise.all([doSomethingAsync(1, 2), doSomethingAsync()]);
          }).rejects.toThrow();
        })
      `},
			{Code: `
        it('should be fixed for async ref to expect', async () => {
          const a = async () => { await doSomethingAsync() };
          await expect(async () => {
            await a();
          }).rejects.toThrow();
        })
      `},
			{Code: `expect(async () => await doSomethingAsync()).rejects.toThrow();`},
			{Code: `expect((async () => { await doSomethingAsync(); })).rejects.toThrow();`},
			{Code: `expect(async function* () { await operation(); }).resolves.toBe(1);`},
			// Keep every original upstream-aligned input verbatim. These calls are
			// unresolved, so the current rule cannot prove that the wrapper is
			// unnecessary and must leave them alone.
			{Code: `
        it('should be fixed', async () => {
          await expect(async () => {
            await doSomethingAsync();
          }).rejects.toThrow();
        })
      `},
			{Code: `
        it('should be fixed', async () => {
          await expect(async function () {
            await doSomethingAsync();
          }).rejects.toThrow();
        })
      `},
			{Code: `
        it('should be fixed for async arrow function', async () => {
          await expect(async () => {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow(); 
        })
      `},
			{Code: `
        it('should be fixed for async normal function', async () => {
          await expect(async function () {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow(); 
        })
      `},
			{Code: `
        it('should be fixed for Promise.all', async () => {
          await expect(async function () {
            await Promise.all([doSomethingAsync(1, 2), doSomethingAsync()]);
          }).rejects.toThrow(); 
        })
      `},
			{Code: `
        it('should be fixed for resolves', async () => {
          await expect(async () => {
            await doSomethingAsync();
          }).resolves.toBe(1);
        })
      `},
			{Code: `
        it('fixes concise async arrow functions', async () => {
          await expect(async () => await doSomethingAsync()).rejects.toThrow();
        })
      `},
			{Code: `
        it('fixes parenthesized async functions', async () => {
          await expect((async () => {
            await doSomethingAsync();
          })).rejects.toThrow();
        })
      `},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `const operation = async () => 1;
expect(async () => /* preserve why */ await operation()).resolves.toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noAsyncWrapperForExpectedPromise"}},
			},
			{
				Code: `const operation = async () => 1;
expect(async () => await operation(), 'custom message').resolves.toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noAsyncWrapperForExpectedPromise"}},
			},
			{
				Code: `const operation = async () => 1;
expect(async () => await operation()).resolves.toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noAsyncWrapperForExpectedPromise"}},
			},
			{
				Code: `const operation = async () => { throw new Error('failure'); };
expect(async () => await operation()).rejects.toThrow('failure');`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noAsyncWrapperForExpectedPromise"}},
			},
		},
	)
}
