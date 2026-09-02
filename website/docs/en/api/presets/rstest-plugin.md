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

No additional Rslint globals are required for `import.meta.rstest`.

### Configure test types separately

To use different rule settings for dedicated test files and in-source tests, create a configuration item for each group. Use `files` to match test and spec filenames in one item, and the source patterns from Rstest's [`includeSource`](https://rstest.rs/config/test/include-source) configuration in the other.

## Presets

| Preset                             | Description  | View rules                                                      |
| ---------------------------------- | ------------ | --------------------------------------------------------------- |
| `rstestPlugin.configs.recommended` | Rstest rules | [View rules →](/rules/?preset=rstestPlugin.configs.recommended) |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
