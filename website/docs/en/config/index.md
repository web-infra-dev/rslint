# Configuration

Rslint uses a flat config format (an array of config entries), aligned with ESLint v10. JS/TS configuration files are the recommended approach.

## Configuration Files

During automatic discovery, Rslint checks config files in the following order:

1. `rslint.config.js`
2. `rslint.config.mjs`
3. `rslint.config.ts`
4. `rslint.config.mts`

Automatic discovery does not consider `.cjs` or `.cts` config files. They can
still be selected explicitly with `--config` or API `overrideConfigFile`.

### Config Discovery

When you run `rslint`, it searches for a config file by walking **upward** from the target file or directory to the filesystem root. It uses the nearest candidate that loads successfully and falls back to an ancestor when a nearer candidate cannot be loaded.

- `rslint src/foo.ts` — searches from `src/` upward
- `rslint src/` — searches from `src/` upward
- `rslint` (no args) — searches from the current working directory upward

In a monorepo, different files can automatically use different config files based on their location:

```
monorepo/
├── rslint.config.ts              ← root config
├── packages/
│   ├── foo/
│   │   ├── rslint.config.ts      ← used for files under foo/
│   │   └── src/
│   └── bar/
│       └── src/                   ← no config, inherits root
```

When linting from the monorepo root, rslint automatically discovers all nested configs and applies the nearest one to each file.

#### Global Ignores and Nested Configs

For directory or no-argument lint runs, global ignores in a parent config prevent nested configs in ignored directories from contributing lint targets.

```ts
// monorepo/rslint.config.ts
export default defineConfig([
  // Global ignore — blocks directory target discovery in these directories
  { ignores: ['**/fixtures/**', 'e2e/**'] },
  ts.configs.recommended,
  // ...
]);
```

With this config, a `rslint.config.ts` inside `e2e/` or any `fixtures/` directory is not used by a root directory traversal. An explicitly named file is still resolved from its nearest config, matching the explicit-target behavior described below.

:::tip
Only **global ignore entries** (entries with only `ignores` and an optional `name`) block directory target discovery. Entry-level ignores do not affect config discovery.
:::

You can also specify a config file explicitly (overrides automatic discovery):

```bash
rslint --config path/to/rslint.config.ts .
```

For automatically discovered configs, relative `files`, `ignores`, and `languageOptions.parserOptions.project` patterns are resolved from the config file's directory.

For a config supplied with `--config`, those patterns are resolved from the current working directory.

To generate a default config, run:

```bash
rslint --init
```

## Basic Configuration

A typical TypeScript project configuration:

```ts
import { defineConfig, globalIgnores, ts } from '@rslint/core';

export default defineConfig([
  // Global ignores — files excluded from all rules
  globalIgnores(['**/dist/**', '**/fixtures/**']),
  // Preset with recommended rules
  ts.configs.recommended,
  // Custom rule overrides
  {
    rules: {
      '@typescript-eslint/no-unused-vars': 'error',
      '@typescript-eslint/array-type': ['warn', { default: 'array-simple' }],
    },
  },
]);
```

For available presets, rule severity, and plugin configuration, see [Rules & Presets](/config/rules-and-presets).

## Configuration Options

### files

- **Type:** `(string | string[])[]`

Glob selectors specifying which files this config entry applies to. Top-level selectors are ORed. Patterns in a nested array are ANDed, so `files: [['**/*.js', '!**/*.test.js']]` selects JavaScript files except test files. If omitted, the entry cascades across files selected by the config's implicit or explicit selectors.

If `files` is present, its outer array must be non-empty. Use an omitted `files` field for shared/default entries; `files: []` is invalid. A nested empty AND group (`files: [[]]`) is valid and matches vacuously.

Lint targets are selected from the CLI/API target range and are limited to Rslint's supported script extensions. Rslint always includes its default extension baseline and adds other supported candidates selected by explicit `files` entries unless the same entry's `ignores` excludes them. A `files` selector cannot make an unsupported source extension lintable. Global ignores then remove targets; CLI and JavaScript API runs also apply `.gitignore`. An entry-level ignore cannot remove a path selected by the baseline or another entry; it only prevents its own selector and config contribution. Every selected target is parsed even when no config entry contributes rules, so syntax diagnostics can still be reported. This includes default-baseline files found by a directory or no-argument scan and explicitly requested supported files that do not match a config entry's `files`.

The implicit default baseline is:

- `.js`
- `.mjs`
- `.cjs`
- `.jsx`
- `.ts`
- `.tsx`
- `.mts`
- `.cts`

This is independent of tsconfig's `include`: a file in tsconfig but outside rslint's lint target set will not run lint rules, while a selected file not covered by a tsconfig declared by its governing config still runs rules that do not require type information.

```ts
{
  files: ['**/*.ts', '**/*.tsx'],
  rules: { /* ... */ },
}
```

:::tip
Selected files not covered by a tsconfig declared by their governing config automatically receive a reduced rule set: only rules that do not require type information run. To enable type-aware rules, add the file to one of that config's tsconfigs.
:::

### ignores

For file exclusion patterns, negation, and `.gitignore` integration, see [Ignoring Files](/config/ignoring-files).

### rules

- **Type:** `RuleSeverity | [RuleSeverity, ...unknown[]]`
- **RuleSeverity:** `'off' | 'warn' | 'error' | 0 | 1 | 2`

