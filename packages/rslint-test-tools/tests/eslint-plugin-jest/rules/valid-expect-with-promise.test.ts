// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('valid-expect-with-promise', {
  valid: [
    ...(
      [
        'expect',
        'expect.hasAssertions',
        'expect.hasAssertions()',
        'expect(a).toBe(b)',
        'expect(Promise.resolve()).resolves.toBe(1)',
        'expect(Promise.resolve()).rejects.toBe(1)',
        `it('is correct', async () => {
  await expect(Promise.resolve()).rejects.toEqual(1);
  await expect(Promise.reject()).resolves.not.toStrictEqual(1);

  expect(await Promise.resolve()).toEqual(1);
});`,
        `interface WithCode { code: string };

class MyError implements WithCode {}

const myError = new MyError();

expect(myError).toBeInstanceOf(Error);

expect(myError).toBe(new Error());
expect(myError).toEqual(new Error());
expect(myError).toStrictEqual(new Error());`,
        `const x: Array<number> = [1, 2, 3];

it('is true', async () => {
  expect(x).toEqual(null);
});`,
        '<T extends Promise<unknown> = Promise<string>>(v: T) => expect(v).resolves.toThrow()',
        '<T = string>(v: T) => expect(v).toBe(1)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 0 valid 11',
      code: `declare class CustomPromise<T> {
  then<R>(
    onFulfilled?: (value: T) => R,
    onRejected?: (error: unknown) => R,
  ): CustomPromise<R>;
}

declare const promised: CustomPromise<string>;

it('works', async () => {
  await expect(promised).resolves.toBe('value');
});`,
      options: [
        {
          checkThenables: true,
        },
      ],
    },
    {
      name: 'run 0 valid 12',
      code: `declare class CustomPromise<T> {
  then<R>(
    onFulfilled?: (value: T) => R,
    onRejected?: (error: unknown) => R,
  ): CustomPromise<R>;
}

declare const promised: CustomPromise<string>;

expect(promised).toEqual(1);`,
    },
    {
      name: 'run 0 valid 13',
      code: 'expect({ then: 1 }).toBe(1)',
    },
    ...(['expect({ then: 1 }).toBe(1)', 'expect(1).toBe(1)'] as const).map(
      (code, index): ValidTestCase => ({
        name: `run 0 valid ${index + 14}`,
        code,
        options: [
          {
            checkThenables: true,
          },
        ],
      }),
    ),
    {
      name: 'run 0 valid 16',
      code: 'expect().toBe(1)',
    },
    {
      name: 'run 0 valid 17',
      code: `declare class Chainable {
  then(next: string): this;
}

declare const chain: Chainable;

expect(chain).toEqual(chain);`,
    },
    ...(
      [
        `declare class Chainable {
  then(next: string): this;
}

declare const chain: Chainable;

expect(chain).toEqual(chain);`,
        `declare class Thenish<T> {
  then<R>(onFulfilled?: (value: T) => R): Thenish<R>;
}

declare const thenish: Thenish<string>;

expect(thenish).toEqual(1);`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 18}`,
      code,
      options: [
        {
          checkThenables: true,
        },
      ],
    })),
    ...(
      [
        'expect',
        'expect.hasAssertions',
        'expect.hasAssertions()',
        'expect(a).toBe(b)',
        'expect(Promise.resolve()).resolves.toBe(1)',
        'expect(Promise.resolve()).rejects.toBe(1)',
        `it('is correct', async () => {
  await expect(Promise.resolve()).rejects.toEqual(1);
  await expect(Promise.reject()).resolves.not.toStrictEqual(1);

  expect(await Promise.resolve()).toEqual(1);
});`,
        `interface WithCode { code: string };

class MyError implements WithCode {}

const myError = new MyError();

expect(myError).toBeInstanceOf(Error);

expect(myError).toBe(new Error());
expect(myError).toEqual(new Error());
expect(myError).toStrictEqual(new Error());`,
        `const x: Array<number> = [1, 2, 3];

it('is true', async () => {
  expect(x).toEqual(null);
});`,
        '<T extends Promise<unknown> = Promise<string>>(v: T) => expect(v).resolves.toThrow()',
        '<T = string>(v: T) => expect(v).toBe(1)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-expect-with-promise / valid ${index}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
    })),
    {
      name: 'upstream valid-expect-with-promise / valid 11',
      code: `declare class CustomPromise<T> {
then<R>(
  onFulfilled?: (value: T) => R,
  onRejected?: (error: unknown) => R,
): CustomPromise<R>;
}

declare const promised: CustomPromise<string>;

      it('works', async () => {
        await expect(promised).resolves.toBe('value');
      });`,
      filename: 'fixtures/upstream/file.ts',
      options: [
        {
          checkThenables: true,
        },
      ],
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
    },
    ...(
      [
        `declare class CustomPromise<T> {
then<R>(
  onFulfilled?: (value: T) => R,
  onRejected?: (error: unknown) => R,
): CustomPromise<R>;
}

declare const promised: CustomPromise<string>;

    expect(promised).toEqual(1);`,
        'expect({ then: 1 }).toBe(1)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-expect-with-promise / valid ${index + 12}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
    })),
    ...(['expect({ then: 1 }).toBe(1)', 'expect(1).toBe(1)'] as const).map(
      (code, index): ValidTestCase => ({
        name: `upstream valid-expect-with-promise / valid ${index + 14}`,
        code,
        filename: 'fixtures/upstream/file.ts',
        options: [
          {
            checkThenables: true,
          },
        ],
        languageOptions: {
          sourceType: 'module',
          parserOptions: {
            project: ['fixtures/upstream/tsconfig.json'],
          },
        },
      }),
    ),
    ...(
      [
        'expect().toBe(1)',
        `declare class Chainable {
  then(next: string): this;
}

declare const chain: Chainable;

expect(chain).toEqual(chain);`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-expect-with-promise / valid ${index + 16}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
    })),
    ...(
      [
        `declare class Chainable {
  then(next: string): this;
}

declare const chain: Chainable;

expect(chain).toEqual(chain);`,
        `declare class Thenish<T> {
  then<R>(onFulfilled?: (value: T) => R): Thenish<R>;
}

declare const thenish: Thenish<string>;

expect(thenish).toEqual(1);`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream valid-expect-with-promise / valid ${index + 18}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      options: [
        {
          checkThenables: true,
        },
      ],
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
    })),
  ],
  invalid: [
    ...(
      [
        ['expect(Promise.resolve()).toBe(1)', 34],
        ['expect(new Promise(r => r())).toBe(1)', 38],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 2',
      code: `const x = 'hello world';

expect(x as Promise<number>).toEqual('hello world');
expect(x as Promise<string[]> & Array<string>).not.toStrictEqual('hello world');
expect(x as Promise<string[]> | Array<string>).not.toStrictEqual('hello world');`,
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 52,
        },
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 80,
        },
      ],
    },
    ...(
      [
        [
          `declare function build<T>(): T;

expect(build<Promise<string>>()).toEqual('hello world');`,
          3,
          56,
        ],
        [
          `class MyPromise extends Promise<unknown> {};

declare function build<T extends Promise<number>>(): T;

expect(build<typeof MyPromise>()).toEqual('hello world');`,
          5,
          57,
        ],
      ] as const
    ).map(([code, line, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 3}`,
      code,
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line,
          column: 1,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `async function promiseValue(value: string): Promise<string> {
  return value;
}

expect(promiseValue('hello world')).toEqual('hello world');
expect(promiseValue()).not.toEqual([]);`,
          5,
          1,
          59,
          6,
          1,
          39,
        ],
        [
          `class PromisedString extends Promise<string> {}

it('works', async () => {
  await expect(PromisedString.resolve("hello sunshine")).toEqual(1);
  await expect(new PromisedString(r => r("value"))).toEqual(1);
});`,
          4,
          9,
          68,
          5,
          9,
          63,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, endColumn1, line2, column2, endColumn2],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index + 5}`,
        code,
        errors: [
          {
            messageId: 'poorlyExpectedPromise',
            message: 'Subject is a promise so resolve or reject should be used',
            line: line1,
            column: column1,
            endLine: line1,
            endColumn: endColumn1,
          },
          {
            messageId: 'poorlyExpectedPromise',
            message: 'Subject is a promise so resolve or reject should be used',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    {
      name: 'run 0 invalid 7',
      code: `declare class CustomPromise<T> {
  then<R>(
    onFulfilled?: (value: T) => R,
    onRejected?: (error: unknown) => R,
  ): CustomPromise<R>;
}

declare const promised: CustomPromise<string>;

expect(promised).toEqual(1);`,
      options: [
        {
          checkThenables: true,
        },
      ],
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 10,
          column: 1,
          endLine: 10,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: `declare class CustomPromise<T> {
  then<R>(
    onFulfilled?: (value: T) => R,
    onRejected?: (error: unknown) => R,
  ): CustomPromise<R>;
}

declare const promised: CustomPromise<string>;

it('works', async () => {
  await expect(promised).resolves.toBe('value');
});`,
      errors: [
        {
          messageId: 'unneededRejectResolve',
          message: 'Subject is not a promise so resolves is not needed',
          line: 11,
          column: 26,
          endLine: 11,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'expect(Promise.resolve()).toBeInstanceOf(Promise);',
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    ...(
      [
        [
          'expect("hello world").resolves.toContain(1)',
          'Subject is not a promise so resolves is not needed',
          1,
          23,
          31,
        ],
        [
          'expect({}).rejects.not.toContain(1)',
          'Subject is not a promise so rejects is not needed',
          1,
          12,
          19,
        ],
        [
          `it('works', async () => {
  await expect(0).resolves.toEqual(1);
});`,
          'Subject is not a promise so resolves is not needed',
          2,
          19,
          27,
        ],
        [
          `class PromisedString extends Promise<string> {}

it('works', async () => {
  const value = await PromisedString.resolve("hello sunshine");

  await expect(value).resolves.toEqual(1);
});`,
          'Subject is not a promise so resolves is not needed',
          6,
          23,
          31,
        ],
      ] as const
    ).map(([code, message, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 10}`,
      code,
      errors: [
        {
          messageId: 'unneededRejectResolve',
          message,
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['expect(Promise.resolve()).toBe(1)', 34],
        ['expect(new Promise(r => r())).toBe(1)', 38],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream valid-expect-with-promise / invalid ${index}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream valid-expect-with-promise / invalid 2',
      code: `const x = 'hello world';

expect(x as Promise<number>).toEqual('hello world');
expect(x as Promise<string[]> & Array<string>).not.toStrictEqual('hello world');
expect(x as Promise<string[]> | Array<string>).not.toStrictEqual('hello world');`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 52,
        },
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 80,
        },
      ],
    },
    ...(
      [
        [
          `declare function build<T>(): T;

expect(build<Promise<string>>()).toEqual('hello world');`,
          3,
          56,
        ],
        [
          `class MyPromise extends Promise<unknown> {};

declare function build<T extends Promise<number>>(): T;

expect(build<typeof MyPromise>()).toEqual('hello world');`,
          5,
          57,
        ],
      ] as const
    ).map(([code, line, endColumn], index): TestCase => ({
      name: `upstream valid-expect-with-promise / invalid ${index + 3}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line,
          column: 1,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `async function promiseValue(value: string): Promise<string> {
  return value;
}

expect(promiseValue('hello world')).toEqual('hello world');
expect(promiseValue()).not.toEqual([]);`,
          5,
          1,
          59,
          6,
          1,
          39,
        ],
        [
          `class PromisedString extends Promise<string> {}

it('works', async () => {
  await expect(PromisedString.resolve("hello sunshine")).toEqual(1);
  await expect(new PromisedString(r => r("value"))).toEqual(1);
});`,
          4,
          9,
          68,
          5,
          9,
          63,
        ],
      ] as const
    ).map(
      (
        [code, line1, column1, endColumn1, line2, column2, endColumn2],
        index,
      ): TestCase => ({
        name: `upstream valid-expect-with-promise / invalid ${index + 5}`,
        code,
        filename: 'fixtures/upstream/file.ts',
        languageOptions: {
          sourceType: 'module',
          parserOptions: {
            project: ['fixtures/upstream/tsconfig.json'],
          },
        },
        errors: [
          {
            messageId: 'poorlyExpectedPromise',
            message: 'Subject is a promise so resolve or reject should be used',
            line: line1,
            column: column1,
            endLine: line1,
            endColumn: endColumn1,
          },
          {
            messageId: 'poorlyExpectedPromise',
            message: 'Subject is a promise so resolve or reject should be used',
            line: line2,
            column: column2,
            endLine: line2,
            endColumn: endColumn2,
          },
        ],
      }),
    ),
    {
      name: 'upstream valid-expect-with-promise / invalid 7',
      code: `declare class CustomPromise<T> {
then<R>(
  onFulfilled?: (value: T) => R,
  onRejected?: (error: unknown) => R,
): CustomPromise<R>;
}

declare const promised: CustomPromise<string>;

      expect(promised).toEqual(1);`,
      filename: 'fixtures/upstream/file.ts',
      options: [
        {
          checkThenables: true,
        },
      ],
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 10,
          column: 7,
          endLine: 10,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'upstream valid-expect-with-promise / invalid 8',
      code: `declare class CustomPromise<T> {
then<R>(
  onFulfilled?: (value: T) => R,
  onRejected?: (error: unknown) => R,
): CustomPromise<R>;
}

declare const promised: CustomPromise<string>;

      it('works', async () => {
        await expect(promised).resolves.toBe('value');
      });`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unneededRejectResolve',
          message: 'Subject is not a promise so resolves is not needed',
          line: 11,
          column: 32,
          endLine: 11,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream valid-expect-with-promise / invalid 9',
      code: 'expect(Promise.resolve()).toBeInstanceOf(Promise);',
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    ...(
      [
        [
          'expect("hello world").resolves.toContain(1)',
          'Subject is not a promise so resolves is not needed',
          1,
          23,
          31,
        ],
        [
          'expect({}).rejects.not.toContain(1)',
          'Subject is not a promise so rejects is not needed',
          1,
          12,
          19,
        ],
        [
          `it('works', async () => {
  await expect(0).resolves.toEqual(1);
});`,
          'Subject is not a promise so resolves is not needed',
          2,
          19,
          27,
        ],
        [
          `class PromisedString extends Promise<string> {}

it('works', async () => {
  const value = await PromisedString.resolve("hello sunshine");

  await expect(value).resolves.toEqual(1);
});`,
          'Subject is not a promise so resolves is not needed',
          6,
          23,
          31,
        ],
      ] as const
    ).map(([code, message, line, column, endColumn], index): TestCase => ({
      name: `upstream valid-expect-with-promise / invalid ${index + 10}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unneededRejectResolve',
          message,
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
  ],
});
