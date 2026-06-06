import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const expectSuggestions = (output: (equalityMatcher: string) => string) =>
  ['toBe', 'toEqual', 'toStrictEqual'].map(equalityMatcher => ({
    messageId: 'suggestEqualityMatcher',
    data: { equalityMatcher },
    output: output(equalityMatcher),
  }));

ruleTester.run('prefer-equality-matcher', {} as never, {
  valid: [
    // ===
    { code: 'expect.hasAssertions' },
    { code: 'expect.hasAssertions()' },
    { code: 'expect.assertions(1)' },
    { code: 'expect(a == 1).toBe(true)' },
    { code: 'expect(1 == a).toBe(true)' },
    { code: 'expect(a == b).toBe(true)' },
    // !==
    { code: 'expect(a != 1).toBe(true)' },
    { code: 'expect(1 != a).toBe(true)' },
    { code: 'expect(a != b).toBe(true)' },
  ],
  invalid: [
    // ===
    {
      code: 'expect(a === b).toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).${equalityMatcher}(b);`,
          ),
          column: 17,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a === b,).toBe(true,);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a,).${equalityMatcher}(b,);`,
          ),
          column: 18,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a === b).toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).not.${equalityMatcher}(b);`,
          ),
          column: 17,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a === b).resolves.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).resolves.${equalityMatcher}(b);`,
          ),
          column: 26,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a === b).resolves.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).resolves.not.${equalityMatcher}(b);`,
          ),
          column: 26,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a === b).not.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).not.${equalityMatcher}(b);`,
          ),
          column: 21,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a === b).not.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).${equalityMatcher}(b);`,
          ),
          column: 21,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a === b).resolves.not.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).resolves.not.${equalityMatcher}(b);`,
          ),
          column: 30,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a === b).resolves.not.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).resolves.${equalityMatcher}(b);`,
          ),
          column: 30,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a === b)["resolves"].not.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).resolves.${equalityMatcher}(b);`,
          ),
          column: 33,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a === b)["resolves"]["not"]["toBe"](false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).resolves.${equalityMatcher}(b);`,
          ),
          column: 36,
          line: 1,
        },
      ],
    },
    // !==
    {
      code: 'expect(a !== b).toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).not.${equalityMatcher}(b);`,
          ),
          column: 17,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a !== b).toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).${equalityMatcher}(b);`,
          ),
          column: 17,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a !== b).resolves.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).resolves.not.${equalityMatcher}(b);`,
          ),
          column: 26,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a !== b).resolves.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).resolves.${equalityMatcher}(b);`,
          ),
          column: 26,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a !== b).not.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).${equalityMatcher}(b);`,
          ),
          column: 21,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a !== b).not.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).not.${equalityMatcher}(b);`,
          ),
          column: 21,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a !== b).resolves.not.toBe(true);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).resolves.${equalityMatcher}(b);`,
          ),
          column: 30,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a !== b).resolves.not.toBe(false);',
      errors: [
        {
          messageId: 'useEqualityMatcher',
          suggestions: expectSuggestions(
            equalityMatcher => `expect(a).resolves.not.${equalityMatcher}(b);`,
          ),
          column: 30,
          line: 1,
        },
      ],
    },
  ],
});
