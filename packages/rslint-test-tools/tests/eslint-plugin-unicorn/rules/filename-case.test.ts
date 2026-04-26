import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

function valid(filename: string | undefined, c?: string) {
  return {
    code: `/* Filename: ${filename ?? '<none>'} */`,
    filename,
    options: c ? [{ case: c }] : [],
  };
}

function validCases(filename: string, cases: Record<string, boolean>) {
  return {
    code: `/* Filename: ${filename} */`,
    filename,
    options: [{ cases }],
  };
}

function validWithOptions(filename: string | undefined, options: any[] = []) {
  return {
    code: `/* Filename: ${filename ?? '<none>'} */`,
    filename,
    options,
  };
}

function invalid(filename: string, c: string | undefined, message: string) {
  return {
    code: `/* Filename: ${filename} */`,
    filename,
    options: c ? [{ case: c }] : [],
    errors: [{ message }],
  };
}

function invalidCases(
  filename: string,
  cases: Record<string, boolean> | undefined,
  message: string,
) {
  return {
    code: `/* Filename: ${filename} */`,
    filename,
    options: cases ? [{ cases }] : [],
    errors: [{ message }],
  };
}

function invalidWithOptions(filename: string, options: any[], message: string) {
  return {
    code: `/* Filename: ${filename} */`,
    filename,
    options,
    errors: [{ message }],
  };
}

