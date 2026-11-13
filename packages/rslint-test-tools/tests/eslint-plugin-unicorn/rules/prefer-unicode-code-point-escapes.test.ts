// Ported from eslint-plugin-unicorn v76.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/prefer-unicode-code-point-escapes.js
import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// eslint-plugin-unicorn v76.0.0 upstream and documentation cases.
new RuleTester(config).run('prefer-unicode-code-point-escapes', {
  valid: [
    "const foo = '\\u{7A}'",
    "const foo = '\\u{1F4A9}'",
    "const foo = '\\n\\t\\r\\\\\\'\\\"'",
    "const foo = '\\0'",
    {
      code: "const foo = '\\8\\9\\08'",
      skip: 'The native TypeScript parser rejects this legacy decimal escape before rule execution; Go source-only tests cover its diagnostics and fixes.',
      languageOptions: {
        sourceType: 'script',
      },
    },
    "const foo = '\\\\u2661'",
    "const foo = '\\\\x7A'",
    'const foo = tag`\\u2661`',
    'const foo = tag`\\123`',
    'const foo = String.raw`\\u2661`',
    'const foo = /\\u{61}/u',
    'const foo = /\\u{61}/v',
    'const foo = /[\\uD83D\\uDCA9]/u',
    'const foo = /[\\uD83D\\uDCA9]/v',
    'const foo = /[[\\uD83D\\uDCA9]\\uD83D\\uDCA9]/v',
    'const foo = /\\u{XYZ}/',
    'const foo = /\\u{110000}/',
    'const foo = /\\u{}/',
    'const foo = /\\u{/',
    'const foo = /\\cK/',
    'const foo = new RegExp("\\\\u0061")',
    `const foo = '\\u{7A}';
const bar = '\\u{2661}';
const baz = '\\u{1F4A9}';`,
    'const foo = `\\u{7A}${bar}\\u{2661}`;',
    'const foo = /\\u{61}/u;',
    'const foo = /\\u{61}/u;',
  ],
  invalid: [
    {
      code: "const foo = '\\x7A'",
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: "const foo = '\\u{7A}'",
    },
    {
      code: 'const foo = "\\x7A"',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const foo = "\\u{7A}"',
    },
    {
      code: "const foo = '\\xa9'",
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: "const foo = '\\u{A9}'",
    },
    {
      code: "const foo = '\\u2661'",
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: "const foo = '\\u{2661}'",
    },
    {
      code: "const foo = '\\uD83D\\uDCA9'",
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 27,
          suggestions: [],
        },
      ],
      output: "const foo = '\\u{1F4A9}'",
    },
    {
      code: "const foo = '\\123'",
      skip: 'The native TypeScript parser rejects this legacy octal escape before rule execution; Go source-only tests cover its diagnostics and fixes.',
      languageOptions: {
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: "const foo = '\\u{53}'",
    },
    {
      code: "const foo = '\\00'",
      skip: 'The native TypeScript parser rejects this legacy octal escape before rule execution; Go source-only tests cover its diagnostics and fixes.',
      languageOptions: {
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 18,
          suggestions: [],
        },
      ],
      output: "const foo = '\\u{0}'",
    },
    {
      code: "const foo = '\\1\\12\\123\\4\\45'",
      skip: 'The native TypeScript parser rejects this legacy octal escape before rule execution; Go source-only tests cover its diagnostics and fixes.',
      languageOptions: {
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 29,
          suggestions: [],
        },
      ],
      output: "const foo = '\\u{1}\\u{A}\\u{53}\\u{4}\\u{25}'",
    },
    {
      code: "const foo = '\\400'",
      skip: 'The native TypeScript parser rejects this legacy octal escape before rule execution; Go source-only tests cover its diagnostics and fixes.',
      languageOptions: {
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: "const foo = '\\u{20}0'",
    },
    {
      code: "const foo = '\\x7A\\u2661\\uD83D\\uDCA9'",
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 37,
          suggestions: [],
        },
      ],
      output: "const foo = '\\u{7A}\\u{2661}\\u{1F4A9}'",
    },
    {
      code: 'const foo = `\\x7A${bar}\\u2661`',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
      output: 'const foo = `\\u{7A}${bar}\\u{2661}`',
    },
    {
      code: 'const foo = `\\\\\\x7A`',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'const foo = `\\\\\\u{7A}`',
    },
    {
      code: 'const foo = /\\x7A/u',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'const foo = /\\u{7A}/u',
    },
    {
      code: 'const foo = /\\u0061/v',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: 'const foo = /\\u{61}/v',
    },
    {
      code: 'const foo = /\\uD83D\\uDCA9/u',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
      output: 'const foo = /\\u{1F4A9}/u',
    },
    {
      code: 'const foo = /\\[\\uD83D\\uDCA9/u',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
      output: 'const foo = /\\[\\u{1F4A9}/u',
    },
    {
      code: 'const foo = /[\\x2D]/u',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: 'const foo = /[\\u{2D}]/u',
    },
    {
      code: 'const foo = /[\\cA]/u',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'const foo = /[\\u{1}]/u',
    },
    {
      code: 'const foo = /\\cA/u',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const foo = /\\u{1}/u',
    },
    {
      code: 'const foo = /\\cA/',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'prefer-unicode-code-point-escapes/add-unicode-flag',
              desc: 'Use Unicode code point escapes and add the `u` flag.',
              output: 'const foo = /\\u{1}/u',
            },
          ],
        },
      ],
    },
    {
      code: 'const foo = /\\u0061/',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'prefer-unicode-code-point-escapes/add-unicode-flag',
              desc: 'Use Unicode code point escapes and add the `u` flag.',
              output: 'const foo = /\\u{61}/u',
            },
          ],
        },
      ],
    },
    {
      code: 'const foo = /\\u{61}/',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'prefer-unicode-code-point-escapes/add-unicode-flag',
              desc: 'Use Unicode code point escapes and add the `u` flag.',
              output: 'const foo = /\\u{61}/u',
            },
          ],
        },
      ],
    },
    {
      code: 'const foo = /\\x7A/g',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'prefer-unicode-code-point-escapes/add-unicode-flag',
              desc: 'Use Unicode code point escapes and add the `u` flag.',
              output: 'const foo = /\\u{7A}/gu',
            },
          ],
        },
      ],
    },
    {
      code: 'const foo = /\\x61\\_/',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const foo = /\\u{61}\\_/',
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
    },
    {
      code: `const foo = '\\x7A';
const bar = '\\u2661';
const baz = '\\uD83D\\uDCA9';
`,
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 21,
          suggestions: [],
        },
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 27,
          suggestions: [],
        },
      ],
      output: `const foo = '\\u{7A}';
const bar = '\\u{2661}';
const baz = '\\u{1F4A9}';
`,
    },
    {
      code: `const foo = \`\\x7A\${bar}\\u2661\`;
`,
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
      output: `const foo = \`\\u{7A}\${bar}\\u{2661}\`;
`,
    },
    {
      code: `const foo = /\\u0061/u;
`,
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: `const foo = /\\u{61}/u;
`,
    },
    {
      code: `const foo = /\\u0061/;
`,
      errors: [
        {
          messageId: 'prefer-unicode-code-point-escapes',
          message: 'Prefer Unicode code point escapes.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'prefer-unicode-code-point-escapes/add-unicode-flag',
              desc: 'Use Unicode code point escapes and add the `u` flag.',
              output: `const foo = /\\u{61}/u;
`,
            },
          ],
        },
      ],
    },
  ],
});
