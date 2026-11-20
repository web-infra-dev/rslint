// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-to-have-length', {
  valid: [
    ...(
      [
        'expect.hasAssertions',
        'expect.hasAssertions()',
        'expect(files).toHaveLength(1);',
        "expect(files.name).toBe('file');",
        "expect(files[`name`]).toBe('file');",
        'expect(users[0]?.permissions?.length).toBe(1);',
        'expect(result).toBe(true);',
        "expect(user.getUserName(5)).resolves.toEqual('Paul')",
        "expect(user.getUserName(5)).rejects.toEqual('Paul')",
        'expect(a);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'expect.hasAssertions',
        'expect.hasAssertions()',
        'expect().toBe(true)',
        'expect(files).toHaveLength(1);',
        "expect(files.name).toBe('file');",
        "expect(files[`name`]).toBe('file');",
        'expect(users[0]?.permissions?.length).toBe(1);',
        'expect(result).toBe(true);',
        "expect(user.getUserName(5)).resolves.toEqual('Paul')",
        "expect(user.getUserName(5)).rejects.toEqual('Paul')",
        'expect(a);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-to-have-length / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          'expect(files["length"]).toBe(1);',
          25,
          29,
          'expect(files).toHaveLength(1);',
        ],
        [
          'expect(files["length"]).toBe(1,);',
          25,
          29,
          'expect(files).toHaveLength(1,);',
        ],
        [
          'expect(files["length"])["not"].toBe(1);',
          32,
          36,
          'expect(files)["not"].toHaveLength(1);',
        ],
        [
          'expect(files["length"])["toBe"](1);',
          25,
          31,
          'expect(files).toHaveLength(1);',
        ],
        [
          'expect(files["length"]).not["toBe"](1);',
          29,
          35,
          'expect(files).not.toHaveLength(1);',
        ],
        [
          'expect(files["length"])["not"]["toBe"](1);',
          32,
          38,
          'expect(files)["not"].toHaveLength(1);',
        ],
        [
          'expect(files.length).toBe(1);',
          22,
          26,
          'expect(files).toHaveLength(1);',
        ],
        [
          'expect(files.length).toEqual(1);',
          22,
          29,
          'expect(files).toHaveLength(1);',
        ],
        [
          'expect(files.length).toStrictEqual(1);',
          22,
          35,
          'expect(files).toHaveLength(1);',
        ],
        [
          'expect(files.length).not.toStrictEqual(1);',
          26,
          39,
          'expect(files).not.toHaveLength(1);',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'useToHaveLength',
          message: 'Use `toHaveLength()` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-to-have-length / invalid 0',
      code: 'expect(files["length"]).toBe(1);',
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useToHaveLength',
          message: 'Use `toHaveLength()` instead',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'expect(files).toHaveLength(1);',
    },
    {
      name: 'upstream prefer-to-have-length / invalid 1',
      code: 'expect(files["length"]).toBe(1,);',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useToHaveLength',
          message: 'Use `toHaveLength()` instead',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'expect(files).toHaveLength(1,);',
    },
    ...(
      [
        [
          'expect(files["length"])["not"].toBe(1);',
          32,
          36,
          'expect(files)["not"].toHaveLength(1);',
        ],
        [
          'expect(files["length"])["toBe"](1);',
          25,
          31,
          'expect(files).toHaveLength(1);',
        ],
        [
          'expect(files["length"]).not["toBe"](1);',
          29,
          35,
          'expect(files).not.toHaveLength(1);',
        ],
        [
          'expect(files["length"])["not"]["toBe"](1);',
          32,
          38,
          'expect(files)["not"].toHaveLength(1);',
        ],
        [
          'expect(files.length).toBe(1);',
          22,
          26,
          'expect(files).toHaveLength(1);',
        ],
        [
          'expect(files.length).toEqual(1);',
          22,
          29,
          'expect(files).toHaveLength(1);',
        ],
        [
          'expect(files.length).toStrictEqual(1);',
          22,
          35,
          'expect(files).toHaveLength(1);',
        ],
        [
          'expect(files.length).not.toStrictEqual(1);',
          26,
          39,
          'expect(files).not.toHaveLength(1);',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream prefer-to-have-length / invalid ${index + 2}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'useToHaveLength',
          message: 'Use `toHaveLength()` instead',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
  ],
});
