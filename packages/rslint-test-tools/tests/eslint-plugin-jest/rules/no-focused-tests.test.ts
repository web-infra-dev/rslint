import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-focused-tests', {} as never, {
  valid: [
    { code: 'describe()' },
    { code: 'it()' },
    { code: 'describe.skip()' },
    { code: 'it.skip()' },
    { code: 'test()' },
    { code: 'test.skip()' },
    { code: 'var appliedOnly = describe.only; appliedOnly.apply(describe)' },
    { code: 'var calledOnly = it.only; calledOnly.call(it)' },
    { code: 'it.each()()' },
    { code: 'it.each`table`()' },
    { code: 'test.each()()' },
    { code: 'test.each`table`()' },
    { code: 'test.concurrent()' },
  ],
  invalid: [
    {
      code: 'describe.only()',
      errors: [
        {
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 14,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'describe()',
            },
          ],
        },
      ],
    },

    // TODO: Fix this test case
    // {
    //   code: 'context.only()',
    //   errors: [
    //     {
    //       line: 1,
    //       column: 9,
    //       endLine: 1,
    //       endColumn: 13,
    //       messageId: 'focusedTest',
    //       suggestions: [
    //         {
    //           messageId: 'suggestRemoveFocus',
    //           output: 'context()',
    //         },
    //       ],
    //     },
    //   ],
    //   settings: { jest: { globalAliases: { describe: ['context'] } } },
    // },

    {
      code: 'describe.only.each()()',
      errors: [
        {
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 14,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'describe.each()()',
            },
          ],
        },
      ],
    },
    {
      code: 'describe.only.each`table`()',
      errors: [
        {
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 14,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'describe.each`table`()',
            },
          ],
        },
      ],
    },
    {
      code: 'describe["only"]()',
      errors: [
        {
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'describe()',
            },
          ],
        },
      ],
    },
    {
      code: 'it.only()',
      errors: [
        {
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'it()',
            },
          ],
        },
      ],
    },
    {
      code: 'it.concurrent.only.each``()',
      errors: [
        {
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'it.concurrent.each``()',
            },
          ],
        },
      ],
    },
    {
      code: 'it.only.each()()',
      errors: [
        {
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'it.each()()',
            },
          ],
        },
      ],
    },
    {
      code: 'it.only.each`table`()',
      errors: [
        {
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'it.each`table`()',
            },
          ],
        },
      ],
    },
    {
      code: 'it["only"]()',
      errors: [
        {
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 10,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'it()',
            },
          ],
        },
      ],
    },
    {
      code: 'test.only()',
      errors: [
        {
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'test()',
            },
          ],
        },
      ],
    },
    {
      code: 'test.concurrent.only.each()()',
      errors: [
        {
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 21,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'test.concurrent.each()()',
            },
          ],
        },
      ],
    },
    {
      code: 'test.only.each()()',
      errors: [
        {
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'test.each()()',
            },
          ],
        },
      ],
    },
    {
      code: 'test.only.each`table`()',
      errors: [
        {
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'test.each`table`()',
            },
          ],
        },
      ],
    },
    {
      code: 'test["only"]()',
      errors: [
        {
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 12,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'test()',
            },
          ],
        },
      ],
    },
    {
      code: 'fdescribe()',
      errors: [
        {
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'describe()',
            },
          ],
        },
      ],
    },
    {
      code: 'fit()',
      errors: [
        {
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'it()',
            },
          ],
        },
      ],
    },
    {
      code: 'fit.each()()',
      errors: [
        {
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'it.each()()',
            },
          ],
        },
      ],
    },
    {
      code: 'fit.each`table`()',
      errors: [
        {
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: 'it.each`table`()',
            },
          ],
        },
      ],
    },
  ],
});

// import test cases
ruleTester.run('no-focused-tests', {} as never, {
  valid: [
    {
      code: `
        import { describe as fdescribe } from '@jest/globals';

        fdescribe()
      `,
    },
  ],

  invalid: [
    {
      code: `
        const { describe } = require('@jest/globals');

        describe.only()
      `,
      errors: [
        {
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 14,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: `
                const { describe } = require('@jest/globals');

                describe()
              `,
            },
          ],
        },
      ],
    },
    {
      code: `
        import { describe as describeThis } from '@jest/globals';

        describeThis.only()
      `,
      errors: [
        {
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 18,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: `
                import { describe as describeThis } from '@jest/globals';

                describeThis()
              `,
            },
          ],
        },
      ],
    },
    {
      code: `
        const { fdescribe } = require('@jest/globals');

        fdescribe()
      `,
      errors: [
        {
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: `
                const { fdescribe } = require('@jest/globals');

                describe()
              `,
            },
          ],
        },
      ],
    },
  ],
});

// alias test cases
ruleTester.run('no-focused-tests', {} as never, {
  valid: [],

  invalid: [
    {
      code: `
        import { describe as describeThis } from '@jest/globals';

        describeThis.only()
      `,
      errors: [
        {
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 18,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: `
                import { describe as describeThis } from '@jest/globals';

                describeThis()
              `,
            },
          ],
        },
      ],
    },
    {
      code: `
        import { fdescribe as describeJustThis } from '@jest/globals';

        describeJustThis()
        describeJustThis.each()()
      `,
      errors: [
        {
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 17,
          messageId: 'focusedTest',
        },
        {
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 17,
          messageId: 'focusedTest',
        },
      ],
    },
    {
      code: `
        import { describe as context } from '@jest/globals';

        context.only.each()()
      `,
      errors: [
        {
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 13,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: `
                import { describe as context } from '@jest/globals';

                context.each()()
              `,
            },
          ],
        },
      ],
    },
  ],
});
