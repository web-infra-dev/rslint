import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-focused-tests', {} as never, {
  valid: [
    { code: 'describe()' },
    { code: 'it()' },
    { code: 'test()' },
    { code: 'describe.skip()' },
    { code: 'it.skip()' },
    { code: 'test.skip()' },
    { code: 'test.todo()' },
    { code: 'test.concurrent()' },
    { code: 'test.fails()' },
    { code: 'test.each()()' },
    { code: 'test.for()()' },
    // Rstest has no fit/fdescribe aliases.
    { code: 'fit()' },
    { code: 'fdescribe()' },
  ],
  invalid: [
    {
      code: 'describe.only()',
      errors: [
        {
          line: 1,
          column: 10,
          endColumn: 14,
          messageId: 'focusedTest',
          suggestions: [
            { messageId: 'suggestRemoveFocus', output: 'describe()' },
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
          endColumn: 8,
          messageId: 'focusedTest',
          suggestions: [{ messageId: 'suggestRemoveFocus', output: 'it()' }],
        },
      ],
    },
    {
      code: 'test.only()',
      errors: [
        {
          line: 1,
          column: 6,
          endColumn: 10,
          messageId: 'focusedTest',
          suggestions: [{ messageId: 'suggestRemoveFocus', output: 'test()' }],
        },
      ],
    },
    {
      code: 'test.concurrent.only()',
      errors: [
        {
          line: 1,
          column: 17,
          endColumn: 21,
          messageId: 'focusedTest',
          suggestions: [
            { messageId: 'suggestRemoveFocus', output: 'test.concurrent()' },
          ],
        },
      ],
    },
    {
      code: 'test.only.for()()',
      errors: [
        {
          line: 1,
          column: 6,
          endColumn: 10,
          messageId: 'focusedTest',
          suggestions: [
            { messageId: 'suggestRemoveFocus', output: 'test.for()()' },
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
          endColumn: 16,
          messageId: 'focusedTest',
          suggestions: [
            { messageId: 'suggestRemoveFocus', output: 'describe()' },
          ],
        },
      ],
    },
    {
      code: 'describe.only.each()()',
      errors: [
        {
          line: 1,
          column: 10,
          endColumn: 14,
          messageId: 'focusedTest',
          suggestions: [
            { messageId: 'suggestRemoveFocus', output: 'describe.each()()' },
          ],
        },
      ],
    },
    {
      code: 'test.runIf(true).only()',
      errors: [
        {
          line: 1,
          column: 18,
          endColumn: 22,
          messageId: 'focusedTest',
          suggestions: [
            { messageId: 'suggestRemoveFocus', output: 'test.runIf(true)()' },
          ],
        },
      ],
    },
  ],
});
