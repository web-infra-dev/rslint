import { checkSkip } from '../../../../rslint-rule-tester/src/cases.ts';
import { test } from 'rstack/test';
// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-empty-named-blocks', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: "import 'mod';",
    },
    {
      name: 'run 0 valid 1',
      code: "import Default from 'mod';",
    },
    {
      name: 'run 0 valid 2',
      code: "import { Named } from 'mod';",
    },
    {
      name: 'run 0 valid 3',
      code: "import Default, { Named } from 'mod';",
    },
    {
      name: 'run 0 valid 4',
      code: "import * as Namespace from 'mod';",
    },
    {
      name: 'run 0 valid 5',
      code: "import type Default from 'mod';",
    },
    {
      name: 'run 0 valid 6',
      code: "import type { Named } from 'mod';",
    },
    {
      name: 'run 0 valid 7',
      code: "import type Default, { Named } from 'mod';",
    },
    {
      name: 'run 0 valid 8',
      code: "import type * as Namespace from 'mod';",
    },
    {
      name: 'run 0 valid 9',
      code: "\n        module.exports = {\n          rules: {\n            'keyword-spacing': ['error', {overrides: {}}],\n          }\n        };\n      ",
    },
    {
      name: 'run 0 valid 10',
      code: "\n        import { DESCRIPTORS, NODE } from '../helpers/constants';\n        // ...\n        import { timeLimitedPromise } from '../helpers/helpers';\n        // ...\n        import { DESCRIPTORS2 } from '../helpers/constants';\n      ",
    },
    {
      name: 'documentation examples / run 1 valid 0',
      code: "import { mod } from 'mod'",
    },
    {
      name: 'documentation examples / run 1 valid 1',
      code: "import Default, { mod } from 'mod'",
    },
    {
      name: 'documentation examples / run 1 valid 2',
      code: "import type { mod } from 'mod'",
    },
    {
      name: 'Flow typeof imports are unsupported / run 2 valid 0',
      code: "import typeof Default from 'mod'; // babel old",
      skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    },
    {
      name: 'Flow typeof imports are unsupported / run 2 valid 1',
      code: "import typeof { Named } from 'mod'; // babel old",
      skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    },
    {
      name: 'Flow typeof imports are unsupported / run 2 valid 2',
      code: "import typeof Default, { Named } from 'mod'; // babel old",
      skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    },
    {
      name: 'Flow typeof imports are unsupported / run 2 valid 3',
      code: "import typeof { mod } from 'mod'",
      skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
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
    {
      name: 'run 0 invalid 1',
      code: "import {} from 'mod';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove unused import',
              output: '',
            },
            {
              messageId: '',
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: "import{}from'mod';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove unused import',
              output: '',
            },
            {
              messageId: '',
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: "import {} from'mod';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove unused import',
              output: '',
            },
            {
              messageId: '',
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: "import {}from 'mod';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove unused import',
              output: '',
            },
            {
              messageId: '',
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: "import type {} from 'mod';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove unused import',
              output: '',
            },
            {
              messageId: '',
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: "import type {}from 'mod';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove unused import',
              output: '',
            },
            {
              messageId: '',
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: "import type{}from 'mod';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove unused import',
              output: '',
            },
            {
              messageId: '',
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: "import type {}from'mod';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove unused import',
              output: '',
            },
            {
              messageId: '',
              desc: 'Remove empty import block',
              output: "import 'mod';",
            },
          ],
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
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
    {
      name: 'documentation examples / run 1 invalid 0',
      code: "import {} from 'mod'",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove unused import',
              output: '',
            },
            {
              messageId: '',
              desc: 'Remove empty import block',
              output: "import 'mod'",
            },
          ],
        },
      ],
    },
    {
      name: 'documentation examples / run 1 invalid 1',
      code: "import type {} from 'mod'",
      errors: [
        {
          messageId: '',
          message: 'Unexpected empty named import block',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: '',
              desc: 'Remove unused import',
              output: '',
            },
            {
              messageId: '',
              desc: 'Remove empty import block',
              output: "import 'mod'",
            },
          ],
        },
      ],
    },
    {
      name: 'documentation examples / run 1 invalid 2',
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
      name: 'documentation examples / run 1 invalid 3',
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
  ],
});

// Preserved upstream inputs/assertions, not executable native expectations.
const skippedInvalidCases = [
  {
    name: 'Flow typeof imports are unsupported / run 2 invalid 0',
    code: "import typeof {} from 'mod';",
    skip: 'TypeScript-Go does not support Flow import typeof declarations.',
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
  },
  {
    name: 'Flow typeof imports are unsupported / run 2 invalid 1',
    code: "import typeof {}from 'mod';",
    skip: 'TypeScript-Go does not support Flow import typeof declarations.',
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
  },
  {
    name: 'Flow typeof imports are unsupported / run 2 invalid 2',
    code: "import typeof {} from'mod';",
    skip: 'TypeScript-Go does not support Flow import typeof declarations.',
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
  },
  {
    name: 'Flow typeof imports are unsupported / run 2 invalid 3',
    code: "import typeof{}from'mod';",
    skip: 'TypeScript-Go does not support Flow import typeof declarations.',
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
  },
  {
    name: 'Flow typeof imports are unsupported / run 2 invalid 4',
    code: "import typeof Default, {} from 'mod';",
    skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    errors: [
      {
        message: 'Unexpected empty named import block',
      },
    ],
    output: "import typeof Default from 'mod';",
  },
  {
    name: 'Flow typeof imports are unsupported / run 2 invalid 5',
    code: "import typeof {} from 'mod'",
    skip: 'TypeScript-Go does not support Flow import typeof declarations.',
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
  },
  {
    name: 'Flow typeof imports are unsupported / run 2 invalid 6',
    code: "import typeof Default, {} from 'mod'",
    skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    errors: [
      {
        message: 'Unexpected empty named import block',
      },
    ],
    output: "import typeof Default from 'mod'",
  },
];
for (const item of skippedInvalidCases) {
  checkSkip(item);
  test.skip(`invalid ${item.name}: ${item.skip}`, () => {});
}
