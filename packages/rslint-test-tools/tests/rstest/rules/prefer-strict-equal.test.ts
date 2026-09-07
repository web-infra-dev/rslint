import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-strict-equal', {} as never, {
  valid: [
    { code: 'expect(something).toStrictEqual(somethingElse);' },
    { code: 'expect(something).to.be.a("string");' },
    { code: "a().toEqual('b')" },
    { code: 'expect(a);' },
  ],
  invalid: [
    {
      code: 'expect(something).toEqual(somethingElse);',
      errors: [
        {
          messageId: 'useToStrictEqual',
          message: 'Use `toStrictEqual()` instead',
          column: 19,
          line: 1,
          endColumn: 26,
          endLine: 1,
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
          endColumn: 26,
          endLine: 1,
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
      code: 'expect(something)["toEqual"](somethingElse);',
      errors: [
        {
          messageId: 'useToStrictEqual',
          column: 19,
          line: 1,
          endColumn: 28,
          endLine: 1,
          suggestions: [
            {
              messageId: 'suggestReplaceWithStrictEqual',
              output: 'expect(something)["toStrictEqual"](somethingElse);',
            },
          ],
        },
      ],
    },
  ],
});
