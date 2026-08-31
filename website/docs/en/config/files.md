# files

- **Type:** `(string | string[])[]`

Glob selectors specifying which files a config entry applies to. Top-level selectors are ORed. Patterns in a nested array are ANDed, so `files: [['**/*.js', '!**/*.test.js']]` selects JavaScript files except test files.

```ts
{
  files: ['**/*.ts', '**/*.tsx'],
  rules: {
    '@typescript-eslint/no-explicit-any': 'error',
  },
}
```

If `files` is omitted, the entry cascades across files selected by the config's implicit or explicit selectors. If `files` is present, its outer array must be non-empty. Use an omitted `files` field for shared or default entries; `files: []` is invalid. A nested empty AND group (`files: [[]]`) is valid and matches vacuously.

## Lint target selection

Lint targets are selected from the CLI or API target range and are limited to Rslint's supported script extensions. Rslint always includes its default extension baseline and adds other supported candidates selected by explicit `files` entries unless the same entry's `ignores` excludes them. A `files` selector cannot make an unsupported source extension lintable.

The implicit default baseline is:

- `.js`
- `.mjs`
- `.cjs`
- `.jsx`
- `.ts`
- `.tsx`
- `.mts`
- `.cts`

Global ignores then remove targets; CLI and JavaScript API runs also apply `.gitignore`. An entry-level ignore cannot remove a path selected by the baseline or another entry. It only prevents its own selector and config contribution. See [`ignores`](/config/ignoring-files) for details.

Every selected target is parsed even when no config entry contributes rules, so syntax diagnostics can still be reported. This includes default-baseline files found by a directory or no-argument scan and explicitly requested supported files that do not match a config entry's `files`.

## TypeScript project coverage

File selection is independent of a tsconfig's `include`. A file in tsconfig but outside Rslint's lint target set will not run lint rules. A selected file not covered by a tsconfig declared by its governing config still runs rules that do not require type information.

:::tip
Selected files not covered by a tsconfig declared by their governing config automatically receive a reduced rule set: only rules that do not require type information run. To enable type-aware rules, add the file to one of that config's tsconfigs. See [`languageOptions.parserOptions.project`](/config/language-options#languageoptionsparseroptionsproject).
:::

When an entry has `basePath`, its `files` patterns resolve from that directory, and its per-file configuration is inactive outside it. See the [`basePath` configuration reference](/config/base-path) for the path-origin and TypeScript project rules.
