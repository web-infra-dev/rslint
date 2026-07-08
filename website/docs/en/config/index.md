# Configuration

Rslint uses a flat config format (an array of config entries), aligned with ESLint v10. JS/TS configuration files are the recommended approach.

## Configuration Files

Rslint looks for config files in the following order:

1. `rslint.config.js`
2. `rslint.config.mjs`
3. `rslint.config.ts`
4. `rslint.config.mts`

### Config Discovery

When you run `rslint`, it searches for the config file by walking **upward** from the target file or directory to the filesystem root, stopping at the first config found.

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

Global ignores in a parent config prevent nested configs in ignored directories from being discovered. This aligns with ESLint v10 behavior.

```ts
// monorepo/rslint.config.ts
export default defineConfig([
  // Global ignore — blocks config discovery in these directories
  { ignores: ['**/fixtures/**', 'e2e/**'] },
  ts.configs.recommended,
  // ...
]);
```

With this config, a `rslint.config.ts` inside `e2e/` or any `fixtures/` directory will **not** be used when linting from the monorepo root. This prevents test fixture configs from interfering with the main lint run.

:::tip
Only **global ignore entries** (entries with only `ignores` and no other fields) block nested config discovery. Entry-level ignores (entries with both `files` and `ignores`) do not affect config discovery.
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

- **Type:** `string[]`

Glob patterns specifying which files this config entry applies to. If omitted, the entry applies to rslint's default lintable file set.

If `files` is present, it must contain at least one pattern. Use an omitted `files` field for shared/default entries; `files: []` is invalid.

Lint targets are selected from the CLI/API target range, then filtered by `files`, global ignores, and `.gitignore`. Explicit file arguments are the exception: an explicit file can still appear as a lint result even when it does not match any `files` pattern, but no rules run for it unless `GetConfigForFile` finds a matching entry. Global ignores and `.gitignore` still remove explicit files from the result set.

Each non-global entry that omits `files` contributes this default set:

- `.js`
- `.mjs`
- `.cjs`
- `.jsx`
- `.ts`
- `.tsx`
- `.mts`
- `.cts`

This is independent of tsconfig's `include`: a file in tsconfig but not selected by rslint's lint scope will not be linted, while a selected file that is not in any tsconfig will still be linted with syntax-only rules (type-aware rules require tsconfig coverage).

```ts
{
  files: ['**/*.ts', '**/*.tsx'],
  rules: { /* ... */ },
}
```

:::tip
Files that match `files` but are not included in any tsconfig automatically receive a reduced rule set — only rules that don't require type information will run. To enable type-aware rules for these files, add them to your tsconfig's `include`.
:::

### ignores

For file exclusion patterns, negation, and `.gitignore` integration, see [Ignoring Files](/config/ignoring-files).

### rules

For rule severity levels, option format, and plugin configuration, see [Rules & Presets](/config/rules-and-presets).

### plugins

- **Type:** `string[] | Record<string, ESLintPlugin>`

Plugins enabled for this entry, in one of two forms.

**Array of names** — built-in (native) plugins. A name declares a rule namespace; its rules become available under the `<plugin>/<rule>` prefix inside `rules`.

Built-in plugin names: `@typescript-eslint`, `import`, `jest`, `jsx-a11y`, `promise`, `react`, `react-hooks`, `unicorn`.

```ts
{
  files: ['**/*.ts'],
  plugins: ['@typescript-eslint'],
  rules: {
    '@typescript-eslint/no-explicit-any': 'error',
  },
}
```

**Object of plugin instances** — community ESLint plugins. Map a prefix key to an imported plugin object, then enable its rules under the `<prefix>/<rule>` namespace. The plugin's JS rules run in a Node worker. Requires a JS/TS config — a live plugin object can't be expressed in JSON.

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

A single entry uses one form. To combine built-in and community plugins, declare them in separate config entries (merged at lint time). A community prefix may not collide with a built-in plugin name.

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

Explicit tsconfig.json paths. Supports glob patterns for monorepos. Files included by these tsconfigs receive full type information, enabling type-aware rules (e.g. `@typescript-eslint/no-floating-promises`, `@typescript-eslint/await-thenable`). Files outside all tsconfigs are still linted but only with syntax-level rules.

```ts
{
  languageOptions: {
    parserOptions: {
      project: ['./tsconfig.json', './packages/*/tsconfig.json'],
    },
  },
}
```

### settings

- **Type:** `Record<string, unknown>`

Shared settings accessible to all rules. Later entries override earlier ones.

## Config Merging

When multiple config entries match a file, they are merged in array order:

1. **Global ignores** — entries with only `ignores` exclude files from all rules
2. **Files matching** — entries whose `files` patterns don't match are skipped
3. **Entry-level ignores** — entries whose `ignores` match are skipped
4. **Rules** — shallow merge, later entries override earlier ones
5. **Plugins** — union from all matching entries
6. **Settings** — shallow merge
7. **Language options** — deep merge at field level

If no entry matches a file, no rules run for it. Explicit file arguments may still be reported as 0-rule results unless global ignores or `.gitignore` exclude them.

## JSON Configuration (Deprecated)

JSON config files (`rslint.json`, `rslint.jsonc`) are deprecated and will be removed in a future version. Run `rslint --init` to automatically migrate your JSON config to a JS/TS config. The migration preserves your custom rules and settings while deduplicating rules already covered by recommended presets.

Key difference: JSON configs automatically enable all core rules and declared plugin rules as `"error"`. JS/TS configs only enable rules explicitly declared in presets or the `rules` field.
