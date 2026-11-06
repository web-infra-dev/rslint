// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-unary-minus.js
import { RuleTester } from '../rule-tester';

new RuleTester().run('prefer-unary-minus', {} as never, {
  valid: [
    // Snapshot cases.
    { code: 'x * 2;', filename: 'src/virtual.js' },
    { code: 'x / 2;', filename: 'src/virtual.js' },
    { code: 'x * -2;', filename: 'src/virtual.js' },
    { code: 'x / -2;', filename: 'src/virtual.js' },
    { code: '-1 * -1;', filename: 'src/virtual.js' },
    { code: '-1 / -1;', filename: 'src/virtual.js' },
    { code: '-1 / x;', filename: 'src/virtual.js' },
    { code: '-1 / (a + b);', filename: 'src/virtual.js' },
    { code: 'x * -1n;', filename: 'src/virtual.js' },
    { code: 'x / -1n;', filename: 'src/virtual.js' },
    { code: '-1n * x;', filename: 'src/virtual.js' },
    { code: '-1n / x;', filename: 'src/virtual.js' },
    { code: 'x ** -1;', filename: 'src/virtual.js' },
    { code: 'x - 1;', filename: 'src/virtual.js' },
    { code: 'x + -1;', filename: 'src/virtual.js' },
    { code: 'x % -1;', filename: 'src/virtual.js' },
    { code: 'x & -1;', filename: 'src/virtual.js' },
    // Documentation examples.
    { code: 'const inverted = -value;', filename: 'src/virtual.js' },
    { code: 'const flipped = -someNumber;', filename: 'src/virtual.js' },
    { code: 'const divided = -value;', filename: 'src/virtual.js' },
  ],
  invalid: [
    // Trailing operator keywords.
    {
      code: 'getObject(/* keep */).value * (-1)in object;',
      filename: 'src/virtual.js',
      output: '-getObject(/* keep */).value in object;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'getObject(/* keep */).value / (-1)instanceof Number;',
      filename: 'src/virtual.js',
      output: '-getObject(/* keep */).value instanceof Number;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over dividing by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '-1 * (getObject(/* keep */).value)in object;',
      filename: 'src/virtual.js',
      output: '-getObject(/* keep */).value in object;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'getObject(/* keep */).value * (-1)instanceof Number;',
      filename: 'src/virtual.js',
      output: '-getObject(/* keep */).value instanceof Number;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    // Preserved operand comments.
    {
      code: 'getValue(/* keep */) * -1;',
      filename: 'src/virtual.js',
      output: '-getValue(/* keep */);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '-1 * values[/* keep */ index];',
      filename: 'src/virtual.js',
      output: '-values[/* keep */ index];',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '(left + /* keep */ right) / -1;',
      filename: 'src/virtual.js',
      output: '-(left + /* keep */ right);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over dividing by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '(condition ? /* keep */ left : right) * -1;',
      filename: 'src/virtual.js',
      output: '-(condition ? /* keep */ left : right);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: 'getValue(\r\n  // keep\r\n) / -1;',
      filename: 'src/virtual.js',
      output: '-getValue(\r\n  // keep\r\n);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over dividing by `-1`.',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 7,
        },
      ],
    },
    // Removed comments and TypeScript.
    {
      code: '(/* keep */ value) * -1;',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '(value /* keep */) / -1;',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over dividing by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'value * -/* keep */1;',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '-1 /* keep */ * value;',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '(value as /* keep */ number) * -1;',
      filename: 'src/virtual.ts',
      output: '-(value as /* keep */ number);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    // Snapshot cases.
    {
      code: 'x * -1;',
      filename: 'src/virtual.js',
      output: '-x;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: '-1 * x;',
      filename: 'src/virtual.js',
      output: '-x;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'x / -1;',
      filename: 'src/virtual.js',
      output: '-x;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over dividing by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'a.b * -1;',
      filename: 'src/virtual.js',
      output: '-a.b;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'foo() * -1;',
      filename: 'src/virtual.js',
      output: '-foo();',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '5 * -1;',
      filename: 'src/virtual.js',
      output: '-5;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: '-1 * 5;',
      filename: 'src/virtual.js',
      output: '-5;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'foo?.bar * -1;',
      filename: 'src/virtual.js',
      output: '-foo?.bar;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'tag`str` * -1;',
      filename: 'src/virtual.js',
      output: '-tag`str`;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '--x * -1;',
      filename: 'src/virtual.js',
      output: '-(--x);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'x++ * -1;',
      filename: 'src/virtual.js',
      output: '-(x++);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: '-y * -1;',
      filename: 'src/virtual.js',
      output: '-(-y);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '+x * -1;',
      filename: 'src/virtual.js',
      output: '-+x;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '(a + b) * -1;',
      filename: 'src/virtual.js',
      output: '-(a + b);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '-1 * (a + b);',
      filename: 'src/virtual.js',
      output: '-(a + b);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '(a ? b : c) * -1;',
      filename: 'src/virtual.js',
      output: '-(a ? b : c);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '(a, b) * -1;',
      filename: 'src/virtual.js',
      output: '-(a, b);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: '(a = b) * -1;',
      filename: 'src/virtual.js',
      output: '-(a = b);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '(a + b) / -1;',
      filename: 'src/virtual.js',
      output: '-(a + b);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over dividing by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '(x) * -1;',
      filename: 'src/virtual.js',
      output: '-x;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: '(x * -1) ** 2;',
      filename: 'src/virtual.js',
      output: '(-x) ** 2;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'a-x*-1;',
      filename: 'src/virtual.js',
      output: 'a- -x;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'a - x * -1;',
      filename: 'src/virtual.js',
      output: 'a - -x;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'foo\nx * -1;',
      filename: 'src/virtual.js',
      output: 'foo\n;-x;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'function f() { return x * -1; }',
      filename: 'src/virtual.js',
      output: 'function f() { return -x; }',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'x * -1 * -1;',
      filename: 'src/virtual.js',
      output: '-(-x);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'a + b * -1;',
      filename: 'src/virtual.js',
      output: 'a + -b;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'x /* c */ * -1;',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'x * -1 /* c */;',
      filename: 'src/virtual.js',
      output: '-x /* c */;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: '(x as number) * -1;',
      filename: 'src/virtual.ts',
      output: '-(x as number);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '-1 * (x as number);',
      filename: 'src/virtual.ts',
      output: '-(x as number);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'x! * -1;',
      filename: 'src/virtual.ts',
      output: '-(x!);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    // Erased assertions before exponentiation.
    // tsgo rejects unparenthesized assertions before ** (TS1005): value * -1 as any ** 2;
    {
      code: '(value * -1 as any) ** 2;',
      filename: 'src/virtual.ts',
      output: '((-value) as any) ** 2;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    // tsgo rejects unparenthesized assertions before ** (TS1005): -1 * value satisfies number as any ** 2;
    // tsgo rejects unparenthesized assertions before ** (TS1005): value / -1 as any ** 2;
    {
      code: 'value * -1 as number;',
      filename: 'src/virtual.ts',
      output: '-value as number;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'value * -1 satisfies number;',
      filename: 'src/virtual.ts',
      output: '-value satisfies number;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '(value * -1) as number;',
      filename: 'src/virtual.ts',
      output: '(-value) as number;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    // tsgo rejects unparenthesized assertions before ** (TS1005): value /* keep */ * -1 as any ** 2;
    // Runtime evaluation cases.
    {
      code: 'getObject(/* keep */).value / (-1)in object;',
      filename: 'src/virtual.js',
      output: '-getObject(/* keep */).value in object;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over dividing by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '-1 * (getObject(/* keep */).value)instanceof Number;',
      filename: 'src/virtual.js',
      output: '-getObject(/* keep */).value instanceof Number;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '-1 * getValue(/* keep */);',
      filename: 'src/virtual.js',
      output: '-getValue(/* keep */);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'getValue(/* keep */) / -1;',
      filename: 'src/virtual.js',
      output: '-getValue(/* keep */);',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over dividing by `-1`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    // Documentation examples.
    {
      code: 'const inverted = value * -1;',
      filename: 'src/virtual.js',
      output: 'const inverted = -value;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'const flipped = -1 * someNumber;',
      filename: 'src/virtual.js',
      output: 'const flipped = -someNumber;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over multiplying by `-1`.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'const divided = value / -1;',
      filename: 'src/virtual.js',
      output: 'const divided = -value;',
      errors: [
        {
          messageId: 'prefer-unary-minus',
          message: 'Prefer the unary minus operator over dividing by `-1`.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
  ],
});
