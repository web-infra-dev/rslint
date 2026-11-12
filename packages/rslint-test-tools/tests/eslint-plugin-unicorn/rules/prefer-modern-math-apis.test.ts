import { RuleTester } from '../rule-tester';

// Upstream: https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-modern-math-apis.js
// Go upstream tests assert complete ranges and edits; this wrapper checks messages.
const ruleTester = new RuleTester();

// upstream line 7
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt(Math.E * Math.E);',
      filename: 'input.js',
      output: 'Math.abs(Math.E);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math.LN10 * Math.LN10);',
      filename: 'input.js',
      output: 'Math.abs(Math.LN10);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math.LN2 * Math.LN2);',
      filename: 'input.js',
      output: 'Math.abs(Math.LN2);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math.LOG10E * Math.LOG10E);',
      filename: 'input.js',
      output: 'Math.abs(Math.LOG10E);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math.LOG2E * Math.LOG2E);',
      filename: 'input.js',
      output: 'Math.abs(Math.LOG2E);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math.PI * Math.PI);',
      filename: 'input.js',
      output: 'Math.abs(Math.PI);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math.SQRT1_2 * Math.SQRT1_2);',
      filename: 'input.js',
      output: 'Math.abs(Math.SQRT1_2);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math.SQRT2 * Math.SQRT2);',
      filename: 'input.js',
      output: 'Math.abs(Math.SQRT2);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(-Math.PI) / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2(-Math.PI);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.LOG10E * Math.log(-Math.E);',
      filename: 'input.js',
      output: 'Math.log10(-Math.E);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = Math.PI; Math.sqrt(value * value);',
      filename: 'input.js',
      output: 'const value = Math.PI; Math.abs(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 33,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = -Math.SQRT2; Math.log(value) / Math.LN10;',
      filename: 'input.js',
      output: 'const value = -Math.SQRT2; Math.log10(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 55,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math.PI * Math.PI + Infinity ** 2);',
      filename: 'input.js',
      output: 'Math.hypot(Math.PI, Infinity);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(NaN ** 2 + Math.SQRT2 ** 2);',
      filename: 'input.js',
      output: 'Math.hypot(NaN, Math.SQRT2);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void Math.sqrt(Math.E ** 2 + Math.PI ** 2);',
      filename: 'input.js',
      output: 'void Math.hypot(Math.E, Math.PI);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt((Math./* keep */PI) * Math.PI);',
      filename: 'input.js',
      output: 'Math.abs((Math./* keep */PI));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'delete (Math.log(getValue()) / Math.LN2);',
      filename: 'input.js',
      output: 'delete (Math.log2(getValue()));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 40,
          suggestions: [],
        },
      ],
    },
    {
      code: 'delete (Math.LOG10E * Math.log(value));',
      filename: 'input.js',
      output: 'delete (Math.log10(value));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 38,
          suggestions: [],
        },
      ],
    },
    {
      code: 'delete Math.sqrt(1e200 * 1e200);',
      filename: 'input.js',
      output: 'delete Math.abs(1e200);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 17,
          suggestions: [],
        },
      ],
    },
    {
      code: 'delete Math.sqrt(3 ** 2 + 4 ** 2);',
      filename: 'input.js',
      output: 'delete Math.hypot(3, 4);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 17,
          suggestions: [],
        },
      ],
    },
    {
      code: 'delete (condition && Math.log(value) / Math.LN2);',
      filename: 'input.js',
      output: 'delete (condition && Math.log2(value));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 48,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 34
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt((Math.PI as number) * Math.PI);',
      filename: 'input.ts',
      output: 'Math.abs((Math.PI as number));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = Math.LN2 as number; Math.sqrt(value * value);',
      filename: 'input.ts',
      output: 'const value = Math.LN2 as number; Math.abs(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(-(Math.SQRT2 as number)) / Math.LN2;',
      filename: 'input.ts',
      output: 'Math.log2(-(Math.SQRT2 as number));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 48
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.log(Math.E) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Math.E);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Math.LN2) * Math.LOG10E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(Math.LN2);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math.PI ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(Math.PI);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math.E * Math.E + Math.PI * Math.PI);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(Math.E, Math.PI);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Math.custom) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Math.custom);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Math.random) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Math.random);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math["PI"] * Math["PI"]);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(Math["PI"]);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Math?.PI) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Math?.PI);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(Math.PI * /* keep */ Math.PI);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'delete Math.sqrt(value ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'delete Math.abs(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'delete (Math.log(/* keep */ value) / Math.LN2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 46,
          suggestions: [],
        },
      ],
    },
    {
      code: 'delete (Math.log(value) / Math.LN2).property;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'delete (Math.log2(value)).property;',
            },
          ],
        },
      ],
    },
  ],
});

