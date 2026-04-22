# Rules & Presets

## Available Presets

| Preset                             | Description                                               |                                                                 |
| ---------------------------------- | --------------------------------------------------------- | --------------------------------------------------------------- |
| `ts.configs.recommended`           | TypeScript recommended rules (includes ESLint core rules) | [View rules →](/rules/?preset=ts.configs.recommended)           |
| `js.configs.recommended`           | JavaScript recommended rules                              | [View rules →](/rules/?preset=js.configs.recommended)           |
| `reactPlugin.configs.recommended`  | React rules                                               | [View rules →](/rules/?preset=reactPlugin.configs.recommended)  |
| `importPlugin.configs.recommended` | Import/export rules                                       | [View rules →](/rules/?preset=importPlugin.configs.recommended) |

Import presets from `@rslint/core`:

```ts
import { defineConfig, ts, js, reactPlugin, importPlugin } from '@rslint/core';
```

## rules

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

## plugins

- **Type:** `string[]`

Plugin names to enable. Available plugins:

| Plugin               | Rules Prefix           |
| -------------------- | ---------------------- |
| `@typescript-eslint` | `@typescript-eslint/*` |
| `import`             | `import/*`             |
| `jest`               | `jest/*`               |
| `promise`            | `promise/*`            |
| `react`              | `react/*`              |

ESLint core rules (e.g. `no-unused-vars`, `prefer-const`, `no-var`) have no prefix and do not belong to any plugin — they can be enabled directly in `rules` without listing anything in `plugins`.

:::tip
When using JS/TS config with presets (e.g., `ts.configs.recommended`), plugins are declared within the preset — you don't need to specify them separately.
:::

For a full list of available rules and their options, see the [Rules](/rules/) reference.
