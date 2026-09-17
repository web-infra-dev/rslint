import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unneeded-async-expect-function', {} as never, {
  valid: [
    { code: 'expect.hasAssertions()' },
    { code: `test('empty expect', async () => { expect(); });` },
    {
      code: `test('direct rejection', async () => { await expect(loadUser()).rejects.toThrow(); });`,
    },
    {
      code: `test('direct resolution', async () => { await expect(loadUser(1, 2)).resolves.toBe(1); });`,
    },
    {
      code: `test('multiple operations', async () => {
  await expect(async () => {
    await createUser();
    await notifyUser();
  }).rejects.toThrow();
});`,
    },
    {
      code: `import { expect as check } from '@rstest/core';
test('direct imported expect', async () => { await check(loadUser()).rejects.toThrow(); });`,
    },
    {
      code: `test('unawaited body call', async () => {
  await expect(async () => { loadUser(); }).rejects.toThrow();
});`,
    },
    {
      code: `test('declaration before await', async () => {
  await expect(async () => { const value = 1; await loadUser(value); }).rejects.toThrow();
});`,
    },
    {
      code: `test('non-async wrapper', async () => {
  await expect(() => { loadUser(); }).rejects.toThrow();
});`,
    },
    {
      code: `test('awaited expect argument', async () => {
  await expect(await loadUser()).rejects.toThrow();
});`,
    },
    {
      code: `test('sync matchers', async () => {
  await expect(await loadUser()).not.toThrow();
  await expect(await loadUser()).toHaveLength(2);
  await expect(await loadUser()).toHaveReturned();
  await expect(await loadUser()).not.toHaveBeenCalled();
  await expect(await loadUser()).not.toBeDefined();
  await expect(await loadUser()).toEqual(2);
});`,
    },
    {
      code: `test('loop body', async () => {
  const callbacks = [async () => Promise.resolve(1), async () => Promise.reject(2)];
  await expect(async () => { for (const callback of callbacks) await callback(); }).rejects.toThrow();
});`,
    },
    {
      code: `test('array body', async () => {
  await expect(async () => [await Promise.reject(2)]).rejects.toThrow(2);
});`,
    },
    {
      code: `test('keeps reject arrow', async () => {
  await expect(async () => { await loadUser(); }).rejects.toThrow();
});`,
    },
    {
      code: `test('keeps concise reject arrow', async () => {
  await expect(async () => await loadUser()).rejects.toThrow();
});`,
    },
    {
      code: `test('keeps reject function', async () => {
  await expect(async function () { await loadUser(); }).rejects.toThrow();
});`,
    },
    {
      code: `test('keeps reject arrow arguments', async () => {
  await expect(async () => { await loadUser(1, 2); }).rejects.toThrow();
});`,
    },
    {
      code: `test('keeps reject function arguments', async () => {
  await expect(async function () { await loadUser(1, 2); }).rejects.toThrow();
});`,
    },
    {
      code: `test('keeps reject Promise.all', async () => {
  await expect(async function () { await Promise.all([loadUser(1), loadUser(2)]); }).rejects.toThrow();
});`,
    },
    {
      code: `test('keeps reject reference', async () => {
  const operation = async () => { await loadUser(); };
  await expect(async () => { await operation(); }).rejects.toThrow();
});`,
    },
    {
      code: `function throwsSynchronously(): Promise<never> {
  throw new Error('sync failure');
}
test('keeps a wrapper that converts a synchronous throw', async () => {
  await expect(async () => { await throwsSynchronously(); }).rejects.toThrow('sync failure');
});`,
    },
  ],
  invalid: [
    {
      code: `expect((async () => { await loadUser(); }) as any).resolves.toBe(1);`,
      errors: [{ messageId: 'noAsyncWrapperForExpectedPromise' }],
    },
    {
      code: `function returnsPlainValue() { return 1; }
test('rejects a function before calling it', async () => {
  await expect(async () => { await returnsPlainValue(); }).resolves.toBe(1);
});`,
      errors: [{ messageId: 'noAsyncWrapperForExpectedPromise' }],
    },
    {
      code: `function throwsSynchronously(): never { throw new Error('sync failure'); }
test('rejects a function before observing its throw', async () => {
  await expect(async () => { await throwsSynchronously(); }).resolves.toBe(1);
});`,
      errors: [{ messageId: 'noAsyncWrapperForExpectedPromise' }],
    },
    {
      code: `test('reports a resolves wrapper', async () => {
  await expect(async () => { await loadUser(); }).resolves.toEqual({ name: 'Ada' });
});`,
      errors: [
        {
          messageId: 'noAsyncWrapperForExpectedPromise',
          message:
            'Avoid wrapping asynchronous expectations in an unnecessary async function.',
        },
      ],
    },
    {
      code: `import { expect as check } from '@rstest/core';
await check(async () => await loadUser()).resolves.toEqual({ name: 'Ada' });`,
      errors: [{ messageId: 'noAsyncWrapperForExpectedPromise' }],
    },
  ],
});
