// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import {
  RuleTester,
  type ValidTestCase,
  type InvalidTestCase,
} from '../rule-tester';

new RuleTester().run(
  'prefer-identifier-import-export-specifiers',
  {} as never,
  {
    valid: [
      ...[
        'import "foo";',
        'import foo from "foo";',
        'import * as foo from "foo";',
        'import {foo} from "foo";',
        'import {foo as bar} from "foo";',
        'import {"a string" as aString} from "foo";',
        'import {"foo-bar" as fooBar} from "foo";',
        'import {"" as empty} from "foo";',
        'const foo = 1;\nexport {foo};',
        'const foo = 1;\nexport {foo as bar};',
        'export {foo} from "foo";',
        'export {foo as bar} from "foo";',
        'const foo = 1;\nexport {foo as "a string"};',
        'const foo = 1;\nexport {foo as "foo-bar"};',
        'export {"a string" as aString} from "foo";',
        'export {"foo-bar" as fooBar} from "foo";',
        'export {"" as empty} from "foo";',
        'export * from "foo";',
        'export * as foo from "foo";',
        'export * as "a string" from "foo";',
        'import foo from "foo" with {type: "json"};',
        'import foo from "foo" with {"foo-bar": "json"};',
        'import foo from "foo" with {"": "json"};',
        'import foo from "foo" with {"0": "json"};',
        'export {foo} from "foo" with {"a string": "x"};',
        "// ✅\nimport {foo as foo} from 'foo';\n",
        'const foo = 1;\n// ✅\nexport {foo as bar};\n',
        "// ✅\nexport {foo as bar} from 'foo';\n",
        "// ✅\nimport foo from 'foo' with {type: 'json'};\n",
        "// ✅\nimport {'a string' as aString} from 'foo';\n",
      ].map((code): ValidTestCase => ({
        code,
        ...{
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
      })),
    ],
    invalid: [
      ...[
        'import {"foo" as foo} from "foo";',
        'import {"foo" as foo} from "foo" with {type: "json"};',
        'import {"foo"as foo} from "foo";',
        'export {"foo" as bar} from "foo";',
        'export {"foo"} from "foo";',
        'export {"foo"as bar} from "foo";',
        'export * as "foo" from "foo";',
        'export * as"foo" from "foo";',
      ].map((code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message: 'Prefer identifier `foo` over string literal `"foo"`.',
              messageId: 'prefer-identifier-import-export-specifiers',
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
      })),
      ...[
        'import {"default" as defaultExport} from "foo";',
        'const foo = 1;\nexport {foo as "default"};',
        'export {"default" as defaultExport} from "foo";',
        'import foo from "foo" with {"default": "json"};',
      ].map((code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message:
                'Prefer identifier `default` over string literal `"default"`.',
              messageId: 'prefer-identifier-import-export-specifiers',
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
      })),
      ...[
        'import {"foo" as foo, "bar" as bar} from "foo";',
        'export {"foo" as "bar"} from "foo";',
        'export {"foo"as"bar"} from "foo";',
      ].map((code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message: 'Prefer identifier `foo` over string literal `"foo"`.',
              messageId: 'prefer-identifier-import-export-specifiers',
            },
            {
              message: 'Prefer identifier `bar` over string literal `"bar"`.',
              messageId: 'prefer-identifier-import-export-specifiers',
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
      })),
      ...['import {"\\u0066oo" as foo} from "foo";'].map(
        (code): InvalidTestCase => ({
          code,
          ...{
            errors: [
              {
                message:
                  'Prefer identifier `foo` over string literal `"\\u0066oo"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
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
        'const foo = 1;\nexport {foo as "bar"};',
        'const foo = 1;\nexport {foo as"bar"};',
      ].map((code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message: 'Prefer identifier `bar` over string literal `"bar"`.',
              messageId: 'prefer-identifier-import-export-specifiers',
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
      })),
      ...['const foo = 1, baz = 2;\nexport {foo as "bar", baz as "qux"};'].map(
        (code): InvalidTestCase => ({
          code,
          ...{
            errors: [
              {
                message: 'Prefer identifier `bar` over string literal `"bar"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
              },
              {
                message: 'Prefer identifier `qux` over string literal `"qux"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
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
      ...['const foo = 1;\nexport {foo as "\\u0062ar"};'].map(
        (code): InvalidTestCase => ({
          code,
          ...{
            errors: [
              {
                message:
                  'Prefer identifier `bar` over string literal `"\\u0062ar"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
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
      ...['export {"foo" as bar, "baz" as qux} from "foo";'].map(
        (code): InvalidTestCase => ({
          code,
          ...{
            errors: [
              {
                message: 'Prefer identifier `foo` over string literal `"foo"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
              },
              {
                message: 'Prefer identifier `baz` over string literal `"baz"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
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
        'import type {"foo" as Foo} from "foo";',
        'export type {"foo" as Foo} from "foo";',
      ].map((code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message: 'Prefer identifier `foo` over string literal `"foo"`.',
              messageId: 'prefer-identifier-import-export-specifiers',
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
      })),
      ...[
        'import foo from "foo" with {"type": "json"};',
        'export {foo} from "foo" with {"type": "json"};',
        'import foo from "foo" with{"type":"json"};',
      ].map((code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message: 'Prefer identifier `type` over string literal `"type"`.',
              messageId: 'prefer-identifier-import-export-specifiers',
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
      })),
      ...['import foo from "foo" with {"type": "json", "other": "x"};'].map(
        (code): InvalidTestCase => ({
          code,
          ...{
            errors: [
              {
                message:
                  'Prefer identifier `type` over string literal `"type"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
              },
              {
                message:
                  'Prefer identifier `other` over string literal `"other"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
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
      ...['import foo from "foo" with {"type": "json"};'].map(
        (code): InvalidTestCase => ({
          code,
          ...{
            errors: [
              {
                message:
                  'Prefer identifier `type` over string literal `"type"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
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
      ...['import {"if" as foo} from "foo";'].map((code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message: 'Prefer identifier `if` over string literal `"if"`.',
              messageId: 'prefer-identifier-import-export-specifiers',
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
      })),
      ...['import {"yield" as foo} from "foo";'].map(
        (code): InvalidTestCase => ({
          code,
          ...{
            errors: [
              {
                message:
                  'Prefer identifier `yield` over string literal `"yield"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
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
      ...['export {"foo" as "foo"} from "foo";'].map(
        (code): InvalidTestCase => ({
          code,
          ...{
            errors: [
              {
                message: 'Prefer identifier `foo` over string literal `"foo"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
              },
              {
                message: 'Prefer identifier `foo` over string literal `"foo"`.',
                messageId: 'prefer-identifier-import-export-specifiers',
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
        "// ❌\nimport {'foo' as foo} from 'foo';\n\n",
        "// ❌\nexport {'foo' as bar} from 'foo';\n\n",
      ].map((code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message: "Prefer identifier `foo` over string literal `'foo'`.",
              messageId: 'prefer-identifier-import-export-specifiers',
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
      })),
      ...["const foo = 1;\n// ❌\nexport {foo as 'bar'};\n\n"].map(
        (code): InvalidTestCase => ({
          code,
          ...{
            errors: [
              {
                message: "Prefer identifier `bar` over string literal `'bar'`.",
                messageId: 'prefer-identifier-import-export-specifiers',
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
      ...["// ❌\nimport foo from 'foo' with {'type': 'json'};\n\n"].map(
        (code): InvalidTestCase => ({
          code,
          ...{
            errors: [
              {
                message:
                  "Prefer identifier `type` over string literal `'type'`.",
                messageId: 'prefer-identifier-import-export-specifiers',
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
  },
);
