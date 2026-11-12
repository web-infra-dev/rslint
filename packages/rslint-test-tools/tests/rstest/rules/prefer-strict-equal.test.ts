import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-strict-equal', {
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
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestReplaceWithStrictEqual',
              desc: 'Replace with `toStrictEqual()`',
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
    {
      code: 'expect(something)["toEqual"](somethingElse);',
      errors: [
        {
          messageId: 'useToStrictEqual',
          message: 'Use `toStrictEqual()` instead',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestReplaceWithStrictEqual',
              desc: 'Replace with `toStrictEqual()`',
              output: 'expect(something)["toStrictEqual"](somethingElse);',
            },
          ],
        },
      ],
    },
  ],
});