// upstream line 67
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.log(Number.NaN) / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2(Number.NaN);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(Number./* keep */NaN) / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2(Number./* keep */NaN);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(Number.POSITIVE_INFINITY) * Math.LOG10E;',
      filename: 'input.js',
      output: 'Math.log10(Number.POSITIVE_INFINITY);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.LOG2E * Math.log(Number.NEGATIVE_INFINITY);',
      filename: 'input.js',
      output: 'Math.log2(Number.NEGATIVE_INFINITY);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(Number.MIN_SAFE_INTEGER) / Math.LN10;',
      filename: 'input.js',
      output: 'Math.log10(Number.MIN_SAFE_INTEGER);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(-Number.MIN_VALUE) / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2(-Number.MIN_VALUE);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(-Number.NEGATIVE_INFINITY) / Math.LN10;',
      filename: 'input.js',
      output: 'Math.log10(-Number.NEGATIVE_INFINITY);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.NaN ** 2);',
      filename: 'input.js',
      output: 'Math.abs(Number.NaN);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.POSITIVE_INFINITY ** 2);',
      filename: 'input.js',
      output: 'Math.abs(Number.POSITIVE_INFINITY);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.NEGATIVE_INFINITY ** 2);',
      filename: 'input.js',
      output: 'Math.abs(Number.NEGATIVE_INFINITY);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt((-Number.POSITIVE_INFINITY) ** 2);',
      filename: 'input.js',
      output: 'Math.abs((-Number.POSITIVE_INFINITY));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt((+Number.POSITIVE_INFINITY) ** 2);',
      filename: 'input.js',
      output: 'Math.abs((+Number.POSITIVE_INFINITY));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.EPSILON * Number.EPSILON);',
      filename: 'input.js',
      output: 'Math.abs(Number.EPSILON);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.MAX_SAFE_INTEGER * Number.MAX_SAFE_INTEGER);',
      filename: 'input.js',
      output: 'Math.abs(Number.MAX_SAFE_INTEGER);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.MIN_SAFE_INTEGER * Number.MIN_SAFE_INTEGER);',
      filename: 'input.js',
      output: 'Math.abs(Number.MIN_SAFE_INTEGER);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.MAX_VALUE * Number.MAX_VALUE + Number.NaN ** 2);',
      filename: 'input.js',
      output: 'Math.hypot(Number.MAX_VALUE, Number.NaN);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.NEGATIVE_INFINITY ** 2 + 3 ** 2);',
      filename: 'input.js',
      output: 'Math.hypot(Number.NEGATIVE_INFINITY, 3);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = Number.NaN; Math.log(value) / Math.LN2;',
      filename: 'input.js',
      output: 'const value = Number.NaN; Math.log2(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 53,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = -Number.POSITIVE_INFINITY; Math.sqrt(value ** 2);',
      filename: 'input.js',
      output: 'const value = -Number.POSITIVE_INFINITY; Math.abs(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 51,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void Math.sqrt(Number.MAX_VALUE ** 2 + Number.MIN_VALUE ** 2);',
      filename: 'input.js',
      output: 'void Math.hypot(Number.MAX_VALUE, Number.MIN_VALUE);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 97
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.log(Number.NaN as number) / Math.LN2;',
      filename: 'input.ts',
      output: 'Math.log2(Number.NaN as number);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = Number.POSITIVE_INFINITY as number; Math.sqrt(value ** 2);',
      filename: 'input.ts',
      output:
        'const value = Number.POSITIVE_INFINITY as number; Math.abs(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 56,
          endLine: 1,
          endColumn: 60,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(-(Number.MIN_VALUE as number)) * Math.LOG2E;',
      filename: 'input.ts',
      output: 'Math.log2(-(Number.MIN_VALUE as number));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 111
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt(Number.MAX_VALUE * Number.MAX_VALUE);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(Number.MAX_VALUE);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.MIN_VALUE * Number.MIN_VALUE);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(Number.MIN_VALUE);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.EPSILON ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(Number.EPSILON);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.NaN ** 2 + Number.POSITIVE_INFINITY ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(Number.NaN, Number.POSITIVE_INFINITY);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Number.EPSILON) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Number.EPSILON);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Number.MIN_VALUE) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Number.MIN_VALUE);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Number.length) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Number.length);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Number.custom) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Number.custom);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Number["NaN"]) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Number["NaN"]);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Number?.NaN) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Number?.NaN);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Number.prototype) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Number.prototype);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(other.NaN) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(other.NaN);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Number.NaN /* keep */ + value) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(Number.NaN /* keep */ + value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.NaN ** /* keep */ 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 132
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt(NaN ** 2 + 3 ** 2);',
      filename: 'input.js',
      output: 'Math.hypot(NaN, 3);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(1e200 * 1e200 + NaN ** 2);',
      filename: 'input.js',
      output: 'Math.hypot(1e200, NaN);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt("not a number" ** 2 + 0 ** 2);',
      filename: 'input.js',
      output: 'Math.hypot("not a number", 0);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = NaN; Math.sqrt((0 ** 2 + value ** 2) + 4 * 4);',
      filename: 'input.js',
      output: 'const value = NaN; Math.hypot(0, value, 4);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 29,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(undefined) / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2(undefined);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = undefined; Math.log(value) * Math.LOG10E;',
      filename: 'input.js',
      output: 'const value = undefined; Math.log10(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 55,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(undefined ** 2);',
      filename: 'input.js',
      output: 'Math.abs(undefined);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(undefined ** 2 + 4 * 4);',
      filename: 'input.js',
      output: 'Math.hypot(undefined, 4);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'for (Math.log(value) / Math.LN2; condition; update()) {}',
      filename: 'input.js',
      output: 'for (Math.log2(value); condition; update()) {}',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 32,
          suggestions: [],
        },
      ],
    },
    {
      code: 'for (; condition; Math.LOG10E * Math.log(getValue())) {}',
      filename: 'input.js',
      output: 'for (; condition; Math.log10(getValue())) {}',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 53,
          suggestions: [],
        },
      ],
    },
    {
      code: 'for (start(); condition; next(), Math.log(value) / Math.LN10) {}',
      filename: 'input.js',
      output: 'for (start(); condition; next(), Math.log10(value)) {}',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 61,
          suggestions: [],
        },
      ],
    },
    {
      code: 'typeof (Math.log(value) * Math.LOG2E);',
      filename: 'input.js',
      output: 'typeof (Math.log2(value));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 37,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const type = typeof (Math.LOG10E * Math.log(getValue()));',
      filename: 'input.js',
      output: 'const type = typeof (Math.log10(getValue()));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 56,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void (condition ? Math.log(value) / Math.LN2 : fallback());',
      filename: 'input.js',
      output: 'void (condition ? Math.log2(value) : fallback());',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 45,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void (condition ? fallback() : Math.log(value) / Math.LN10);',
      filename: 'input.js',
      output: 'void (condition ? fallback() : Math.log10(value));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 59,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void (condition && Math.log(value) / Math.LN2);',
      filename: 'input.js',
      output: 'void (condition && Math.log2(value));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 46,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const type = typeof (condition || Math.log(value) * Math.LOG10E);',
      filename: 'input.js',
      output: 'const type = typeof (condition || Math.log10(value));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 64,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const result = (condition ?? Math.log(value) / Math.LN2, next());',
      filename: 'input.js',
      output: 'const result = (condition ?? Math.log2(value), next());',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 56,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void (first ? second && (next(), Math.log(getValue()) / Math.LN2) : fallback());',
      filename: 'input.js',
      output:
        'void (first ? second && (next(), Math.log2(getValue())) : fallback());',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 65,
          suggestions: [],
        },
      ],
    },
    {
      code: 'for (; condition; flag ? Math.log(value) / Math.LN2 : fallback()) {}',
      filename: 'input.js',
      output: 'for (; condition; flag ? Math.log2(value) : fallback()) {}',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 52,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void Math.sqrt(4 ** 2);',
      filename: 'input.js',
      output: 'void Math.abs(4);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void Math.sqrt(1e200 * 1e200);',
      filename: 'input.js',
      output: 'void Math.abs(1e200);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [],
        },
      ],
    },
    {
      code: 'typeof Math.sqrt(3 * 3 + 4 * 4);',
      filename: 'input.js',
      output: 'typeof Math.hypot(3, 4);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 17,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void Math.sqrt(NaN ** 2 + Infinity ** 2);',
      filename: 'input.js',
      output: 'void Math.hypot(NaN, Infinity);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const result = (Math.sqrt(3 ** 2 + 4 ** 2), next());',
      filename: 'input.js',
      output: 'const result = (Math.hypot(3, 4), next());',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 167
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt((NaN as number) ** 2 + 3 ** 2);',
      filename: 'input.ts',
      output: 'Math.hypot((NaN as number), 3);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(undefined as undefined) / Math.LN2;',
      filename: 'input.ts',
      output: 'Math.log2(undefined as undefined);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
    },
    {
      code: 'for ((Math.log(value) / Math.LN2) as number; condition; update()) {}',
      filename: 'input.ts',
      output: 'for ((Math.log2(value)) as number; condition; update()) {}',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 33,
          suggestions: [],
        },
      ],
    },
    {
      code: 'typeof ((Math.log(value) / Math.LN2) satisfies number);',
      filename: 'input.ts',
      output: 'typeof ((Math.log2(value)) satisfies number);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void ((condition ? (Math.log(value) / Math.LN2) as number : fallback()) satisfies number);',
      filename: 'input.ts',
      output:
        'void ((condition ? (Math.log2(value)) as number : fallback()) satisfies number);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 47,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 183
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt(NaN ** 2 + value ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(NaN, value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(undefined ** 2 + Infinity ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(undefined, Infinity);',
            },
          ],
        },
      ],
    },
    {
      code: 'const result = condition ? Math.log(value) / Math.LN2 : fallback();',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 54,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output:
                'const result = condition ? Math.log2(value) : fallback();',
            },
          ],
        },
      ],
    },
    {
      code: 'const result = condition && Math.log(value) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 55,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'const result = condition && Math.log2(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'void ((Math.log(value) / Math.LN2) ? first() : second());',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'void ((Math.log2(value)) ? first() : second());',
            },
          ],
        },
      ],
    },
    {
      code: 'void ((Math.log(value) / Math.LN2) && next());',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'void ((Math.log2(value)) && next());',
            },
          ],
        },
      ],
    },
    {
      code: 'for (; Math.log(value) / Math.LN2; update()) {}',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'for (; Math.log2(value); update()) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'for (const value = Math.log(input) / Math.LN2; condition; update()) {}',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 46,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output:
                'for (const value = Math.log2(input); condition; update()) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'for (; condition; update()) { Math.log(value) / Math.LN2; }',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 57,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'for (; condition; update()) { Math.log2(value); }',
            },
          ],
        },
      ],
    },
    {
      code: 'void Math.sqrt(getValue() ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'void Math.abs(getValue());',
            },
          ],
        },
      ],
    },
    {
      code: 'void Math.sqrt(1n ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'void Math.abs(1n);',
            },
          ],
        },
      ],
    },
    {
      code: 'typeof Math.sqrt(value * value);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'typeof Math.abs(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'void (Math.sqrt(1e-200 * 1e-200) && next());',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'void (Math.abs(1e-200) && next());',
            },
          ],
        },
      ],
    },
    {
      code: 'void (Math.sqrt(1e-200 * 1e-200) ? first() : second());',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'void (Math.abs(1e-200) ? first() : second());',
            },
          ],
        },
      ],
    },
    {
      code: 'void Math.sqrt(getValue() ** 2 + 3 ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'void Math.hypot(getValue(), 3);',
            },
          ],
        },
      ],
    },
    {
      code: 'void Math.sqrt(4 ** /* keep */ 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [],
        },
      ],
    },
    {
      code: 'for (Math.log(/* keep */ value) / Math.LN2; condition; update()) {}',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 207
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt(0 ** 2);',
      filename: 'input.js',
      output: 'Math.abs(0);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt((-0) ** 2);',
      filename: 'input.js',
      output: 'Math.abs((-0));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Infinity ** 2);',
      filename: 'input.js',
      output: 'Math.abs(Infinity);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt((-Infinity) ** 2);',
      filename: 'input.js',
      output: 'Math.abs((-Infinity));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(NaN ** 2);',
      filename: 'input.js',
      output: 'Math.abs(NaN);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt("4" * "4");',
      filename: 'input.js',
      output: 'Math.abs("4");',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(true * true);',
      filename: 'input.js',
      output: 'Math.abs(true);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = -0; Math.sqrt(value ** 2);',
      filename: 'input.js',
      output: 'const value = -0; Math.abs(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = Infinity; Math.sqrt(value ** 2);',
      filename: 'input.js',
      output: 'const value = Infinity; Math.abs(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 34,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(0 ** 2 + 0 ** 2);',
      filename: 'input.js',
      output: 'Math.hypot(0, 0);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt((0 * 0 + (-0) ** 2) + 0 ** 2);',
      filename: 'input.js',
      output: 'Math.hypot(0, (-0), 0);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(Infinity ** 2 + 3 * 3);',
      filename: 'input.js',
      output: 'Math.hypot(Infinity, 3);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(3 ** 2 + (-Infinity) ** 2);',
      filename: 'input.js',
      output: 'Math.hypot(3, (-Infinity));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = -Infinity; Math.sqrt(4 * 4 + value ** 2);',
      filename: 'input.js',
      output: 'const value = -Infinity; Math.hypot(4, value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(1e999 * 1e999 + 1e200 * 1e200);',
      filename: 'input.js',
      output: 'Math.hypot(1e999, 1e200);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(true) / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2(true);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(false) * Math.LOG10E;',
      filename: 'input.js',
      output: 'Math.log10(false);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.LOG2E * Math.log(null);',
      filename: 'input.js',
      output: 'Math.log2(null);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log("1") / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2("1");',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(`1`) / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2(`1`);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(`0` ** 2);',
      filename: 'input.js',
      output: 'Math.abs(`0`);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(`not a number` ** 2);',
      filename: 'input.js',
      output: 'Math.abs(`not a number`);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log("-1") * Math.LOG2E;',
      filename: 'input.js',
      output: 'Math.log2("-1");',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log("not a number") / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2("not a number");',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log("Infinity") / Math.LN10;',
      filename: 'input.js',
      output: 'Math.log10("Infinity");',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = true; Math.log(value) / Math.LN2;',
      filename: 'input.js',
      output: 'const value = true; Math.log2(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 47,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void (Math.log(value) / Math.LN2);',
      filename: 'input.js',
      output: 'void (Math.log2(value));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 33,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void (Math.log(getValue()) * Math.LOG10E);',
      filename: 'input.js',
      output: 'void (Math.log10(getValue()));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 41,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void (Math.LOG2E * Math.log(value));',
      filename: 'input.js',
      output: 'void (Math.log2(value));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const result = (Math.log(value) / Math.LN10, next());',
      filename: 'input.js',
      output: 'const result = (Math.log10(value), next());',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const result = (Math.LOG10E * Math.log(value), next());',
      filename: 'input.js',
      output: 'const result = (Math.log10(value), next());',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 46,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const result = ((next(), Math.log(value) / Math.LN2), final());',
      filename: 'input.js',
      output: 'const result = ((next(), Math.log2(value)), final());',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 52,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void (next(), Math.log(value) / Math.LN2);',
      filename: 'input.js',
      output: 'void (next(), Math.log2(value));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 41,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void (Math.log((/* keep */ value)) / Math.LN2);',
      filename: 'input.js',
      output: 'void (Math.log2((/* keep */ value)));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 46,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void (Math.log(1n) / Math.LN2);',
      filename: 'input.js',
      output: 'void (Math.log2(1n));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
    },
    {
      code: 'async function calculate() { return void (Math.log(await getValue()) / Math.LN2); }',
      filename: 'input.js',
      output:
        'async function calculate() { return void (Math.log2(await getValue())); }',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 80,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 253
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt((0 as number) ** 2);',
      filename: 'input.ts',
      output: 'Math.abs((0 as number));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = Infinity as number; Math.sqrt(value ** 2 + 4 ** 2);',
      filename: 'input.ts',
      output: 'const value = Infinity as number; Math.hypot(value, 4);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(true as boolean) / Math.LN2;',
      filename: 'input.ts',
      output: 'Math.log2(true as boolean);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void ((Math.log(value) / Math.LN2) as number);',
      filename: 'input.ts',
      output: 'void ((Math.log2(value)) as number);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 34,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const result = ((Math.log(value) / Math.LN2) satisfies number, next());',
      filename: 'input.ts',
      output: 'const result = ((Math.log2(value)) satisfies number, next());',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 269
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt(Infinity ** 2 + "not a number" ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(Infinity, "not a number");',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(NaN ** 2 + (-Infinity) ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(NaN, (-Infinity));',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(Infinity ** 2 + value ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(Infinity, value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(0 ** 2 + getValue() ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(0, getValue());',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(1 ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(1);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((-1) ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs((-1));',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(0 ** /* keep */ 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(2n) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(2n);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(/1/) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(/1/);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log({valueOf() { return 1; }}) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2({valueOf() { return 1; }});',
            },
          ],
        },
      ],
    },
    {
      code: 'const result = (next(), Math.log(value) / Math.LN2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 51,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'const result = (next(), Math.log2(value));',
            },
          ],
        },
      ],
    },
    {
      code: 'void (Math.log(value) / Math.LN2 + next());',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'void (Math.log2(value) + next());',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(value) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'void (Math.log(/* keep */ value) / Math.LN2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
    },
    {
      code: 'void Math.sqrt(value ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'void Math.abs(value);',
            },
          ],
        },
      ],
    },
  ],
});

