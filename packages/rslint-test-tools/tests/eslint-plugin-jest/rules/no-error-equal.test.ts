// Upstream reference: eslint-plugin-jest@29.16.6; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Shared source templates; reporting ranges remain explicit in each matrix.
const assertionSetup = `
const myError = new MyError();

expect(myError).toBeInstanceOf(Error);

expect(myError).toBe(new Error());
expect(myError).toEqual(new Error());
expect(myError).toStrictEqual(new Error());`;

new RuleTester(fixtureConfig()).run('no-error-equal', {
  valid: [
    ...(
      [
        'expect',
        'expect.hasAssertions',
        'expect.hasAssertions()',
        'expect(a).toBe(b)',
        'expect(a).toThrow(b)',
        'expect(a).toThrowError(b)',
        `class MyError {}

expect(new MyError()).toBeInstanceOf(Error);

expect(new MyError()).toBe(new Error());
expect(new MyError()).toEqual(new Error());
expect(new MyError()).toStrictEqual(new Error());`,
        `class MyError {}
` + assertionSetup,
        `interface WithCode { code: string };

class MyError implements WithCode {}
` + assertionSetup,
        `function buildError(): Error | { code: string } | null {
  return new Error('oh noes');
}

const x: Array<number> = [1, 2, 3];

expect(buildError()).toEqual(null);
expect(x).toEqual(null);
expect(buildError()).toEqual(new Error());`,
        `type Mx = Array<number>;

const x: Mx = [1, 2, 3];

expect(buildError()).toEqual(null);
expect(x).toEqual(null);
expect(buildError()).toEqual(new Error());`,
        '<T = Error>(v) => expect(v as T).toEqual(new Error())',
        'expect(new AggregateError()).toBe(0)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
      filename: 'src/virtual.ts',
    })),
    ...(
      [
        'expect',
        'expect.hasAssertions',
        'expect.hasAssertions()',
        'expect(a).toBe(b)',
        'expect(a).toThrow(b)',
        'expect(a).toThrowError(b)',
        `class MyError {}

expect(new MyError()).toBeInstanceOf(Error);

expect(new MyError()).toBe(new Error());
expect(new MyError()).toEqual(new Error());
expect(new MyError()).toStrictEqual(new Error());`,
        `class MyError {}
` + assertionSetup,
        `interface WithCode { code: string };

class MyError implements WithCode {}
` + assertionSetup,
        `function buildError(): Error | { code: string } | null {
  return new Error('oh noes');
}

const x: Array<number> = [1, 2, 3];

expect(buildError()).toEqual(null);
expect(x).toEqual(null);
expect(buildError()).toEqual(new Error());`,
        `type Mx = Array<number>;

const x: Mx = [1, 2, 3];

expect(buildError()).toEqual(null);
expect(x).toEqual(null);
expect(buildError()).toEqual(new Error());`,
        '<T = Error>(v) => expect(v as T).toEqual(new Error())',
        'expect(new AggregateError()).toBe(0)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-error-equal / valid ${index}`,
      code,
      filename: 'fixtures/upstream/file.ts',
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
        [
          'expect(new Error("hello world")).toEqual(new Error("hello world"))',
          67,
        ],
        ['expect(Error("hello world")).toEqual(new Error("hello world"))', 63],
        [
          'expect(new Error("hello world")).toStrictEqual(new Error("hello world"))',
          73,
        ],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `class MyError extends Error {}

expect(new MyError()).toBeInstanceOf(Error);

expect(new MyError()).toBe(new Error());
expect(new MyError()).toEqual(new Error());
expect(new MyError()).toStrictEqual(new Error());`,
          6,
          43,
          7,
          49,
        ],
        [
          `class MyError extends Error {}
` + assertionSetup,
          8,
          37,
          9,
          43,
        ],
        [
          `expect(new AggregateError([], 'hello world')).toBe(new Error());
expect(new AggregateError([], 'hello world')).toEqual(new Error());
expect(new AggregateError([], 'hello world')).toStrictEqual(new Error());`,
          2,
          67,
          3,
          73,
        ],
        [
          `const x = 'hello world';

expect(x as Error).toEqual('hello world');
expect(x as Error & {}).not.toStrictEqual('hello world');`,
          3,
          42,
          4,
          57,
        ],
      ] as const
    ).map(([code, line1, endColumn1, line2, endColumn2], index): TestCase => ({
      name: `run 0 invalid ${index + 3}`,
      code,
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: line1,
          column: 1,
          endLine: line1,
          endColumn: endColumn1,
        },
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: line2,
          column: 1,
          endLine: line2,
          endColumn: endColumn2,
        },
      ],
    })),
    {
      name: 'run 0 invalid 7',
      code: `declare function buildError<T>(): T;

expect(buildError<Error>()).toEqual('hello world');`,
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 51,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: `declare function addCode<T extends Error>(err: T, code: string): T & { code: string };

expect(addCode(new Error('hello world'), 'MODULE_NOT_FOUND')).toEqual('hello world');
expect(addCode(new AggregateError([], 'hello world'), 'MODULE_NOT_FOUND')).toEqual('hello world');`,
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 85,
        },
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 98,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: '<T extends Error>(v: any) => expect(v as T).toStrictEqual(new Error())',
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 71,
        },
      ],
    },
    ...(
      [
        `function buildError(msg: string) {
  return new Error(msg);
}

expect(buildError('hello world')).toEqual('hello world');
expect(buildError('hello world')).toEqual(new Error('hello world'));`,
        `function buildError(msg: string): Error {
  return new Error(msg);
}

expect(buildError('hello world')).toEqual('hello world');
expect(buildError('hello world')).toEqual(new Error('hello world'));`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index + 10}`,
      code,
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 57,
        },
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 68,
        },
      ],
    })),
    {
      name: 'run 0 invalid 12',
      code: `class MyGenericError extends Error {}
class MySpecificError extends MyGenericError {}

expect(new MySpecificError('hello world')).toEqual('hello world');
expect(MySpecificError('hello world')).toStrictEqual(new Error('hello world'));`,
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 66,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: `interface Mx { code: string }

class MyError extends Error implements Mx {}

expect(new MyError('hello world')).toEqual('hello world');
expect(new MyError('hello world')).not.toStrictEqual(new Error('hello world'));`,
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 58,
        },
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 79,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: `it('works', async () => {
  const err = await Promise.resolve(new Error('oh noes'));

  expect(err).toEqual(new Error('oh noes'));
});`,
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 44,
        },
      ],
    },
    ...(
      [
        [
          'expect(new Error("hello world")).toEqual(new Error("hello world"))',
          67,
        ],
        ['expect(Error("hello world")).toEqual(new Error("hello world"))', 63],
        [
          'expect(new Error("hello world")).toStrictEqual(new Error("hello world"))',
          73,
        ],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-error-equal / invalid ${index}`,
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
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `class MyError extends Error {}

expect(new MyError()).toBeInstanceOf(Error);

expect(new MyError()).toBe(new Error());
expect(new MyError()).toEqual(new Error());
expect(new MyError()).toStrictEqual(new Error());`,
          6,
          43,
          7,
          49,
        ],
        [
          `class MyError extends Error {}
` + assertionSetup,
          8,
          37,
          9,
          43,
        ],
        [
          `expect(new AggregateError([], 'hello world')).toBe(new Error());
expect(new AggregateError([], 'hello world')).toEqual(new Error());
expect(new AggregateError([], 'hello world')).toStrictEqual(new Error());`,
          2,
          67,
          3,
          73,
        ],
        [
          `const x = 'hello world';

expect(x as Error).toEqual('hello world');
expect(x as Error & {}).not.toStrictEqual('hello world');`,
          3,
          42,
          4,
          57,
        ],
      ] as const
    ).map(([code, line1, endColumn1, line2, endColumn2], index): TestCase => ({
      name: `upstream no-error-equal / invalid ${index + 3}`,
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
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: line1,
          column: 1,
          endLine: line1,
          endColumn: endColumn1,
        },
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: line2,
          column: 1,
          endLine: line2,
          endColumn: endColumn2,
        },
      ],
    })),
    {
      name: 'upstream no-error-equal / invalid 7',
      code: `declare function buildError<T>(): T;

expect(buildError<Error>()).toEqual('hello world');`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 51,
        },
      ],
    },
    {
      name: 'upstream no-error-equal / invalid 8',
      code: `declare function addCode<T extends Error>(err: T, code: string): T & { code: string };

expect(addCode(new Error('hello world'), 'MODULE_NOT_FOUND')).toEqual('hello world');
expect(addCode(new AggregateError([], 'hello world'), 'MODULE_NOT_FOUND')).toEqual('hello world');`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 85,
        },
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 98,
        },
      ],
    },
    {
      name: 'upstream no-error-equal / invalid 9',
      code: '<T extends Error>(v: any) => expect(v as T).toStrictEqual(new Error())',
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 71,
        },
      ],
    },
    ...(
      [
        `function buildError(msg: string) {
  return new Error(msg);
}

expect(buildError('hello world')).toEqual('hello world');
expect(buildError('hello world')).toEqual(new Error('hello world'));`,
        `function buildError(msg: string): Error {
  return new Error(msg);
}

expect(buildError('hello world')).toEqual('hello world');
expect(buildError('hello world')).toEqual(new Error('hello world'));`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream no-error-equal / invalid ${index + 10}`,
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
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 57,
        },
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 68,
        },
      ],
    })),
    {
      name: 'upstream no-error-equal / invalid 12',
      code: `class MyGenericError extends Error {}
class MySpecificError extends MyGenericError {}

expect(new MySpecificError('hello world')).toEqual('hello world');
expect(MySpecificError('hello world')).toStrictEqual(new Error('hello world'));`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 66,
        },
      ],
    },
    {
      name: 'upstream no-error-equal / invalid 13',
      code: `interface Mx { code: string }

class MyError extends Error implements Mx {}

expect(new MyError('hello world')).toEqual('hello world');
expect(new MyError('hello world')).not.toStrictEqual(new Error('hello world'));`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 58,
        },
        {
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 79,
        },
      ],
    },
    {
      name: 'upstream no-error-equal / invalid 14',
      code: `it('works', async () => {
  const err = await Promise.resolve(new Error('oh noes'));

  expect(err).toEqual(new Error('oh noes'));
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
          messageId: 'equalError',
          message: 'Avoid using equality matchers to check errors',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 44,
        },
      ],
    },
  ],
});
