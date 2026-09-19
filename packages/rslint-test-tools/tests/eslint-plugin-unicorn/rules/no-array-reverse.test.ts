// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import { RuleTester } from '../rule-tester';

interface UpstreamCase {
  code: string;
  filename?: string;
  options?: unknown[];
}
function normalizeCase(value: string | UpstreamCase): UpstreamCase {
  return typeof value === 'string'
    ? { code: value, filename: 'src/virtual.js' }
    : { filename: 'src/virtual.js', ...value };
}

const valid: (string | UpstreamCase)[] = [
  {
    code: 'function f(foo: Set<number>) { foo.reverse(); }',
    filename: 'src/virtual.ts',
  },
  'reversed =[...array].toReversed()',
  'reversed =array.toReversed()',
  'reversed =[...array].reverse',
  'reversed =[...array].reverse?.()',
  'array.reverse()',
  'array.reverse?.()',
  'array?.reverse()',
  'if (true) array.reverse()',
  'reversed = array.reverse(extraArgument)',
  'const reversed = [...array].toReversed();',
];

const invalid0: (string | UpstreamCase)[] = [
  'reversed = [...array].reverse()',
  'reversed = [...array]?.reverse()',
  'reversed = array.reverse()',
  'reversed = array?.reverse()',
  { code: 'array.reverse()', options: [{ allowExpressionStatement: false }] },
  { code: 'array?.reverse()', options: [{ allowExpressionStatement: false }] },
  {
    code: '[...array].reverse()',
    options: [{ allowExpressionStatement: false }],
  },
  'reversed = [...(0, array)].reverse()',
  'reversed = [...a + b].reverse()',
  'reversed = [...a ? b : c].reverse()',
  'reversed = [...(a + b)].reverse()',
  'reversed = [...new Set(array)].reverse()',
  'const reversed = [...array].reverse();',
  { code: 'array.reverse();', options: [{ allowExpressionStatement: false }] },
];

new RuleTester().run('no-array-reverse', {} as never, {
  valid: valid.map(normalizeCase),
  invalid: [
    ...invalid0.map((value) => ({
      ...normalizeCase(value),
      errors: [
        {
          messageId: 'error',
          message: 'Use `Array#toReversed()` instead of `Array#reverse()`.',
        },
      ],
    })),
  ],
});
