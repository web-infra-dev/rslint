# rstestPlugin

`rstestPlugin` is Rslint's [built-in Rstest-specific plugin](/rules/?group=rstest).

```ts
import { defineConfig, rstestPlugin } from '@rslint/core';

export default defineConfig([
  {
    ...rstestPlugin.configs.recommended,
    files: ['**/*.{test,spec}.{js,mjs,jsx,ts,tsx,mts}'],
  },
]);
```

`rstestPlugin.configs.recommended` does not declare `files`; the example scopes it to test and spec files.

## Presets

| Preset                             | Description  | View rules                                                      |
| ---------------------------------- | ------------ | --------------------------------------------------------------- |
| `rstestPlugin.configs.recommended` | Rstest rules | [View rules →](/rules/?preset=rstestPlugin.configs.recommended) |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
