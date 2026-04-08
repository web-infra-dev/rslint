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

To generate a default config, run:

```bash
rslint --init
```

## Basic Configuration

A typical TypeScript project configuration:

```ts
import { defineConfig, ts } from '@rslint/core';

export default defineConfig([
  // Global ignores — files excluded from all rules
  {
    ignores: ['**/dist/**', '**/fixtures/**'],
  },
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

Glob patterns specifying which files this config entry applies to. If omitted, the entry applies to all files matched by other entries.

The `files` field determines the **lint scope** — only files matching at least one entry's `files` pattern will be linted. This is independent of tsconfig's `include`: a file in tsconfig but not matching any `files` pattern will not be linted, while a file matching `files` but not in any tsconfig will still be linted with syntax-only rules (type-aware rules require tsconfig coverage).

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

If no entry matches a file, it is not linted.

## JSON Configuration (Deprecated)

JSON config files (`rslint.json`, `rslint.jsonc`) are deprecated and will be removed in a future version. Run `rslint --init` to automatically migrate your JSON config to a JS/TS config. The migration preserves your custom rules and settings while deduplicating rules already covered by recommended presets.

Key difference: JSON configs automatically enable all core rules and declared plugin rules as `"error"`. JS/TS configs only enable rules explicitly declared in presets or the `rules` field.