// upstream line 291
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt(4 * 4);',
      filename: 'input.js',
      output: 'Math.abs(4);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(0 * 0);',
      filename: 'input.js',
      output: 'Math.abs(0);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(0.5 * 0.5);',
      filename: 'input.js',
      output: 'Math.abs(0.5);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(1.5 * 1.5);',
      filename: 'input.js',
      output: 'Math.abs(1.5);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(1e150 * 1e150);',
      filename: 'input.js',
      output: 'Math.abs(1e150);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(1e-150 * 1e-150);',
      filename: 'input.js',
      output: 'Math.abs(1e-150);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(1e999 * 1e999);',
      filename: 'input.js',
      output: 'Math.abs(1e999);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = 4; Math.sqrt(value * value);',
      filename: 'input.js',
      output: 'const value = 4; Math.abs(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 27,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = -4; Math.sqrt(value * value);',
      filename: 'input.js',
      output: 'const value = -4; Math.abs(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = -0; Math.sqrt(value * value);',
      filename: 'input.js',
      output: 'const value = -0; Math.abs(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(((4)) * ((4)),);',
      filename: 'input.js',
      output: 'Math.abs(((4)),);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt((/* keep */ 4) * 4);',
      filename: 'input.js',
      output: 'Math.abs((/* keep */ 4));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(1) * Math.LOG10E;',
      filename: 'input.js',
      output: 'Math.log10(1);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.LOG10E * Math.log(0);',
      filename: 'input.js',
      output: 'Math.log10(0);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(-1) / Math.LN10;',
      filename: 'input.js',
      output: 'Math.log10(-1);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(-0) * Math.LOG2E;',
      filename: 'input.js',
      output: 'Math.log2(-0);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.LOG2E * Math.log(1);',
      filename: 'input.js',
      output: 'Math.log2(1);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(1) / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2(1);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(+1) / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2(+1);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(1e999) / Math.LN10;',
      filename: 'input.js',
      output: 'Math.log10(1e999);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(-1e999) * Math.LOG2E;',
      filename: 'input.js',
      output: 'Math.log2(-1e999);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = 1; Math.log(value) / Math.LN2;',
      filename: 'input.js',
      output: 'const value = 1; Math.log2(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = -1; Math.LOG10E * Math.log(value);',
      filename: 'input.js',
      output: 'const value = -1; Math.log10(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 48,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log((/* keep */ 1)) / Math.LN2;',
      filename: 'input.js',
      output: 'Math.log2((/* keep */ 1));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 326
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt(1e200 * 1e200);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(1e200);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(1e-200 * 1e-200);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(1e-200);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(1e-160 * 1e-160);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(1e-160);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(4 ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(4);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(3 * 3 + 4 * 4);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(3, 4);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(2) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(2);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log("2") / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2("2");',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(1n) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(1n);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log({valueOf() { return 1; }}) * Math.LOG10E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10({valueOf() { return 1; }});',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(getValue()) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(getValue());',
            },
          ],
        },
      ],
    },
    {
      code: 'let value = 1; value = 2; Math.log(value) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 53,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'let value = 1; value = 2; Math.log2(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'const value = getValue(); Math.log(value) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 53,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'const value = getValue(); Math.log2(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'const value = 4; const alias = value; Math.sqrt(alias * alias);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 48,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'const value = 4; const alias = value; Math.abs(alias);',
            },
          ],
        },
      ],
    },
    {
      code: 'const value = 2 + 2; Math.sqrt(value * value);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'const value = 2 + 2; Math.abs(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'const value = {number: 4}; Math.sqrt(value.number * value.number);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'const value = {number: 4}; Math.abs(value.number);',
            },
          ],
        },
      ],
    },
    {
      code: 'const value = 4; function calculate(value) { return Math.sqrt(value * value); }',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 58,
          endLine: 1,
          endColumn: 62,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output:
                'const value = 4; function calculate(value) { return Math.abs(value); }',
            },
          ],
        },
      ],
    },
    {
      code: 'const value = 1; { let value = 2; Math.log(value) / Math.LN2; }',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 61,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'const value = 1; { let value = 2; Math.log2(value); }',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(4 * /* keep */ 4);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(/* keep */ 1) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 351
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.log(1 as number) / Math.LN2;',
      filename: 'input.ts',
      output: 'Math.log2(1 as number);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt((4 as number) * (4 as number));',
      filename: 'input.ts',
      output: 'Math.abs((4 as number));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value: number = 4; Math.sqrt(value * value);',
      filename: 'input.ts',
      output: 'const value: number = 4; Math.abs(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = 4; Math.sqrt((value as number) * value);',
      filename: 'input.ts',
      output: 'const value = 4; Math.abs((value as number));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 27,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = 1; Math.log(value as number) / Math.LN2;',
      filename: 'input.ts',
      output: 'const value = 1; Math.log2(value as number);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 54,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const value = -1 as number; Math.log(value) / Math.LN10;',
      filename: 'input.ts',
      output: 'const value = -1 as number; Math.log10(value);',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 56,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log((1 satisfies number)) * Math.LOG2E;',
      filename: 'input.ts',
      output: 'Math.log2((1 satisfies number));',
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 369
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'const x = 1e200, y = 1; Math.sqrt(x * x + y * y);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'const x = 1e200, y = 1; Math.hypot(x, y);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(3) * Math.LOG10E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(3);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(1000) / Math.LN10;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(1000);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(9) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(9);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(1e200 ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(1e200);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(1e-200 ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(1e-200);',
            },
          ],
        },
      ],
    },
  ],
});