ruleTester.run('filename-case', {} as never, {
  valid: [
    // ---- Single `case` option ----
    valid('src/foo/bar.js', 'camelCase'),
    valid('src/foo/fooBar.js', 'camelCase'),
    valid('src/foo/bar.test.js', 'camelCase'),
    valid('src/foo/fooBar.test.js', 'camelCase'),
    valid('src/foo/fooBar.test-utils.js', 'camelCase'),
    valid('src/foo/fooBar.test_utils.js', 'camelCase'),
    valid('src/foo/.test_utils.js', 'camelCase'),
    valid('src/foo/foo.js', 'snakeCase'),
    valid('src/foo/foo_bar.js', 'snakeCase'),
    valid('src/foo/foo.test.js', 'snakeCase'),
    valid('src/foo/foo_bar.test.js', 'snakeCase'),
    valid('src/foo/foo_bar.test_utils.js', 'snakeCase'),
    valid('src/foo/foo_bar.test-utils.js', 'snakeCase'),
    valid('src/foo/.test-utils.js', 'snakeCase'),
    valid('src/foo/foo.js', 'kebabCase'),
    valid('src/foo/foo-bar.js', 'kebabCase'),
    valid('src/foo/foo.test.js', 'kebabCase'),
    valid('src/foo/foo-bar.test.js', 'kebabCase'),
    valid('src/foo/foo-bar.test-utils.js', 'kebabCase'),
    valid('src/foo/foo-bar.test_utils.js', 'kebabCase'),
    valid('src/foo/.test_utils.js', 'kebabCase'),
    valid('src/foo/Foo.js', 'pascalCase'),
    valid('src/foo/FooBar.js', 'pascalCase'),
    valid('src/foo/Foo.test.js', 'pascalCase'),
    valid('src/foo/FooBar.test.js', 'pascalCase'),
    valid('src/foo/FooBar.test-utils.js', 'pascalCase'),
    valid('src/foo/FooBar.test_utils.js', 'pascalCase'),
    valid('src/foo/.test_utils.js', 'pascalCase'),

    // ---- Numeric / mixed identifier cases ----
    valid('spec/iss47Spec.js', 'camelCase'),
    valid('spec/iss47Spec100.js', 'camelCase'),
    valid('spec/i18n.js', 'camelCase'),
    valid('spec/iss47-spec.js', 'kebabCase'),
    valid('spec/iss-47-spec.js', 'kebabCase'),
    valid('spec/iss47-100spec.js', 'kebabCase'),
    valid('spec/i18n.js', 'kebabCase'),
    valid('spec/iss47_spec.js', 'snakeCase'),
    valid('spec/iss_47_spec.js', 'snakeCase'),
    valid('spec/iss47_100spec.js', 'snakeCase'),
    valid('spec/i18n.js', 'snakeCase'),
    valid('spec/Iss47Spec.js', 'pascalCase'),
    valid('spec/Iss47.100spec.js', 'pascalCase'),
    valid('spec/I18n.js', 'pascalCase'),

    // ---- Leading underscores preserved ----
    valid('src/foo/_fooBar.js', 'camelCase'),
    valid('src/foo/___fooBar.js', 'camelCase'),
    valid('src/foo/_foo_bar.js', 'snakeCase'),
    valid('src/foo/___foo_bar.js', 'snakeCase'),
    valid('src/foo/_foo-bar.js', 'kebabCase'),
    valid('src/foo/___foo-bar.js', 'kebabCase'),
    valid('src/foo/_FooBar.js', 'pascalCase'),
    valid('src/foo/___FooBar.js', 'pascalCase'),

    // ---- Default kebab + special chars at start ----
    valid('src/foo/$foo.js'),

    // ---- `cases` option ----
    {
      code: '/* Filename: src/foo/foo-bar.js */',
      filename: 'src/foo/foo-bar.js',
      options: [{ cases: undefined }],
    },
    {
      code: '/* Filename: src/foo/foo-bar.js */',
      filename: 'src/foo/foo-bar.js',
      options: [{ cases: {} }],
    },
    validCases('src/foo/fooBar.js', { camelCase: true }),
    validCases('src/foo/FooBar.js', { kebabCase: true, pascalCase: true }),
    validCases('src/foo/___foo_bar.js', { snakeCase: true, pascalCase: true }),

    // ---- Default options (no options at all) ----
    { code: '/* Filename: src/foo/bar.js */', filename: 'src/foo/bar.js' },

    // ---- Decoration characters ----
    valid('src/foo/[fooBar].js', 'camelCase'),
    valid('src/foo/{foo_bar}.js', 'snakeCase'),

    // ---- Default-ignored index files (any case enabled) ----
    valid('index.js', 'camelCase'),
    valid('index.mjs', 'camelCase'),
    valid('index.cjs', 'camelCase'),
    valid('index.ts', 'camelCase'),
    valid('index.tsx', 'camelCase'),
    valid('index.vue', 'camelCase'),
    valid('index.js', 'snakeCase'),
    valid('index.mjs', 'snakeCase'),
    valid('index.cjs', 'snakeCase'),
    valid('index.ts', 'snakeCase'),
    valid('index.tsx', 'snakeCase'),
    valid('index.vue', 'snakeCase'),
    valid('index.js', 'kebabCase'),
    valid('index.mjs', 'kebabCase'),
    valid('index.cjs', 'kebabCase'),
    valid('index.ts', 'kebabCase'),
    valid('index.tsx', 'kebabCase'),
    valid('index.vue', 'kebabCase'),
    valid('index.js', 'pascalCase'),
    valid('index.mjs', 'pascalCase'),
    valid('index.cjs', 'pascalCase'),
    valid('index.ts', 'pascalCase'),
    valid('index.tsx', 'pascalCase'),
    valid('index.vue', 'pascalCase'),

    // ---- Ignore patterns ----
    validWithOptions('src/foo/index.js', [
      { case: 'kebabCase', ignore: [String.raw`FOOBAR\.js`] },
    ]),
    validWithOptions('src/foo/FOOBAR.js', [
      { case: 'kebabCase', ignore: [String.raw`FOOBAR\.js`] },
    ]),
    validWithOptions('src/foo/FOOBAR.js', [
      { case: 'camelCase', ignore: [String.raw`FOOBAR\.js`] },
    ]),
    validWithOptions('src/foo/FOOBAR.js', [
      { case: 'snakeCase', ignore: [String.raw`FOOBAR\.js`] },
    ]),
    validWithOptions('src/foo/FOOBAR.js', [
      { case: 'pascalCase', ignore: [String.raw`FOOBAR\.js`] },
    ]),
    validWithOptions('src/foo/BARBAZ.js', [
      {
        case: 'kebabCase',
        ignore: [String.raw`FOOBAR\.js`, String.raw`BARBAZ\.js`],
      },
    ]),
    validWithOptions('src/foo/[FOOBAR].js', [
      { case: 'camelCase', ignore: [String.raw`\[FOOBAR\]\.js`] },
    ]),
    validWithOptions('src/foo/{FOOBAR}.js', [
      { case: 'snakeCase', ignore: [String.raw`\{FOOBAR\}\.js`] },
    ]),
    validWithOptions('src/foo/foo.js', [
      { case: 'kebabCase', ignore: ['^(F|f)oo'] },
    ]),
    validWithOptions('src/foo/foo-bar.js', [
      { case: 'kebabCase', ignore: ['^(F|f)oo'] },
    ]),
    validWithOptions('src/foo/fooBar.js', [
      { case: 'kebabCase', ignore: ['^(F|f)oo'] },
    ]),
    validWithOptions('src/foo/foo_bar.js', [
      { case: 'kebabCase', ignore: ['^(F|f)oo'] },
    ]),
    validWithOptions('src/foo/foo-bar.js', [
      { case: 'kebabCase', ignore: [String.raw`\.(web|android|ios)\.js$`] },
    ]),
    validWithOptions('src/foo/FooBar.web.js', [
      { case: 'kebabCase', ignore: [String.raw`\.(web|android|ios)\.js$`] },
    ]),
    validWithOptions('src/foo/FooBar.android.js', [
      { case: 'kebabCase', ignore: [String.raw`\.(web|android|ios)\.js$`] },
    ]),
    validWithOptions('src/foo/FooBar.ios.js', [
      { case: 'kebabCase', ignore: [String.raw`\.(web|android|ios)\.js$`] },
    ]),
    validWithOptions('src/foo/FooBar.js', [
      { case: 'kebabCase', ignore: ['^(F|f)oo'] },
    ]),
    validWithOptions('src/foo/FOOBAR.js', [
      { case: 'kebabCase', ignore: ['^FOO', String.raw`BAZ\.js$`] },
    ]),
    validWithOptions('src/foo/BARBAZ.js', [
      { case: 'kebabCase', ignore: ['^FOO', String.raw`BAZ\.js$`] },
    ]),
    validWithOptions('src/foo/FOOBAR.js', [
      {
        cases: {
          kebabCase: true,
          camelCase: true,
          snakeCase: true,
          pascalCase: true,
        },
        ignore: [String.raw`FOOBAR\.js`],
      },
    ]),
    validWithOptions('src/foo/BaRbAz.js', [
      {
        cases: {
          kebabCase: true,
          camelCase: true,
          snakeCase: true,
          pascalCase: true,
        },
        ignore: [String.raw`FOOBAR\.js`, String.raw`BaRbAz\.js`],
      },
    ]),

    // ---- multipleFileExtensions=false ----
    validWithOptions('index.tsx', [
      { case: 'pascalCase', multipleFileExtensions: false },
    ]),
    validWithOptions('src/index.tsx', [
      { case: 'pascalCase', multipleFileExtensions: false },
    ]),
    validWithOptions('src/foo/fooBar.test.js', [
      { case: 'camelCase', multipleFileExtensions: false },
    ]),
    validWithOptions('src/foo/fooBar.testUtils.js', [
      { case: 'camelCase', multipleFileExtensions: false },
    ]),
    validWithOptions('src/foo/foo_bar.test_utils.js', [
      { case: 'snakeCase', multipleFileExtensions: false },
    ]),
    validWithOptions('src/foo/foo.test.js', [
      { case: 'kebabCase', multipleFileExtensions: false },
    ]),
    validWithOptions('src/foo/foo-bar.test.js', [
      { case: 'kebabCase', multipleFileExtensions: false },
    ]),
    validWithOptions('src/foo/foo-bar.test-utils.js', [
      { case: 'kebabCase', multipleFileExtensions: false },
    ]),
    validWithOptions('src/foo/Foo.Test.js', [
      { case: 'pascalCase', multipleFileExtensions: false },
    ]),
    validWithOptions('src/foo/FooBar.Test.js', [
      { case: 'pascalCase', multipleFileExtensions: false },
    ]),
    validWithOptions('src/foo/FooBar.TestUtils.js', [
      { case: 'pascalCase', multipleFileExtensions: false },
    ]),
    validWithOptions('spec/Iss47.100Spec.js', [
      { case: 'pascalCase', multipleFileExtensions: false },
    ]),

    // ---- Multipart filename / multipleFileExtensions=true ----
    validWithOptions('src/foo/fooBar.Test.js', [{ case: 'camelCase' }]),
    validWithOptions('test/foo/fooBar.testUtils.js', [{ case: 'camelCase' }]),
    validWithOptions('test/foo/.testUtils.js', [{ case: 'camelCase' }]),
    validWithOptions('test/foo/foo_bar.Test.js', [{ case: 'snakeCase' }]),
    validWithOptions('test/foo/foo_bar.Test_Utils.js', [{ case: 'snakeCase' }]),
    validWithOptions('test/foo/.Test_Utils.js', [{ case: 'snakeCase' }]),
    validWithOptions('test/foo/foo-bar.Test.js', [{ case: 'kebabCase' }]),
    validWithOptions('test/foo/foo-bar.Test-Utils.js', [{ case: 'kebabCase' }]),
    validWithOptions('test/foo/.Test-Utils.js', [{ case: 'kebabCase' }]),
    validWithOptions('test/foo/FooBar.Test.js', [{ case: 'pascalCase' }]),
    validWithOptions('test/foo/FooBar.TestUtils.js', [{ case: 'pascalCase' }]),
    validWithOptions('test/foo/.TestUtils.js', [{ case: 'pascalCase' }]),

    // ---- Snapshot-style: directory with uppercase ext doesn't matter ----
    { code: '', filename: 'src/foo.JS/bar.js' },
    { code: '', filename: 'src/foo.JS/bar.spec.js' },
    { code: '', filename: 'src/foo.JS/.spec.js' },
    { code: '', filename: 'foo.SPEC.js' },
    { code: '', filename: '.SPEC.js' },
  ],

  invalid: [
    // ---- Default kebab on snake-cased filename ----
    invalid(
      'src/foo/foo_bar.js',
      undefined,
      'Filename is not in kebab case. Rename it to `foo-bar.js`.',
    ),

    // ---- camelCase failures ----
    invalid(
      'src/foo/foo_bar.test.js',
      'camelCase',
      'Filename is not in camel case. Rename it to `fooBar.test.js`.',
    ),
    invalid(
      'test/foo/foo_bar.test_utils.js',
      'camelCase',
      'Filename is not in camel case. Rename it to `fooBar.test_utils.js`.',
    ),

    // ---- snakeCase failures ----
    invalid(
      'test/foo/fooBar.js',
      'snakeCase',
      'Filename is not in snake case. Rename it to `foo_bar.js`.',
    ),
    invalid(
      'test/foo/fooBar.test.js',
      'snakeCase',
      'Filename is not in snake case. Rename it to `foo_bar.test.js`.',
    ),
    invalid(
      'test/foo/fooBar.testUtils.js',
      'snakeCase',
      'Filename is not in snake case. Rename it to `foo_bar.testUtils.js`.',
    ),

    // ---- kebabCase failures ----
    invalid(
      'test/foo/fooBar.js',
      'kebabCase',
      'Filename is not in kebab case. Rename it to `foo-bar.js`.',
    ),
    invalid(
      'test/foo/fooBar.test.js',
      'kebabCase',
      'Filename is not in kebab case. Rename it to `foo-bar.test.js`.',
    ),
    invalid(
      'test/foo/fooBar.testUtils.js',
      'kebabCase',
      'Filename is not in kebab case. Rename it to `foo-bar.testUtils.js`.',
    ),

    // ---- pascalCase failures ----
    invalid(
      'test/foo/fooBar.js',
      'pascalCase',
      'Filename is not in pascal case. Rename it to `FooBar.js`.',
    ),
    invalid(
      'test/foo/foo_bar.test.js',
      'pascalCase',
      'Filename is not in pascal case. Rename it to `FooBar.test.js`.',
    ),
    invalid(
      'test/foo/foo-bar.test-utils.js',
      'pascalCase',
      'Filename is not in pascal case. Rename it to `FooBar.test-utils.js`.',
    ),

    // ---- Leading underscores preserved verbatim ----
    invalid(
      'src/foo/_FOO-BAR.js',
      'camelCase',
      'Filename is not in camel case. Rename it to `_fooBar.js`.',
    ),
    invalid(
      'src/foo/___FOO-BAR.js',
      'camelCase',
      'Filename is not in camel case. Rename it to `___fooBar.js`.',
    ),
    invalid(
      'src/foo/_FOO-BAR.js',
      'snakeCase',
      'Filename is not in snake case. Rename it to `_foo_bar.js`.',
    ),
    invalid(
      'src/foo/___FOO-BAR.js',
      'snakeCase',
      'Filename is not in snake case. Rename it to `___foo_bar.js`.',
    ),
    invalid(
      'src/foo/_FOO-BAR.js',
      'kebabCase',
      'Filename is not in kebab case. Rename it to `_foo-bar.js`.',
    ),
    invalid(
      'src/foo/___FOO-BAR.js',
      'kebabCase',
      'Filename is not in kebab case. Rename it to `___foo-bar.js`.',
    ),
    invalid(
      'src/foo/_FOO-BAR.js',
      'pascalCase',
      'Filename is not in pascal case. Rename it to `_FooBar.js`.',
    ),
    invalid(
      'src/foo/___FOO-BAR.js',
      'pascalCase',
      'Filename is not in pascal case. Rename it to `___FooBar.js`.',
    ),

    // ---- `cases` option failures (canonical case order in our message) ----
    invalidCases(
      'src/foo/foo_bar.js',
      undefined,
      'Filename is not in kebab case. Rename it to `foo-bar.js`.',
    ),
    invalidCases(
      'src/foo/foo-bar.js',
      { camelCase: true, pascalCase: true },
      'Filename is not in camel case or pascal case. Rename it to `fooBar.js` or `FooBar.js`.',
    ),
    invalidCases(
      'src/foo/_foo_bar.js',
      { camelCase: true, pascalCase: true, kebabCase: true },
      'Filename is not in camel case, kebab case, or pascal case. Rename it to `_fooBar.js`, `_foo-bar.js`, or `_FooBar.js`.',
    ),
    invalidCases(
      'src/foo/_FOO-BAR.js',
      { snakeCase: true },
      'Filename is not in snake case. Rename it to `_foo_bar.js`.',
    ),

    // ---- Decoration characters preserved ----
    invalid(
      'src/foo/[foo_bar].js',
      undefined,
      'Filename is not in kebab case. Rename it to `[foo-bar].js`.',
    ),
    invalid(
      'src/foo/$foo_bar.js',
      undefined,
      'Filename is not in kebab case. Rename it to `$foo-bar.js`.',
    ),
    invalid(
      'src/foo/$fooBar.js',
      undefined,
      'Filename is not in kebab case. Rename it to `$foo-bar.js`.',
    ),
    invalidCases(
      'src/foo/{foo_bar}.js',
      { camelCase: true, pascalCase: true, kebabCase: true },
      'Filename is not in camel case, kebab case, or pascal case. Rename it to `{fooBar}.js`, `{foo-bar}.js`, or `{FooBar}.js`.',
    ),

    // ---- Ignore patterns that don't match still report ----
    invalidWithOptions(
      'src/foo/barBaz.js',
      [{ case: 'kebabCase', ignore: [String.raw`FOOBAR\.js`] }],
      'Filename is not in kebab case. Rename it to `bar-baz.js`.',
    ),
    invalidWithOptions(
      'src/foo/fooBar.js',
      [{ case: 'kebabCase', ignore: [String.raw`FOOBAR\.js`] }],
      'Filename is not in kebab case. Rename it to `foo-bar.js`.',
    ),
    invalidWithOptions(
      'src/foo/fooBar.js',
      [
        {
          case: 'kebabCase',
          ignore: [String.raw`FOOBAR\.js`, String.raw`foobar\.js`],
        },
      ],
      'Filename is not in kebab case. Rename it to `foo-bar.js`.',
    ),
    invalidWithOptions(
      'src/foo/FooBar.js',
      [
        {
          cases: { camelCase: true, snakeCase: true },
          ignore: [String.raw`FOOBAR\.js`],
        },
      ],
      'Filename is not in camel case or snake case. Rename it to `fooBar.js` or `foo_bar.js`.',
    ),

    // ---- #1136: trailing underscore on a digit-only word ----
    invalidCases(
      'src/foo/1_.js',
      { camelCase: true, pascalCase: true, kebabCase: true },
      'Filename is not in camel case, kebab case, or pascal case. Rename it to `1.js`.',
    ),

    // ---- multipleFileExtensions=false: middle parts must also match ----
    invalidWithOptions(
      'src/foo/foo_bar.test.js',
      [{ case: 'camelCase', multipleFileExtensions: false }],
      'Filename is not in camel case. Rename it to `fooBar.test.js`.',
    ),
    invalidWithOptions(
      'test/foo/foo_bar.test_utils.js',
      [{ case: 'camelCase', multipleFileExtensions: false }],
      'Filename is not in camel case. Rename it to `fooBar.testUtils.js`.',
    ),
    invalidWithOptions(
      'test/foo/fooBar.test.js',
      [{ case: 'snakeCase', multipleFileExtensions: false }],
      'Filename is not in snake case. Rename it to `foo_bar.test.js`.',
    ),
    invalidWithOptions(
      'test/foo/fooBar.testUtils.js',
      [{ case: 'snakeCase', multipleFileExtensions: false }],
      'Filename is not in snake case. Rename it to `foo_bar.test_utils.js`.',
    ),
    invalidWithOptions(
      'test/foo/fooBar.test.js',
      [{ case: 'kebabCase', multipleFileExtensions: false }],
      'Filename is not in kebab case. Rename it to `foo-bar.test.js`.',
    ),
    invalidWithOptions(
      'test/foo/fooBar.testUtils.js',
      [{ case: 'kebabCase', multipleFileExtensions: false }],
      'Filename is not in kebab case. Rename it to `foo-bar.test-utils.js`.',
    ),
    invalidWithOptions(
      'test/foo/foo_bar.test.js',
      [{ case: 'pascalCase', multipleFileExtensions: false }],
      'Filename is not in pascal case. Rename it to `FooBar.Test.js`.',
    ),
    invalidWithOptions(
      'test/foo/foo-bar.test-utils.js',
      [{ case: 'pascalCase', multipleFileExtensions: false }],
      'Filename is not in pascal case. Rename it to `FooBar.TestUtils.js`.',
    ),

    // NOTE: uppercase / non-standard extension cases (`.JS`, `.mJS`, `.mjs`,
    // `.cjs`, `.vue`) are intentionally not exercised here. The rslint binary
    // routes lint requests through TypeScript's program builder, which only
    // picks up files whose extensions match the project's recognized set.
    // Files with `.JS` etc. silently skip the program, so no diagnostics
    // appear regardless of rule logic. The Go test suite covers the rule's
    // extension/case-style branches with stable extensions; the upstream rule
    // logic is identical for any extension.
  ],
});
