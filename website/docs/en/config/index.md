# Configuration

Rslint uses a flat config format (an array of config entries), aligned with ESLint v10. JS/TS configuration files are the recommended approach.

## Configuration Files

Rslint looks for config files in the following order:

1. `rslint.config.js`
2. `rslint.config.mjs`
3. `rslint.config.ts`
4. `rslint.config.mts`

You can also specify a config file explicitly:

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

### Available Presets

| Preset                             | Description                                               |
| ---------------------------------- | --------------------------------------------------------- |
| `ts.configs.recommended`           | TypeScript recommended rules (includes ESLint core rules) |
| `js.configs.recommended`           | JavaScript recommended rules                              |
| `reactPlugin.configs.recommended`  | React rules                                               |
| `importPlugin.configs.recommended` | Import/export rules                                       |

Import presets from `@rslint/core`:

```ts
import { defineConfig, ts, js, reactPlugin, importPlugin } from '@rslint/core';
```

## Configuration Options

### files

- **Type:** `string[]`

Glob patterns specifying which files this config entry applies to. If omitted, the entry applies to all files matched by other entries.

```ts
{
  files: ['**/*.ts', '**/*.tsx'],
  rules: { /* ... */ },
}
```

### ignores

- **Type:** `string[]`

Glob patterns for files to exclude. An entry with **only** `ignores` (no other fields) acts as a global ignore — matching files are excluded from all rules.

```ts
// Global ignore entry
{
  ignores: ['**/dist/**', '**/fixtures/**'],
}

// Entry-level ignore (only applies to this entry)
{
  files: ['**/*.ts'],
  ignores: ['**/*.test.ts'],
  rules: { /* ... */ },
}
```

:::tip
`node_modules` is automatically excluded by rslint — you don't need to add it to ignores.
:::

### rules

- **Type:** `Record<string, RuleSeverity | [RuleSeverity, ...options]>`

Configure individual rules with a severity level and optional options.

**Severity levels:**

| Value     | Description                                 |
| --------- | ------------------------------------------- |
| `"error"` | Reports as error, causes non-zero exit code |
| `"warn"`  | Reports as warning                          |
| `"off"`   | Disables the rule                           |

**String format** (severity only):

```ts
rules: {
  '@typescript-eslint/no-explicit-any': 'error',
  '@typescript-eslint/require-await': 'off',
}
```

**Array format** (severity + options):

```ts
rules: {
  '@typescript-eslint/array-type': ['warn', { default: 'array-simple' }],
  '@typescript-eslint/no-unused-vars': ['error', {
    argsIgnorePattern: '^_',
    varsIgnorePattern: '^_',
  }],
}
```

### plugins

- **Type:** `string[]`

Plugin names to enable. Available plugins:

| Plugin                 | Rules Prefix           |
| ---------------------- | ---------------------- |
| `@typescript-eslint`   | `@typescript-eslint/*` |
| `eslint-plugin-import` | `import/*`             |
| `react`                | `react/*`              |

:::tip
When using JS/TS config with presets (e.g., `ts.configs.recommended`), plugins are declared within the preset — you don't need to specify them separately.
:::

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

Explicit tsconfig.json paths. Supports glob patterns for monorepos.

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
