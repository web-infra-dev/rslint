# filename-case

## Rule Details

Enforces all linted files to have their names in a certain case style and a lowercase file extension. The default is `kebabCase`.

Files named `index.js`, `index.mjs`, `index.cjs`, `index.ts`, `index.tsx`, `index.vue` are ignored as they can't change case (only a problem with `pascalCase`).

Characters in the filename other than `a-z`, `A-Z`, `0-9`, `-`, and `_` are ignored.

### Cases

#### `kebabCase`

- `foo-bar.js`
- `foo-bar.test.js`
- `foo-bar.test-utils.js`

#### `camelCase`

- `fooBar.js`
- `fooBar.test.js`
- `fooBar.testUtils.js`

#### `snakeCase`

- `foo_bar.js`
- `foo_bar.test.js`
- `foo_bar.test_utils.js`

#### `pascalCase`

- `FooBar.js`
- `FooBar.Test.js`
- `FooBar.TestUtils.js`

## Options

### case

Type: `"camelCase" | "snakeCase" | "kebabCase" | "pascalCase"`

Set a single allowed case style:

```json
{ "unicorn/filename-case": ["error", { "case": "kebabCase" }] }
```

### cases

Type: `{ camelCase?: boolean; snakeCase?: boolean; kebabCase?: boolean; pascalCase?: boolean }`

Allow several case styles at once. Setting a key to `true` enables that style; the file passes if it matches any enabled style.

```json
{ "unicorn/filename-case": ["error", { "cases": { "camelCase": true, "pascalCase": true } }] }
```

### ignore

Type: `string[]`\
Default: `[]`

A list of regular-expression patterns (as strings) that match filenames the rule should skip.

```json
{
  "unicorn/filename-case": [
    "error",
    {
      "case": "kebabCase",
      "ignore": ["^FOOBAR\\.js$", "^(B|b)az", "\\.SOMETHING\\.js$"]
    }
  ]
}
```

### multipleFileExtensions

Type: `boolean`\
Default: `true`

When `true`, additional `.`-separated parts of the basename are treated as part of the extension and are not subject to case checking. When `false`, only the final extension is treated as such, and the rest of the basename must match the chosen case styles.

## Differences from ESLint

- When the `cases` option enables more than one style, the diagnostic message lists the case names and rename suggestions in a fixed order: camel case, snake case, kebab case, pascal case. ESLint lists them in the order the keys appear in your `cases` object. So `{ pascalCase: true, camelCase: true }` against `foo-bar.js` reports `Filename is not in camel case or pascal case. Rename it to ` `` `fooBar.js` `` ` or ` `` `FooBar.js` `` `.` here, and `Filename is not in pascal case or camel case. Rename it to ` `` `FooBar.js` `` ` or ` `` `fooBar.js` `` `.` in ESLint.
- A regular-expression literal in `ignore` (for example `/^vendor/i`) does not match anything. Write the pattern as a string — for example `"^vendor"` for `/^vendor/`, or `"(?i)^vendor"` for `/^vendor/i`.

## Original Documentation

- ESLint plugin documentation: <https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/filename-case.md>
- Source code: <https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/rules/filename-case.js>
