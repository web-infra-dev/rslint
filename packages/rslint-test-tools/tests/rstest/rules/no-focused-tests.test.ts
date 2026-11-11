import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-focused-tests', {
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
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
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
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
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
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: 'test()',
            },
          ],
        },
      ],
    },
    {
      code: 'test.concurrent.only()',
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: 'test.concurrent()',
            },
          ],
        },
      ],
    },
    {
      code: 'test.only.for()()',
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: 'test.for()()',
            },
          ],
        },
      ],
    },
    {
      code: 'describe["only"]()',
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: 'describe()',
            },
          ],
        },
      ],
    },
    {
      code: 'test?.only()',
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 11,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: 'test?.()',
            },
          ],
        },
      ],
    },
    {
      code: 'test?.["only"]()',
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: 'test?.()',
            },
          ],
        },
      ],
    },
    {
      code: 'describe.only.each()()',
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: 'describe.each()()',
            },
          ],
        },
      ],
    },
    {
      code: 'test.runIf(true).only()',
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: 'test.runIf(true)()',
            },
          ],
        },
      ],
    },
    {
      code: `const focused = test.only;
focused('case', fn)`,
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 8,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: "const focused = test;\nfocused('case', fn)",
            },
          ],
        },
      ],
    },
    {
      code: `const focused = test.only;
focused.skip('case', fn)`,
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 8,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: "const focused = test;\nfocused.skip('case', fn)",
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
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 8,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output:
                "const focused = test;\nfocused('first', fn);\nfocused('second', fn)",
            },
          ],
        },
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 8,
        },
      ],
    },
    {
      code: `import.meta.rstest.test.only('case', fn)`,
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: "import.meta.rstest.test('case', fn)",
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
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 2,
          column: 6,
          endLine: 2,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output:
                "import { test } from '@rstest/playwright';\ntest('case', fn)",
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
          message: 'Unexpected focused test',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: 'test/* keep */ ()',
            },
          ],
        },
      ],
    },
  ],
});
