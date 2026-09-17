// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import {
  RuleTester,
  type ValidTestCase,
  type InvalidTestCase,
} from '../rule-tester';

new RuleTester().run('prefer-global-number-constants', {} as never, {
  valid: [
    ...[
      'const foo = NaN;',
      'const foo = Infinity;',
      'const foo = -Infinity;',
      'const foo = Number.MAX_SAFE_INTEGER;',
      'const foo = object.Number.NaN;',
      'Number.NaN = 1;',
      'Number.POSITIVE_INFINITY ||= 1;',
      '[Number.NEGATIVE_INFINITY] = [];',
      'const Number = {\n\tNaN: 1,\n\tPOSITIVE_INFINITY: 2,\n\tNEGATIVE_INFINITY: -2,\n};\nconst foo = Number.NaN + Number.POSITIVE_INFINITY + Number.NEGATIVE_INFINITY;',
      'function foo() {\n\tconst NaN = 1;\n\tconst value = Number.NaN;\n}',
      'function foo() {\n\tconst Infinity = 1;\n\tconst positive = Number.POSITIVE_INFINITY;\n\tconst negative = Number.NEGATIVE_INFINITY;\n}',
      '// ✅\nconst foo = NaN;\n',
      '// ✅\nconst foo = Infinity;\n',
      '// ✅\nconst foo = -Infinity;\n',
    ].map((code): ValidTestCase => ({
      code,
      ...{
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...[
      'const NaN = 1; const value = Number.NaN;',
      'const Infinity = 1; const positive = Number.POSITIVE_INFINITY; const negative = Number.NEGATIVE_INFINITY;',
    ].map((code): ValidTestCase => ({
      code,
      ...{
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'script',
        },
      },
    })),
  ],
  invalid: [
    ...[
      'const foo = Number.NaN;',
      'const foo = window.Number.NaN;',
      'const foo = Number["NaN"];',
      'const foo = {value: Number.NaN};',
      'const foo = Number /* comment */ .NaN;',
      '// ❌\nconst foo = Number.NaN;\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Prefer `NaN` over `Number.NaN`.',
            messageId: 'prefer-global-number-constants',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...[
      'const foo = Number.POSITIVE_INFINITY;',
      '// ❌\nconst foo = Number.POSITIVE_INFINITY;\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Prefer `Infinity` over `Number.POSITIVE_INFINITY`.',
            messageId: 'prefer-global-number-constants',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...[
      'const foo = Number.NEGATIVE_INFINITY;',
      'const foo = Number.NEGATIVE_INFINITY.toString();',
      '// ❌\nconst foo = Number.NEGATIVE_INFINITY;\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Prefer `-Infinity` over `Number.NEGATIVE_INFINITY`.',
            messageId: 'prefer-global-number-constants',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...[
      'const foo = {[Number.POSITIVE_INFINITY]: Number.NEGATIVE_INFINITY};',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Prefer `Infinity` over `Number.POSITIVE_INFINITY`.',
            messageId: 'prefer-global-number-constants',
          },
          {
            message: 'Prefer `-Infinity` over `Number.NEGATIVE_INFINITY`.',
            messageId: 'prefer-global-number-constants',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
  ],
});
