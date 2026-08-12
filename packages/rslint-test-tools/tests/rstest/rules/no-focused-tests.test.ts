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
    { code: "import { test } from 'node:test'; test.only()" },
    { code: "const test = createRunner(); test.only('case', fn)" },
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
      code: 'test?.only()',
      errors: [
        {
          line: 1,
          column: 7,
          endColumn: 11,
          messageId: 'focusedTest',
          suggestions: [
            { messageId: 'suggestRemoveFocus', output: 'test?.()' },
          ],
        },
      ],
    },
    {
      code: 'test?.["only"]()',
      errors: [
        {
          line: 1,
          column: 8,
          endColumn: 14,
          messageId: 'focusedTest',
          suggestions: [
            { messageId: 'suggestRemoveFocus', output: 'test?.()' },
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
    {
      code: `const focused = test.only;
focused('case', fn)`,
      errors: [
        {
          line: 2,
          column: 1,
          endColumn: 8,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: `const focused = test;
focused('case', fn)`,
            },
          ],
        },
      ],
    },
    {
      code: `const focused = test.only;
focused('first', fn);
focused('second', fn)`,
      errors: [
        {
          line: 2,
          column: 1,
          endColumn: 8,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: `const focused = test;
focused('first', fn);
focused('second', fn)`,
            },
          ],
        },
        {
          line: 3,
          column: 1,
          endColumn: 8,
          messageId: 'focusedTest',
        },
      ],
    },
    {
      code: `import.meta.rstest.test.only('case', fn)`,
      errors: [
        {
          line: 1,
          column: 25,
          endColumn: 29,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: `import.meta.rstest.test('case', fn)`,
            },
          ],
        },
      ],
    },
    {
      code: `import { test } from '@rstest/playwright';
test.only('case', fn)`,
      errors: [
        {
          line: 2,
          column: 6,
          endColumn: 10,
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: `import { test } from '@rstest/playwright';
test('case', fn)`,
            },
          ],
        },
      ],
    },
    {
      code: `test[/* keep */ "only"]()`,
      errors: [
        {
          messageId: 'focusedTest',
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              output: `test/* keep */ ()`,
            },
          ],
        },
      ],
    },
  ],
});
