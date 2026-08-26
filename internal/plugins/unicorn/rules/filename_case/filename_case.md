# filename-case

## Rule Details

Enforces filenames and directory names of linted files to use a certain case style and a lowercase file extension. The default is `kebabCase`.

Directory names are checked only when the file is inside the current working directory. Files outside the current working directory only have their filename checked.

Files named `index.js`, `index.mjs`, `index.cjs`, `index.ts`, `index.tsx`, and `index.vue` are ignored because they cannot change case. Their parent directories are still checked.

Characters other than `a-z`, `A-Z`, `0-9`, `-`, and `_` are ignored for casing and kept as-is in suggested names. Path segments starting with `$` are also ignored, as they are commonly used for route parameters.

### Cases

#### `kebabCase`

- `foo-bar.js`
- `foo-bar.test.js`
- `foo-bar.test-utils.js`

#### `camelCase`

- `fooBar.js`
- `fooBar.test.js`
- `fooBar.testUtils.js`

#### `camelCaseWithAcronyms`

- `innerHTML.js`
- `getDOMRangeRect.js`
- `apiURL.js`

This is still lower camel case. Leading acronyms are lowercased, so `HTMLParser.js` should be `htmlParser.js`.

#### `snakeCase`

- `foo_bar.js`
- `foo_bar.test.js`
- `foo_bar.test_utils.js`

#### `pascalCase`

- `FooBar.js`
- `FAQPage.js`
- `FooBar.Test.js`
- `FooBar.TestUtils.js`

## Options

### case

Type: `"camelCase" | "camelCaseWithAcronyms" | "snakeCase" | "kebabCase" | "pascalCase"`

Set a single allowed case style:

```json
{ "unicorn/filename-case": ["error", { "case": "kebabCase" }] }
```

### cases

Type: `{ camelCase?: boolean; camelCaseWithAcronyms?: boolean; snakeCase?: boolean; kebabCase?: boolean; pascalCase?: boolean }`

Allow several case styles at once. Setting a key to `true` enables that style; the file passes if it matches any enabled style.

```json
{ "unicorn/filename-case": ["error", { "cases": { "camelCase": true, "pascalCase": true } }] }
```

### ignore

Type: `string[]`\
Default: `[]`

A list of regular-expression patterns (as strings) matched against every path segment. If any directory or the filename matches, the file is skipped.

```json
{
  "unicorn/filename-case": [
    "error",
    {
      "case": "kebabCase",
      "ignore": ["^FOOBAR\\.js$", "^vendor$", "^(B|b)az", "\\.SOMETHING\\.js$"]
    }
  ]
}
```

### checkDirectories

Type: `boolean`\
Default: `true`

Whether to check directory names. Filenames are always checked.

### multipleFileExtensions

Type: `boolean`\
Default: `true`

When `true`, additional `.`-separated parts of the basename are treated as part of the extension and are not subject to case checking. When `false`, only the final extension is treated as such, and the rest of the basename must match the chosen case styles.

This option only affects filenames. Directory names are always checked as complete path segments when `checkDirectories` is enabled.

## Differences from ESLint

- When several entries in `cases` are enabled, rslint uses a fixed order for
  diagnostic text and suggestions: `camelCase`, `camelCaseWithAcronyms`,
  `kebabCase`, `snakeCase`, then `pascalCase`. ESLint derives this presentation
  order from `Object.keys(options.cases)`. This does not change whether a path
  is valid or which replacements are suggested.

- JavaScript `RegExp` objects do not survive the native configuration's JSON boundary, so a regular-expression literal in `ignore` (for example `/^vendor/i`) does not match anything. Write the pattern as a string — for example `"^vendor"` for `/^vendor/`, or `"(?i)^vendor"` for `/^vendor/i`.

## Original Documentation

- [eslint-plugin-unicorn: filename-case](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/filename-case.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/filename-case.js)
