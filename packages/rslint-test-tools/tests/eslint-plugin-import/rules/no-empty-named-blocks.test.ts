import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { config } from '../config';

// Migrated from the existing suite. Each active case pins the full diagnostic and edits.
// Skipped cases retain their upstream records; re-enabling them also requires complete expectations.
new RuleTester(config).run('no-empty-named-blocks', {
  valid: [
    {
      code: "import 'mod';",
    },
    {
      code: "import Default from 'mod';",
    },
    {
      code: "import { Named } from 'mod';",
    },
    {
      code: "import Default, { Named } from 'mod';",
    },
    {
      code: "import * as Namespace from 'mod';",
    },
    {
      code: "import type Default from 'mod';",
    },
    {
      code: "import type { Named } from 'mod';",
    },
    {
      code: "import type Default, { Named } from 'mod';",
    },
    {
      code: "import type * as Namespace from 'mod';",
    },
    {
      code: "\n        module.exports = {\n          rules: {\n            'keyword-spacing': ['error', {overrides: {}}],\n          }\n        };\n      ",
    },
    {
      code: "\n        import { DESCRIPTORS, NODE } from '../helpers/constants';\n        // ...\n        import { timeLimitedPromise } from '../helpers/helpers';\n        // ...\n        import { DESCRIPTORS2 } from '../helpers/constants';\n      ",
    },
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
    {
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
      name: 'documentation examples',
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
      name: 'documentation examples',
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