// upstream line 388
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.log(`2`) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(`2`);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(`${1}`) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(`${1}`);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(`${0}` ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(`${0}`);',
            },
          ],
        },
      ],
    },
    {
      code: 'const {length} = "Infinity"; Math.log(length) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 57,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'const {length} = "Infinity"; Math.log2(length);',
            },
          ],
        },
      ],
    },
    {
      code: 'const [value] = "4"; Math.sqrt(value * value);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'const [value] = "4"; Math.abs(value);',
            },
          ],
        },
      ],
    },
  ],
});

// upstream line 406
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.sqrt((a ** 2 + b ** 2) + (c ** 2 + d ** 2));',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b, c, d);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a ** 2 + ((b ** 2 + c ** 2) + (d * d + e ** 2)));',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b, c, d, e);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(((a ** 2 + b ** 2) + (c ** 2 + d ** 2)) + ((e ** 2 + f ** 2) + (g ** 2 + h ** 2)));',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b, c, d, e, f, g, h);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((((a ** 2 + b ** 2))) + (((c ** 2 + d ** 2))),);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b, c, d,);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(((a as number) ** 2 + b ** 2) + (c ** 2 + d ** 2));',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot((a as number), b, c, d);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((a ** 2 + b ** 2) /* keep */ + (c ** 2 + d ** 2));',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b /* keep */, c, d);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((a ** /* keep */ 2 + b ** 2) + (c ** 2 + d ** 2));',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream line 428
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [
    {
      code: 'Math.log(x) * Math.log(x)',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) * Math.log(x)',
      filename: 'input.js',
    },
    {
      code: 'Math.LOG10E * Math.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math.LOG2E * Math.LOG2E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) * Math[LOG10E]',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) * Math[LOG10E]',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) * LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) * LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math[log](x) * Math.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math[log](x) * Math.LOG2E',
      filename: 'input.js',
    },
    {
      code: 'foo.Math.log(x) * Math.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'foo.Math.log(x) * Math.LOG2E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) * foo.Math.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) * foo.Math.LOG2E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) * Math.NOT_LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) * Math.NOT_LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) * Math?.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) * Math?.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math?.log(x) * Math.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math?.log(x) * Math.LOG2E',
      filename: 'input.js',
    },
    {
      code: 'log(x) * Math.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'log(x) * Math.LOG2E',
      filename: 'input.js',
    },
    {
      code: 'new Math.log(x) * Math.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'new Math.log(x) * Math.LOG2E',
      filename: 'input.js',
    },
    {
      code: 'Math.not_log(x) + Math.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math.not_log(x) + Math.LOG2E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x)[Math.LOG10E]',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x)[Math.LOG2E]',
      filename: 'input.js',
    },
    {
      code: 'Math.log() * Math.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math.log() * Math.LOG2E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x, extraArgument) * Math.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x, extraArgument) * Math.LOG2E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(...x) * Math.LOG10E',
      filename: 'input.js',
    },
    {
      code: 'Math.log(...x) * Math.LOG2E',
      filename: 'input.js',
    },
    {
      code: 'Math.LN10 / Math.LN10',
      filename: 'input.js',
    },
    {
      code: 'Math.LN2 / Math.LN2',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) /Math[LN10]',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) /Math[LN10]',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) / LN10',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) / LN10',
      filename: 'input.js',
    },
    {
      code: 'Math[log](x) / Math.LN10',
      filename: 'input.js',
    },
    {
      code: 'Math[log](x) / Math.LN2',
      filename: 'input.js',
    },
    {
      code: 'foo.Math.log(x) / Math.LN10',
      filename: 'input.js',
    },
    {
      code: 'foo.Math.log(x) / Math.LN2',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) / foo.Math.LN10',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) / foo.Math.LN2',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) / Math.log(x)',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) / Math.log(x)',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) / Math.NOT_LN10',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) / Math.NOT_LN10',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) / Math?.LN10',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) / Math?.LN10',
      filename: 'input.js',
    },
    {
      code: 'Math?.log(x) / Math.LN10',
      filename: 'input.js',
    },
    {
      code: 'Math?.log(x) / Math.LN2',
      filename: 'input.js',
    },
    {
      code: 'log(x) / Math.LN10',
      filename: 'input.js',
    },
    {
      code: 'log(x) / Math.LN2',
      filename: 'input.js',
    },
    {
      code: 'new Math.log(x) / Math.LN10',
      filename: 'input.js',
    },
    {
      code: 'new Math.log(x) / Math.LN2',
      filename: 'input.js',
    },
    {
      code: 'Math.not_log(x) + Math.LN10',
      filename: 'input.js',
    },
    {
      code: 'Math.not_log(x) + Math.LN2',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x)[Math.LN10]',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x)[Math.LN2]',
      filename: 'input.js',
    },
    {
      code: 'Math.log() / Math.LN10',
      filename: 'input.js',
    },
    {
      code: 'Math.log() / Math.LN2',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x, extraArgument) / Math.LN10',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x, extraArgument) / Math.LN2',
      filename: 'input.js',
    },
    {
      code: 'Math.log(...x) / Math.LN10',
      filename: 'input.js',
    },
    {
      code: 'Math.log(...x) / Math.LN2',
      filename: 'input.js',
    },
  ],
  invalid: [
    {
      code: 'Math.log(x) * Math.LOG10E',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(x) * Math.LOG2E',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG10E * Math.log(x)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG2E * Math.log(x)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(x) / Math.LN10',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(x) / Math.LN2',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log((( 0,x ))) * Math.LOG10E',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10((( 0,x )))',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log((( 0,x ))) * Math.LOG2E',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2((( 0,x )))',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG10E * Math.log((( 0,x )))',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10((( 0,x )))',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG2E * Math.log((( 0,x )))',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2((( 0,x )))',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log((( 0,x ))) / Math.LN10',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10((( 0,x )))',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log((( 0,x ))) / Math.LN2',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2((( 0,x )))',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo(x) {\n\treturn (\n\t\tMath.log(x)\n\t\t\t/ Math.LN10\n\t);\n}',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 3,
          column: 3,
          endLine: 4,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output:
                'function foo(x) {\n\treturn (\n\t\tMath.log10(x)\n\t);\n}',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo(x) {\n\treturn (\n\t\tMath.log(x)\n\t\t\t/ Math.LN2\n\t);\n}',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 3,
          column: 3,
          endLine: 4,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output:
                'function foo(x) {\n\treturn (\n\t\tMath.log2(x)\n\t);\n}',
            },
          ],
        },
      ],
    },
  ],
});

