// Ported from eslint-plugin-unicorn v76.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/explicit-length-check.js
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Upstream cases; diagnostic and edit expectations also run in the Go suite.
ruleTester.run('explicit-length-check', {} as never, {
  valid: [
    // Upstream valid #1
    {
      code: 'if (foo.notLength) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #2
    {
      code: 'if (length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #3
    {
      code: 'if (foo[length]) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #4
    {
      code: 'if (foo["length"]) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #5
    {
      code: 'foo.length === 0',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #6
    {
      code: 'foo.length > 0',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #7
    {
      code: 'const bar = foo.length',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #8
    {
      code: 'const bar = +foo.length',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #9
    {
      code: 'const x = Boolean(foo.length, foo.length)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #10
    {
      code: 'const x = new Boolean(foo.length)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #11
    {
      code: 'const x = NotBoolean(foo.length)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #12
    {
      code: 'const Boolean = value => value; const isNotEmpty = Boolean(foo.length)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #13
    {
      code: 'function unicorn(Boolean) { if (Boolean(foo.length)) {} }',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #14
    {
      code: 'const length = foo.length ?? 0',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #15
    {
      code: 'if (foo.length ?? bar) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #16
    {
      code: 'if (foo.length > 0) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #17
    {
      code: 'if (foo.length > 0) {}',
      options: [{ 'non-zero': 'greater-than' }],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #18
    {
      code: 'if (foo.length !== 0) {}',
      options: [{ 'non-zero': 'not-equal' }],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #19
    {
      code: 'const object: {size: number | null} = {size: 123}; if (object.size && object.size > 0) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #20
    {
      code: 'if (foo.length!) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #21
    {
      code: 'const object: {length: number | undefined} = {length: 123}; if (object.length && object.length > 0) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #22
    {
      code: 'const object: {size: number | undefined} = {size: 123}; if (object.size && object.size !== 0) {}',
      options: [{ 'non-zero': 'not-equal' }],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #23
    {
      code: 'const object: {size: number | undefined} = {size: 123}; if (object.size && object.size! > 0) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #24
    {
      code: 'const object: {size: number | undefined} = {size: 123}; if (object.size && (object.size as number) > 0) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #25
    {
      code: 'const object: {size: number | undefined} = {size: 123}; if (object.size && (<number>object.size) > 0) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #26
    {
      code: 'const object = {size: 123}; if (object.size && (object.size satisfies number) > 0) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #27
    {
      code: 'const object: {size: number | undefined} = {size: 123}; if (object.size! >= 1) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #28
    {
      code: 'const object: {size: number | undefined} = {size: 123}; if ((object.size as number) >= 1) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #29
    {
      code: 'const object: {size: number | undefined} = {size: 123}; if ((<number>object.size) >= 1) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #30
    {
      code: 'const object = {size: 123}; if ((object.size satisfies number) >= 1) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #31
    {
      code: 'const object: {size: number | undefined} = {size: 123}; if (object.size && object.size! >= 1) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #32
    {
      code: 'if (foo.length === 0) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #33
    {
      code: 'const bar = foo.length === 0 ? 1 : 2',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #34
    {
      code: 'while (foo.length > 0) {\n\tfoo.pop();\n}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #35
    {
      code: 'do {\n\tfoo.pop();\n} while (foo.length > 0);',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #36
    {
      code: 'for (; foo.length > 0; foo.pop());',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #37
    {
      code: 'if (foo.length !== 1) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #38
    {
      code: 'if (foo.length > 1) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #39
    {
      code: 'if (foo.length < 2) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #40
    {
      code: 'const foo = { size: "small" }; if (foo.size) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #41
    {
      code: 'const foo = { length: -1 }; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #42
    {
      code: 'const foo = { length: 1.5 }; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #43
    {
      code: 'const foo = { length: NaN }; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #44
    {
      code: 'const foo = { length: Infinity }; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #45
    {
      code: 'const x = foo.length || 2',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #46
    {
      code: 'const A_NUMBER = 2; const x = foo.length || A_NUMBER',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #47
    {
      code: 'const x = foo.length || "bar"',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #48
    {
      code: 'const x = foo.length || `bar`',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #49
    {
      code: 'const A_STRING = "bar"; const x = foo.length || A_STRING',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #50
    {
      code: 'const size = props.size || "mini"',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #51
    {
      code: 'const x = foo.length || unknown',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #52
    {
      code: 'something(options.length || 500)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #53
    {
      code: 'const itemCount = result.totalCount || result.length',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #54
    {
      code: 'if (\n\tdimensions.width &&\n\tdimensions.height &&\n\tdimensions.length\n) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #55
    {
      code: 'if (packagingData.dimensions.width && packagingData.dimensions.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #56
    {
      code: 'if (dimensions.width && dimensions.size) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #57
    {
      code: 'if (dimensions.height && dimensions.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Upstream valid #58
    {
      code: 'if (dimensions.depth && dimensions.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
  ],
  invalid: [
    // Upstream invalid #1
    {
      code: 'if (!foo.length > 0) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: null,
    },
    // Upstream invalid #2
    {
      code: 'if (!foo.length === 0) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: null,
    },
    // Upstream invalid #3
    {
      code: '() => foo.length && bar()',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace `.length` with `.length > 0`.',
              output: '() => foo.length > 0 && bar()',
            },
          ],
        },
      ],
      output: null,
    },
    // Upstream invalid #4
    {
      code: 'alert(foo.length && bar())',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace `.length` with `.length > 0`.',
              output: 'alert(foo.length > 0 && bar())',
            },
          ],
        },
      ],
      output: null,
    },
    // Upstream invalid #5
    {
      code: 'if (items.length && items.every(Boolean)) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 17,
          suggestions: [],
        },
      ],
      output: 'if (items.length > 0 && items.every(Boolean)) {}',
    },
    // Upstream invalid #6
    {
      code: 'if (items.map && items.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
      output: 'if (items.map && items.length > 0) {}',
    },
    // Upstream invalid #7
    {
      code: 'if (text.trim && text.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 29,
          suggestions: [],
        },
      ],
      output: 'if (text.trim && text.length > 0) {}',
    },
    // Upstream invalid #8
    {
      code: 'if (set.has && set.size) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 24,
          suggestions: [],
        },
      ],
      output: 'if (set.has && set.size > 0) {}',
    },
    // Upstream invalid #9
    {
      code: 'if (bytes.subarray && bytes.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
      output: 'if (bytes.subarray && bytes.length > 0) {}',
    },
    // Upstream invalid #10
    {
      code: 'if (typedArray.BYTES_PER_ELEMENT && typedArray.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 54,
          suggestions: [],
        },
      ],
      output: 'if (typedArray.BYTES_PER_ELEMENT && typedArray.length > 0) {}',
    },
    // Upstream invalid #11
    {
      code: 'if (container.width && items.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
      output: 'if (container.width && items.length > 0) {}',
    },
    // Upstream invalid #12
    {
      code: 'if (dimensions.width > 0 && dimensions.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 46,
          suggestions: [],
        },
      ],
      output: 'if (dimensions.width > 0 && dimensions.length > 0) {}',
    },
    // Upstream invalid #13
    {
      code: 'if (object.size && object.size >= 1) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
      output: 'if (object.size && object.size > 0) {}',
    },
    // Upstream invalid #14
    {
      code: 'if (0 < object.size && object.size) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'if (object.size > 0 && object.size) {}',
    },
    // Upstream invalid #15
    {
      code: 'if (object.size && object.size > 0) {}',
      options: [{ 'non-zero': 'not-equal' }],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size !== 0` when checking size is not zero.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
      output: 'if (object.size && object.size !== 0) {}',
    },
    // Upstream invalid #16
    {
      code: 'if (object.size && other.size > 0) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: 'if (object.size > 0 && other.size > 0) {}',
    },
    // Upstream invalid #17
    {
      code: 'if (object.size && !(object.size === 0)) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 40,
          suggestions: [],
        },
      ],
      output: 'if (object.size && object.size > 0) {}',
    },
    // Upstream invalid #18
    {
      code: 'if (object.size && Boolean(object.size !== 0)) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 46,
          suggestions: [],
        },
      ],
      output: 'if (object.size && object.size > 0) {}',
    },
    // Upstream invalid #19
    {
      code: 'if (object.size && Boolean(object.size > 0)) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
      output: 'if (object.size && object.size > 0) {}',
    },
    // Upstream invalid #20
    {
      code: 'if (object.size && !Boolean(object.size === 0)) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 47,
          suggestions: [],
        },
      ],
      output: 'if (object.size && object.size > 0) {}',
    },
    // Upstream invalid #21
    {
      code: 'if (object.size >= 1 && object.size !== 0) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
      output: 'if (object.size > 0 && object.size > 0) {}',
    },
    // Upstream invalid #22
    {
      code: 'if (!!object.size && object.size > 0) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
          suggestions: [],
        },
      ],
      output: 'if (object.size > 0 && object.size > 0) {}',
    },
    // Upstream invalid #23
    {
      code: 'if (!object.size && object.size > 0) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.size === 0` when checking size is zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 17,
          suggestions: [],
        },
      ],
      output: 'if (object.size === 0 && object.size > 0) {}',
    },
  ],
});

// Snapshots cases; diagnostic and edit expectations also run in the Go suite.
ruleTester.run('explicit-length-check', {} as never, {
  valid: [
    // Snapshots valid #1
    {
      code: 'class A {\n\ta() {\n\t\tif (this.length);\n\t\twhile (!this.size || foo);\n\t}\n}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #2
    {
      code: "const foo = {length: 123}; mutate(); if (foo.length) {} function mutate() { Object.defineProperty(foo, 'length', {get() { return 'x'; }}); }",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #3
    {
      code: 'const foo = {length: -1}; function mutate() { foo.length = 123; } if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #4
    {
      code: "const foo = {length: 123}; if (foo.length) {} mutate(); function mutate() { Object.defineProperty(foo, 'length', {get() { return 'x'; }}); }",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #5
    {
      code: "const foo = {length: -1}; foo.length = 'x'; if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #6
    {
      code: "const foo = {length: 123}; foo.length = 'x'; if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #7
    {
      code: "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 'x'}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #8
    {
      code: "const foo = {length: 123}; Object.defineProperty(foo, 'length', {value: 'x'}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #9
    {
      code: "const foo = {length: -1}; const descriptor = {value: 'x'}; Object.defineProperty(foo, 'length', descriptor); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #10
    {
      code: "const foo = {length: -1}; Object.assign(foo, {length: 'x'}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #11
    {
      code: "const foo = {length: -1}; const values = {length: 'x'}; Object.assign(foo, values); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #12
    {
      code: "const foo = {length: -1}; const values = {length: 'x'}; Object.assign(foo, {length: 123}, values); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #13
    {
      code: "const foo = {length: -1}; Object.assign(foo, {length: 123, length: 'x'}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #14
    {
      code: "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 123, value: 'x'}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #15
    {
      code: "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 123, ...descriptor}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #16
    {
      code: "const foo = {length: -1}; Object.defineProperties(foo, {length: {value: 123}, length: {value: 'x'}}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #17
    {
      code: "const foo = {length: -1}; ({length: foo.length} = {length: 123, length: 'x'}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #18
    {
      code: 'const foo = {length: -1}; ({length: foo.length = 123} = {length: 456}); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #19
    {
      code: 'const foo = {length: -1}; const {value = (foo.length = 123)} = {value: 0}; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #20
    {
      code: 'const foo = {length: -1}; try { if (condition) throw new Error(); foo.length = 123; } catch {} if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #21
    {
      code: 'const foo = {length: -1}; if (condition) foo.length = 123; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #22
    {
      code: 'const foo = {length: -1}; condition && (foo.length = 123); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #23
    {
      code: 'const foo = {length: -1}; condition ? foo.length = 123 : 0; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #24
    {
      code: 'const foo = {length: -1}; while (condition) foo.length = 123; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #25
    {
      code: 'const foo = {length: -1}; for (; condition;) foo.length = 123; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #26
    {
      code: 'const foo = {length: -1}; switch (value) { case 1: foo.length = 123; } if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #27
    {
      code: 'const foo = {length: -1}; try {} catch { foo.length = 123; } if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #28
    {
      code: 'const foo = {length: -1}; Object.assign?.(foo, {length: 123}); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #29
    {
      code: 'const foo = {length: -1}; object?.method(foo.length = 123); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #30
    {
      code: 'const foo = {length: -1}; Object.defineProperties(foo, definitions); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #31
    {
      code: 'const foo = {length: -1}; Object.defineProperties(foo, {...definitions}); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #32
    {
      code: 'const foo = {length: -1}; Object.defineProperty(foo, propertyName, {value: 123}); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #33
    {
      code: 'const foo = {length: -1}; Object.assign(foo, {other: 123}); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #34
    {
      code: 'const foo = {length: -1}; maybe?.[foo.length = 123]; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #35
    {
      code: 'const foo = {length: -1}; maybe?.property[foo.length = 123]; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #36
    {
      code: 'const foo = {length: -1}; maybe?.property.method(foo.length = 123); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #37
    {
      code: "const foo = {length: 123}; Object.assign?.(foo, {length: 'x'}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #38
    {
      code: 'const foo = {length: -1}; if (foo.length) {} foo.length = 123;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #39
    {
      code: 'const foo = {length: -1}; if (foo.length) {} Object.assign(foo, {length: 123});',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #40
    {
      code: "const foo = {length: -1}; if (foo.length) {} Object.defineProperty(foo, 'length', {value: 123});",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #41
    {
      code: 'const foo = {length: -1}; class A {field = (foo.length = "x");} new A(); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #42
    {
      code: 'const foo = {length: -1}; class A {[foo.length = "x"]() {}} if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #43
    {
      code: 'const foo = {length: -1}; class A {accessor field = (foo.length = "x");} new A(); if (foo.length) {}',
      options: [],
      filename: 'case.ts',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #44
    {
      code: 'const foo = {length: -1}; [...foo.length] = [[123]]; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #45
    {
      code: 'const foo = {length: -1}; for (foo.length in {key: true}) {} if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #46
    {
      code: 'const foo = {length: 123}; for (foo.length in {key: true}) {} if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Snapshots valid #47
    {
      code: "const foo = {length: -1}; for (foo.length of ['x']) {} if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
  ],
  invalid: [
    // Snapshots invalid #1
    {
      code: 'if (\n\t!!!(\n\t\t!foo.length &&\n\t\tfoo.length == 0 &&\n\t\tfoo.length < 1 &&\n\t\tfoo.length <= 0 &&\n\t\t0 === foo.length &&\n\t\t0 == foo.length &&\n\t\t1 > foo.length &&\n\t\t0 >= foo.length\n\t) ||\n\t!(\n\t\tfoo.length ||\n\t\t!!foo.length ||\n\t\tfoo.length !== 0 ||\n\t\tfoo.length != 0 ||\n\t\tfoo.length >= 1 ||\n\t\t0 !== foo.length ||\n\t\t0 != foo.length ||\n\t\t0 < foo.length ||\n\t\t1 <= foo.length\n\t)\n) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 14,
          suggestions: [],
        },
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 18,
          suggestions: [],
        },
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 17,
          suggestions: [],
        },
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 18,
          suggestions: [],
        },
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 7,
          column: 3,
          endLine: 7,
          endColumn: 19,
          suggestions: [],
        },
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 18,
          suggestions: [],
        },
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 9,
          column: 3,
          endLine: 9,
          endColumn: 17,
          suggestions: [],
        },
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 10,
          column: 3,
          endLine: 10,
          endColumn: 18,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 13,
          column: 3,
          endLine: 13,
          endColumn: 13,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 14,
          column: 3,
          endLine: 14,
          endColumn: 15,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 15,
          column: 3,
          endLine: 15,
          endColumn: 19,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 16,
          column: 3,
          endLine: 16,
          endColumn: 18,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 17,
          column: 3,
          endLine: 17,
          endColumn: 18,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 18,
          column: 3,
          endLine: 18,
          endColumn: 19,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 19,
          column: 3,
          endLine: 19,
          endColumn: 18,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 20,
          column: 3,
          endLine: 20,
          endColumn: 17,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 21,
          column: 3,
          endLine: 21,
          endColumn: 18,
          suggestions: [],
        },
      ],
      output:
        'if (\n\t!!!(\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0\n\t) ||\n\t!(\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0\n\t)\n) {}',
    },
    // Snapshots invalid #2
    {
      code: 'if (\n\tfoo.length ||\n\t!!foo.length ||\n\tfoo.length != 0 ||\n\tfoo.length > 0 ||\n\tfoo.length >= 1 ||\n\t0 !== foo.length ||\n\t0 != foo.length ||\n\t0 < foo.length ||\n\t1 <= foo.length\n) {}',
      options: [{ 'non-zero': 'not-equal' }],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length !== 0` when checking length is not zero.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 12,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length !== 0` when checking length is not zero.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 14,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length !== 0` when checking length is not zero.',
          line: 4,
          column: 2,
          endLine: 4,
          endColumn: 17,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length !== 0` when checking length is not zero.',
          line: 5,
          column: 2,
          endLine: 5,
          endColumn: 16,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length !== 0` when checking length is not zero.',
          line: 6,
          column: 2,
          endLine: 6,
          endColumn: 17,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length !== 0` when checking length is not zero.',
          line: 7,
          column: 2,
          endLine: 7,
          endColumn: 18,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length !== 0` when checking length is not zero.',
          line: 8,
          column: 2,
          endLine: 8,
          endColumn: 17,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length !== 0` when checking length is not zero.',
          line: 9,
          column: 2,
          endLine: 9,
          endColumn: 16,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length !== 0` when checking length is not zero.',
          line: 10,
          column: 2,
          endLine: 10,
          endColumn: 17,
          suggestions: [],
        },
      ],
      output:
        'if (\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0\n) {}',
    },
    // Snapshots invalid #3
    {
      code: 'const foo = { length: 123 }; if (foo.length) {}',
      options: [{ 'non-zero': 'not-equal' }],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length !== 0` when checking length is not zero.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
      output: 'const foo = { length: 123 }; if (foo.length !== 0) {}',
    },
    // Snapshots invalid #4
    {
      code: 'const foo = {length: -1}; foo.length = 123; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 59,
          suggestions: [],
        },
      ],
      output:
        'const foo = {length: -1}; foo.length = 123; if (foo.length > 0) {}',
    },
    // Snapshots invalid #5
    {
      code: 'const foo = {length: -1}; Object.assign(foo, {length: 123}); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 66,
          endLine: 1,
          endColumn: 76,
          suggestions: [],
        },
      ],
      output:
        'const foo = {length: -1}; Object.assign(foo, {length: 123}); if (foo.length > 0) {}',
    },
    // Snapshots invalid #6
    {
      code: "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 123}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 83,
          endLine: 1,
          endColumn: 93,
          suggestions: [],
        },
      ],
      output:
        "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 123}); if (foo.length > 0) {}",
    },
    // Snapshots invalid #7
    {
      code: 'const foo = {length: -1}; [foo.length] = [123]; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 53,
          endLine: 1,
          endColumn: 63,
          suggestions: [],
        },
      ],
      output:
        'const foo = {length: -1}; [foo.length] = [123]; if (foo.length > 0) {}',
    },
    // Snapshots invalid #8
    {
      code: 'const foo = {length: -1}; ({length: foo.length} = {length: 123}); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 71,
          endLine: 1,
          endColumn: 81,
          suggestions: [],
        },
      ],
      output:
        'const foo = {length: -1}; ({length: foo.length} = {length: 123}); if (foo.length > 0) {}',
    },
    // Snapshots invalid #9
    {
      code: 'const foo = {length: -1}; for (foo.length of [123]) {} if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 60,
          endLine: 1,
          endColumn: 70,
          suggestions: [],
        },
      ],
      output:
        'const foo = {length: -1}; for (foo.length of [123]) {} if (foo.length > 0) {}',
    },
    // Snapshots invalid #10
    {
      code: 'if (foo.bar && foo.bar.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
      output: 'if (foo.bar && foo.bar.length > 0) {}',
    },
    // Snapshots invalid #11
    {
      code: 'if (foo.length || foo.bar()) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 15,
          suggestions: [],
        },
      ],
      output: 'if (foo.length > 0 || foo.bar()) {}',
    },
    // Snapshots invalid #12
    {
      code: 'if (!!(!!foo.length)) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'if (foo.length > 0) {}',
    },
    // Snapshots invalid #13
    {
      code: 'if (!(foo.length === 0)) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 24,
          suggestions: [],
        },
      ],
      output: 'if (foo.length > 0) {}',
    },
    // Snapshots invalid #14
    {
      code: 'while (foo.length >= 1) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output: 'while (foo.length > 0) {}',
    },
    // Snapshots invalid #15
    {
      code: 'do {} while (foo.length);',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 24,
          suggestions: [],
        },
      ],
      output: 'do {} while (foo.length > 0);',
    },
    // Snapshots invalid #16
    {
      code: 'for (let i = 0; (bar && !foo.length); i ++) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
      output: 'for (let i = 0; (bar && foo.length === 0); i ++) {}',
    },
    // Snapshots invalid #17
    {
      code: 'const isEmpty = foo.length < 1;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
      output: 'const isEmpty = foo.length === 0;',
    },
    // Snapshots invalid #18
    {
      code: 'const isEmpty = foo.length < 1 ? true : false;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
      output: 'const isEmpty = foo.length === 0 ? true : false;',
    },
    // Snapshots invalid #19
    {
      code: 'const isEmpty = foo.length <= 0;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 32,
          suggestions: [],
        },
      ],
      output: 'const isEmpty = foo.length === 0;',
    },
    // Snapshots invalid #20
    {
      code: 'if (0 >= foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'if (foo.length === 0) {}',
    },
    // Snapshots invalid #21
    {
      code: 'bar(foo.length >= 1)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'bar(foo.length > 0)',
    },
    // Snapshots invalid #22
    {
      code: 'bar(!foo.length || foo.length)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: 'bar(foo.length === 0 || foo.length)',
    },
    // Snapshots invalid #23
    {
      code: 'const bar = void !foo.length;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 29,
          suggestions: [],
        },
      ],
      output: 'const bar = void (foo.length === 0);',
    },
    // Snapshots invalid #24
    {
      code: 'const isNotEmpty = Boolean(foo.length)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 39,
          suggestions: [],
        },
      ],
      output: 'const isNotEmpty = foo.length > 0',
    },
    // Snapshots invalid #25
    {
      code: 'if (!!Boolean(foo.length)) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
      output: 'if (foo.length > 0) {}',
    },
    // Snapshots invalid #26
    {
      code: 'const isNotEmpty = Boolean(foo.length || bar)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 38,
          suggestions: [],
        },
      ],
      output: 'const isNotEmpty = Boolean(foo.length > 0 || bar)',
    },
    // Snapshots invalid #27
    {
      code: 'const isEmpty = Boolean(!foo.length)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 37,
          suggestions: [],
        },
      ],
      output: 'const isEmpty = foo.length === 0',
    },
    // Snapshots invalid #28
    {
      code: 'const isEmpty = Boolean(foo.length === 0)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
      output: 'const isEmpty = foo.length === 0',
    },
    // Snapshots invalid #29
    {
      code: 'const isNotEmpty = !Boolean(foo.length === 0)',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 46,
          suggestions: [],
        },
      ],
      output: 'const isNotEmpty = foo.length > 0',
    },
    // Snapshots invalid #30
    {
      code: 'const isEmpty = !Boolean(!Boolean(foo.length === 0))',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 53,
          suggestions: [],
        },
      ],
      output: 'const isEmpty = foo.length === 0',
    },
    // Snapshots invalid #31
    {
      code: 'if (foo.size) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
      output: 'if (foo.size > 0) {}',
    },
    // Snapshots invalid #32
    {
      code: 'if (foo.size && bar.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.size > 0` when checking size is not zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 27,
          suggestions: [],
        },
      ],
      output: 'if (foo.size > 0 && bar.length > 0) {}',
    },
    // Snapshots invalid #33
    {
      code: 'function foo() {return!foo.length}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 34,
          suggestions: [],
        },
      ],
      output: 'function foo() {return foo.length === 0}',
    },
    // Snapshots invalid #34
    {
      code: 'function foo() {throw!foo.length}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 33,
          suggestions: [],
        },
      ],
      output: 'function foo() {throw foo.length === 0}',
    },
    // Snapshots invalid #35
    {
      code: 'async function foo() {await!foo.length}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 39,
          suggestions: [],
        },
      ],
      output: 'async function foo() {await (foo.length === 0)}',
    },
    // Snapshots invalid #36
    {
      code: 'function * foo() {yield!foo.length}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
      output: 'function * foo() {yield foo.length === 0}',
    },
    // Snapshots invalid #37
    {
      code: 'function * foo() {yield*!foo.length}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
      output: 'function * foo() {yield*foo.length === 0}',
    },
    // Snapshots invalid #38
    {
      code: 'delete!foo.length',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 18,
          suggestions: [],
        },
      ],
      output: 'delete (foo.length === 0)',
    },
    // Snapshots invalid #39
    {
      code: 'typeof!foo.length',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 18,
          suggestions: [],
        },
      ],
      output: 'typeof (foo.length === 0)',
    },
    // Snapshots invalid #40
    {
      code: 'void!foo.length',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: 'void (foo.length === 0)',
    },
    // Snapshots invalid #41
    {
      code: 'a instanceof!foo.length',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 24,
          suggestions: [],
        },
      ],
      output: 'a instanceof foo.length === 0',
    },
    // Snapshots invalid #42
    {
      code: 'a in!foo.length',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: 'a in foo.length === 0',
    },
    // Snapshots invalid #43
    {
      code: 'export default!foo.length',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
      output: 'export default foo.length === 0',
    },
    // Snapshots invalid #44
    {
      code: 'if(true){}else!foo.length',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
      output: 'if(true){}else foo.length === 0',
    },
    // Snapshots invalid #45
    {
      code: 'do!foo.length;while(true) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 14,
          suggestions: [],
        },
      ],
      output: 'do foo.length === 0;while(true) {}',
    },
    // Snapshots invalid #46
    {
      code: 'switch(foo){case!foo.length:{}}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
      output: 'switch(foo){case foo.length === 0:{}}',
    },
    // Snapshots invalid #47
    {
      code: 'for(const a of!foo.length);',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
      output: 'for(const a of foo.length === 0);',
    },
    // Snapshots invalid #48
    {
      code: 'for(const a in!foo.length);',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
      output: 'for(const a in foo.length === 0);',
    },
    // Snapshots invalid #49
    {
      code: 'const foo = {length: -1}; if (true) foo.length = 123; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 59,
          endLine: 1,
          endColumn: 69,
          suggestions: [],
        },
      ],
      output:
        'const foo = {length: -1}; if (true) foo.length = 123; if (foo.length > 0) {}',
    },
    // Snapshots invalid #50
    {
      code: 'const foo = {length: -1}; if (false) {} else foo.length = 123; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 68,
          endLine: 1,
          endColumn: 78,
          suggestions: [],
        },
      ],
      output:
        'const foo = {length: -1}; if (false) {} else foo.length = 123; if (foo.length > 0) {}',
    },
    // Snapshots invalid #51
    {
      code: 'const foo = {length: -1}; true ? foo.length = 123 : 0; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 60,
          endLine: 1,
          endColumn: 70,
          suggestions: [],
        },
      ],
      output:
        'const foo = {length: -1}; true ? foo.length = 123 : 0; if (foo.length > 0) {}',
    },
    // Snapshots invalid #52
    {
      code: 'const foo = {length: -1}; false ? 0 : foo.length = 123; if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 61,
          endLine: 1,
          endColumn: 71,
          suggestions: [],
        },
      ],
      output:
        'const foo = {length: -1}; false ? 0 : foo.length = 123; if (foo.length > 0) {}',
    },
    // Snapshots invalid #53
    {
      code: "const foo = {length: -1}; Object.assign(foo, {length: 'x', length: 123}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 79,
          endLine: 1,
          endColumn: 89,
          suggestions: [],
        },
      ],
      output:
        "const foo = {length: -1}; Object.assign(foo, {length: 'x', length: 123}); if (foo.length > 0) {}",
    },
    // Snapshots invalid #54
    {
      code: "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 'x', value: 123}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 95,
          endLine: 1,
          endColumn: 105,
          suggestions: [],
        },
      ],
      output:
        "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 'x', value: 123}); if (foo.length > 0) {}",
    },
    // Snapshots invalid #55
    {
      code: "const foo = {length: -1}; Object.defineProperties(foo, {length: {value: 'x'}, length: {value: 123}}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 107,
          endLine: 1,
          endColumn: 117,
          suggestions: [],
        },
      ],
      output:
        "const foo = {length: -1}; Object.defineProperties(foo, {length: {value: 'x'}, length: {value: 123}}); if (foo.length > 0) {}",
    },
    // Snapshots invalid #56
    {
      code: "const foo = {length: -1}; ({length: foo.length} = {length: 'x', length: 123}); if (foo.length) {}",
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 84,
          endLine: 1,
          endColumn: 94,
          suggestions: [],
        },
      ],
      output:
        "const foo = {length: -1}; ({length: foo.length} = {length: 'x', length: 123}); if (foo.length > 0) {}",
    },
    // Snapshots invalid #57
    {
      code: 'const foo = {length: 123}; Object.assign(foo); if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 52,
          endLine: 1,
          endColumn: 62,
          suggestions: [],
        },
      ],
      output:
        'const foo = {length: 123}; Object.assign(foo); if (foo.length > 0) {}',
    },
    // Snapshots invalid #58
    {
      code: 'const foo = {length: -1}; switch (value) { default: foo.length = 123; } if (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 77,
          endLine: 1,
          endColumn: 87,
          suggestions: [],
        },
      ],
      output:
        'const foo = {length: -1}; switch (value) { default: foo.length = 123; } if (foo.length > 0) {}',
    },
  ],
});

// Docs cases; diagnostic and edit expectations also run in the Go suite.
ruleTester.run('explicit-length-check', {} as never, {
  valid: [
    // Docs valid #9
    {
      code: '// ✅\nconst isEmpty = foo.length === 0;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Docs valid #11
    {
      code: '// ✅\nconst isEmptySet = foo.size === 0;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Docs valid #22
    {
      code: '// ✅\nconst isNotEmpty = foo.length > 0;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Docs valid #24
    {
      code: '// ✅\nif (foo.length > 0 || bar.length > 0) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Docs valid #26
    {
      code: '// ✅\nconst unicorn = foo.length > 0 ? 1 : 2;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Docs valid #28
    {
      code: '// ✅\nwhile (foo.length > 0) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Docs valid #30
    {
      code: '// ✅\ndo {} while (foo.length > 0);',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Docs valid #32
    {
      code: '// ✅\nfor (; foo.length > 0; ) {};',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Docs valid #34
    {
      code: 'if (bothNotEmpty(foo, bar)) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Docs valid #35
    {
      code: 'const bothNotEmpty = (a, b) => a.length > 0 && b.length > 0;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
    // Docs valid #36
    {
      code: 'if (bothNotEmpty(foo, bar)) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
    },
  ],
  invalid: [
    // Docs invalid #1
    {
      code: '// ❌\nconst isEmpty = !foo.length;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 28,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isEmpty = foo.length === 0;',
    },
    // Docs invalid #2
    {
      code: '// ❌\nconst isEmpty = foo.length == 0;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 32,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isEmpty = foo.length === 0;',
    },
    // Docs invalid #3
    {
      code: '// ❌\nconst isEmpty = foo.length < 1;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 31,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isEmpty = foo.length === 0;',
    },
    // Docs invalid #4
    {
      code: '// ❌\nconst isEmpty = foo.length <= 0;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 32,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isEmpty = foo.length === 0;',
    },
    // Docs invalid #5
    {
      code: '// ❌\nconst isEmpty = 0 === foo.length;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 33,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isEmpty = foo.length === 0;',
    },
    // Docs invalid #6
    {
      code: '// ❌\nconst isEmpty = 0 == foo.length;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 32,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isEmpty = foo.length === 0;',
    },
    // Docs invalid #7
    {
      code: '// ❌\nconst isEmpty = 1 > foo.length;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 31,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isEmpty = foo.length === 0;',
    },
    // Docs invalid #8
    {
      code: '// ❌\n// Negative style is disallowed too\nconst isEmpty = !(foo.length > 0);',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 34,
          suggestions: [],
        },
      ],
      output:
        '// ❌\n// Negative style is disallowed too\nconst isEmpty = foo.length === 0;',
    },
    // Docs invalid #10
    {
      code: '// ❌\nconst isEmptySet = !foo.size;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.size === 0` when checking size is zero.',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 29,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isEmptySet = foo.size === 0;',
    },
    // Docs invalid #13
    {
      code: '// ❌\nconst isNotEmpty = foo.length !== 0;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 36,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isNotEmpty = foo.length > 0;',
    },
    // Docs invalid #14
    {
      code: '// ❌\nconst isNotEmpty = foo.length != 0;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 35,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isNotEmpty = foo.length > 0;',
    },
    // Docs invalid #15
    {
      code: '// ❌\nconst isNotEmpty = foo.length >= 1;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 35,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isNotEmpty = foo.length > 0;',
    },
    // Docs invalid #16
    {
      code: '// ❌\nconst isNotEmpty = 0 !== foo.length;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 36,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isNotEmpty = foo.length > 0;',
    },
    // Docs invalid #17
    {
      code: '// ❌\nconst isNotEmpty = 0 != foo.length;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 35,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isNotEmpty = foo.length > 0;',
    },
    // Docs invalid #18
    {
      code: '// ❌\nconst isNotEmpty = 0 < foo.length;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 34,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isNotEmpty = foo.length > 0;',
    },
    // Docs invalid #19
    {
      code: '// ❌\nconst isNotEmpty = 1 <= foo.length;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 35,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isNotEmpty = foo.length > 0;',
    },
    // Docs invalid #20
    {
      code: '// ❌\nconst isNotEmpty = Boolean(foo.length);',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 39,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isNotEmpty = foo.length > 0;',
    },
    // Docs invalid #21
    {
      code: '// ❌\n// Negative style is disallowed too\nconst isNotEmpty = !(foo.length === 0);',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 3,
          column: 20,
          endLine: 3,
          endColumn: 39,
          suggestions: [],
        },
      ],
      output:
        '// ❌\n// Negative style is disallowed too\nconst isNotEmpty = foo.length > 0;',
    },
    // Docs invalid #23
    {
      code: '// ❌\nif (foo.length || bar.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 15,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 29,
          suggestions: [],
        },
      ],
      output: '// ❌\nif (foo.length > 0 || bar.length > 0) {}',
    },
    // Docs invalid #25
    {
      code: '// ❌\nconst unicorn = foo.length ? 1 : 2;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 27,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst unicorn = foo.length > 0 ? 1 : 2;',
    },
    // Docs invalid #27
    {
      code: '// ❌\nwhile (foo.length) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 18,
          suggestions: [],
        },
      ],
      output: '// ❌\nwhile (foo.length > 0) {}',
    },
    // Docs invalid #29
    {
      code: '// ❌\ndo {} while (foo.length);',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 24,
          suggestions: [],
        },
      ],
      output: '// ❌\ndo {} while (foo.length > 0);',
    },
    // Docs invalid #31
    {
      code: '// ❌\nfor (; foo.length; ) {};',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 18,
          suggestions: [],
        },
      ],
      output: '// ❌\nfor (; foo.length > 0; ) {};',
    },
    // Docs invalid #33
    {
      code: 'const bothNotEmpty = (a, b) => a.length && b.length;',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 40,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace `.length` with `.length > 0`.',
              output:
                'const bothNotEmpty = (a, b) => a.length > 0 && b.length;',
            },
          ],
        },
        {
          messageId: 'non-zero',
          message: 'Use `.length > 0` when checking length is not zero.',
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 52,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Replace `.length` with `.length > 0`.',
              output:
                'const bothNotEmpty = (a, b) => a.length && b.length > 0;',
            },
          ],
        },
      ],
      output: null,
    },
    // Docs invalid #37
    {
      code: '// ❌\nif (!foo.length > 0) {}',
      options: [],
      filename: 'case.js',
      languageOptions: { sourceType: 'module' },
      errors: [
        {
          messageId: 'zero',
          message: 'Use `.length === 0` when checking length is zero.',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: null,
    },
  ],
});

// The native parser does not supply Vue template ASTs or parser services.
test.skip.each([
  {
    code: '<not-template><div v-if="foo.length"></div></not-template>',
    options: [],
  },
  {
    code: '<template><div v-not-if="foo.length"></div></template>',
    options: [],
  },
  {
    code: '<template><div v-if="foo.notLength"></div></template>',
    options: [],
  },
  { code: '<template><div v-SHoW="foo.length"></div></template>', options: [] },
  {
    code: '<template><div hidden="!foo.length"></div></template>',
    options: [],
  },
  { code: '<template><img :width="foo.length"/></template>', options: [] },
  { code: '<template><div v-if="foo.length"></div></template>', options: [] },
  {
    code: '<template>\n\t<div>\n\t\t<div v-if="foo"></div>\n\t\t<div v-else-if="bar.length"></div>\n\t</div>\n</template>',
    options: [],
  },
  { code: '<template><div v-if="foo.length"></div></template>', options: [] },
  {
    code: '<template><div v-if="foo.length"></div></template>',
    options: [{ 'non-zero': 'not-equal' }],
  },
  {
    code: '<template><div v-if="foo.length"></div></template>',
    options: [{ 'non-zero': 'greater-than' }],
  },
  {
    code: '<template><div v-if="foo.length && bar"></div></template>',
    options: [],
  },
  { code: '<script>if (foo.length) {}</script>', options: [] },
  { code: '<template><div v-show="foo.length"></div></template>', options: [] },
  {
    code: '<template><div :hidden="foo.length >= 1"></div></template>',
    options: [],
  },
  {
    code: '<template><div @click="foo.length >= 1"></div></template>',
    options: [],
  },
  {
    code: '<template><div @click="method($event, foo.length >= 1)"></div></template>',
    options: [],
  },
  {
    code: '<template><div v-bind:hidden="0 === foo.length"></div></template>',
    options: [],
  },
  {
    code: '<template><input :disabled="Boolean(foo.length)"></template>',
    options: [],
  },
  {
    code: '<template><custom-component :custom-property="!foo.length"></custom-component></template>',
    options: [],
  },
  {
    code: '<template>\n\t<!-- ❌ -->\n\t<div v-if="!foo.length">Vue</div>\n\n\t<!-- ✅ -->\n\t<div v-if="foo.length === 0">Vue</div>\n</template>',
    options: [],
  },
])('Vue template: $code', () => {});
