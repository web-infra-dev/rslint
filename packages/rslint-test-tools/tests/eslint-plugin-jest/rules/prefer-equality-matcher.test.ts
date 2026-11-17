// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function equalitySuggestions(
  output: (matcher: string, index: number) => string,
) {
  return ['toBe', 'toEqual', 'toStrictEqual'].map((matcher, index) => ({
    messageId: 'suggestEqualityMatcher',
    desc: `Use \`${matcher}\``,
    output: output(matcher, index),
  }));
}

new RuleTester(fixtureConfig()).run('prefer-equality-matcher', {
  valid: [
    ...(
      [
        'expect.hasAssertions',
        'expect.hasAssertions()',
        'expect.assertions(1)',
        'expect(true).toBe(...true)',
        'expect(a == 1).toBe(true)',
        'expect(1 == a).toBe(true)',
        'expect(a == b).toBe(true)',
        'expect(a != 1).toBe(true)',
        'expect(1 != a).toBe(true)',
        'expect(a != b).toBe(true)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'expect.hasAssertions',
        'expect.hasAssertions()',
        'expect.assertions(1)',
        'expect()',
        'expect().toStrictEqual({})',
        'expect(true).toBe(...true)',
        'expect(a == 1).toBe(true)',
        'expect(1 == a).toBe(true)',
        'expect(a == b).toBe(true)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-equality-matcher: === / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      [
        'expect(a != 1).toBe(true)',
        'expect(1 != a).toBe(true)',
        'expect(a != b).toBe(true)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-equality-matcher: !== / valid ${index + 6}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          'expect(a === b).toBe(true);',
          17,
          21,
          'expect(a).toBe(b);',
          'expect(a).toEqual(b);',
          'expect(a).toStrictEqual(b);',
        ],
        [
          'expect(a === b,).toBe(true,);',
          18,
          22,
          'expect(a,).toBe(b,);',
          'expect(a,).toEqual(b,);',
          'expect(a,).toStrictEqual(b,);',
        ],
        [
          'expect(a === b).toBe(false);',
          17,
          21,
          'expect(a).not.toBe(b);',
          'expect(a).not.toEqual(b);',
          'expect(a).not.toStrictEqual(b);',
        ],
        [
          'expect(a === b).resolves.toBe(true);',
          26,
          30,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
        [
          'expect(a === b).resolves.toBe(false);',
          26,
          30,
          'expect(a).resolves.not.toBe(b);',
          'expect(a).resolves.not.toEqual(b);',
          'expect(a).resolves.not.toStrictEqual(b);',
        ],
        [
          'expect(a === b).not.toBe(true);',
          21,
          25,
          'expect(a).not.toBe(b);',
          'expect(a).not.toEqual(b);',
          'expect(a).not.toStrictEqual(b);',
        ],
        [
          'expect(a === b).not.toBe(false);',
          21,
          25,
          'expect(a).toBe(b);',
          'expect(a).toEqual(b);',
          'expect(a).toStrictEqual(b);',
        ],
        [
          'expect(a === b).resolves.not.toBe(true);',
          30,
          34,
          'expect(a).resolves.not.toBe(b);',
          'expect(a).resolves.not.toEqual(b);',
          'expect(a).resolves.not.toStrictEqual(b);',
        ],
        [
          'expect(a === b).resolves.not.toBe(false);',
          30,
          34,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
        [
          'expect(a === b)["resolves"].not.toBe(false);',
          33,
          37,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
        [
          'expect(a === b)["resolves"]["not"]["toBe"](false);',
          36,
          42,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
        [
          'expect(a !== b).toBe(true);',
          17,
          21,
          'expect(a).not.toBe(b);',
          'expect(a).not.toEqual(b);',
          'expect(a).not.toStrictEqual(b);',
        ],
        [
          'expect(a !== b).toBe(false);',
          17,
          21,
          'expect(a).toBe(b);',
          'expect(a).toEqual(b);',
          'expect(a).toStrictEqual(b);',
        ],
        [
          'expect(a !== b).resolves.toBe(true);',
          26,
          30,
          'expect(a).resolves.not.toBe(b);',
          'expect(a).resolves.not.toEqual(b);',
          'expect(a).resolves.not.toStrictEqual(b);',
        ],
        [
          'expect(a !== b).resolves.toBe(false);',
          26,
          30,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
        [
          'expect(a !== b).not.toBe(true);',
          21,
          25,
          'expect(a).toBe(b);',
          'expect(a).toEqual(b);',
          'expect(a).toStrictEqual(b);',
        ],
        [
          'expect(a !== b).not.toBe(false);',
          21,
          25,
          'expect(a).not.toBe(b);',
          'expect(a).not.toEqual(b);',
          'expect(a).not.toStrictEqual(b);',
        ],
        [
          'expect(a !== b).resolves.not.toBe(true);',
          30,
          34,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
        [
          'expect(a !== b).resolves.not.toBe(false);',
          30,
          34,
          'expect(a).resolves.not.toBe(b);',
          'expect(a).resolves.not.toEqual(b);',
          'expect(a).resolves.not.toStrictEqual(b);',
        ],
      ] as const
    ).map(
      (
        [
          code,
          column,
          endColumn,
          suggestionOutput1,
          suggestionOutput2,
          suggestionOutput3,
        ],
        index,
      ): TestCase => ({
        name: `run 0 invalid ${index}`,
        code,
        errors: [
          {
            messageId: 'useEqualityMatcher',
            message: 'Prefer using one of the equality matchers instead',
            line: 1,
            column,
            endLine: 1,
            endColumn,
            suggestions: equalitySuggestions(
              (_, index) =>
                [suggestionOutput1, suggestionOutput2, suggestionOutput3][
                  index
                ],
            ),
          },
        ],
      }),
    ),
    {
      name: 'upstream prefer-equality-matcher: === / invalid 0',
      code: 'expect(a === b).toBe(true);',
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 21,
          suggestions: equalitySuggestions(
            (matcher) => `expect(a).${matcher}(b);`,
          ),
        },
      ],
    },
    {
      name: 'upstream prefer-equality-matcher: === / invalid 1',
      code: 'expect(a === b,).toBe(true,);',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 22,
          suggestions: equalitySuggestions(
            (matcher) => `expect(a,).${matcher}(b,);`,
          ),
        },
      ],
    },
    ...(
      [
        [
          'expect(a === b).toBe(false);',
          17,
          21,
          'expect(a).not.toBe(b);',
          'expect(a).not.toEqual(b);',
          'expect(a).not.toStrictEqual(b);',
        ],
        [
          'expect(a === b).resolves.toBe(true);',
          26,
          30,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
        [
          'expect(a === b).resolves.toBe(false);',
          26,
          30,
          'expect(a).resolves.not.toBe(b);',
          'expect(a).resolves.not.toEqual(b);',
          'expect(a).resolves.not.toStrictEqual(b);',
        ],
        [
          'expect(a === b).not.toBe(true);',
          21,
          25,
          'expect(a).not.toBe(b);',
          'expect(a).not.toEqual(b);',
          'expect(a).not.toStrictEqual(b);',
        ],
        [
          'expect(a === b).not.toBe(false);',
          21,
          25,
          'expect(a).toBe(b);',
          'expect(a).toEqual(b);',
          'expect(a).toStrictEqual(b);',
        ],
        [
          'expect(a === b).resolves.not.toBe(true);',
          30,
          34,
          'expect(a).resolves.not.toBe(b);',
          'expect(a).resolves.not.toEqual(b);',
          'expect(a).resolves.not.toStrictEqual(b);',
        ],
        [
          'expect(a === b).resolves.not.toBe(false);',
          30,
          34,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
        [
          'expect(a === b)["resolves"].not.toBe(false);',
          33,
          37,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
        [
          'expect(a === b)["resolves"]["not"]["toBe"](false);',
          36,
          42,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
      ] as const
    ).map(
      (
        [
          code,
          column,
          endColumn,
          suggestionOutput1,
          suggestionOutput2,
          suggestionOutput3,
        ],
        index,
      ): TestCase => ({
        name: `upstream prefer-equality-matcher: === / invalid ${index + 2}`,
        code,
        languageOptions: {
          ecmaVersion: 2015,
        },
        errors: [
          {
            messageId: 'useEqualityMatcher',
            message: 'Prefer using one of the equality matchers instead',
            line: 1,
            column,
            endLine: 1,
            endColumn,
            suggestions: equalitySuggestions(
              (_, index) =>
                [suggestionOutput1, suggestionOutput2, suggestionOutput3][
                  index
                ],
            ),
          },
        ],
      }),
    ),
    ...(
      [
        [
          'expect(a !== b).toBe(true);',
          17,
          'expect(a).not.toBe(b);',
          'expect(a).not.toEqual(b);',
          'expect(a).not.toStrictEqual(b);',
        ],
        [
          'expect(a !== b).toBe(false);',
          17,
          'expect(a).toBe(b);',
          'expect(a).toEqual(b);',
          'expect(a).toStrictEqual(b);',
        ],
        [
          'expect(a !== b).resolves.toBe(true);',
          26,
          'expect(a).resolves.not.toBe(b);',
          'expect(a).resolves.not.toEqual(b);',
          'expect(a).resolves.not.toStrictEqual(b);',
        ],
        [
          'expect(a !== b).resolves.toBe(false);',
          26,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
        [
          'expect(a !== b).not.toBe(true);',
          21,
          'expect(a).toBe(b);',
          'expect(a).toEqual(b);',
          'expect(a).toStrictEqual(b);',
        ],
        [
          'expect(a !== b).not.toBe(false);',
          21,
          'expect(a).not.toBe(b);',
          'expect(a).not.toEqual(b);',
          'expect(a).not.toStrictEqual(b);',
        ],
        [
          'expect(a !== b).resolves.not.toBe(true);',
          30,
          'expect(a).resolves.toBe(b);',
          'expect(a).resolves.toEqual(b);',
          'expect(a).resolves.toStrictEqual(b);',
        ],
        [
          'expect(a !== b).resolves.not.toBe(false);',
          30,
          'expect(a).resolves.not.toBe(b);',
          'expect(a).resolves.not.toEqual(b);',
          'expect(a).resolves.not.toStrictEqual(b);',
        ],
      ] as const
    ).map(
      (
        [code, column, suggestionOutput1, suggestionOutput2, suggestionOutput3],
        index,
      ): TestCase => ({
        name: `upstream prefer-equality-matcher: !== / invalid ${index}`,
        code,
        languageOptions: {
          ecmaVersion: 2015,
        },
        errors: [
          {
            messageId: 'useEqualityMatcher',
            message: 'Prefer using one of the equality matchers instead',
            line: 1,
            column,
            endLine: 1,
            endColumn: column + 4,
            suggestions: equalitySuggestions(
              (_, index) =>
                [suggestionOutput1, suggestionOutput2, suggestionOutput3][
                  index
                ],
            ),
          },
        ],
      }),
    ),
  ],
});
