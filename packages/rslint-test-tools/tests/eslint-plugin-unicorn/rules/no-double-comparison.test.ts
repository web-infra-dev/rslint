// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-double-comparison.js
import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

const defaults = {
  filename: 'src/virtual.js',
  languageOptions: { sourceType: 'module' },
} satisfies Pick<ValidTestCase, 'filename' | 'languageOptions'>;

// The wrapper checks messages; Go tests also assert ranges and suggestion edits.
function invalid(
  code: string,
  replacement: string,
  typescript = false,
): InvalidTestCase {
  return {
    ...defaults,
    filename: typescript ? 'src/virtual.ts' : defaults.filename,
    code,
    output: null,
    errors: [
      {
        messageId: 'no-double-comparison',
        message: `These comparisons can be simplified to \`${replacement}\`.`,
      },
    ],
  };
}

new RuleTester().run('no-double-comparison', {} as never, {
  valid: [
    // Snapshot cases
    { ...defaults, code: 'x == y || x < y' },
    { ...defaults, code: 'x == y || x > y' },
    { ...defaults, code: 'x != y && x <= y' },
    { ...defaults, code: 'x <= y || x < y' },
    { ...defaults, code: 'x < y || x <= y' },
    { ...defaults, code: 'x >= y || x > y' },
    { ...defaults, code: 'x === y || x <= y' },
    { ...defaults, code: 'x === y || x < z' },
    { ...defaults, code: 'a === b || c < d' },
    { ...defaults, code: 'x < y || x < y' },
    { ...defaults, code: 'x === y || x === y' },
    { ...defaults, code: 'x <= y || x >= y' },
    { ...defaults, code: 'x < y || y > x' },
    { ...defaults, code: 'x > y || y < x' },
    { ...defaults, code: 'x <= y && y >= x' },
    { ...defaults, code: 'f() === y || f() < y' },
    { ...defaults, code: 'x === f() || x < f()' },
    { ...defaults, code: 'x < y && x > y' },
    { ...defaults, code: 'x <= y || x > y' },
    { ...defaults, code: 'x === y || x !== y' },
    { ...defaults, code: 'x === y && x !== y' },
    { ...defaults, code: 'x === y || z || x < y' },
    { ...defaults, code: 'x?.a === y || x?.a < y' },
    { ...defaults, code: 'x?.a <= y && x?.a >= y' },
    { ...defaults, code: 'x === y || foo(x, y)' },
    { ...defaults, code: 'x === y || x' },
    { ...defaults, code: 'x === y ?? x < y' },
    { ...defaults, code: 'const y = /foo/; /foo/ === y || /foo/ < y' },
    { ...defaults, code: 'const y = /foo/; /foo/ < y || /foo/ === y' },
    { ...defaults, code: 'x === /foo/ || x < /foo/' },
    { ...defaults, code: 'const a = /foo/, b = /foo/; a === b || a < b' },
    { ...defaults, code: 'const a = /foo/; a === b || a < b' },
    { ...defaults, code: 'const a = /foo/g; a === b || a < b' },
    // Documentation examples
    { ...defaults, code: 'x <= y;' },
    { ...defaults, code: 'x >= y;' },
    { ...defaults, code: 'x === y;' },
    { ...defaults, code: 'x < y;' },
    { ...defaults, code: 'x > y;' },
    { ...defaults, code: 'x !== y;' },
    { ...defaults, code: 'x <= y || x < y;' },
    { ...defaults, code: 'x < y || y > x;' },
  ],
  invalid: [
    // Explicit TypeScript cases
    invalid(
      'x === y as number || x < (y as number)',
      'x <= (y as number)',
      true,
    ),
    invalid(
      'x === y as number || x > (y as number)',
      'x >= (y as number)',
      true,
    ),
    invalid('x as number === y || (x as number) < y', 'x as number <= y', true),
    invalid(
      'x === (y /* keep */ as number) || x < (y as number)',
      'x <= (y /* keep */ as number)',
      true,
    ),
    invalid(
      'foo()\nx as number === y || (x as number) < y',
      'x as number <= y',
      true,
    ),
    invalid(
      'if (ready)\nx as number === y || (x as number) < y',
      'x as number <= y',
      true,
    ),
    invalid(
      'x === y satisfies number || x < (y satisfies number)',
      'x <= (y satisfies number)',
      true,
    ),
    invalid(
      'x === y satisfies number || x > (y satisfies number)',
      'x >= (y satisfies number)',
      true,
    ),
    invalid(
      'x satisfies number === y || (x satisfies number) < y',
      'x satisfies number <= y',
      true,
    ),
    invalid(
      'x === (y /* keep */ satisfies number) || x < (y satisfies number)',
      'x <= (y /* keep */ satisfies number)',
      true,
    ),
    invalid(
      'foo()\nx satisfies number === y || (x satisfies number) < y',
      'x satisfies number <= y',
      true,
    ),
    invalid(
      'if (ready)\nx satisfies number === y || (x satisfies number) < y',
      'x satisfies number <= y',
      true,
    ),
    // Snapshot cases
    invalid('x === y || x < y', 'x <= y'),
    invalid('x < y || x === y', 'x <= y'),
    invalid('x === y || x > y', 'x >= y'),
    invalid('x > y || x === y', 'x >= y'),
    invalid('x <= y && x >= y', 'x === y'),
    invalid('x >= y && x <= y', 'x === y'),
    invalid('x <= y && x !== y', 'x < y'),
    invalid('x !== y && x <= y', 'x < y'),
    invalid('x >= y && x !== y', 'x > y'),
    invalid('x !== y && x >= y', 'x > y'),
    invalid('x < y || x > y', 'x !== y'),
    invalid('x > y || x < y', 'x !== y'),
    invalid(
      'undefined === undefined || undefined < undefined',
      'undefined <= undefined',
    ),
    invalid(
      'undefined <= undefined && undefined >= undefined',
      'undefined === undefined',
    ),
    invalid(
      'const symbol = Symbol(); symbol === symbol || symbol < symbol',
      'symbol <= symbol',
    ),
    invalid(
      'const symbol = Symbol(); symbol <= symbol && symbol >= symbol',
      'symbol === symbol',
    ),
    invalid('null <= 0 && null !== 0', 'null < 0'),
    invalid('null >= 0 && null !== 0', 'null > 0'),
    invalid('x === y || y < x', 'x >= y'),
    invalid('y < x || x === y', 'y <= x'),
    invalid('x <= y && y <= x', 'x === y'),
    invalid('a.b === c || c > a.b', 'a.b <= c'),
    invalid('x === 5 || x < 5', 'x <= 5'),
    invalid('a.b === c || a.b < c', 'a.b <= c'),
    invalid('this.x === y || this.x > y', 'this.x >= y'),
    invalid('a[b] === c || a[b] > c', 'a[b] >= c'),
    invalid('x === y || x < y || z', 'x <= y'),
    invalid('x < y || x === y || x > y', 'x <= y'),
    invalid('(x === y || x < y).foo', 'x <= y'),
    invalid('if (x === y || x < y) {}', 'x <= y'),
    invalid('const result = x >= y && x !== y;', 'x > y'),
    invalid('x === y || /* keep */ x < y', 'x <= y'),
    invalid('x === y || x /* keep */ < y', 'x <= y'),
    invalid('(a /* keep */) === c || a < c', '(a /* keep */) <= c'),
    invalid('(x === y) || (x < y)', 'x <= y'),
    invalid(
      '(x as number) === y || (x as number) < y',
      '(x as number) <= y',
      true,
    ),
    invalid('x! === y || x! < y', 'x! <= y', true),
    invalid(
      '(x satisfies number) === y || (x satisfies number) < y',
      '(x satisfies number) <= y',
      true,
    ),
    // Documentation examples
    invalid('x === y || x < y;', 'x <= y'),
    invalid('x === y || x > y;', 'x >= y'),
    invalid('x <= y && x >= y;', 'x === y'),
    invalid('x <= y && x !== y;', 'x < y'),
    invalid('x === y || y < x;', 'x >= y'),
    invalid(
      'undefined === undefined || undefined < undefined',
      'undefined <= undefined',
    ),
    invalid("const x = '5'; x === 5 || x < 5", 'x <= 5'),
    invalid('x < y || x === y || x > y', 'x <= y'),
  ],
});
