// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-strict-equal', {
  valid: [
    ...(
      [
        'expect(something).toStrictEqual(somethingElse);',
        "a().toEqual('b')",
        'expect(a);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
  ],
  invalid: [
    ...(
      [
        [
          'expect(something).toEqual(somethingElse);',
          26,
          'expect(something).toStrictEqual(somethingElse);',
        ],
        [
          'expect(something).toEqual(somethingElse,);',
          26,
          'expect(something).toStrictEqual(somethingElse,);',
        ],
        [
          'expect(something)["toEqual"](somethingElse);',
          28,
          'expect(something)["toStrictEqual"](somethingElse);',
        ],
      ] as const
    ).map(([code, endColumn, suggestionOutput], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'useToStrictEqual',
          message: 'Use `toStrictEqual()` instead',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn,
          suggestions: [
            {
              messageId: 'suggestReplaceWithStrictEqual',
              desc: 'Replace with `toStrictEqual()`',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    {
      name: 'run 0 invalid 3',
      code: `const toEqual = 'toEqual';
expect(something)[toEqual](somethingElse);`,
      errors: [
        {
          messageId: 'useToStrictEqual',
          message: 'Use `toStrictEqual()` instead',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: `expect(something).
  toEqual(somethingElse);`,
      errors: [
        {
          messageId: 'useToStrictEqual',
          message: 'Use `toStrictEqual()` instead',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestReplaceWithStrictEqual',
              desc: 'Replace with `toStrictEqual()`',
              output: `expect(something).
  toStrictEqual(somethingElse);`,
            },
          ],
        },
      ],
    },
    {
      name: 'upstream prefer-strict-equal / invalid 1',
      code: 'expect(something).toEqual(somethingElse,);',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useToStrictEqual',
          message: 'Use `toStrictEqual()` instead',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestReplaceWithStrictEqual',
              desc: 'Replace with `toStrictEqual()`',
              output: 'expect(something).toStrictEqual(somethingElse,);',
            },
          ],
        },
      ],
    },
  ],
});
