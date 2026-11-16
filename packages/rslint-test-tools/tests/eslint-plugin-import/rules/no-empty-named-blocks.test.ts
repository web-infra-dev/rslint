// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-empty-named-blocks.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-empty-named-blocks.md
import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { config } from '../config';

function suggestionCases(
  codes: string[],
  output = "import 'mod';",
  name?: string,
): TestCase[] {
  return codes.map((code) => ({
    code,
    ...(name ? { name } : {}),
    errors: [
      {
        messageId: '',
        message: 'Unexpected empty named import block',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: code.length + 1,
        suggestions: [
          { messageId: '', desc: 'Remove unused import', output: '' },
          { messageId: '', desc: 'Remove empty import block', output },
        ],
      },
    ],
  }));
}

// Skipped cases retain their upstream records; re-enabling them also requires complete expectations.
new RuleTester(config).run('no-empty-named-blocks', {
  valid: [
    "import 'mod';",
    "import Default from 'mod';",
    "import { Named } from 'mod';",
    "import Default, { Named } from 'mod';",
    "import * as Namespace from 'mod';",
    // TypeScript.
    "import type Default from 'mod';",
    "import type { Named } from 'mod';",
    "import type Default, { Named } from 'mod';",
    "import type * as Namespace from 'mod';",
    `
        module.exports = {
          rules: {
            'keyword-spacing': ['error', {overrides: {}}],
          }
        };
      `,
    `
        import { DESCRIPTORS, NODE } from '../helpers/constants';
        // ...
        import { timeLimitedPromise } from '../helpers/helpers';
        // ...
        import { DESCRIPTORS2 } from '../helpers/constants';
      `,
    {
      name: 'documentation examples',
      code: "import { mod } from 'mod'",
    },
    {
      name: 'documentation examples',
      code: "import Default, { mod } from 'mod'",
    },
    {
      name: 'documentation examples',
      code: "import type { mod } from 'mod'",
    },
    {
      name: 'Flow typeof imports are unsupported',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof Default from 'mod'; // babel old",
    },
    {
      name: 'Flow typeof imports are unsupported',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof { Named } from 'mod'; // babel old",
    },
    {
      name: 'Flow typeof imports are unsupported',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof Default, { Named } from 'mod'; // babel old",
    },
    {
      name: 'Flow typeof imports are unsupported',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof { mod } from 'mod'",
    },
  ],
  invalid: [
    {
      code: "import Default, {} from 'mod';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: "import Default from 'mod';",
    },
    ...suggestionCases([
      "import {} from 'mod';",
      "import{}from'mod';",
      "import {} from'mod';",
      "import {}from 'mod';",
    ]),
    ...suggestionCases([
      "import type {} from 'mod';",
      "import type {}from 'mod';",
      "import type{}from 'mod';",
      "import type {}from'mod';",
    ]),
    {
      code: "import type Default, {} from 'mod';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
      output: "import type Default from 'mod';",
    },
    ...suggestionCases(
      ["import {} from 'mod'", "import type {} from 'mod'"],
      "import 'mod'",
      'documentation examples',
    ),
    {
      name: 'documentation examples',
      code: "import Default, {} from 'mod'",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
      output: "import Default from 'mod'",
    },
    {
      name: 'documentation examples',
      code: "import type Default, {} from 'mod'",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
      output: "import type Default from 'mod'",
    },
    {
      name: 'Flow typeof imports are unsupported',
      skip: 'The native TypeScript parser does not support Flow typeof imports.',
      code: "import typeof {} from 'mod';",
      errors: [
        {
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              desc: 'Remove unused import',
              output: '',
            },
            {
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    } as unknown as TestCase,
    {
      name: 'Flow typeof imports are unsupported',
      skip: 'The native TypeScript parser does not support Flow typeof imports.',
      code: "import typeof {}from 'mod';",
      errors: [
        {
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              desc: 'Remove unused import',
              output: '',
            },
            {
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    } as unknown as TestCase,
    {
      name: 'Flow typeof imports are unsupported',
      skip: 'The native TypeScript parser does not support Flow typeof imports.',
      code: "import typeof {} from'mod';",
      errors: [
        {
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              desc: 'Remove unused import',
              output: '',
            },
            {
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    } as unknown as TestCase,
    {
      name: 'Flow typeof imports are unsupported',
      skip: 'The native TypeScript parser does not support Flow typeof imports.',
      code: "import typeof{}from'mod';",
      errors: [
        {
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              desc: 'Remove unused import',
              output: '',
            },
            {
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    } as unknown as TestCase,
    {
      name: 'Flow typeof imports are unsupported',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof Default, {} from 'mod';",
      errors: [
        {
          message: 'Unexpected empty named import block',
        },
      ],
      output: "import typeof Default from 'mod';",
    } as unknown as TestCase,
    {
      name: 'Flow typeof imports are unsupported',
      skip: 'The native TypeScript parser does not support Flow typeof imports.',
      code: "import typeof {} from 'mod'",
      errors: [
        {
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              desc: 'Remove unused import',
              output: '',
            },
            {
              desc: 'Remove empty import block',
              output: "import 'mod'",
            },
          ],
        },
      ],
    } as unknown as TestCase,
    {
      name: 'Flow typeof imports are unsupported',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof Default, {} from 'mod'",
      errors: [
        {
          message: 'Unexpected empty named import block',
        },
      ],
      output: "import typeof Default from 'mod'",
    } as unknown as TestCase,
  ],
});
