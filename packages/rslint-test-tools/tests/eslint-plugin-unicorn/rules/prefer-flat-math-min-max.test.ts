// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const filename = 'src/virtual.js';

const valid = [
  'Math.max(a, b, c);',
  'Math.min(a, b, c);',
  'Math.max(Math.min(a, b), c);',
  'Math.min(Math.max(a, b), c);',
  'max(max(a, b), c);',
  'Math.max.apply(Math, [Math.max(a, b), c]);',
  'Math["max"](Math["max"](a, b), c);',
  'Math.max?.(Math.max(a, b), c);',
  'Math?.max(Math.max(a, b), c);',
  'Math.max(Math.max?.(a, b), c);',
  'Math.max(Math?.max(a, b), c);',
  'globalThis.Math.max(Math.max(a, b), c);',
  'Number.max(Number.max(a, b), c);',
  'Math.hypot(Math.hypot(a, b), c);',
].map((code) => ({ code, filename }));

const invalid = [
  ['Math.max(Math.max(a, b), c);', 'max'],
  ['Math.min(a, Math.min(b, c));', 'min'],
  ['Math.max(Math.max(a, b), Math.max(c, d), e);', 'max'],
  ['Math.min(Math.min(Math.min(a, b), c), d);', 'min'],
  ['Math.max(Math.max(a, b));', 'max'],
  ['Math.min(Math.min());', 'min'],
  ['Math.max(Math.max(...values), fallback);', 'max'],
  ['const value = Math.max(foo, Math.max(bar, baz)).toString();', 'max'],
  ['const value = Math.min((Math.min(a, b)), c);', 'min'],
  [
    'const value = Math.max(\n\tMath.max(a, b),\n\tc,\n);',
    'max',
  ],
  [
    'const value = Math.max(\n\tMath.max(\n\t\ta,\n\t\tb,\n\t),\n\tc,\n);',
    'max',
  ],
  [
    'const value = Math.max(\n\tMath.max(/* keep */ a, b),\n\tc,\n);',
    'max',
  ],
].map(([code, method]) => ({
  code,
  filename,
  errors: [
    {
      messageId: 'prefer-flat-math-min-max',
      message: `Prefer a flat \`Math.${method}()\` call instead of nested calls.`,
    },
  ],
}));

ruleTester.run('prefer-flat-math-min-max', null as never, {
  valid,
  invalid,
});
