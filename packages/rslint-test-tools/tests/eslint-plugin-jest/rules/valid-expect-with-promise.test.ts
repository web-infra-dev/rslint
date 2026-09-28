import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const customPromiseDeclaration = `declare class CustomPromise<T> {
  then<R>(
    onFulfilled?: (value: T) => R,
    onRejected?: (error: unknown) => R,
  ): CustomPromise<R>;
}

declare const promised: CustomPromise<string>;`;

ruleTester.run('valid-expect-with-promise', {} as never, {
  valid: [
    { code: 'expect' },
    { code: 'expect.hasAssertions' },
    { code: 'expect.hasAssertions()' },
    { code: 'expect(a).toBe(b)' },
    { code: 'expect(Promise.resolve()).resolves.toBe(1)' },
    { code: 'expect(Promise.resolve()).rejects.toBe(1)' },
    {
      code: `it('is correct', async () => {
  await expect(Promise.resolve()).rejects.toEqual(1);
  await expect(Promise.reject()).resolves.not.toStrictEqual(1);

  expect(await Promise.resolve()).toEqual(1);
});`,
    },
    {
      code: `interface WithCode { code: string };

class MyError implements WithCode {}

const myError = new MyError();

expect(myError).toBeInstanceOf(Error);

expect(myError).toBe(new Error());
expect(myError).toEqual(new Error());
expect(myError).toStrictEqual(new Error());`,
    },
    {
      code: `const x: Array<number> = [1, 2, 3];

it('is true', async () => {
  expect(x).toEqual(null);
});`,
    },
    {
      code: '<T extends Promise<unknown> = Promise<string>>(v: T) => expect(v).resolves.toThrow()',
    },
    { code: '<T = string>(v: T) => expect(v).toBe(1)' },
    {
      code: `${customPromiseDeclaration}

it('works', async () => {
  await expect(promised).resolves.toBe('value');
});`,
      options: [{ checkThenables: true }],
    },
    {
      code: `${customPromiseDeclaration}

expect(promised).toEqual(1);`,
    },
    { code: 'expect({ then: 1 }).toBe(1)' },
    {
      code: 'expect({ then: 1 }).toBe(1)',
      options: [{ checkThenables: true }],
    },
    { code: 'expect(1).toBe(1)', options: [{ checkThenables: true }] },
    { code: 'expect().toBe(1)' },
    {
      code: `declare class Chainable {
  then(next: string): this;
}

declare const chain: Chainable;

expect(chain).toEqual(chain);`,
    },
    {
      code: `declare class Chainable {
  then(next: string): this;
}

declare const chain: Chainable;

expect(chain).toEqual(chain);`,
      options: [{ checkThenables: true }],
    },
    {
      code: `declare class Thenish<T> {
  then<R>(onFulfilled?: (value: T) => R): Thenish<R>;
}

declare const thenish: Thenish<string>;

expect(thenish).toEqual(1);`,
      options: [{ checkThenables: true }],
    },
  ],
  invalid: [
    {
      code: 'expect(Promise.resolve()).toBe(1)',
      errors: [
        {
          message: 'Subject is a promise so resolve or reject should be used',
          messageId: 'poorlyExpectedPromise',
          line: 1,
          column: 1,
        },
      ],
    },
    {
      code: 'expect(new Promise(r => r())).toBe(1)',
      errors: [{ messageId: 'poorlyExpectedPromise', line: 1, column: 1 }],
    },
    {
      code: `const x = 'hello world';

expect(x as Promise<number>).toEqual('hello world');
expect(x as Promise<string[]> & Array<string>).not.toStrictEqual('hello world');
expect(x as Promise<string[]> | Array<string>).not.toStrictEqual('hello world');`,
      errors: [
        { messageId: 'poorlyExpectedPromise', line: 3, column: 1 },
        { messageId: 'poorlyExpectedPromise', line: 4, column: 1 },
      ],
    },
    {
      code: `declare function build<T>(): T;

expect(build<Promise<string>>()).toEqual('hello world');`,
      errors: [{ messageId: 'poorlyExpectedPromise', line: 3, column: 1 }],
    },
    {
      code: `class MyPromise extends Promise<unknown> {};

declare function build<T extends Promise<number>>(): T;

expect(build<typeof MyPromise>()).toEqual('hello world');`,
      errors: [{ messageId: 'poorlyExpectedPromise', line: 5, column: 1 }],
    },
    {
      code: `async function promiseValue(value: string): Promise<string> {
  return value;
}

expect(promiseValue('hello world')).toEqual('hello world');
expect(promiseValue()).not.toEqual([]);`,
      errors: [
        { messageId: 'poorlyExpectedPromise', line: 5, column: 1 },
        { messageId: 'poorlyExpectedPromise', line: 6, column: 1 },
      ],
    },
    {
      code: `class PromisedString extends Promise<string> {}

it('works', async () => {
  await expect(PromisedString.resolve("hello sunshine")).toEqual(1);
  await expect(new PromisedString(r => r("value"))).toEqual(1);
});`,
      errors: [
        { messageId: 'poorlyExpectedPromise', line: 4, column: 9 },
        { messageId: 'poorlyExpectedPromise', line: 5, column: 9 },
      ],
    },
    {
      code: `${customPromiseDeclaration}

expect(promised).toEqual(1);`,
      options: [{ checkThenables: true }],
      errors: [{ messageId: 'poorlyExpectedPromise', line: 10, column: 1 }],
    },
    {
      code: `${customPromiseDeclaration}

it('works', async () => {
  await expect(promised).resolves.toBe('value');
});`,
      errors: [{ messageId: 'unneededRejectResolve', line: 11, column: 26 }],
    },
    {
      code: 'expect(Promise.resolve()).toBeInstanceOf(Promise);',
      errors: [{ messageId: 'poorlyExpectedPromise', line: 1, column: 1 }],
    },
    {
      code: 'expect("hello world").resolves.toContain(1)',
      errors: [
        {
          message: 'Subject is not a promise so resolves is not needed',
          messageId: 'unneededRejectResolve',
          line: 1,
          column: 23,
        },
      ],
    },
    {
      code: 'expect({}).rejects.not.toContain(1)',
      errors: [
        {
          message: 'Subject is not a promise so rejects is not needed',
          messageId: 'unneededRejectResolve',
          line: 1,
          column: 12,
        },
      ],
    },
    {
      code: `it('works', async () => {
  await expect(0).resolves.toEqual(1);
});`,
      errors: [{ messageId: 'unneededRejectResolve', line: 2, column: 19 }],
    },
    {
      code: `class PromisedString extends Promise<string> {}

it('works', async () => {
  const value = await PromisedString.resolve("hello sunshine");

  await expect(value).resolves.toEqual(1);
});`,
      errors: [{ messageId: 'unneededRejectResolve', line: 6, column: 23 }],
    },
  ],
});
