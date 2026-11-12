import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-equality-matcher', {
  valid: [
    { code: 'expect.hasAssertions' },
    { code: 'expect.hasAssertions()' },
    { code: 'expect.assertions(1)' },
    { code: 'expect(true).toBe(...true)' },
    { code: 'expect(a).to.be.a("string");' },
    { code: 'expect(a == 1).toBe(true)' },
    { code: 'expect(1 == a).toBe(true)' },
    { code: 'expect(a == b).toBe(true)' },
    { code: 'expect(a != 1).toBe(true)' },
    { code: 'expect(1 != a).toBe(true)' },
    { code: 'expect(a != b).toBe(true)' },
  ],
  invalid: [
    {
      code: 'expect(a === b).toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a === b,).toBe(true,);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a,).toBe(b,);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a,).toEqual(b,);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a,).toStrictEqual(b,);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a === b).toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).not.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).not.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).not.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a === b).resolves.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).resolves.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).resolves.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).resolves.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a === b).resolves.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).resolves.not.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).resolves.not.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).resolves.not.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a === b).not.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).not.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).not.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).not.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a === b).not.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a === b).resolves.not.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).resolves.not.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).resolves.not.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).resolves.not.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a === b).resolves.not.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).resolves.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).resolves.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).resolves.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a === b)["resolves"].not.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a)["resolves"].toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a)["resolves"].toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a)["resolves"].toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a === b)["resolves"]["not"]["toBe"](false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 42,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a)["resolves"]["toBe"](b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a)["resolves"]["toEqual"](b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a)["resolves"]["toStrictEqual"](b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a !== b).toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).not.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).not.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).not.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a !== b).toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a !== b).resolves.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).resolves.not.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).resolves.not.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).resolves.not.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a !== b).resolves.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).resolves.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).resolves.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).resolves.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a !== b).not.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a !== b).not.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).not.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).not.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).not.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a !== b).resolves.not.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).resolves.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).resolves.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).resolves.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(a !== b).resolves.not.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          message: 'Prefer using one of the equality matchers instead',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toBe`',
              output: 'expect(a).resolves.not.toBe(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toEqual`',
              output: 'expect(a).resolves.not.toEqual(b);',
            },
            {
              messageId: 'suggestEqualityMatcher',
              desc: 'Use `toStrictEqual`',
              output: 'expect(a).resolves.not.toStrictEqual(b);',
            },
          ],
        },
      ],
    },
  ],
});
