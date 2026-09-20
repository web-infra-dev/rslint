// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import {
  RuleTester,
  type ValidTestCase,
  type InvalidTestCase,
} from '../rule-tester';

new RuleTester().run('no-unnecessary-array-splice-count', {} as never, {
  valid: [
    ...[
      'foo.splice?.(1, foo.length)',
      'foo.splice(foo.length, 1)',
      'foo.splice()',
      'foo.splice(1)',
      'foo.splice(1, foo.length - 1)',
      'foo.splice(1, foo.length, extraArgument)',
      'foo.splice(...[1], foo.length)',
      'foo.not_splice(1, foo.length)',
      'new foo.splice(1, foo.length)',
      'splice(1, foo.length)',
      'foo.splice(1, foo.notLength)',
      'foo.splice(1, length)',
      'foo[splice](1, foo.length)',
      'foo.splice(1, foo[length])',
      'foo.splice(1, bar.length)',
      'foo?.splice(1, NotInfinity)',
      'foo?.splice(1, Number.NOT_POSITIVE_INFINITY)',
      'foo?.splice(1, Not_Number.POSITIVE_INFINITY)',
      'foo?.splice(1, Number?.POSITIVE_INFINITY)',
      'foo().splice(1, foo().length)',
      'foo.toSpliced?.(1, foo.length)',
      'foo.toSpliced(foo.length, 1)',
      'foo.toSpliced()',
      'foo.toSpliced(1)',
      'foo.toSpliced(1, foo.length - 1)',
      'foo.toSpliced(1, foo.length, extraArgument)',
      'foo.toSpliced(...[1], foo.length)',
      'foo.not_toSpliced(1, foo.length)',
      'new foo.toSpliced(1, foo.length)',
      'toSpliced(1, foo.length)',
      'foo.toSpliced(1, foo.notLength)',
      'foo.toSpliced(1, length)',
      'foo[toSpliced](1, foo.length)',
      'foo.toSpliced(1, foo[length])',
      'foo.toSpliced(1, bar.length)',
      'foo?.toSpliced(1, NotInfinity)',
      'foo?.toSpliced(1, Number.NOT_POSITIVE_INFINITY)',
      'foo?.toSpliced(1, Not_Number.POSITIVE_INFINITY)',
      'foo?.toSpliced(1, Number?.POSITIVE_INFINITY)',
      'foo().toSpliced(1, foo().length)',
      '// ❌\nconst foo = array.toSpliced(1, string.length);\n\n',
      '// ✅\nconst foo = array.toSpliced(1);\n',
      '// ✅\nconst foo = array.toSpliced(1);\n',
      '// ✅\nconst foo = array.toSpliced(1);\n',
      '// ❌\narray.splice(1, string.length);\n\n',
      '// ✅\narray.splice(1);\n',
      '// ✅\narray.splice(1);\n',
      '// ✅\narray.splice(1);\n',
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
      '(foo as any[]).splice(1, bar.length)',
      'foo!.splice(1, bar!.length)',
      '(foo as any[]).toSpliced(1, bar.length)',
      'foo!.toSpliced(1, bar!.length)',
      'function f(foo: {splice(start: number, deleteCount: number): void; length: number}) { foo.splice(1, foo.length); }',
      'function f(foo: Set<number>) { foo.splice(1, Infinity); }',
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
      'foo.splice(1, foo.length)',
      'foo?.splice(1, foo.length)',
      'foo.splice(1, foo.length,)',
      'foo.splice(1, (( foo.length )))',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `foo.length` as the `deleteCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['foo.splice(1, foo?.length)', 'foo?.splice(1, foo?.length)'].map(
      (code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message:
                'Passing `foo?.length` as the `deleteCount` argument is unnecessary.',
              messageId: 'no-unnecessary-array-splice-count',
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
      'foo?.splice(1, Infinity)',
      '// ❌\narray.splice(1, Infinity);\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `Infinity` as the `deleteCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
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
      'foo?.splice(1, Number.POSITIVE_INFINITY)',
      '// ❌\narray.splice(1, Number.POSITIVE_INFINITY);\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `Number.POSITIVE_INFINITY` as the `deleteCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['foo.bar.splice(1, foo.bar.length)'].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `….length` as the `deleteCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
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
      '(foo as any[]).splice(1, (foo as any[]).length)',
      'foo!.splice(1, foo!.length)',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `….length` as the `deleteCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
          },
        ],
        filename: 'src/virtual.ts',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['(array as string[])?.splice(1, (array as string[])?.length)'].map(
      (code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message:
                'Passing `…?.length` as the `deleteCount` argument is unnecessary.',
              messageId: 'no-unnecessary-array-splice-count',
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
      '(foo as number[]).splice(1, Infinity)',
      'foo!.splice(1, Infinity)',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `Infinity` as the `deleteCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
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
      '(bar as any[]).splice(1, Number.POSITIVE_INFINITY)',
      'bar!.splice(1, Number.POSITIVE_INFINITY)',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `Number.POSITIVE_INFINITY` as the `deleteCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
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
      'foo.toSpliced(1, foo.length)',
      'foo?.toSpliced(1, foo.length)',
      'foo.toSpliced(1, foo.length,)',
      'foo.toSpliced(1, (( foo.length )))',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `foo.length` as the `skipCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['foo.toSpliced(1, foo?.length)', 'foo?.toSpliced(1, foo?.length)'].map(
      (code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message:
                'Passing `foo?.length` as the `skipCount` argument is unnecessary.',
              messageId: 'no-unnecessary-array-splice-count',
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
      'foo?.toSpliced(1, Infinity)',
      '// ❌\nconst foo = array.toSpliced(1, Infinity);\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `Infinity` as the `skipCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
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
      'foo?.toSpliced(1, Number.POSITIVE_INFINITY)',
      '// ❌\nconst foo = array.toSpliced(1, Number.POSITIVE_INFINITY);\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `Number.POSITIVE_INFINITY` as the `skipCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['foo.bar.toSpliced(1, foo.bar.length)'].map(
      (code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message:
                'Passing `….length` as the `skipCount` argument is unnecessary.',
              messageId: 'no-unnecessary-array-splice-count',
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
      '(foo as any[]).toSpliced(1, (foo as any[]).length)',
      'foo!.toSpliced(1, foo!.length)',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `….length` as the `skipCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
          },
        ],
        filename: 'src/virtual.ts',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['(array as string[])?.toSpliced(1, (array as string[])?.length)'].map(
      (code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message:
                'Passing `…?.length` as the `skipCount` argument is unnecessary.',
              messageId: 'no-unnecessary-array-splice-count',
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
      '(foo as number[]).toSpliced(1, Infinity)',
      'foo!.toSpliced(1, Infinity)',
      'function f(foo: Uint8Array) { foo.toSpliced(1, Infinity); }',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `Infinity` as the `skipCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
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
      '(bar as any[]).toSpliced(1, Number.POSITIVE_INFINITY)',
      'bar!.toSpliced(1, Number.POSITIVE_INFINITY)',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `Number.POSITIVE_INFINITY` as the `skipCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
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
      'function f(foo: number[]) { foo.splice(1, foo.length); }',
      'function f(foo: Uint8Array) { foo.splice(1, foo.length); }',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `foo.length` as the `deleteCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
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
      'function f(foo: readonly number[]) { foo.toSpliced(1, foo.length); }',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message:
              'Passing `foo.length` as the `skipCount` argument is unnecessary.',
            messageId: 'no-unnecessary-array-splice-count',
          },
        ],
        filename: 'src/virtual.ts',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
  ],
});
