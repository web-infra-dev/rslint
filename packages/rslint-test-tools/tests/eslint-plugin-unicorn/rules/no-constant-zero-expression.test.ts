// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-constant-zero-expression.js
import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

const defaults = {
  filename: 'src/virtual.js',
  languageOptions: { sourceType: 'module' },
} satisfies Pick<ValidTestCase, 'filename' | 'languageOptions'>;

// The wrapper asserts messages; Go tests also assert ranges and suggestion edits.
function invalid(code: string, typescript = false): InvalidTestCase {
  return {
    ...defaults,
    filename: typescript ? 'src/virtual.ts' : defaults.filename,
    code,
    errors: [
      {
        messageId: 'no-constant-zero-expression/error',
        message:
          'This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.',
      },
    ],
  };
}

new RuleTester().run('no-constant-zero-expression', {} as never, {
  valid: [
    // Snapshot cases
    { ...defaults, code: 'x * 1;' },
    { ...defaults, code: 'x + 0;' },
    { ...defaults, code: 'x - 0;' },
    { ...defaults, code: 'x / 1;' },
    { ...defaults, code: 'x | 0;' },
    { ...defaults, code: 'x ^ 0;' },
    { ...defaults, code: 'x << 0;' },
    { ...defaults, code: 'x >> 0;' },
    { ...defaults, code: 'x / 0;' },
    { ...defaults, code: 'x % 0;' },
    { ...defaults, code: '0 % x;' },
    { ...defaults, code: 'x ** 0;' },
    { ...defaults, code: '0 ** x;' },
    { ...defaults, code: 'x - x;' },
    { ...defaults, code: 'x ^ x;' },
    { ...defaults, code: '0n & x;' },
    { ...defaults, code: 'x & 0n;' },
    { ...defaults, code: 'x * 0n;' },
    { ...defaults, code: '-0 * x;' },
    { ...defaults, code: 'x * -0;' },
    { ...defaults, code: 'x === 0;' },
    { ...defaults, code: 'x !== 0;' },
    // Documentation examples
    { ...defaults, code: 'const total = 0;' },
    { ...defaults, code: 'const flags = 0;' },
    { ...defaults, code: 'const ratio = 0;' },
  ],
  invalid: [
    // Static template cases
    invalid('Math[`PI`] * 0;'),
    invalid('`abc`.length * 0;'),
    invalid('const text = `abc`; text.length * 0;'),
    // Snapshot cases
    invalid('x * 0;'),
    invalid('0 * x;'),
    invalid('0.0 * x;'),
    invalid('x & 0;'),
    invalid('0 & x;'),
    invalid('0 / x;'),
    invalid('const y = value * 0;'),
    invalid('a + b * 0;'),
    invalid('(a + b) * 0;'),
    invalid('(x) & 0;'),
    invalid('foo.bar * 0;'),
    invalid('foo() * 0;'),
    invalid(
      "const object = {value: 1};\nObject.defineProperty(object, 'value', {get() { return 1n; }});\nobject.value * 0;",
    ),
    invalid(
      "const object = {value: 1};\nObject.defineProperty(object, 'value', {get() { return 1n; }});\nconst value = object.value;\nvalue * 0;",
    ),
    invalid('const a = 5;\na * 0;'),
    invalid('5 * 0;'),
    invalid('0 * 5;'),
    invalid('7 & 0;'),
    invalid('0 & 3;'),
    invalid('0 / 5;'),
    invalid('-5 * 0;'),
    invalid('0 / -5;'),
    invalid('0 / 0;'),
    invalid('5 /* keep */ * 0;'),
    invalid('(x as number) * 0;', true),
    invalid('x! & 0;', true),
    // Unsafe forward alias
    invalid(
      'const alias = condition; var condition = true; (alias ? 1 : NaN) * 0;',
    ),
    // Documentation examples
    invalid('const total = price * 0;'),
    invalid('const flags = mask & 0;'),
    invalid('const ratio = 0 / count;'),
  ],
});
