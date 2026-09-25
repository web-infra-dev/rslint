// Ported from eslint-plugin-unicorn v76.0.0 test/no-useless-continue.js and documentation; see LICENSE.
import { RuleTester } from '../rule-tester';

new RuleTester().run('no-useless-continue', {} as never, {
  valid: [
    {
      code: 'for (const x of xs) {\n\tif (skip(x)) {\n\t\tcontinue;\n\t}\n\n\tprocess(x);\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\tif (a) {\n\t\tcontinue;\n\t}\n\n\tdoMore();\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\tif (a) {\n\t\tcontinue;\n\t} else {\n\t\tdoMore();\n\t}\n\n\tuse(x);\n}\n',
      filename: 'src/virtual.js',
    },
    { code: 'while (cond) continue;', filename: 'src/virtual.js' },
    { code: 'for (;;) continue;', filename: 'src/virtual.js' },
    { code: 'for (const x of xs) continue;', filename: 'src/virtual.js' },
    {
      code: 'outer: for (const x of xs) {\n\tfor (const y of ys) {\n\t\tcontinue outer;\n\t}\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'loop: for (const x of xs) {\n\tdoX();\n\tcontinue loop;\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\tswitch (x) {\n\t\tcase 1:\n\t\t\tcontinue;\n\t}\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\ttry {\n\t\tcontinue;\n\t} finally {\n\t\tcleanup();\n\t}\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\ttry {\n\t\tdoX();\n\t\tcontinue;\n\t} catch {}\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\ttry {\n\t\tdoX();\n\t} catch {\n\t\tcontinue;\n\t}\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\ttry {\n\t\tdoX();\n\t} finally {\n\t\tcontinue;\n\t}\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\tfor (const y of ys) {\n\t\tif (skip(y)) {\n\t\t\tcontinue;\n\t\t}\n\n\t\tprocess(y);\n\t}\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\tcontinue;\n\tdoX();\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\tcontinue;\n\tfunction f() {}\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\tif (a) {\n\t\tcontinue;\n\t}\n\t;\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\t{\n\t\tcontinue;\n\t}\n\n\tdoMore();\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\tblock: {\n\t\tcontinue;\n\t}\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const x of xs) {\n\tif (a) {\n\t\tif (b) {\n\t\t\tcontinue;\n\t\t}\n\t}\n\n\tdoMore();\n}\n',
      filename: 'src/virtual.js',
    },
    // Examples from docs/rules/no-useless-continue.md.
    {
      code: 'for (const item of items) {\n\tprocess(item);\n}\n',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const item of items) {\n\tif (shouldSkip(item)) {\n\t\tcontinue;\n\t}\n\tprocess(item);\n}\n',
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    {
      code: 'for (const x of xs) {\n\tprocess(x);\n\tcontinue;\n}\n',
      filename: 'src/virtual.js',
      output: 'for (const x of xs) {\n\tprocess(x);\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tcontinue;\n}\n',
      filename: 'src/virtual.js',
      output: 'for (const x of xs) {\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'while (cond) {\n\tdoX();\n\tcontinue;\n}\n',
      filename: 'src/virtual.js',
      output: 'while (cond) {\n\tdoX();\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'do {\n\tdoX();\n\tcontinue;\n} while (cond);\n',
      filename: 'src/virtual.js',
      output: 'do {\n\tdoX();\n} while (cond);\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'for (const x in object) {\n\tdoX();\n\tcontinue;\n}\n',
      filename: 'src/virtual.js',
      output: 'for (const x in object) {\n\tdoX();\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'for (let i = 0; i < n; i++) {\n\tdoX();\n\tcontinue;\n}\n',
      filename: 'src/virtual.js',
      output: 'for (let i = 0; i < n; i++) {\n\tdoX();\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'for (const x of xs) if (a) { continue; }',
      filename: 'src/virtual.js',
      output: 'for (const x of xs) if (a) {  }',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tif (a) {\n\t\tcontinue;\n\t}\n}\n',
      filename: 'src/virtual.js',
      output: 'for (const x of xs) {\n\tif (a) {\n\t}\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tif (a) {\n\t\tif (b) {\n\t\t\tcontinue;\n\t\t}\n\t}\n}\n',
      filename: 'src/virtual.js',
      output:
        'for (const x of xs) {\n\tif (a) {\n\t\tif (b) {\n\t\t}\n\t}\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 4,
          column: 4,
          endLine: 4,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t\tcontinue;\n\t} else {\n\t\tdoY();\n\t}\n}\n',
      filename: 'src/virtual.js',
      output:
        'for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else {\n\t\tdoY();\n\t}\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else {\n\t\tdoY();\n\t\tcontinue;\n\t}\n}\n',
      filename: 'src/virtual.js',
      output:
        'for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else {\n\t\tdoY();\n\t}\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else if (b) {\n\t\tdoY();\n\t\tcontinue;\n\t}\n}\n',
      filename: 'src/virtual.js',
      output:
        'for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else if (b) {\n\t\tdoY();\n\t}\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'async function run() {\n\tfor await (const x of xs) {\n\t\tprocess(x);\n\t\tcontinue;\n\t}\n}\n',
      filename: 'src/virtual.js',
      output:
        'async function run() {\n\tfor await (const x of xs) {\n\t\tprocess(x);\n\t}\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\t{\n\t\tcontinue;\n\t}\n}\n',
      filename: 'src/virtual.js',
      output: 'for (const x of xs) {\n\t{\n\t}\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tfor (const y of ys) {\n\t\tprocess(y);\n\t\tcontinue;\n\t}\n}\n',
      filename: 'src/virtual.js',
      output:
        'for (const x of xs) {\n\tfor (const y of ys) {\n\t\tprocess(y);\n\t}\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tfor (const y of ys) {\n\t\tcontinue;\n\t}\n\n\tcontinue;\n}\n',
      filename: 'src/virtual.js',
      output: 'for (const x of xs) {\n\tfor (const y of ys) {\n\t}\n\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 12,
        },
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 6,
          column: 2,
          endLine: 6,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'outer: for (const x of xs) {\n\tdoX();\n\tcontinue;\n}\n',
      filename: 'src/virtual.js',
      output: 'outer: for (const x of xs) {\n\tdoX();\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tcontinue;\n\tcontinue;\n}\n',
      filename: 'src/virtual.js',
      output: 'for (const x of xs) {\n\tcontinue;\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else {\n\t\tif (b) {\n\t\t\tcontinue;\n\t\t}\n\t}\n}\n',
      filename: 'src/virtual.js',
      output:
        'for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else {\n\t\tif (b) {\n\t\t}\n\t}\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 6,
          column: 4,
          endLine: 6,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'do {\n\tif (a) {\n\t\tcontinue;\n\t}\n} while (cond);\n',
      filename: 'src/virtual.js',
      output: 'do {\n\tif (a) {\n\t}\n} while (cond);\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tdoX();\n\tcontinue; // trailing comment\n}\n',
      filename: 'src/virtual.js',
      output: 'for (const x of xs) {\n\tdoX();\n\t // trailing comment\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'for (const x of xs) {\n\tdoX();\n\t// leading comment\n\tcontinue;\n}\n',
      filename: 'src/virtual.js',
      output: 'for (const x of xs) {\n\tdoX();\n\t// leading comment\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 4,
          column: 2,
          endLine: 4,
          endColumn: 11,
        },
      ],
    },
    // Example from docs/rules/no-useless-continue.md.
    {
      code: 'for (const item of items) {\n\tprocess(item);\n\tcontinue;\n}\n',
      filename: 'src/virtual.js',
      output: 'for (const item of items) {\n\tprocess(item);\n}\n',
      errors: [
        {
          messageId: 'no-useless-continue',
          message: 'Unnecessary `continue` statement.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
  ],
});
