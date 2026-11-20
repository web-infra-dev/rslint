// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-unnecessary-assertion', {
  valid: [
    ...(
      [
        'expect(a).toBe(b)',
        'expect.toBeNull()',
        'declare function mx(): string | null; expect(mx()).toBeNull();',
        'declare function mx(): string | undefined; expect(mx()).not.toBeUndefined();',
        'declare function mx(): unknown; expect(mx()).toBeDefined();',
        'declare function mx(): string | number; expect(mx()).toBeNaN();',
        'declare function mx(): Promise<string>; async function run() { await expect(mx()).resolves.toBeNull(); }',
        "import { expect } from 'vitest'; expect('hello').toBeNull();",
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'expect',
        'expect.hasAssertions',
        'expect.hasAssertions()',
        'expect(a).toBe(b)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-unnecessary-assertion (general) / valid ${index}`,
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
        'expect.toBeNull',
        'expect.toBeNull()',
        `const add = (a, b) => a + b;

expect(add(1, 1)).toBe(2);`,
        `const add = (a: number, b: number) => a + b;

expect(add(1, 1)).toBe(2);`,
        `const add = (a: number, b: number): number => a + b;

expect(add(1, 1)).toBe(2);`,
        `const add = (a: number, b: number): number | null => a + b;

expect(add(1, 1)).not.toBeNull();`,
        `declare function mx(): any;

expect(mx()).toBeNull();
expect(mx()).not.toBeNull();`,
        `declare function mx(): unknown;

expect(mx()).toBeNull();
expect(mx()).not.toBeNull();`,
        `declare function mx(): string | null;

expect(mx()).toBeNull();
expect(mx()).not.toBeNull();`,
        `declare function mx<T>(p: T): T | null;

expect(mx(null)).toBeNull();
expect(mx(null)).not.toBeNull();

expect(mx('hello')).toBeNull();
expect(mx('world')).not.toBeNull();`,
        'expect(null).not.toBeNull()',
        'expect("hello" as null).toBeNull();',
        `declare function mx(): Promise<string | null>;

it('is async', async () => {
  await expect(mx()).resolves.toBeNull();
});`,
        `declare function mx(): Promise<string | null>;

it('is async', async () => {
  await expect(mx()).rejects.toBeNull();
});`,
        `declare function mx(): Promise<string | null>;

it('is async', async () => {
  await expect(mx()).rejects.not.toBeNull();
});`,
        `declare function mx(): Promise<string | null>;

it('is async', async () => {
  expect(await mx()).not.toBeNull();
});`,
        `declare function mx(): Promise<string>;

it('is async', async () => {
  await expect(mx()).resolves.toBeNull();
});`,
        `declare function mx(): Promise<string>;

it('is async', async () => {
  await expect(mx()).rejects.not.toBeNull();
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-unnecessary-assertion (toBeNull) / valid ${index}`,
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
        [
          'expect.toBeDefined',
          'upstream no-unnecessary-assertion (toBeDefined) / valid 0',
        ],
        [
          'expect.toBeDefined()',
          'upstream no-unnecessary-assertion (toBeDefined) / valid 1',
        ],
        [
          `const add = (a: number, b: number): number | undefined => a + b;

expect(add(1, 1)).not.toBeDefined();`,
          'upstream no-unnecessary-assertion (toBeDefined) / valid 5',
        ],
        [
          `declare function mx(): any;

expect(mx()).toBeDefined();
expect(mx()).not.toBeDefined();`,
          'upstream no-unnecessary-assertion (toBeDefined) / valid 6',
        ],
        [
          `declare function mx(): unknown;

expect(mx()).toBeDefined();
expect(mx()).not.toBeDefined();`,
          'upstream no-unnecessary-assertion (toBeDefined) / valid 7',
        ],
        [
          `declare function mx(): string | undefined;

expect(mx()).toBeDefined();
expect(mx()).not.toBeDefined();`,
          'upstream no-unnecessary-assertion (toBeDefined) / valid 8',
        ],
        [
          `declare function mx<T>(p: T): T | undefined;

expect(mx(undefined)).toBeDefined();
expect(mx(undefined)).not.toBeDefined();

expect(mx('hello')).toBeDefined();
expect(mx('world')).not.toBeDefined();`,
          'upstream no-unnecessary-assertion (toBeDefined) / valid 9',
        ],
        [
          'expect(undefined).not.toBeDefined()',
          'upstream no-unnecessary-assertion (toBeDefined) / valid 10',
        ],
        [
          'expect("hello" as undefined).toBeDefined();',
          'upstream no-unnecessary-assertion (toBeDefined) / valid 11',
        ],
        [
          `declare function mx(): Promise<string | undefined>;

it('is async', async () => {
  await expect(mx()).resolves.toBeDefined();
});`,
          'upstream no-unnecessary-assertion (toBeDefined) / valid 12',
        ],
        [
          `declare function mx(): Promise<string | undefined>;

it('is async', async () => {
  await expect(mx()).rejects.toBeDefined();
});`,
          'upstream no-unnecessary-assertion (toBeDefined) / valid 13',
        ],
        [
          `declare function mx(): Promise<string | undefined>;

it('is async', async () => {
  await expect(mx()).rejects.not.toBeDefined();
});`,
          'upstream no-unnecessary-assertion (toBeDefined) / valid 14',
        ],
        [
          `declare function mx(): Promise<string | undefined>;

it('is async', async () => {
  expect(await mx()).not.toBeDefined();
});`,
          'upstream no-unnecessary-assertion (toBeDefined) / valid 15',
        ],
        [
          `declare function mx(): Promise<string>;

it('is async', async () => {
  await expect(mx()).resolves.toBeDefined();
});`,
          'upstream no-unnecessary-assertion (toBeDefined) / valid 16',
        ],
        [
          `declare function mx(): Promise<string>;

it('is async', async () => {
  await expect(mx()).rejects.not.toBeDefined();
});`,
          'upstream no-unnecessary-assertion (toBeDefined) / valid 17',
        ],
      ] as const
    ).map(([code, name]): ValidTestCase => ({
      name,
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
        [
          'expect.toBeUndefined',
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 0',
        ],
        [
          'expect.toBeUndefined()',
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 1',
        ],
        [
          `const add = (a: number, b: number): number | undefined => a + b;

expect(add(1, 1)).not.toBeUndefined();`,
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 5',
        ],
        [
          `declare function mx(): any;

expect(mx()).toBeUndefined();
expect(mx()).not.toBeUndefined();`,
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 6',
        ],
        [
          `declare function mx(): unknown;

expect(mx()).toBeUndefined();
expect(mx()).not.toBeUndefined();`,
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 7',
        ],
        [
          `declare function mx(): string | undefined;

expect(mx()).toBeUndefined();
expect(mx()).not.toBeUndefined();`,
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 8',
        ],
        [
          `declare function mx<T>(p: T): T | undefined;

expect(mx(undefined)).toBeUndefined();
expect(mx(undefined)).not.toBeUndefined();

expect(mx('hello')).toBeUndefined();
expect(mx('world')).not.toBeUndefined();`,
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 9',
        ],
        [
          'expect(undefined).not.toBeUndefined()',
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 10',
        ],
        [
          'expect("hello" as undefined).toBeUndefined();',
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 11',
        ],
        [
          `declare function mx(): Promise<string | undefined>;

it('is async', async () => {
  await expect(mx()).resolves.toBeUndefined();
});`,
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 12',
        ],
        [
          `declare function mx(): Promise<string | undefined>;

it('is async', async () => {
  await expect(mx()).rejects.toBeUndefined();
});`,
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 13',
        ],
        [
          `declare function mx(): Promise<string | undefined>;

it('is async', async () => {
  await expect(mx()).rejects.not.toBeUndefined();
});`,
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 14',
        ],
        [
          `declare function mx(): Promise<string | undefined>;

it('is async', async () => {
  expect(await mx()).not.toBeUndefined();
});`,
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 15',
        ],
        [
          `declare function mx(): Promise<string>;

it('is async', async () => {
  await expect(mx()).resolves.toBeUndefined();
});`,
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 16',
        ],
        [
          `declare function mx(): Promise<string>;

it('is async', async () => {
  await expect(mx()).rejects.not.toBeUndefined();
});`,
          'upstream no-unnecessary-assertion (toBeUndefined) / valid 17',
        ],
      ] as const
    ).map(([code, name]): ValidTestCase => ({
      name,
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
        [
          'expect.toBeNaN',
          'upstream no-unnecessary-assertion (toBeNaN) / valid 0',
        ],
        [
          'expect.toBeNaN()',
          'upstream no-unnecessary-assertion (toBeNaN) / valid 1',
        ],
        [
          'expect(0).toBeNaN()',
          'upstream no-unnecessary-assertion (toBeNaN) / valid 2',
        ],
        [
          'expect(0).not.toBeNaN()',
          'upstream no-unnecessary-assertion (toBeNaN) / valid 3',
        ],
        [
          `const x = 0;

expect(x).toBeNaN()
expect(x).not.toBeNaN()`,
          'upstream no-unnecessary-assertion (toBeNaN) / valid 4',
        ],
        [
          `const add = (a: number, b: number) => a.toString() + b.toString();

expect(add(1, 1)).toBe(2);`,
          'upstream no-unnecessary-assertion (toBeNaN) / valid 6',
        ],
        [
          `const add = (a: number, b: number): string => a.toString() + b.toString();

expect(add(1, 1)).toBe(2);`,
          'upstream no-unnecessary-assertion (toBeNaN) / valid 7',
        ],
        [
          `const add = (a: number, b: number): string | number => a.toString() + b.toString();

expect(add(1, 1)).toBe(2);`,
          'upstream no-unnecessary-assertion (toBeNaN) / valid 8',
        ],
        [
          `const add = (a: number, b: number): string | number => a.toString() + b.toString();

expect(add(1, 1)).toBeNaN();`,
          'upstream no-unnecessary-assertion (toBeNaN) / valid 9',
        ],
        [
          `const add = (a: number, b: number): string | number => a + b;

expect(add(1, 1)).not.toBeNaN();`,
          'upstream no-unnecessary-assertion (toBeNaN) / valid 10',
        ],
        [
          `declare function mx(): string | number;

expect(mx()).toBeNaN();
expect(mx()).not.toBeNaN();`,
          'upstream no-unnecessary-assertion (toBeNaN) / valid 11',
        ],
        [
          `declare function mx<T>(p: T): T | number;

expect(mx(42)).toBeNaN();
expect(mx(4.2)).toBeNaN();
expect(mx(Infinity)).not.toBeNaN();

expect(mx('hello')).toBeNaN();
expect(mx('world')).not.toBeNaN();`,
          'upstream no-unnecessary-assertion (toBeNaN) / valid 12',
        ],
        [
          'expect("hello" as number).toBeNaN();',
          'upstream no-unnecessary-assertion (toBeNaN) / valid 13',
        ],
      ] as const
    ).map(([code, name]): ValidTestCase => ({
      name,
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
          'expect(0).toBeNull()',
          'Unnecessary assertion, subject cannot be null',
          1,
          21,
        ],
        [
          'expect([]).not.toBeUndefined()',
          'Unnecessary assertion, subject cannot be undefined',
          1,
          31,
        ],
        [
          'expect("hello world").toBeNaN()',
          'Unnecessary assertion, subject cannot be a number',
          1,
          32,
        ],
        [
          "import { expect as check } from '@jest/globals'; check('hello')['toBeDefined']();",
          'Unnecessary assertion, subject cannot be undefined',
          50,
          81,
        ],
        [
          'function check<T>(value: T) { expect(value).toBeNull(); }',
          'Unnecessary assertion, subject cannot be null',
          31,
          55,
        ],
      ] as const
    ).map(([code, message, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message,
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream no-unnecessary-assertion (general) / invalid 0',
      code: 'expect(x).toBe(y);',
      filename: 'fixtures/upstream/unstrict/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/unstrict/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'noStrictNullCheck',
          message:
            'This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
    },
    ...(
      [
        ['expect(0).toBeNull()', 21],
        ['expect("hello world").toBeNull()', 33],
        ['expect({}).toBeNull()', 22],
        ['expect([]).not.toBeNull()', 26],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeNull) / invalid ${index}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream no-unnecessary-assertion (toBeNull) / invalid 4',
      code: `const x = 0;

expect(x).toBeNull()
expect(x).not.toBeNull()`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 21,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 25,
        },
      ],
    },
    ...(
      [
        [
          `const add = (a: number, b: number) => a + b;

expect(add(1, 1)).toBeNull();`,
          29,
        ],
        [
          `const add = (a: number, b: number): number => a + b;

expect(add(1, 1)).toBeNull();`,
          29,
        ],
        [
          `const add = (a: number, b: number): number => a + b;

expect(add(1, 1)).not.toBeNull();`,
          33,
        ],
        [
          `declare function mx(): never;

expect(mx()).not.toBeNull();`,
          28,
        ],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeNull) / invalid ${index + 5}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn,
        },
      ],
    })),
    ...(
      [
        `const result = "hello world".match("sunshine") ?? [];

expect(result).not.toBeNull();
expect(result).toBeNull();`,
        `const result = "hello world".match("sunshine") || [];

expect(result).not.toBeNull();
expect(result).toBeNull();`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeNull) / invalid ${index + 9}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 30,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 26,
        },
      ],
    })),
    {
      name: 'upstream no-unnecessary-assertion (toBeNull) / invalid 11',
      code: `declare function mx(): Promise<string>;

it('is async', async () => {
  expect(await mx()).toBeNull();
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 32,
        },
      ],
    },
    ...(
      [
        [
          `declare function mx(): string | number;

expect(mx()).toBeNull();
expect(mx()).not.toBeNull();`,
          3,
          24,
          4,
          28,
        ],
        [
          `declare function mx<T>(p: T): T;

expect(mx(null)).toBeNull();
expect(mx(null)).not.toBeNull();

expect(mx('hello')).toBeNull();
expect(mx('world')).not.toBeNull();`,
          6,
          31,
          7,
          35,
        ],
        [
          `declare function mx<T>(p: T): T extends string ? null : T;

expect(mx('hello')).toBeNull();
expect(mx('world')).not.toBeNull();

expect(mx({})).toBeNull();
expect(mx({})).not.toBeNull();`,
          6,
          26,
          7,
          30,
        ],
      ] as const
    ).map(([code, line1, endColumn1, line2, endColumn2], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeNull) / invalid ${index + 12}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: line1,
          column: 1,
          endLine: line1,
          endColumn: endColumn1,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: line2,
          column: 1,
          endLine: line2,
          endColumn: endColumn2,
        },
      ],
    })),
    {
      name: 'upstream no-unnecessary-assertion (toBeNull) / invalid 15',
      code: `declare function mx(): string | null;

expect(mx()!).toBeNull();
expect(mx()!).not.toBeNull();

expect(mx() as string).toBeNull();
expect(mx() as string).not.toBeNull();

expect(mx() as number).toBeNull();
expect(mx() as number).not.toBeNull();`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 25,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 29,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 34,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 7,
          column: 1,
          endLine: 7,
          endColumn: 38,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 9,
          column: 1,
          endLine: 9,
          endColumn: 34,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 10,
          column: 1,
          endLine: 10,
          endColumn: 38,
        },
      ],
    },
    ...(
      [
        ['expect(0).toBeDefined()', 24],
        ['expect("hello world").toBeDefined()', 36],
        ['expect({}).toBeDefined()', 25],
        ['expect([]).not.toBeDefined()', 29],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeDefined) / invalid ${index}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream no-unnecessary-assertion (toBeDefined) / invalid 4',
      code: `const x = 0;

expect(x).toBeDefined()
expect(x).not.toBeDefined()`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 24,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 28,
        },
      ],
    },
    ...(
      [
        [
          `const add = (a: number, b: number) => a + b;

expect(add(1, 1)).toBeDefined();`,
          32,
        ],
        [
          `const add = (a: number, b: number): number => a + b;

expect(add(1, 1)).toBeDefined();`,
          32,
        ],
        [
          `const add = (a: number, b: number): number => a + b;

expect(add(1, 1)).not.toBeDefined();`,
          36,
        ],
        [
          `declare function mx(): never;

expect(mx()).not.toBeDefined();`,
          31,
        ],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeDefined) / invalid ${index + 5}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn,
        },
      ],
    })),
    ...(
      [
        `const result = "hello world".match("sunshine") ?? [];

expect(result).not.toBeDefined();
expect(result).toBeDefined();`,
        `const result = "hello world".match("sunshine") || [];

expect(result).not.toBeDefined();
expect(result).toBeDefined();`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeDefined) / invalid ${index + 9}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 33,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 29,
        },
      ],
    })),
    {
      name: 'upstream no-unnecessary-assertion (toBeDefined) / invalid 11',
      code: `declare function mx(): Promise<string>;

it('is async', async () => {
  expect(await mx()).toBeDefined();
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 35,
        },
      ],
    },
    ...(
      [
        [
          `declare function mx(): string | number;

expect(mx()).toBeDefined();
expect(mx()).not.toBeDefined();`,
          3,
          27,
          4,
          31,
        ],
        [
          `declare function mx<T>(p: T): T;

expect(mx(undefined)).toBeDefined();
expect(mx(undefined)).not.toBeDefined();

expect(mx('hello')).toBeDefined();
expect(mx('world')).not.toBeDefined();`,
          6,
          34,
          7,
          38,
        ],
        [
          `declare function mx<T>(p: T): T extends string ? undefined : T;

expect(mx('hello')).toBeDefined();
expect(mx('world')).not.toBeDefined();

expect(mx({})).toBeDefined();
expect(mx({})).not.toBeDefined();`,
          6,
          29,
          7,
          33,
        ],
      ] as const
    ).map(([code, line1, endColumn1, line2, endColumn2], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeDefined) / invalid ${index + 12}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: line1,
          column: 1,
          endLine: line1,
          endColumn: endColumn1,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: line2,
          column: 1,
          endLine: line2,
          endColumn: endColumn2,
        },
      ],
    })),
    {
      name: 'upstream no-unnecessary-assertion (toBeDefined) / invalid 15',
      code: `declare function mx(): string | undefined;

expect(mx()!).toBeDefined();
expect(mx()!).not.toBeDefined();

expect(mx() as string).toBeDefined();
expect(mx() as string).not.toBeDefined();

expect(mx() as number).toBeDefined();
expect(mx() as number).not.toBeDefined();`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 28,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 32,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 37,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 7,
          column: 1,
          endLine: 7,
          endColumn: 41,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 9,
          column: 1,
          endLine: 9,
          endColumn: 37,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 10,
          column: 1,
          endLine: 10,
          endColumn: 41,
        },
      ],
    },
    ...(
      [
        ['expect(0).toBeUndefined()', 26],
        ['expect("hello world").toBeUndefined()', 38],
        ['expect({}).toBeUndefined()', 27],
        ['expect([]).not.toBeUndefined()', 31],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeUndefined) / invalid ${index}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream no-unnecessary-assertion (toBeUndefined) / invalid 4',
      code: `const x = 0;

expect(x).toBeUndefined()
expect(x).not.toBeUndefined()`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 26,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 30,
        },
      ],
    },
    ...(
      [
        [
          `const add = (a: number, b: number) => a + b;

expect(add(1, 1)).toBeUndefined();`,
          34,
        ],
        [
          `const add = (a: number, b: number): number => a + b;

expect(add(1, 1)).toBeUndefined();`,
          34,
        ],
        [
          `const add = (a: number, b: number): number => a + b;

expect(add(1, 1)).not.toBeUndefined();`,
          38,
        ],
        [
          `declare function mx(): never;

expect(mx()).not.toBeUndefined();`,
          33,
        ],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeUndefined) / invalid ${index + 5}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn,
        },
      ],
    })),
    ...(
      [
        `const result = "hello world".match("sunshine") ?? [];

expect(result).not.toBeUndefined();
expect(result).toBeUndefined();`,
        `const result = "hello world".match("sunshine") || [];

expect(result).not.toBeUndefined();
expect(result).toBeUndefined();`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeUndefined) / invalid ${index + 9}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 35,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 31,
        },
      ],
    })),
    {
      name: 'upstream no-unnecessary-assertion (toBeUndefined) / invalid 11',
      code: `declare function mx(): Promise<string>;

it('is async', async () => {
  expect(await mx()).toBeUndefined();
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 37,
        },
      ],
    },
    ...(
      [
        [
          `declare function mx(): string | number;

expect(mx()).toBeUndefined();
expect(mx()).not.toBeUndefined();`,
          3,
          29,
          4,
          33,
        ],
        [
          `declare function mx<T>(p: T): T;

expect(mx(undefined)).toBeUndefined();
expect(mx(undefined)).not.toBeUndefined();

expect(mx('hello')).toBeUndefined();
expect(mx('world')).not.toBeUndefined();`,
          6,
          36,
          7,
          40,
        ],
        [
          `declare function mx<T>(p: T): T extends string ? undefined : T;

expect(mx('hello')).toBeUndefined();
expect(mx('world')).not.toBeUndefined();

expect(mx({})).toBeUndefined();
expect(mx({})).not.toBeUndefined();`,
          6,
          31,
          7,
          35,
        ],
      ] as const
    ).map(([code, line1, endColumn1, line2, endColumn2], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeUndefined) / invalid ${index + 12}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: line1,
          column: 1,
          endLine: line1,
          endColumn: endColumn1,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: line2,
          column: 1,
          endLine: line2,
          endColumn: endColumn2,
        },
      ],
    })),
    {
      name: 'upstream no-unnecessary-assertion (toBeUndefined) / invalid 15',
      code: `declare function mx(): string | undefined;

expect(mx()!).toBeUndefined();
expect(mx()!).not.toBeUndefined();

expect(mx() as string).toBeUndefined();
expect(mx() as string).not.toBeUndefined();

expect(mx() as number).toBeUndefined();
expect(mx() as number).not.toBeUndefined();`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 30,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 34,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 39,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 7,
          column: 1,
          endLine: 7,
          endColumn: 43,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 9,
          column: 1,
          endLine: 9,
          endColumn: 39,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 10,
          column: 1,
          endLine: 10,
          endColumn: 43,
        },
      ],
    },
    ...(
      [
        ['expect("hello world").toBeNaN()', 32],
        ['expect({}).toBeNaN()', 21],
        ['expect([]).not.toBeNaN()', 25],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeNaN) / invalid ${index}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be a number',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream no-unnecessary-assertion (toBeNaN) / invalid 3',
      code: `const x = 'hello world';

expect(x).toBeNaN()
expect(x).not.toBeNaN()`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be a number',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 20,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be a number',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 24,
        },
      ],
    },
    ...(
      [
        [
          `const join = (a: string, b: string) => a + b;

expect(join('hello', 'world')).toBeNaN();`,
          41,
        ],
        [
          `const join = (a: string, b: string): string => a + b;

expect(join('hello', 'world')).toBeNaN();`,
          41,
        ],
        [
          `const join = (a: string, b: string): string => a + b;

expect(join('hello', 'world')).not.toBeNaN();`,
          45,
        ],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeNaN) / invalid ${index + 4}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be a number',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `const result = "hello world".match("sunshine") ?? [];

expect(result).not.toBeNaN();
expect(result).toBeNaN();`,
          3,
          29,
          4,
          25,
        ],
        [
          `const result = "hello world".match("sunshine") || [];

expect(result).not.toBeNaN();
expect(result).toBeNaN();`,
          3,
          29,
          4,
          25,
        ],
        [
          `declare function mx(): string | null;

expect(mx()).toBeNaN();
expect(mx()).not.toBeNaN();`,
          3,
          23,
          4,
          27,
        ],
        [
          `declare function mx<T>(p: T): T;

expect(mx(0)).toBeNaN();
expect(mx(1)).not.toBeNaN();
expect(mx(NaN)).not.toBeNaN();

expect(mx('hello')).toBeNaN();
expect(mx('world')).not.toBeNaN();`,
          7,
          30,
          8,
          34,
        ],
        [
          `declare function mx<T>(p: T): T extends string ? number : T;

expect(mx('hello')).toBeNaN();
expect(mx('world')).not.toBeNaN();

expect(mx({})).toBeNaN();
expect(mx({})).not.toBeNaN();`,
          6,
          25,
          7,
          29,
        ],
        [
          `declare function mx(): string | number;

expect(mx() as string).toBeNaN();
expect(mx() as string).not.toBeNaN();

expect(mx() as number).toBeNaN();
expect(mx() as number).not.toBeNaN();`,
          3,
          33,
          4,
          37,
        ],
      ] as const
    ).map(([code, line1, endColumn1, line2, endColumn2], index): TestCase => ({
      name: `upstream no-unnecessary-assertion (toBeNaN) / invalid ${index + 7}`,
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
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be a number',
          line: line1,
          column: 1,
          endLine: line1,
          endColumn: endColumn1,
        },
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be a number',
          line: line2,
          column: 1,
          endLine: line2,
          endColumn: endColumn2,
        },
      ],
    })),
  ],
});
