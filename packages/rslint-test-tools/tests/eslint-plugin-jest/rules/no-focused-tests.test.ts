// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-focused-tests', {
  valid: [
    ...(
      [
        'describe()',
        'it()',
        'describe.skip()',
        'it.skip()',
        'test()',
        'test.skip()',
        'var appliedOnly = describe.only; appliedOnly.apply(describe)',
        'var calledOnly = it.only; calledOnly.call(it)',
        'it.each()()',
        'it.each`table`()',
        'test.each()()',
        'test.each`table`()',
        'test.concurrent()',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 1 valid 0',
      code: `
        import { describe as fdescribe } from '@jest/globals';

        fdescribe()
     \u0020`,
    },
    ...(
      [
        'describe()',
        'it()',
        'describe.skip()',
        'it.skip()',
        'test()',
        'test.skip()',
        'var appliedOnly = describe.only; appliedOnly.apply(describe)',
        'var calledOnly = it.only; calledOnly.call(it)',
        'it.each()()',
        'it.each`table`()',
        'test.each()()',
        'test.each`table`()',
        'test.concurrent()',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-focused-tests / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 6,
      },
    })),
    {
      name: 'upstream no-focused-tests (with imports) / valid 0',
      code: `import { describe as fdescribe } from '@jest/globals';

fdescribe()`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 6,
      },
    },
  ],
  invalid: [
    ...(
      [
        ['describe.only()', 10, 14, 'describe()'],
        ['describe.only.each()()', 10, 14, 'describe.each()()'],
        ['describe.only.each`table`()', 10, 14, 'describe.each`table`()'],
        ['describe["only"]()', 10, 16, 'describe()'],
        ['it.only()', 4, 8, 'it()'],
        ['it.concurrent.only.each``()', 15, 19, 'it.concurrent.each``()'],
        ['it.only.each()()', 4, 8, 'it.each()()'],
        ['it.only.each`table`()', 4, 8, 'it.each`table`()'],
        ['it["only"]()', 4, 10, 'it()'],
        ['test.only()', 6, 10, 'test()'],
        ['test.concurrent.only.each()()', 17, 21, 'test.concurrent.each()()'],
        ['test.only.each()()', 6, 10, 'test.each()()'],
        ['test.only.each`table`()', 6, 10, 'test.each`table`()'],
        ['test["only"]()', 6, 12, 'test()'],
        ['fdescribe()', 1, 10, 'describe()'],
        ['fit()', 1, 4, 'it()'],
        ['fit.each()()', 1, 4, 'it.each()()'],
        ['fit.each`table`()', 1, 4, 'it.each`table`()'],
      ] as const
    ).map(([code, column, endColumn, suggestionOutput], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column,
          endLine: 1,
          endColumn,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    ...(
      [
        [
          `
        const { describe } = require('@jest/globals');

        describe.only()
     \u0020`,
          18,
          22,
          `
        const { describe } = require('@jest/globals');

        describe()
     \u0020`,
        ],
        [
          `
        import { describe as describeThis } from '@jest/globals';

        describeThis.only()
     \u0020`,
          22,
          26,
          `
        import { describe as describeThis } from '@jest/globals';

        describeThis()
     \u0020`,
        ],
        [
          `
        const { fdescribe } = require('@jest/globals');

        fdescribe()
     \u0020`,
          9,
          18,
          `
        const { fdescribe } = require('@jest/globals');

        describe()
     \u0020`,
        ],
      ] as const
    ).map(([code, column, endColumn, suggestionOutput], index): TestCase => ({
      name: `run 1 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 4,
          column,
          endLine: 4,
          endColumn,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    {
      name: 'run 2 invalid 0',
      code: `
        import { describe as describeThis } from '@jest/globals';

        describeThis.only()
     \u0020`,
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 4,
          column: 22,
          endLine: 4,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: `
        import { describe as describeThis } from '@jest/globals';

        describeThis()
     \u0020`,
            },
          ],
        },
      ],
    },
    {
      name: 'run 2 invalid 1',
      code: `
        import { fdescribe as describeJustThis } from '@jest/globals';

        describeJustThis()
        describeJustThis.each()()
     \u0020`,
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 25,
        },
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 2 invalid 2',
      code: `
        import { describe as context } from '@jest/globals';

        context.only.each()()
     \u0020`,
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 4,
          column: 17,
          endLine: 4,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: `
        import { describe as context } from '@jest/globals';

        context.each()()
     \u0020`,
            },
          ],
        },
      ],
    },
    {
      name: 'upstream no-focused-tests / invalid 0',
      code: 'describe.only()',
      languageOptions: {
        ecmaVersion: 6,
      },
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
      name: 'upstream no-focused-tests / invalid 1',
      code: 'context.only()',
      settings: {
        jest: {
          globalAliases: {
            describe: ['context'],
          },
        },
      },
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 13,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: 'context()',
            },
          ],
        },
      ],
    },
    ...(
      [
        ['describe.only.each()()', 10, 14, 'describe.each()()'],
        ['describe.only.each`table`()', 10, 14, 'describe.each`table`()'],
        ['describe["only"]()', 10, 16, 'describe()'],
        ['it.only()', 4, 8, 'it()'],
        ['it.concurrent.only.each``()', 15, 19, 'it.concurrent.each``()'],
        ['it.only.each()()', 4, 8, 'it.each()()'],
        ['it.only.each`table`()', 4, 8, 'it.each`table`()'],
        ['it["only"]()', 4, 10, 'it()'],
        ['test.only()', 6, 10, 'test()'],
        ['test.concurrent.only.each()()', 17, 21, 'test.concurrent.each()()'],
        ['test.only.each()()', 6, 10, 'test.each()()'],
        ['test.only.each`table`()', 6, 10, 'test.each`table`()'],
        ['test["only"]()', 6, 12, 'test()'],
        ['fdescribe()', 1, 10, 'describe()'],
        ['fit()', 1, 4, 'it()'],
        ['fit.each()()', 1, 4, 'it.each()()'],
        ['fit.each`table`()', 1, 4, 'it.each`table`()'],
      ] as const
    ).map(([code, column, endColumn, suggestionOutput], index): TestCase => ({
      name: `upstream no-focused-tests / invalid ${index + 2}`,
      code,
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 1,
          column,
          endLine: 1,
          endColumn,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: suggestionOutput,
            },
          ],
        },
      ],
    })),
    {
      name: 'upstream no-focused-tests (with imports) / invalid 0',
      code: `const { describe } = require('@jest/globals');

describe.only()`,
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: `const { describe } = require('@jest/globals');

describe()`,
            },
          ],
        },
      ],
    },
    {
      name: 'upstream no-focused-tests (with imports) / invalid 1',
      code: `import { describe as describeThis } from '@jest/globals';

describeThis.only()`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: `import { describe as describeThis } from '@jest/globals';

describeThis()`,
            },
          ],
        },
      ],
    },
    {
      name: 'upstream no-focused-tests (with imports) / invalid 2',
      code: `const { fdescribe } = require('@jest/globals');

fdescribe()`,
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: `const { fdescribe } = require('@jest/globals');

describe()`,
            },
          ],
        },
      ],
    },
    {
      name: 'upstream no-focused-tests (aliases) / invalid 1',
      code: `import { fdescribe as describeJustThis } from '@jest/globals';

describeJustThis()
describeJustThis.each()()`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 17,
        },
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 17,
        },
      ],
    },
    {
      name: 'upstream no-focused-tests (aliases) / invalid 2',
      code: `import { describe as context } from '@jest/globals';

context.only.each()()`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'focusedTest',
          message: 'Unexpected focused test',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 13,
          suggestions: [
            {
              messageId: 'suggestRemoveFocus',
              desc: 'Suggest removing focus from test',
              output: `import { describe as context } from '@jest/globals';

context.each()()`,
            },
          ],
        },
      ],
    },
  ],
});
