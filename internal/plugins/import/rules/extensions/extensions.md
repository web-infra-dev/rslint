# extensions

## Rule Details

Enforce consistent file extensions in import paths, re-exports, dynamic imports,
and `require()` calls. The default mode is `"never"`. Builtins and package roots,
including names such as `decimal.js`, are exempt by default.

Examples of **incorrect** code for this rule, when both paths resolve to the same
file:

```javascript
import value from './value.js';
export { value } from './value.js';
```

Examples of **correct** code for this rule:

```javascript
import value from './value';
export { value } from './value';
```

If both `value.js` and `value.json` exist, `./value.json` remains valid when
`./value` resolves to `value.js`. The rule only forbids extensions that can be
removed without changing the resolution result. Two unresolved paths also count
as the same result, matching upstream.

Query strings are excluded when resolving a path and included in diagnostics.
The rule does not provide automatic fixes or suggestions.

## Options

The first option selects the default mode:

- `"never"` (default): forbid extensions that can be omitted.
- `"always"`: require extensions.
- `"ignorePackages"`: require extensions outside package imports.

An extension map overrides the default mode:

```javascript
export default [
  {
    plugins: ['import'],
    rules: {
      'import/extensions': ['error', 'never', { json: 'always' }],
    },
  },
];
```

The object form also accepts:

| Option | Default | Behavior |
| --- | --- | --- |
| `pattern` | `{}` | Map extensions to modes, such as `{ js: 'never', json: 'always' }`. |
| `ignorePackages` | `false` | Exempt package imports from required extensions. Forbidden extensions are still checked. |
| `checkTypeImports` | `false` | Check missing extensions in `import type` and `export type` declarations. Forbidden extensions are checked regardless. |
| `pathGroupOverrides` | `[]` | Apply the first matching glob's `ignore` or `enforce` action. |

Each path override has a `pattern`, optional `patternOptions`, and an `action`.
`ignore` skips the import; `enforce` applies the configured mode even to builtins
and packages. Glob options default to `{ nocomment: true }` when omitted.

For example, require extensions for aliases that resemble packages:

```javascript
export default [{
  plugins: ['import'],
  rules: {
    'import/extensions': ['error', 'always', {
      ignorePackages: true,
      pathGroupOverrides: [
        { pattern: 'app/**', action: 'enforce' },
        { pattern: './**/*.css', action: 'ignore' },
      ],
    }],
  },
}];
```

As in upstream v2.32.0, include `pattern`, `ignorePackages`, or `checkTypeImports`
in an object containing `pathGroupOverrides`; otherwise its overrides have no
effect. The per-extension mode `ignorePackages` disables checking that extension.

Examples of **incorrect** code with `"always"`:

```javascript
export { value } from './value';
```

Examples of **correct** code with `"always"`:

```javascript
export { value } from './value.js';
```

## Resolution settings

The default Node resolver tries `.mjs`, `.js`, `.json`, and `.node`. Configure
`settings['import/resolver']` to use `node` with an `extensions` array, or
`typescript` to use the TypeScript project selected by rslint. The legacy
`settings['import/resolve']` also accepts Node options.

## Differences from upstream

- Supported resolvers are `node` and `typescript`, also accepted as
  `eslint-import-resolver-node` and `eslint-import-resolver-typescript`.
  For example, `settings['import/resolver'] = 'webpack'` reports a resolver error.
  For bundler aliases, use `typescript` with matching `paths` in `tsconfig.json`,
  or exclude those imports with `pathGroupOverrides`.
- The `typescript` resolver follows the TypeScript project selected by rslint.
  Its `project`, `alwaysTryTypes`, and other resolver-specific options have no
  effect. For example, `{ typescript: { project: 'tsconfig.app.json' } }` does not
  select that config; use `languageOptions.parserOptions.project` instead.
- Multiple resolvers in an object run alphabetically. Use an array such as
  `['typescript', 'node']` to specify their order.
- On Windows, use `/` separators in glob patterns, such as `app/**`.
  Backslashes escape pattern characters rather than separating directories.
- Extensions such as `constructor` and `toString` follow the configured mode.
  For example, `import './file.constructor'` reports with `"never"`, while
  upstream implicitly ignores it. Set `{ constructor: 'ignorePackages' }` to
  retain that exemption.
- If an import path contains a lone surrogate escape such as `\uD800`, the
  diagnostic displays replacement characters for that escape. Regular Unicode
  paths, including emoji, retain their spelling.

## Original Documentation

- [eslint-plugin-import: extensions](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/extensions.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/extensions.js)
