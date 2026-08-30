# Configuration File

Rslint uses JS/TS module configuration with a flat config array aligned with ESLint v10.

## Supported filenames

During automatic discovery, Rslint checks config files in the following order:

1. `rslint.config.js`
2. `rslint.config.mjs`
3. `rslint.config.ts`
4. `rslint.config.mts`

Automatic discovery does not consider `.cjs` or `.cts` config files. They can still be selected explicitly with `--config` or API `overrideConfigFile`.

## Config discovery

When you run `rslint`, it searches for a config file by walking **upward** from the target file or directory to the filesystem root. It uses the nearest candidate that loads successfully and falls back to an ancestor when a nearer candidate cannot be loaded.

- `rslint src/foo.ts` — searches from `src/` upward
- `rslint src/` — searches from `src/` upward
- `rslint` (no args) — searches from the current working directory upward

In a monorepo, different files can automatically use different config files based on their location:

```text
monorepo/
├── rslint.config.ts              ← root config
├── packages/
│   ├── foo/
│   │   ├── rslint.config.ts      ← used for files under foo/
│   │   └── src/
│   └── bar/
│       └── src/                   ← no config, inherits root
```

When linting from the monorepo root, Rslint automatically discovers all nested configs and applies the nearest one to each file.

### Global ignores and nested configs

For directory or no-argument lint runs, global ignores in a parent config prevent nested configs in ignored directories from contributing lint targets.

```ts
// monorepo/rslint.config.ts
export default defineConfig([
  // Global ignore — blocks directory target discovery in these directories
  { ignores: ['**/fixtures/**', 'e2e/**'] },
  js.configs.recommended,
  ts.configs.recommended,
]);
```

With this config, a `rslint.config.ts` inside `e2e/` or any `fixtures/` directory is not used by a root directory traversal. An explicitly named file is still resolved from its nearest config.

:::tip
Only **global ignore entries** (entries containing only `ignores`) block directory target discovery. Entry-level ignores do not affect config discovery. See [`ignores`](/config/ignoring-files) for the distinction.
:::

You can specify a config file explicitly, which overrides automatic discovery:

```bash
rslint --config path/to/rslint.config.ts .
```

Relative `files`, `ignores`, and `languageOptions.parserOptions.project` patterns are resolved from the config file's directory, whether the config is discovered automatically or supplied with `--config`.

To generate a default config, run:

```bash
rslint --init
```

## Basic configuration

A typical TypeScript project configuration:

```ts
import { defineConfig, globalIgnores, js, ts } from '@rslint/core';

export default defineConfig([
  // Files excluded from all rules
  globalIgnores(['**/dist/**', '**/fixtures/**']),
  // Presets with recommended rules
  js.configs.recommended,
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

:::tip
When using both JavaScript and TypeScript recommended presets, place `js.configs.recommended` before `ts.configs.recommended`. The TypeScript preset disables ESLint core rules that are handled by TypeScript-aware rules, and later config entries override earlier ones.
:::

See the [Configuration overview](/config/) for every available option and [Rules & Presets](/config/rules-and-presets) for the available presets.

## Config merging

When multiple config entries match a file, they are merged in array order:

1. **Global ignores** — entries containing only `ignores` remove files from the target set
2. **Selector union** — the implicit default baseline and effective explicit `files` entries decide whether the config selects the file
3. **Files matching** — entries whose explicit `files` patterns don't match are skipped; entries without `files` cascade across the selector union
4. **Entry-level ignores** — matching entries do not select or configure the file, but cannot remove a target selected elsewhere
5. **Rules** — later entries override earlier ones; a severity-only value retains earlier options
6. **Plugins** — union from all matching entries
7. **Settings** — ordinary nested objects merge recursively; arrays and scalar values are replaced
8. **Language options** — ordinary nested objects merge recursively; arrays and scalar values are replaced

If no entry matches a selected file, no lint rules run for it, but the file is still parsed and included in the result so parser diagnostics remain visible. This applies to default-baseline files found during directory discovery as well as explicitly requested supported files. Global ignores remove matching targets; CLI and JavaScript API runs apply `.gitignore` as an additional global ignore source.

## Migrating a legacy JSON configuration

Rslint no longer loads `rslint.json` or `rslint.jsonc` while linting. Passing one to `--config` is rejected; automatic discovery ignores those filenames.

Run `rslint --init` in a project that still has a legacy JSON/JSONC file to migrate it to a JS/TS module config. The migration preserves custom rules and settings while deduplicating rules already covered by recommended presets.