The numeric levels follow ESLint: `0` disables a rule, `1` reports warnings, and `2` reports errors. Array entries pass every item after the severity to the rule as positional options, so configurations such as `['error', 'always', { exceptRange: true }]` are supported. Invalid severities and rule value shapes are rejected while the configuration is loaded.

When a later matching entry changes only the severity, the rule keeps options from the earlier entry. Supplying any positional option in the later array replaces the earlier options.

For available rules, presets, and plugin configuration, see [Rules & Presets](/config/rules-and-presets).

### plugins

- **Type:** `string[] | Record<string, ESLintPlugin>`

Plugins enabled for this entry, in one of two forms.

**Array of names** — built-in (native) plugins. A name declares a rule namespace; its rules become available under the `<plugin>/<rule>` prefix inside `rules`.

Built-in plugin names: `@typescript-eslint`, `import`, `jest`, `jsx-a11y`, `promise`, `react`, `react-hooks`, `rstest`, `unicorn`.

```ts
{
  files: ['**/*.ts'],
  plugins: ['@typescript-eslint'],
  rules: {
    '@typescript-eslint/no-explicit-any': 'error',
  },
}
```

**Object of plugin instances** — third-party ESLint plugins. Map a prefix key to an imported plugin object, then enable its rules under the `<prefix>/<rule>` namespace. These JavaScript rules run in the Node plugin worker and are routed through the same per-file flat config as native rules. This form requires a JS/TS config because JSON cannot carry a live plugin object.

```ts
import examplePlugin from 'eslint-plugin-example';

{
  files: ['**/*.ts'],
  plugins: { example: examplePlugin },
  rules: {
    'example/some-rule': 'error',
  },
}
```

A single entry uses one form. To combine built-in and third-party plugins, declare them in separate config entries; matching entries are merged before linting. A third-party prefix may not collide with a built-in plugin name.

ESLint core rules (unprefixed names like `no-unused-vars` or `prefer-const`) are not part of any plugin and can be enabled directly in `rules` without listing anything here. Presets like `ts.configs.recommended` already include their own `plugins` entry, so you only need this field when configuring plugin rules outside a preset.

See [ESLint plugin compatibility](/guide/eslint-plugins) for the supported and unsupported ESLint APIs.

### languageOptions

- **Type:** `object`

#### languageOptions.parserOptions.projectService

- **Type:** `boolean`

Enable TypeScript's project service for automatic tsconfig discovery. This is the default in `ts.configs.recommended`.

```ts
{
  languageOptions: {
    parserOptions: {
      projectService: true,
    },
  },
}
```

#### languageOptions.parserOptions.project

- **Type:** `string | string[]`

Explicit tsconfig.json paths. Supports glob patterns for monorepos. Files included by these tsconfigs receive full type information, enabling type-aware rules (e.g. `@typescript-eslint/no-floating-promises`, `@typescript-eslint/await-thenable`). Files outside all tsconfigs are still linted, but only rules that do not require type information run.

```ts
{
  languageOptions: {
    parserOptions: {
      project: ['./tsconfig.json', './packages/*/tsconfig.json'],
    },
  },
}
```

#### languageOptions.globals

- **Type:** `Record<string, boolean | null | 'true' | 'false' | 'readonly' | 'readable' | 'writable' | 'writeable' | 'off'>`

Declares globals available to matching files. Values are normalized before rules or third-party plugins receive the scope:

- Writable: `true`, `'true'`, `'writable'`, `'writeable'`
- Read-only: `false`, `null`, `'false'`, `'readonly'`, `'readable'`
- Disabled: `'off'`

A disabled value removes a declaration inherited from an earlier matching entry, including an ECMAScript built-in in the third-party plugin scope.

```ts
{
  languageOptions: {
    globals: {
      BUILD_ID: 'readonly',
      testRuntime: 'writable',
    },
  },
}
```

### settings

- **Type:** `Record<string, unknown>`

Shared settings accessible to all rules. Ordinary nested objects are merged recursively; later arrays and scalar values replace earlier values.

## Config Merging

When multiple config entries match a file, they are merged in array order:

1. **Global ignores** — entries containing only `ignores` and an optional `name` remove files from the target set
2. **Selector union** — the implicit default baseline and effective explicit `files` entries decide whether the config selects the file
3. **Files matching** — entries whose explicit `files` patterns don't match are skipped; entries without `files` cascade across the selector union
4. **Entry-level ignores** — matching entries do not select or configure the file, but cannot remove a target selected elsewhere
5. **Rules** — later entries override earlier ones; a severity-only value retains earlier options
6. **Plugins** — union from all matching entries
7. **Settings** — ordinary nested objects merge recursively; arrays and scalar values are replaced
8. **Language options** — ordinary nested objects merge recursively; arrays and scalar values are replaced

If no entry matches a selected file, no lint rules run for it, but the file is still parsed and included in the result so parser diagnostics remain visible. This applies to default-baseline files found during directory discovery as well as explicitly requested supported files. Global ignores remove matching targets; CLI and JavaScript API runs apply `.gitignore` as an additional global ignore source.

## JSON Configuration (Deprecated)

JSON config files (`rslint.json`, `rslint.jsonc`) are deprecated and will be removed in a future version. Run `rslint --init` to automatically migrate your JSON config to a JS/TS config. The migration preserves your custom rules and settings while deduplicating rules already covered by recommended presets.

Key difference: JSON configs automatically enable all core rules and declared plugin rules as `"error"`. JS/TS configs only enable rules explicitly declared in presets or the `rules` field.
