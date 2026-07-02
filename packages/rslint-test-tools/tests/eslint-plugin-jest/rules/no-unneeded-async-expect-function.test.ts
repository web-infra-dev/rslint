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
  ],
  invalid: [
    {
      code: `
        it('should be fixed', async () => {
          await expect(async () => {
            await doSomethingAsync();
          }).rejects.toThrow(); 
        })
      `,
      output: `
        it('should be fixed', async () => {
          await expect(doSomethingAsync()).rejects.toThrow(); 
        })
      `,
      errors: [
        {
          endColumn: 4,
          column: 16,
          messageId: 'noAsyncWrapperForExpectedPromise',
        },
      ],
    },
    {
      code: `
        it('should be fixed', async () => {
          await expect(async function () {
            await doSomethingAsync();
          }).rejects.toThrow(); 
        })
      `,
      output: `
        it('should be fixed', async () => {
          await expect(doSomethingAsync()).rejects.toThrow(); 
        })
    `,
      errors: [
        {
          endColumn: 4,
          column: 16,
          messageId: 'noAsyncWrapperForExpectedPromise',
        },
      ],
    },
    {
      code: `
        it('should be fixed for async arrow function', async () => {
          await expect(async () => {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow(); 
        })
      `,
      output: `
        it('should be fixed for async arrow function', async () => {
          await expect(doSomethingAsync(1, 2)).rejects.toThrow(); 
        })
      `,
      errors: [
        {
          endColumn: 4,
          column: 16,
          messageId: 'noAsyncWrapperForExpectedPromise',
        },
      ],
    },
    {
      code: `
        it('should be fixed for async normal function', async () => {
          await expect(async function () {
            await doSomethingAsync(1, 2);
          }).rejects.toThrow(); 
        })
      `,
      output: `
        it('should be fixed for async normal function', async () => {
          await expect(doSomethingAsync(1, 2)).rejects.toThrow(); 
        })
      `,
      errors: [
        {
          endColumn: 4,
          column: 16,
          messageId: 'noAsyncWrapperForExpectedPromise',
        },
      ],
    },
    {
      code: `
        it('should be fixed for Promise.all', async () => {
          await expect(async function () {
            await Promise.all([doSomethingAsync(1, 2), doSomethingAsync()]);
          }).rejects.toThrow(); 
        })
      `,
      output: `
        it('should be fixed for Promise.all', async () => {
          await expect(Promise.all([doSomethingAsync(1, 2), doSomethingAsync()])).rejects.toThrow(); 
        })
      `,
      errors: [
        {
          endColumn: 4,
          column: 16,
          messageId: 'noAsyncWrapperForExpectedPromise',
        },
      ],
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
      output: `
        it('should be fixed for async ref to expect', async () => {
          const a = async () => { await doSomethingAsync() };
          await expect(a()).rejects.toThrow();
        })
      `,
      errors: [
        {
          endColumn: 4,
          column: 16,
          messageId: 'noAsyncWrapperForExpectedPromise',
        },
      ],
    },
    {
      code: `
        it('should be fixed for resolves', async () => {
          await expect(async () => {
            await doSomethingAsync();
          }).resolves.toBe(1);
        })
      `,
      output: `
        it('should be fixed for resolves', async () => {
          await expect(doSomethingAsync()).resolves.toBe(1);
        })
      `,
      errors: [
        {
          endColumn: 4,
          column: 16,
          messageId: 'noAsyncWrapperForExpectedPromise',
        },
      ],
    },
  ],
});
