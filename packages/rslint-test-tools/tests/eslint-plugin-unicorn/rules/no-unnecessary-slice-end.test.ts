// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import {
  RuleTester,
  type ValidTestCase,
  type InvalidTestCase,
} from '../rule-tester';

new RuleTester().run('no-unnecessary-slice-end', {} as never, {
  valid: [
    ...[
      'foo.slice?.(1, foo.length)',
      'foo.slice(foo.length, 1)',
      'foo.slice()',
      'foo.slice(1)',
      'foo.slice(1, foo.length - 1)',
      'foo.slice(1, foo.length, extraArgument)',
      'foo.slice(...[1], foo.length)',
      'foo.not_slice(1, foo.length)',
      'new foo.slice(1, foo.length)',
      'slice(1, foo.length)',
      'foo.slice(1, foo.notLength)',
      'foo.slice(1, length)',
      'foo[slice](1, foo.length)',
      'foo.slice(1, foo[length])',
      'foo.slice(1, bar.length)',
      'foo?.slice(1, NotInfinity)',
      'foo?.slice(1, Number.NOT_POSITIVE_INFINITY)',
      'foo?.slice(1, Not_Number.POSITIVE_INFINITY)',
      'foo?.slice(1, Number?.POSITIVE_INFINITY)',
      'foo().slice(1, foo().length)',
      '// ✅\nconst foo = string.slice(1);\n',
      '// ✅\nconst foo = string.slice(1);\n',
      '// ✅\nconst foo = string.slice(1);\n',
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
      '(foo as any[]).slice(1, bar.length)',
      'foo!.slice(1, bar!.length)',
    ].map((code): ValidTestCase => ({
      code,
      ...{
        filename: 'src/virtual.ts',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
  ],
  invalid: [
    ...[
      'foo.slice(1, foo.length)',
      'foo?.slice(1, foo.length)',
      'foo.slice(1, foo.length,)',
      'foo.slice(1, (( foo.length )))',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `foo.length` as the `end` argument is unnecessary.',
            messageId: 'no-unnecessary-slice-end',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['foo.slice(1, foo?.length)', 'foo?.slice(1, foo?.length)'].map(
      (code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message:
                'Passing `foo?.length` as the `end` argument is unnecessary.',
              messageId: 'no-unnecessary-slice-end',
            },
          ],
          filename: 'src/virtual.js',
          languageOptions: {
            globals: {
              global: 'readonly',
              self: 'readonly',
              window: 'readonly',
            },
            sourceType: 'module',
          },
        },
      }),
    ),
    ...[
      'foo?.slice(1, Infinity)',
      '// ❌\nconst foo = string.slice(1, Infinity);\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Passing `Infinity` as the `end` argument is unnecessary.',
            messageId: 'no-unnecessary-slice-end',
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
      'foo?.slice(1, Number.POSITIVE_INFINITY)',
      '// ❌\nconst foo = string.slice(1, Number.POSITIVE_INFINITY);\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `Number.POSITIVE_INFINITY` as the `end` argument is unnecessary.',
            messageId: 'no-unnecessary-slice-end',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['foo.bar.slice(1, foo.bar.length)'].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Passing `….length` as the `end` argument is unnecessary.',
            messageId: 'no-unnecessary-slice-end',
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
      '(foo as any[]).slice(1, (foo as any[]).length)',
      'foo!.slice(1, foo!.length)',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Passing `….length` as the `end` argument is unnecessary.',
            messageId: 'no-unnecessary-slice-end',
          },
        ],
        filename: 'src/virtual.ts',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['(array as string[])?.slice(1, (array as string[])?.length)'].map(
      (code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message:
                'Passing `…?.length` as the `end` argument is unnecessary.',
              messageId: 'no-unnecessary-slice-end',
            },
          ],
          filename: 'src/virtual.ts',
          languageOptions: {
            globals: {
              global: 'readonly',
              self: 'readonly',
              window: 'readonly',
            },
            sourceType: 'module',
          },
        },
      }),
    ),
    ...['(foo as number[]).slice(1, Infinity)', 'foo!.slice(1, Infinity)'].map(
      (code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message:
                'Passing `Infinity` as the `end` argument is unnecessary.',
              messageId: 'no-unnecessary-slice-end',
            },
          ],
          filename: 'src/virtual.ts',
          languageOptions: {
            globals: {
              global: 'readonly',
              self: 'readonly',
              window: 'readonly',
            },
            sourceType: 'module',
          },
        },
      }),
    ),
    ...[
      '(bar as any[]).slice(1, Number.POSITIVE_INFINITY)',
      'bar!.slice(1, Number.POSITIVE_INFINITY)',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `Number.POSITIVE_INFINITY` as the `end` argument is unnecessary.',
            messageId: 'no-unnecessary-slice-end',
          },
        ],
        filename: 'src/virtual.ts',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...[
      'function f(string: string) { return string.slice(1, string.length); }',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `string.length` as the `end` argument is unnecessary.',
            messageId: 'no-unnecessary-slice-end',
          },
        ],
        filename: 'src/virtual.ts',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...[
      'function f(bytes: Uint8Array) { return bytes.slice(1, bytes.length); }',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `bytes.length` as the `end` argument is unnecessary.',
            messageId: 'no-unnecessary-slice-end',
          },
        ],
        filename: 'src/virtual.ts',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['// ❌\nconst foo = string.slice(1, string.length);\n\n'].map(
      (code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message:
                'Passing `string.length` as the `end` argument is unnecessary.',
              messageId: 'no-unnecessary-slice-end',
            },
          ],
          filename: 'src/virtual.js',
          languageOptions: {
            globals: {
              global: 'readonly',
              self: 'readonly',
              window: 'readonly',
            },
            sourceType: 'module',
          },
        },
      }),
    ),
  ],
});
