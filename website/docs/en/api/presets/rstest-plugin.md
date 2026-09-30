# rstestPlugin

`rstestPlugin` provides Rslint's [rules for Rstest](/rules/?group=rstest). Enable the recommended preset to catch common mistakes in test definitions, assertions, mocks, and snapshots.

## Usage

### Dedicated test files

For projects that keep tests in dedicated test and spec files, apply the preset to those files:

```ts
import { defineConfig, rstestPlugin } from '@rslint/core';

export default defineConfig([
  {
    ...rstestPlugin.configs.recommended,
    files: ['**/*.{test,spec}.{js,mjs,jsx,ts,tsx,mts}'],
  },
]);
```

### In-source tests

When using Rstest's [in-source testing](https://rstest.rs/config/test/include-source), tests live alongside production code and access the test API through `import.meta.rstest`. Apply the preset to every linted file to check both dedicated test files and in-source tests:

```ts
import { defineConfig, rstestPlugin } from '@rslint/core';

export default defineConfig([rstestPlugin.configs.recommended]);
```

No additional Rslint globals are required for imported APIs or `import.meta.rstest`.

Rstest does not expose global APIs by default. When the Rstest `globals` option is enabled, add the separate `env` preset alongside a rule preset:

```ts
import { defineConfig, rstestPlugin } from '@rslint/core';

export default defineConfig([
  rstestPlugin.configs.env,
  rstestPlugin.configs.recommended,
]);
```

The `style` preset contains the common style rules implemented for Jest, Vitest, and Rstest: `prefer-to-be`, `prefer-to-contain`, and `prefer-to-have-length`. The `all` preset configures every Rstest rule currently supported by Rslint. It follows `@vitest/eslint-plugin` severities for corresponding rules, including keeping conflicting rules disabled, and enables additional Rstest rules as warnings.

### Rstack CLI projects

Projects driven by the [Rstack CLI](https://rstack.rs/) import the test API from `rstack/test`, which re-exports it. Every rule recognizes that specifier alongside `@rstest/core`, so the preset needs no extra configuration and a file may use either or both:

```ts
import { describe, expect, test } from 'rstack/test';
```

### Configure test types separately

To use different rule settings for dedicated test files and in-source tests, create a configuration item for each group. Use `files` to match test and spec filenames in one item, and the source patterns from Rstest's [`includeSource`](https://rstest.rs/config/test/include-source) configuration in the other.

## Presets

| Preset                             | Description                                  | View rules                                                      |
| ---------------------------------- | -------------------------------------------- | --------------------------------------------------------------- |
| `rstestPlugin.configs.all`         | All Rstest rules, with 5 disabled by default | [View rules →](/rules/?preset=rstestPlugin.configs.all)         |
| `rstestPlugin.configs.env`         | Rstest globals                               |                                                                 |
| `rstestPlugin.configs.recommended` | Recommended Rstest rules                     | [View rules →](/rules/?preset=rstestPlugin.configs.recommended) |
| `rstestPlugin.configs.style`       | Rstest style rules                           | [View rules →](/rules/?preset=rstestPlugin.configs.style)       |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
