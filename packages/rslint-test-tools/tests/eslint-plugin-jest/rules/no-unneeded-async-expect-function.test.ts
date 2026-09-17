import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unneeded-async-expect-function', {} as never, {
  valid: [
    { code: 'expect.hasAssertions()' },
    {
      code: `
      it('pass', async () => {
        expect();
      })
    `,
    },
    {
      code: `
      it('pass', async () => {
        await expect(doSomethingAsync()).rejects.toThrow();
      })
    `,
    },
    {
      code: `
      it('pass', async () => {
        await expect(doSomethingAsync(1, 2)).resolves.toBe(1);
      })
    `,
    },
    {
      code: `
      it('pass', async () => {
        await expect(async () => {
          await doSomethingAsync();
          await doSomethingTwiceAsync(1, 2);
        }).rejects.toThrow();
      })
    `,
    },
    {
      code: `
        import { expect as pleaseExpect } from '@jest/globals';
        it('pass', async () => {
          await pleaseExpect(doSomethingAsync()).rejects.toThrow();
        })
      `,
    },
    {
      code: `
      it('pass', async () => {
        await expect(async () => {
          doSomethingAsync();
        }).rejects.toThrow();
      })
    `,
    },
    {
      code: `
      it('pass', async () => {
        await expect(async () => {
          const a = 1;
          await doSomethingAsync(a);
        }).rejects.toThrow();
      })
    `,
    },
    {
      code: `
      it('pass for non-async expect', async () => {
        await expect(() => {
          doSomethingSync(a);
        }).rejects.toThrow();
      })
    `,
    },
    {
      code: `
      it('pass for await in expect', async () => {
        await expect(await doSomethingAsync()).rejects.toThrow();
      })
    `,
    },
    {
      code: `
      it('pass for different matchers', async () => {
        await expect(await doSomething()).not.toThrow();
        await expect(await doSomething()).toHaveLength(2);
        await expect(await doSomething()).toHaveReturned();
        await expect(await doSomething()).not.toHaveBeenCalled();
        await expect(await doSomething()).not.toBeDefined();
        await expect(await doSomething()).toEqual(2);
      })
    `,
    },
    {
      code: `
      it('pass for using await within for-loop', async () => {
        const b = [async () => Promise.resolve(1), async () => Promise.reject(2)];
        await expect(async() => {
          for (const a of b) {
            await b();
          }
        }).rejects.toThrow();
      })
    `,
    },
    {
      code: `
      it('pass for using await within array', async () => {
        await expect(async() => [await Promise.reject(2)]).rejects.toThrow(2);
      })
    `,
    },
    {
      code: `
      it('does not unwrap awaited identifiers', async () => {
        const promise = doSomethingAsync();
        await expect(async () => {
          await promise;
        }).rejects.toThrow();
      })
    `,
    },
    {
      code: `expect(async () => { /* preserve why */ await doSomethingAsync(); }).rejects.toThrow();`,
    },
    {
      code: `expect(async () => { await doSomethingAsync(); }, makeMessage()).rejects.toThrow();`,
    },
    {
      code: `expect(async function () { await client.arguments(); }).rejects.toThrow();`,
    },
    {
      code: `function returnsPlainValue() { return 1; }
expect(async () => { await returnsPlainValue(); }).resolves.toBe(1);`,
    },
    {
      code: `function throwsSynchronously() { throw new Error('sync failure'); }
expect(async () => { await throwsSynchronously(); }).rejects.toThrow('sync failure');`,
    },
    {
      code: `expect(async () => { await doSomethingAsync(1, 2); }).rejects.toThrow();`,
    },
    {
      code: `const operation = async (value) => value;
expect(async () => { await operation(1); }).resolves.toBe(1);`,
    },
    {
      code: `let operation = async () => 1;
expect(async () => { await operation(); }).resolves.toBe(1);`,
    },
    {
      code: `async function operation() { return 1; }
operation = () => 1;
expect(async () => { await operation(); }).resolves.toBe(1);`,
    },
    {
      code: `async function operation() { return 1; }
expect(async (value = 1) => { await operation(); }).resolves.toBe(1);`,
    },
    {
      code: `async function operation() { return 1; }
expect(async () => { await operation(); }).resolves.toBeUndefined();`,
    },
    {
      code: `async function operation() { return 1; }
expect(async () => { await operation(); }).rejects.toThrow();`,
    },
    {
      code: `async function operation() { return 1; }
expect(async () => await operation()).resolves.toBe(1);`,
    },
    {
      code: `const operation = async () => 1;
expect(operation).resolves.toBe(1);`,
    },
    {
      code: `const operation = async () => { throw new Error('failure'); };
expect(operation).rejects.toThrow('failure');`,
    },
    {
      code: `expect(async () => await operation()).resolves.toBe(1);
const operation = async () => 1;`,
    },
    {
      code: `const operation = async <Value,>() => undefined as Value;
expect(async () => await operation<number>()).resolves.toBe(1);`,
    },
    {
      code: `
        it('keeps calls with arguments', async () => {
          await expect(async () => {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow();
        })
      `,
    },
    {
      code: `
        it('keeps function calls with arguments', async () => {
          await expect(async function () {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow();
        })
      `,
    },
    {
      code: `
        it('keeps member calls', async () => {
          await expect(async function () {
            await Promise.all([doSomethingAsync(1, 2), doSomethingAsync()]);
          }).rejects.toThrow();
        })
      `,
    },
    {
      code: `
        it('should be fixed for async ref to expect', async () => {
          const a = async () => { await doSomethingAsync() };
          await expect(async () => {
            await a();
          }).rejects.toThrow();
        })
      `,
    },
    {
      code: `expect(async () => await doSomethingAsync()).rejects.toThrow();`,
    },
    {
      code: `expect((async () => { await doSomethingAsync(); })).rejects.toThrow();`,
    },
    {
      code: `expect(async function* () { await operation(); }).resolves.toBe(1);`,
    },
    // Keep every original upstream-aligned input verbatim. These calls are
    // unresolved, so the current rule cannot prove that the wrapper is
    // unnecessary and must leave them alone.
    {
      code: `
        it('should be fixed', async () => {
          await expect(async () => {
            await doSomethingAsync();
          }).rejects.toThrow();
        })
      `,
    },
    {
      code: `
        it('should be fixed', async () => {
          await expect(async function () {
            await doSomethingAsync();
          }).rejects.toThrow();
        })
      `,
    },
    {
      code: `
        it('should be fixed for async arrow function', async () => {
          await expect(async () => {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow();
        })
      `,
    },
    {
      code: `
        it('should be fixed for async normal function', async () => {
          await expect(async function () {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow(); 
        })
      `,
    },
    {
      code: `
        it('should be fixed for Promise.all', async () => {
          await expect(async function () {
            await Promise.all([doSomethingAsync(1, 2), doSomethingAsync()]);
          }).rejects.toThrow(); 
        })
      `,
    },
    {
      code: `
        it('should be fixed for resolves', async () => {
          await expect(async () => {
            await doSomethingAsync();
          }).resolves.toBe(1);
        })
      `,
    },
  ],
  invalid: [
    {
      code: `const operation = async () => 1;
expect(async () => /* preserve why */ await operation()).resolves.toBe(1);`,
      errors: [{ messageId: 'noAsyncWrapperForExpectedPromise' }],
    },
    {
      code: `const operation = async () => 1;
expect(async () => await operation(), 'custom message').resolves.toBe(1);`,
      errors: [{ messageId: 'noAsyncWrapperForExpectedPromise' }],
    },
    {
      code: `const operation = async () => 1;
expect(async () => await operation()).resolves.toBe(1);`,
      errors: [{ messageId: 'noAsyncWrapperForExpectedPromise' }],
    },
    {
      code: `const operation = async () => { throw new Error('failure'); };
expect(async () => await operation()).rejects.toThrow('failure');`,
      errors: [{ messageId: 'noAsyncWrapperForExpectedPromise' }],
    },
  ],
});