// upstream line 486
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [
    {
      code: 'Math.notSqrt(a ** 2 + b ** 2)',
      filename: 'input.js',
    },
    {
      code: 'NotMath.sqrt(a ** 2 + b ** 2)',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt(a ** 2 - b ** 2)',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt(a ** 2 + 2 ** b)',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt(a * c + b * c)',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt((++a) * (++a))',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt(Math.pow(a, 2) + Math.pow(b, 2))',
      filename: 'input.js',
    },
  ],
  invalid: [
    {
      code: 'Math.sqrt(a * a + b * b)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a ** 2 + b ** 2)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a * a + b ** 2)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a * a + b * b + c * c)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b, c)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a ** 2 + b ** 2 + c ** 2)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b, c)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a * a)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(a)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a ** 2)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(a)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a * a,)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(a,)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a ** 2,)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(a,)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((a, b) ** 2)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs((a, b))',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((++a) ** 2)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs((++a))',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a * a + b * b,)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b,)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a ** 2 + b ** 2,)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b,)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((( a ** 2 )) + (( b ** 2 + c ** 2 )) + (( d )) * (( d )) + (( e )) ** (( 2 )))',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot((( a )),  b, c , (( d )), (( e )))',
            },
          ],
        },
      ],
    },
  ],
});

