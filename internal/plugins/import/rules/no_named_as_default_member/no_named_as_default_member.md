# no-named-as-default-member

## Rule Details

Warns when a property accessed on a default import has the same name as a named
export of that module. Named exports are not automatically properties of the
default export.

This rule is enabled as a warning in `importPlugin.configs.recommended`.

Given this module:

```javascript
// fruit.js
export default 'apple';
export const pear = 'pear';
```

Examples of **incorrect** code for this rule:

```javascript
import fruit from './fruit.js';
const pear = fruit.pear;
const { pear: anotherPear } = fruit;
```

Examples of **correct** code for this rule:

```javascript
import fruit, { pear } from './fruit.js';
```

The rule checks names declared directly by the imported module, including local
export lists and namespace aliases. Names reached through `export { name } from`
or `export * from` are not checked. Unresolved, ignored and CommonJS-only modules
are skipped.

Like upstream, matching uses identifier names: `fruit[pear]` is checked, while
`fruit['pear']` is not. A shadowed variable named `fruit` is also checked. The
property `default` is always allowed. Only default import specifiers are checked;
`import { default as fruit }` is not.

## Options

This rule has no options. It does not provide automatic fixes or suggestions.

## Differences from upstream

- **Checking imports requires access to the imported files.** If you disable both
  `languageOptions.parserOptions.project` and `projectService`, include imported
  files in the lint targets too. For example, checking `app.ts` alone will not
  warn about `fruit.pear` from an unselected `fruit.ts`. Keep either project
  option enabled to check dependencies automatically.
- **Imported-file syntax errors are not reported here.** If an imported file
  contains `export default {}; export const foo = ;`, upstream reports a parse
  error. This rule skips that file without warning about `obj.foo`. Check that
  file separately for syntax errors.
- **TypeScript imports work without configuring an ESLint parser.** For example,
  named exports from `fruit.ts` are checked without an `import/parsers` setting.
  An explicit `import/extensions` list still restricts which files are checked;
  extensions listed under `import/parsers` are also allowed.
- **TypeScript's default interop settings apply.** With `module: 'nodenext'`,
  rslint can warn about `obj.foo` when a library exposes `foo` through an
  `export as namespace` declaration. Upstream only checks those members when
  `esModuleInterop: true` is set explicitly. Set `esModuleInterop: false` to
  disable this behavior in both tools.

## Original Documentation

- [eslint-plugin-import: no-named-as-default-member](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-named-as-default-member.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-named-as-default-member.js)
