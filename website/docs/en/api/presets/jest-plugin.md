# jestPlugin

`jestPlugin` contains Rslint's bundled Jest presets.

```ts
import { defineConfig, jestPlugin } from '@rslint/core';

export default defineConfig([
  {
    ...jestPlugin.configs.recommended,
    files: ['**/*.{test,spec}.{js,mjs,jsx,ts,tsx,mts}'],
  },
]);
```

`jestPlugin.configs.recommended` does not declare `files`; the example scopes it to test and spec files.

## Presets

| Preset                           | Description      | View rules                                                    |
| -------------------------------- | ---------------- | ------------------------------------------------------------- |
| `jestPlugin.configs.recommended` | Jest rules       | [View rules →](/rules/?preset=jestPlugin.configs.recommended) |
| `jestPlugin.configs.style`       | Jest style rules | [View rules →](/rules/?preset=jestPlugin.configs.style)       |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