// upstream line 516
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.log(/* keep */ x) * Math.LOG10E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(x) // keep\n\t* Math.LOG10E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 15,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(x) / Math.LN10;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(x);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(x) / /* keep */ Math.LN10;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(a ** /* keep */ 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(a ** // keep\n\t2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(x /* keep */ + 1) * Math.LOG10E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(x /* keep */ + 1);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log((/* keep */ x)) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2((/* keep */ x));',
            },
          ],
        },
      ],
    },
  ],
});

// upstream line 532
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.log(value) * Math.LOG10E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG10E * Math.log(value);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(value) * Math.LOG2E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG2E * Math.log(value);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(value) / Math.LN10;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(value) / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(value * value);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(value ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(first * first + second * second);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(first, second);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(first ** 2 + second ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(first, second);',
            },
          ],
        },
      ],
    },
  ],
});

// upstream line 555
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'const x = 1e-200, y = 1e-200; Math.sqrt(x * x + y * y);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 40,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'const x = 1e-200, y = 1e-200; Math.hypot(x, y);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(Infinity ** 2 + NaN ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(Infinity, NaN);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(Number.MAX_VALUE ** 2 + Number.MIN_VALUE ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(Number.MAX_VALUE, Number.MIN_VALUE);',
            },
          ],
        },
      ],
    },
    {
      code: 'const x = 1e200; Math.sqrt(x * x);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'const x = 1e200; Math.abs(x);',
            },
          ],
        },
      ],
    },
    {
      code: 'const x = 1e-200; Math.sqrt(x * x);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'const x = 1e-200; Math.abs(x);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG2E * Math.log(9);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(9);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG10E * Math.log(3);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(3);',
            },
          ],
        },
      ],
    },
  ],
});

