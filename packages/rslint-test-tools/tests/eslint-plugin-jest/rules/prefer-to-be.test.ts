// Upstream reference: eslint-plugin-jest@29.16.1; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-to-be', {
  valid: [
    ...(
      [
        'expect(null).toBeNull();',
        'expect(null).not.toBeNull();',
        'expect(null).toBe(1);',
        'expect(null).toBe(-1);',
        'expect(null).toBe(...1);',
        'expect(obj).toStrictEqual([ x, 1 ]);',
        'expect(obj).toStrictEqual({ x: 1 });',
        'expect(obj).not.toStrictEqual({ x: 1 });',
        'expect(value).toMatchSnapshot();',
        "expect(catchError()).toStrictEqual({ message: 'oh noes!' })",
        'expect("something");',
        'expect(token).toStrictEqual(/[abc]+/g);',
        "expect(token).toStrictEqual(new RegExp('[abc]+', 'g'));",
        'expect(value).toEqual(dedent`my string`);',
        'expect(null).toBeNull();',
        'expect(null).not.toBeNull();',
        'expect(null).toBe(1);',
        'expect(obj).toStrictEqual([ x, 1 ]);',
        'expect(obj).toStrictEqual({ x: 1 });',
        'expect(obj).not.toStrictEqual({ x: 1 });',
        'expect(value).toMatchSnapshot();',
        "expect(catchError()).toStrictEqual({ message: 'oh noes!' })",
        'expect("something");',
        'expect(null).not.toEqual();',
        'expect(null).toBe();',
        'expect(null).toMatchSnapshot();',
        'expect("a string").toMatchSnapshot(null);',
        'expect("a string").not.toMatchSnapshot();',
        'expect(null).toBe',
        'expect(undefined).toBeUndefined();',
        'expect(true).toBeDefined();',
        'expect({}).toEqual({});',
        'expect(something).toBe()',
        'expect(something).toBe(somethingElse)',
        'expect(something).toEqual(somethingElse)',
        'expect(something).not.toBe(somethingElse)',
        'expect(something).not.toEqual(somethingElse)',
        'expect(undefined).toBe',
        'expect("something");',
        'expect(NaN).toBeNaN();',
        'expect(true).not.toBeNaN();',
        'expect({}).toEqual({});',
        'expect(something).toBe()',
        'expect(something).toBe(somethingElse)',
        'expect(something).toEqual(somethingElse)',
        'expect(something).not.toBe(somethingElse)',
        'expect(something).not.toEqual(somethingElse)',
        'expect(undefined).toBe',
        'expect("something");',
        "(expect('Model must be bound to an array if the multiple property is true') as any).toHaveBeenTipped()",
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'expect(null).toBeNull();',
        'expect(null).not.toBeNull();',
        'expect(null).toBe(1);',
        'expect(null).toBe(-1);',
        'expect(null).toBe(...1);',
        'expect(obj).toStrictEqual([ x, 1 ]);',
        'expect(obj).toStrictEqual({ x: 1 });',
        'expect(obj).not.toStrictEqual({ x: 1 });',
        'expect(value).toMatchSnapshot();',
        "expect(catchError()).toStrictEqual({ message: 'oh noes!' })",
        'expect("something");',
        'expect(token).toStrictEqual(/[abc]+/g);',
        "expect(token).toStrictEqual(new RegExp('[abc]+', 'g'));",
        'expect(value).toEqual(dedent`my string`);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-to-be / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      [
        'expect(null).not.toEqual();',
        'expect(null).toBe();',
        'expect(null).toMatchSnapshot();',
        'expect("a string").toMatchSnapshot(null);',
        'expect("a string").not.toMatchSnapshot();',
        'expect(null).toBe',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-to-be: null / valid ${index + 9}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      [
        'expect(undefined).toBeUndefined();',
        'expect(true).toBeDefined();',
        'expect({}).toEqual({});',
        'expect(something).toBe()',
        'expect(something).toBe(somethingElse)',
        'expect(something).toEqual(somethingElse)',
        'expect(something).not.toBe(somethingElse)',
        'expect(something).not.toEqual(somethingElse)',
        'expect(undefined).toBe',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-to-be: undefined / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(['expect(NaN).toBeNaN();', 'expect(true).not.toBeNaN();'] as const).map(
      (code, index): ValidTestCase => ({
        name: `upstream prefer-to-be: NaN / valid ${index}`,
        code,
        languageOptions: {
          ecmaVersion: 2015,
        },
      }),
    ),
  ],
  invalid: [
    ...(
      [
        [
          'expect(value).toEqual("my string");',
          15,
          22,
          'expect(value).toBe("my string");',
        ],
        [
          'expect(value).toStrictEqual("my string");',
          15,
          28,
          'expect(value).toBe("my string");',
        ],
        ['expect(value).toStrictEqual(1);', 15, 28, 'expect(value).toBe(1);'],
        ['expect(value).toStrictEqual(1,);', 15, 28, 'expect(value).toBe(1,);'],
        ['expect(value).toStrictEqual(-1);', 15, 28, 'expect(value).toBe(-1);'],
        [
          'expect(value).toEqual(`my string`);',
          15,
          22,
          'expect(value).toBe(`my string`);',
        ],
        [
          'expect(value)["toEqual"](`my string`);',
          15,
          24,
          "expect(value)['toBe'](`my string`);",
        ],
        [
          'expect(value).toStrictEqual(`my ${string}`);',
          15,
          28,
          'expect(value).toBe(`my ${string}`);',
        ],
        [
          'expect(loadMessage()).resolves.toStrictEqual("hello world");',
          32,
          45,
          'expect(loadMessage()).resolves.toBe("hello world");',
        ],
        [
          'expect(loadMessage()).resolves["toStrictEqual"]("hello world");',
          32,
          47,
          'expect(loadMessage()).resolves[\'toBe\']("hello world");',
        ],
        [
          'expect(loadMessage())["resolves"].toStrictEqual("hello world");',
          35,
          48,
          'expect(loadMessage())["resolves"].toBe("hello world");',
        ],
        [
          'expect(loadMessage()).resolves.toStrictEqual(false);',
          32,
          45,
          'expect(loadMessage()).resolves.toBe(false);',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'useToBe',
          message: 'Use `toBe` when expecting primitive literals',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        ['expect(null).toBe(null);', 14, 18, 'expect(null).toBeNull();'],
        ['expect(null).toEqual(null);', 14, 21, 'expect(null).toBeNull();'],
        ['expect(null).toEqual(null,);', 14, 21, 'expect(null).toBeNull();'],
        [
          'expect(null).toStrictEqual(null);',
          14,
          27,
          'expect(null).toBeNull();',
        ],
        [
          'expect("a string").not.toBe(null);',
          24,
          28,
          'expect("a string").not.toBeNull();',
        ],
        [
          'expect("a string").not["toBe"](null);',
          24,
          30,
          'expect("a string").not[\'toBeNull\']();',
        ],
        [
          'expect("a string")["not"]["toBe"](null);',
          27,
          33,
          'expect("a string")["not"][\'toBeNull\']();',
        ],
        [
          'expect("a string").not.toEqual(null);',
          24,
          31,
          'expect("a string").not.toBeNull();',
        ],
        [
          'expect("a string").not.toStrictEqual(null);',
          24,
          37,
          'expect("a string").not.toBeNull();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 12}`,
      code,
      errors: [
        {
          messageId: 'useToBeNull',
          message: 'Use `toBeNull` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        ['expect(undefined).toBe(undefined);', 23],
        ['expect(undefined).toEqual(undefined);', 26],
        ['expect(undefined).toStrictEqual(undefined);', 32],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 21}`,
      code,
      errors: [
        {
          messageId: 'useToBeUndefined',
          message: 'Use `toBeUndefined` instead',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn,
        },
      ],
      output: 'expect(undefined).toBeUndefined();',
    })),
    ...(
      [
        [
          'expect("a string").not.toBe(undefined);',
          24,
          28,
          'expect("a string").toBeDefined();',
        ],
        [
          'expect("a string").rejects.not.toBe(undefined);',
          32,
          36,
          'expect("a string").rejects.toBeDefined();',
        ],
        [
          'expect("a string").rejects.not["toBe"](undefined);',
          32,
          38,
          'expect("a string").rejects[\'toBeDefined\']();',
        ],
        [
          'expect("a string").not.toEqual(undefined);',
          24,
          31,
          'expect("a string").toBeDefined();',
        ],
        [
          'expect("a string").not.toStrictEqual(undefined);',
          24,
          37,
          'expect("a string").toBeDefined();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 24}`,
      code,
      errors: [
        {
          messageId: 'useToBeDefined',
          message: 'Use `toBeDefined` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        ['expect(NaN).toBe(NaN);', 13, 17, 'expect(NaN).toBeNaN();'],
        ['expect(NaN).toEqual(NaN);', 13, 20, 'expect(NaN).toBeNaN();'],
        ['expect(NaN).toStrictEqual(NaN);', 13, 26, 'expect(NaN).toBeNaN();'],
        [
          'expect("a string").not.toBe(NaN);',
          24,
          28,
          'expect("a string").not.toBeNaN();',
        ],
        [
          'expect("a string").rejects.not.toBe(NaN);',
          32,
          36,
          'expect("a string").rejects.not.toBeNaN();',
        ],
        [
          'expect("a string")["rejects"].not.toBe(NaN);',
          35,
          39,
          'expect("a string")["rejects"].not.toBeNaN();',
        ],
        [
          'expect("a string").not.toEqual(NaN);',
          24,
          31,
          'expect("a string").not.toBeNaN();',
        ],
        [
          'expect("a string").not.toStrictEqual(NaN);',
          24,
          37,
          'expect("a string").not.toBeNaN();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 29}`,
      code,
      errors: [
        {
          messageId: 'useToBeNaN',
          message: 'Use `toBeNaN` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'expect(undefined).not.toBeDefined();',
          23,
          34,
          'expect(undefined).toBeUndefined();',
        ],
        [
          'expect(undefined).resolves.not.toBeDefined();',
          32,
          43,
          'expect(undefined).resolves.toBeUndefined();',
        ],
        [
          'expect(undefined).resolves.toBe(undefined);',
          28,
          32,
          'expect(undefined).resolves.toBeUndefined();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 37}`,
      code,
      errors: [
        {
          messageId: 'useToBeUndefined',
          message: 'Use `toBeUndefined` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'expect("a string").not.toBeUndefined();',
          24,
          'expect("a string").toBeDefined();',
        ],
        [
          'expect("a string").rejects.not.toBeUndefined();',
          32,
          'expect("a string").rejects.toBeDefined();',
        ],
      ] as const
    ).map(([code, column, output], index): TestCase => ({
      name: `run 0 invalid ${index + 40}`,
      code,
      errors: [
        {
          messageId: 'useToBeDefined',
          message: 'Use `toBeDefined` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 13,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'expect(null).toEqual(1 as unknown as string as unknown as any);',
          14,
          21,
          'expect(null).toBe(1 as unknown as string as unknown as any);',
        ],
        [
          'expect(null).toEqual(-1 as unknown as string as unknown as any);',
          14,
          21,
          'expect(null).toBe(-1 as unknown as string as unknown as any);',
        ],
        [
          'expect("a string").not.toStrictEqual("string" as number);',
          24,
          37,
          'expect("a string").not.toBe("string" as number);',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 42}`,
      code,
      errors: [
        {
          messageId: 'useToBe',
          message: 'Use `toBe` when expecting primitive literals',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'expect(null).toBe(null as unknown as string as unknown as any);',
          14,
          18,
          'expect(null).toBeNull();',
        ],
        [
          'expect("a string").not.toEqual(null as number);',
          24,
          31,
          'expect("a string").not.toBeNull();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 45}`,
      code,
      errors: [
        {
          messageId: 'useToBeNull',
          message: 'Use `toBeNull` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'expect(undefined).toBe(undefined as unknown as string as any);',
          19,
          23,
          'expect(undefined).toBeUndefined();',
        ],
        [
          'expect("a string").toEqual(undefined as number);',
          20,
          27,
          'expect("a string").toBeUndefined();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 47}`,
      code,
      errors: [
        {
          messageId: 'useToBeUndefined',
          message: 'Use `toBeUndefined` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'expect(value).toEqual("my string");',
          22,
          'expect(value).toBe("my string");',
        ],
        [
          'expect(value).toStrictEqual("my string");',
          28,
          'expect(value).toBe("my string");',
        ],
        ['expect(value).toStrictEqual(1);', 28, 'expect(value).toBe(1);'],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `upstream prefer-to-be / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useToBe',
          message: 'Use `toBe` when expecting primitive literals',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-to-be / invalid 3',
      code: 'expect(value).toStrictEqual(1,);',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useToBe',
          message: 'Use `toBe` when expecting primitive literals',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'expect(value).toBe(1,);',
    },
    ...(
      [
        ['expect(value).toStrictEqual(-1);', 15, 28, 'expect(value).toBe(-1);'],
        [
          'expect(value).toEqual(`my string`);',
          15,
          22,
          'expect(value).toBe(`my string`);',
        ],
        [
          'expect(value)["toEqual"](`my string`);',
          15,
          24,
          "expect(value)['toBe'](`my string`);",
        ],
        [
          'expect(value).toStrictEqual(`my ${string}`);',
          15,
          28,
          'expect(value).toBe(`my ${string}`);',
        ],
        [
          'expect(loadMessage()).resolves.toStrictEqual("hello world");',
          32,
          45,
          'expect(loadMessage()).resolves.toBe("hello world");',
        ],
        [
          'expect(loadMessage()).resolves["toStrictEqual"]("hello world");',
          32,
          47,
          'expect(loadMessage()).resolves[\'toBe\']("hello world");',
        ],
        [
          'expect(loadMessage())["resolves"].toStrictEqual("hello world");',
          35,
          48,
          'expect(loadMessage())["resolves"].toBe("hello world");',
        ],
        [
          'expect(loadMessage()).resolves.toStrictEqual(false);',
          32,
          45,
          'expect(loadMessage()).resolves.toBe(false);',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream prefer-to-be / invalid ${index + 4}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useToBe',
          message: 'Use `toBe` when expecting primitive literals',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        ['expect(null).toBe(null);', 18],
        ['expect(null).toEqual(null);', 21],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream prefer-to-be: null / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useToBeNull',
          message: 'Use `toBeNull` instead',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn,
        },
      ],
      output: 'expect(null).toBeNull();',
    })),
    {
      name: 'upstream prefer-to-be: null / invalid 2',
      code: 'expect(null).toEqual(null,);',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useToBeNull',
          message: 'Use `toBeNull` instead',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'expect(null).toBeNull();',
    },
    ...(
      [
        [
          'expect(null).toStrictEqual(null);',
          14,
          27,
          'expect(null).toBeNull();',
        ],
        [
          'expect("a string").not.toBe(null);',
          24,
          28,
          'expect("a string").not.toBeNull();',
        ],
        [
          'expect("a string").not["toBe"](null);',
          24,
          30,
          'expect("a string").not[\'toBeNull\']();',
        ],
        [
          'expect("a string")["not"]["toBe"](null);',
          27,
          33,
          'expect("a string")["not"][\'toBeNull\']();',
        ],
        [
          'expect("a string").not.toEqual(null);',
          24,
          31,
          'expect("a string").not.toBeNull();',
        ],
        [
          'expect("a string").not.toStrictEqual(null);',
          24,
          37,
          'expect("a string").not.toBeNull();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream prefer-to-be: null / invalid ${index + 3}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useToBeNull',
          message: 'Use `toBeNull` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        ['expect(undefined).toBe(undefined);', 23],
        ['expect(undefined).toEqual(undefined);', 26],
        ['expect(undefined).toStrictEqual(undefined);', 32],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream prefer-to-be: undefined / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useToBeUndefined',
          message: 'Use `toBeUndefined` instead',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn,
        },
      ],
      output: 'expect(undefined).toBeUndefined();',
    })),
    ...(
      [
        [
          'expect("a string").not.toBe(undefined);',
          24,
          28,
          'expect("a string").toBeDefined();',
        ],
        [
          'expect("a string").rejects.not.toBe(undefined);',
          32,
          36,
          'expect("a string").rejects.toBeDefined();',
        ],
        [
          'expect("a string").rejects.not["toBe"](undefined);',
          32,
          38,
          'expect("a string").rejects[\'toBeDefined\']();',
        ],
        [
          'expect("a string").not.toEqual(undefined);',
          24,
          31,
          'expect("a string").toBeDefined();',
        ],
        [
          'expect("a string").not.toStrictEqual(undefined);',
          24,
          37,
          'expect("a string").toBeDefined();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream prefer-to-be: undefined / invalid ${index + 3}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useToBeDefined',
          message: 'Use `toBeDefined` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        ['expect(NaN).toBe(NaN);', 13, 17, 'expect(NaN).toBeNaN();'],
        ['expect(NaN).toEqual(NaN);', 13, 20, 'expect(NaN).toBeNaN();'],
        ['expect(NaN).toStrictEqual(NaN);', 13, 26, 'expect(NaN).toBeNaN();'],
        [
          'expect("a string").not.toBe(NaN);',
          24,
          28,
          'expect("a string").not.toBeNaN();',
        ],
        [
          'expect("a string").rejects.not.toBe(NaN);',
          32,
          36,
          'expect("a string").rejects.not.toBeNaN();',
        ],
        [
          'expect("a string")["rejects"].not.toBe(NaN);',
          35,
          39,
          'expect("a string")["rejects"].not.toBeNaN();',
        ],
        [
          'expect("a string").not.toEqual(NaN);',
          24,
          31,
          'expect("a string").not.toBeNaN();',
        ],
        [
          'expect("a string").not.toStrictEqual(NaN);',
          24,
          37,
          'expect("a string").not.toBeNaN();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream prefer-to-be: NaN / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useToBeNaN',
          message: 'Use `toBeNaN` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'expect(undefined).not.toBeDefined();',
          23,
          34,
          'expect(undefined).toBeUndefined();',
        ],
        [
          'expect(undefined).resolves.not.toBeDefined();',
          32,
          43,
          'expect(undefined).resolves.toBeUndefined();',
        ],
        [
          'expect(undefined).resolves.toBe(undefined);',
          28,
          32,
          'expect(undefined).resolves.toBeUndefined();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream prefer-to-be: undefined vs defined / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useToBeUndefined',
          message: 'Use `toBeUndefined` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'expect("a string").not.toBeUndefined();',
          24,
          'expect("a string").toBeDefined();',
        ],
        [
          'expect("a string").rejects.not.toBeUndefined();',
          32,
          'expect("a string").rejects.toBeDefined();',
        ],
      ] as const
    ).map(([code, column, output], index): TestCase => ({
      name: `upstream prefer-to-be: undefined vs defined / invalid ${index + 3}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useToBeDefined',
          message: 'Use `toBeDefined` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 13,
        },
      ],
      output,
    })),
  ],
});
