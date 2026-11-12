// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-while-loop-condition.js
import { RuleTester } from '../rule-tester';

new RuleTester().run('prefer-while-loop-condition', {} as never, {
  valid: [
    // test.snapshot valid.
    {
      code: 'while (hasMore()) {\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'while (false) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'while (1) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (; false;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (let index = 0;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (;; update()) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} while (false);',
      filename: 'src/virtual.js',
    },
    {
      code: 'let done = false;\n\nwhile (true) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tlet done = true;\n\twork();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'let done = false;\n\nfor (;;) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tconst done = true;\n\twork();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'let done = false;\n\ndo {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tclass done {}\n\twork();\n} while (true);',
      filename: 'src/virtual.js',
    },
    {
      code: 'let done = false;\n\nwhile (true) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tfunction done() {}\n\twork();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'loop: for (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'loop: do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} while (true);',
      filename: 'src/virtual.js',
    },
    {
      code: 'while (true) {\n\tsetup();\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t} else {\n\t\tprocessNext();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function process() {\n\twhile (true) {\n\t\tif (!hasMore()) {\n\t\t\treturn;\n\t\t}\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t\tprocessNext();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'loop: while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'loop: {\n\twhile (true) {\n\t\tif (!hasMore()) {\n\t\t\tbreak loop;\n\t\t}\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tif (done) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\t{\n\t\tif (done) {\n\t\t\tbreak;\n\t\t}\n\t}\n}',
      filename: 'src/virtual.js',
    },
    // documentation example 2.
    {
      code: '// ✅\nwhile (hasMore()) {\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
    },
    // documentation example 4.
    {
      code: '// ✅\nwhile (hasMore()) {\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
    },
    // documentation example 6.
    {
      code: '// ✅\nwhile (!done) {\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
    },
    // documentation example 8.
    {
      code: '// ✅\nwhile (hasMore()) {\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    // test.snapshot invalid.
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tfor (const item of items) {\n\t\tif (done(item)) {\n\t\t\tbreak;\n\t\t}\n\t}\n}',
      filename: 'src/virtual.js',
      output:
        'while (hasMore()) {\n\n\tfor (const item of items) {\n\t\tif (done(item)) {\n\t\t\tbreak;\n\t\t}\n\t}\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tswitch (value) {\n\t\tcase 1:\n\t\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
      output:
        'while (hasMore()) {\n\n\tswitch (value) {\n\t\tcase 1:\n\t\t\tbreak;\n\t}\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tfunction stop() {\n\t\twhile (done) {\n\t\t\tbreak;\n\t\t}\n\t}\n}',
      filename: 'src/virtual.js',
      output:
        'while (hasMore()) {\n\n\tfunction stop() {\n\t\twhile (done) {\n\t\t\tbreak;\n\t\t}\n\t}\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: 'while (hasMore()) {\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) break;\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: 'while (hasMore()) {\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: 'while (!done) {\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (done || cancelled) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: 'while (!(done || cancelled)) {\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'for (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: 'while (hasMore()) {\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'for (; true;) {\n\tif (!hasMore()) break;\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: 'while (hasMore()) {\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'for (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tif (skip) {\n\t\tcontinue;\n\t}\n\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output:
        'while (hasMore()) {\n\n\tif (skip) {\n\t\tcontinue;\n\t}\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n} while (true);',
      filename: 'src/virtual.js',
      output: 'while (hasMore()) {\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tif (skip) {\n\t\tcontinue;\n\t}\n\n\tprocessNext();\n} while (true);',
      filename: 'src/virtual.js',
      output:
        'while (hasMore()) {\n\n\tif (skip) {\n\t\tcontinue;\n\t}\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'do {\n\tif (!hasMore()) break;\n\tprocessNext();\n} while (true);',
      filename: 'src/virtual.js',
      output: 'while (hasMore()) {\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'do {\n\tif (done || cancelled) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n} while (true);',
      filename: 'src/virtual.js',
      output: 'while (!(done || cancelled)) {\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tvar done = true;\n\twork();\n}',
      filename: 'src/virtual.js',
      output: 'while (!done) {\n\n\tvar done = true;\n\twork();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
      output: 'while (hasMore()) {\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'for (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
      output: 'while (hasMore()) {\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'for (; true;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
      output: 'while (hasMore()) {\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} while (true);',
      filename: 'src/virtual.js',
      output: 'while (hasMore()) {\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (!value!) {\n\t\tbreak;\n\t}\n\n\tprocess(value);\n}',
      filename: 'src/virtual.ts',
      output: 'while (value!) {\n\n\tprocess(value);\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (value!) {\n\t\tbreak;\n\t}\n\n\tprocess(value);\n}',
      filename: 'src/virtual.ts',
      output: 'while (!(value!)) {\n\n\tprocess(value);\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (/* keep */ true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'for (;;) {\n\tif (\n\t\t// Check before processing.\n\t\t!hasMore()\n\t) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 7,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'for /* keep */ (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'for (; /* keep */ true;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} /* keep */ while (true);',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'do /* keep */ {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} while (true);',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} while (true); // Preserve this comment.',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (\n\t\t// Check before processing.\n\t\t!hasMore()\n\t) {\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 7,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\t// Preserve this comment.\n\t\tbreak;\n\t}\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 5,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\t// Preserve this comment.\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 3,
          column: 2,
          endLine: 5,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) break; // Preserve this comment.\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t} // Preserve this comment.\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\t// Process the remaining work.\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output:
        'while (hasMore()) {\n\n\t// Process the remaining work.\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'do {\r\n  if (done) {\r\n    break;\r\n  }\r\n  processNext(/* Keep this argument comment. */ value);\r\n} while (true);',
      filename: 'src/virtual.js',
      output:
        'while (!done) {\r\n  processNext(/* Keep this argument comment. */ value);\r\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 2,
          column: 3,
          endLine: 4,
          endColumn: 4,
        },
      ],
    },
    // documentation example 1.
    {
      code: '// ❌\nwhile (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: '// ❌\nwhile (hasMore()) {\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 3,
          column: 2,
          endLine: 5,
          endColumn: 3,
        },
      ],
    },
    // documentation example 3.
    {
      code: '// ❌\ndo {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n} while (true);',
      filename: 'src/virtual.js',
      output: '// ❌\nwhile (hasMore()) {\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 3,
          column: 2,
          endLine: 5,
          endColumn: 3,
        },
      ],
    },
    // documentation example 5.
    {
      code: '// ❌\nwhile (true) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: '// ❌\nwhile (!done) {\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 3,
          column: 2,
          endLine: 5,
          endColumn: 3,
        },
      ],
    },
    // documentation example 7.
    {
      code: '// ❌\nfor (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}',
      filename: 'src/virtual.js',
      output: '// ❌\nwhile (hasMore()) {\n\n\tprocessNext();\n}',
      errors: [
        {
          messageId: 'prefer-while-loop-condition',
          message: 'Prefer putting the condition in the `while` statement.',
          line: 3,
          column: 2,
          endLine: 5,
          endColumn: 3,
        },
      ],
    },
  ],
});