// upstream line 569
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [
    {
      code: 'Math.log?.(x) * Math.LOG10E;',
      filename: 'input.js',
    },
    {
      code: 'Math.LOG2E * Math.log?.(x);',
      filename: 'input.js',
    },
    {
      code: 'Math.log?.(x) / Math.LN2;',
      filename: 'input.js',
    },
    {
      code: 'Math["log"](x) * Math.LOG10E;',
      filename: 'input.js',
    },
    {
      code: 'Math.log(x) / Math["LN10"];',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt();',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt(a ** 2, b ** 2);',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt(...squares);',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt?.(a ** 2 + b ** 2);',
      filename: 'input.js',
    },
    {
      code: 'Math?.sqrt(a ** 2 + b ** 2);',
      filename: 'input.js',
    },
    {
      code: 'Math["sqrt"](a ** 2 + b ** 2);',
      filename: 'input.js',
    },
    {
      code: 'new Math.sqrt(a ** 2 + b ** 2);',
      filename: 'input.js',
    },
    {
      code: 'math.sqrt(a ** 2 + b ** 2);',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt(a ** 2 + b);',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt(a ** 2 + b ** 3);',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt(a ** 2 * b ** 2);',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt(getValue() * getValue());',
      filename: 'input.js',
    },
    {
      code: 'Math.sqrt(a * b + c * c);',
      filename: 'input.js',
    },
  ],
  invalid: [],
});

// upstream line 594
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'const result = 2 * (Math.log(value) / Math.LN10);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 48,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'const result = 2 * (Math.log10(value));',
            },
          ],
        },
      ],
    },
    {
      code: 'const result = (Math.log(value) * Math.LOG2E) ** 2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 45,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'const result = (Math.log2(value)) ** 2;',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(getValue()) * Math.LOG10E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(getValue());',
            },
          ],
        },
      ],
    },
    {
      code: 'async function calculate() { return Math.log(await getValue()) / Math.LN2; }',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 74,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output:
                'async function calculate() { return Math.log2(await getValue()); }',
            },
          ],
        },
      ],
    },
    {
      code: 'function * calculate() { return Math.log(yield value) * Math.LOG2E; }',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 67,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output:
                'function * calculate() { return Math.log2(yield value); }',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(value = nextValue) / Math.LN10;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(value = nextValue);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG10E * Math.log(condition ? first : second);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(condition ? first : second);',
            },
          ],
        },
      ],
    },
    {
      code: '(Math).log(value,) / (Math.LN2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'const distance = Math.sqrt(point.x * point.x + point.y * point.y);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'const distance = Math.hypot(point.x, point.y);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(values[index] * values[index] + other ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(values[index], other);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((first + second) ** 2 + (third ? fourth : fifth) ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot((first + second), (third ? fourth : fifth));',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(getValue() ** 2 + other ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(getValue(), other);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((first, second) ** 2 + third ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot((first, second), third);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(value ** 2).toFixed(2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(value).toFixed(2);',
            },
          ],
        },
      ],
    },
    {
      code: 'const result = <span>{Math.log(value) / Math.LN10}</span>;',
      filename: 'input.jsx',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 50,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'const result = <span>{Math.log10(value)}</span>;',
            },
          ],
        },
      ],
    },
    {
      code: 'const result = <>{Math.sqrt(first ** 2 + second ** 2)}</>;',
      filename: 'input.jsx',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'const result = <>{Math.hypot(first, second)}</>;',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Math.sqrt(value ** 2)) / Math.LN10;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(Math.sqrt(value ** 2));',
            },
          ],
        },
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.log(Math.abs(value)) / Math.LN10;',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((Math.log(value) * Math.LOG2E) ** 2 + other ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot((Math.log(value) * Math.LOG2E), other);',
            },
          ],
        },
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 40,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.sqrt((Math.log2(value)) ** 2 + other ** 2);',
            },
          ],
        },
      ],
    },
  ],
});

