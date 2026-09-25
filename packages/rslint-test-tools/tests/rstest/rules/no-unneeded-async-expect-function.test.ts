import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unneeded-async-expect-function', {} as never, {
  valid: [
    'expect.hasAssertions()',
    `it('pass', async () => {
  expect();
})`,
    `it('pass', async () => {
  await expect(doSomethingAsync()).rejects.toThrow();
})`,
    `it('pass', async () => {
  await expect(doSomethingAsync(1, 2)).resolves.toBe(1);
})`,
    `it('pass', async () => {
  await expect(async () => {
    await doSomethingAsync();
    await doSomethingTwiceAsync(1, 2);
  }).rejects.toThrow();
})`,
    `import { expect as pleaseExpect } from '@rstest/core';
it('pass', async () => {
  await pleaseExpect(doSomethingAsync()).rejects.toThrow();
})`,
    `it('pass', async () => {
  await expect(async () => {
    doSomethingAsync();
  }).rejects.toThrow();
})`,
    `it('pass', async () => {
  await expect(async () => {
    const a = 1;
    await doSomethingAsync(a);
  }).rejects.toThrow();
})`,
    `it('pass for non-async expect', async () => {
  await expect(() => {
    doSomethingSync(a);
  }).rejects.toThrow();
})`,
    `it('pass for await in expect', async () => {
  await expect(await doSomethingAsync()).rejects.toThrow();
})`,
    `it('pass for different matchers', async () => {
  await expect(await doSomething()).not.toThrow();
  await expect(await doSomething()).toHaveLength(2);
  await expect(await doSomething()).toHaveReturned();
  await expect(await doSomething()).not.toHaveBeenCalled();
  await expect(await doSomething()).not.toBeDefined();
  await expect(await doSomething()).toEqual(2);
})`,
    `it('pass for using await within for-loop', async () => {
  const b = [async () => Promise.resolve(1), async () => Promise.reject(2)];
  await expect(async() => {
    for (const a of b) {
      await b();
    }
  }).rejects.toThrow();
})`,
    `it('pass for using await within array', async () => {
  await expect(async() => [await Promise.reject(2)]).rejects.toThrow(2);
})`,
    // The assertion is about the function itself, so the wrapper stays.
    `it('pass without a promise modifier', async () => {
  expect(async () => {
    await doSomethingAsync();
  }).toBeInstanceOf(Function);
})`,
  ].map((code) => ({ code })),
  invalid: [
    {
      code: `it('is reported', async () => {
  await expect(async () => {
    await doSomethingAsync();
  }).rejects.toThrow();
})`,
      output: null,
      errors: [
        {
          messageId: 'noAsyncWrapperForExpectedPromise',
          line: 2,
          column: 16,
          endLine: 4,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestRemoveAsyncWrapper',
              output: `it('is reported', async () => {
  await expect(doSomethingAsync()).rejects.toThrow();
})`,
            },
          ],
        },
      ],
    },
    {
      code: `it('is reported', async () => {
  await expect(async () => await doSomethingAsync()).rejects.toThrow();
})`,
      output: null,
      errors: [
        {
          messageId: 'noAsyncWrapperForExpectedPromise',
          line: 2,
          column: 16,
          endLine: 2,
          endColumn: 52,
          suggestions: [
            {
              messageId: 'suggestRemoveAsyncWrapper',
              output: `it('is reported', async () => {
  await expect(doSomethingAsync()).rejects.toThrow();
})`,
            },
          ],
        },
      ],
    },
    {
      code: `it('is reported', async () => {
  await expect(async function () {
    await doSomethingAsync(1, 2);
  }).rejects.toThrow();
})`,
      output: null,
      errors: [
        {
          messageId: 'noAsyncWrapperForExpectedPromise',
          line: 2,
          column: 16,
          endLine: 4,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestRemoveAsyncWrapper',
              output: `it('is reported', async () => {
  await expect(doSomethingAsync(1, 2)).rejects.toThrow();
})`,
            },
          ],
        },
      ],
    },
    {
      code: `it('is reported for Promise.all', async () => {
  await expect(async function () {
    await Promise.all([doSomethingAsync(1, 2), doSomethingAsync()]);
  }).rejects.toThrow();
})`,
      output: null,
      errors: [
        {
          messageId: 'noAsyncWrapperForExpectedPromise',
          line: 2,
          column: 16,
          endLine: 4,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestRemoveAsyncWrapper',
              output: `it('is reported for Promise.all', async () => {
  await expect(Promise.all([doSomethingAsync(1, 2), doSomethingAsync()])).rejects.toThrow();
})`,
            },
          ],
        },
      ],
    },
    {
      code: `it('is reported for a local async binding', async () => {
  const a = async () => { await doSomethingAsync() };
  await expect(async () => {
    await a();
  }).rejects.toThrow();
})`,
      output: null,
      errors: [
        {
          messageId: 'noAsyncWrapperForExpectedPromise',
          line: 3,
          column: 16,
          endLine: 5,
          endColumn: 4,
          suggestions: [
            {
              messageId: 'suggestRemoveAsyncWrapper',
              output: `it('is reported for a local async binding', async () => {
  const a = async () => { await doSomethingAsync() };
  await expect(a()).rejects.toThrow();
})`,
            },
          ],
        },
      ],
    },
  ],
});
