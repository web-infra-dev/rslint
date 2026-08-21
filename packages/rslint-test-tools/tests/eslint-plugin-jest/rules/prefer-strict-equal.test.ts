import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-strict-equal', {} as never, {
  valid: [
    { code: 'expect(something).toStrictEqual(somethingElse);' },
    { code: "a().toEqual('b')" },
    { code: 'expect(a);' },
  ],
  invalid: [
    {
      code: 'expect(something).toEqual(somethingElse);',
      errors: [
        {
          messageId: 'useToStrictEqual',
          column: 19,
          line: 1,
          suggestions: [
            {
              messageId: 'suggestReplaceWithStrictEqual',
              output: 'expect(something).toStrictEqual(somethingElse);',
            },
          ],
        },
      ],
    },
    {
      code: 'expect(something).toEqual(somethingElse,);',
      errors: [
        {
          messageId: 'useToStrictEqual',
          column: 19,
          line: 1,
          suggestions: [
            {
              messageId: 'suggestReplaceWithStrictEqual',
              output: 'expect(something).toStrictEqual(somethingElse,);',
            },
          ],
        },
      ],
    },
    {
      // The suggestion keeps the double quotes; it used to normalise them to
      // single quotes, which disagreed with jest/no-alias-methods and could
      // introduce a `quotes` violation.
      code: 'expect(something)["toEqual"](somethingElse);',
      errors: [
        {
          messageId: 'useToStrictEqual',
          column: 19,
          line: 1,
          suggestions: [
            {
              messageId: 'suggestReplaceWithStrictEqual',
              output: 'expect(something)["toStrictEqual"](somethingElse);',
            },
          ],
        },
      ],
    },
    {
      // `toEqual` is a variable here, so the matcher is reported but gets no
      // suggestion; rewriting it would point the computed key at an identifier
      // that does not exist.
      code: "const toEqual = 'toEqual';\nexpect(something)[toEqual](somethingElse);",
      errors: [
        {
          messageId: 'useToStrictEqual',
          column: 19,
          line: 2,
        },
      ],
    },
    {
      code: 'expect(something).\n  toEqual(somethingElse);',
      errors: [
        {
          messageId: 'useToStrictEqual',
          column: 3,
          line: 2,
          suggestions: [
            {
              messageId: 'suggestReplaceWithStrictEqual',
              output: 'expect(something).\n  toStrictEqual(somethingElse);',
            },
          ],
        },
      ],
    },
  ],
});