// upstream line 625
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.log(value as number) * Math.LOG10E;',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(value as number);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG2E * Math.log(<number>value);',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(<number>value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(value!) / Math.LN2;',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(value!);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(value satisfies number) / Math.LN10;',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(value satisfies number);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(getValue<number>()) * Math.LOG2E;',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(getValue<number>());',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((value as number) ** 2);',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs((value as number));',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((<number>value) ** 2);',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs((<number>value));',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(value! * value!);',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(value!);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((value satisfies number) ** 2);',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs((value satisfies number));',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(first! ** 2 + (second as number) ** 2);',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(first!, (second as number));',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(getValue<number>() ** 2 + second ** 2);',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(getValue<number>(), second);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(value ** 2) satisfies number;',
      filename: 'input.ts',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(value) satisfies number;',
            },
          ],
        },
      ],
    },
  ],
});

// upstream line 644
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'Math.LOG10E /* keep */ * Math.log(value);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.LOG2E * Math.log(/* keep */ value);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(value) /* keep */ / Math.LN2;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math /* keep */ .log(value) * Math.LOG10E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log(value) * Math./* keep */LOG2E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(value * /* keep */ value);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt(first ** 2 + second ** /* keep */ 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.sqrt((first ** 2 + second ** 2) + (third * /* keep */ third + fourth ** 2));',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [],
        },
      ],
    },
    {
      code: 'Math.log((/* keep */ value)) * Math.LOG2E;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2((/* keep */ value));',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG10E * Math.log(value /* keep */ + offset);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(value /* keep */ + offset);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(/* keep */ first ** 2 + second ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(/* keep */ first, second);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((/* keep */ first) ** 2 + second ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot((/* keep */ first), second);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(first ** 2 + /* keep */ second ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(first, /* keep */ second);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math /* keep */ .sqrt(value ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math /* keep */ .abs(value);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(first ** 2 // keep\n  + second ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(first // keep\n, second);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(first ** 2 + // keep\n  second ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(first, // keep\n  second);',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(Math.sqrt(value ** 2)) /* keep */ / Math.LN10;',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
          suggestions: [],
        },
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.log(Math.abs(value)) /* keep */ / Math.LN10;',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt((Math.log(/* keep */ value) * Math.LOG2E) ** 2 + other ** 2);',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output:
                'Math.hypot((Math.log(/* keep */ value) * Math.LOG2E), other);',
            },
          ],
        },
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 51,
          suggestions: [],
        },
      ],
    },
    {
      code: 'function distance() {\r\n  return Math.sqrt(\r\n    first ** 2 +\r\n    second ** 2\r\n  );\r\n}',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output:
                'function distance() {\r\n  return Math.hypot(\r\n    first,\r\n    second\r\n  );\r\n}',
            },
          ],
        },
      ],
    },
    {
      code: 'function logarithm() {\r\n  return (\r\n    Math.log(value /* keep */ + offset)\r\n      / Math.LN10\r\n  );\r\n}',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 3,
          column: 5,
          endLine: 4,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output:
                'function logarithm() {\r\n  return (\r\n    Math.log10(value /* keep */ + offset)\r\n  );\r\n}',
            },
          ],
        },
      ],
    },
  ],
});

// documentation examples
ruleTester.run('prefer-modern-math-apis', {} as never, {
  valid: [
    {
      code: 'Math.log10(x)',
      filename: 'input.js',
    },
    {
      code: 'Math.log2(x)',
      filename: 'input.js',
    },
    {
      code: 'Math.hypot(a, b)',
      filename: 'input.js',
    },
    {
      code: 'Math.abs(x)',
      filename: 'input.js',
    },
  ],
  invalid: [
    {
      code: 'Math.log(x) * Math.LOG10E',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG10E * Math.log(x)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(x) / Math.LN10',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log10(…)`.',
              output: 'Math.log10(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(x) * Math.LOG2E',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.LOG2E * Math.log(x)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.log(x) / Math.LN2',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.log2(…)`.',
              output: 'Math.log2(x)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a * a + b * b)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(a ** 2 + b ** 2)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.hypot(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.hypot(…)`.',
              output: 'Math.hypot(a, b)',
            },
          ],
        },
      ],
    },
    {
      code: 'Math.sqrt(x ** 2)',
      filename: 'input.js',
      output: null,
      errors: [
        {
          messageId: 'prefer-modern-math-apis',
          message: 'Prefer `Math.abs(…)` over `Math.sqrt(…)`.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace with `Math.abs(…)`.',
              output: 'Math.abs(x)',
            },
          ],
        },
      ],
    },
  ],
});
