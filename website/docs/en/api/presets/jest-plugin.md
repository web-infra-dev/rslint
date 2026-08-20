# jestPlugin

`jestPlugin` exposes Rslint's built-in implementation of supported rules from [`eslint-plugin-jest` 29.x](https://github.com/jest-community/eslint-plugin-jest/tree/v29.16.0). Its presets follow the corresponding upstream flat configurations for the rules Rslint currently supports.

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

| Preset                           | Description      | View rules                                                    | Source                                                                                                                               |
| -------------------------------- | ---------------- | ------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `jestPlugin.configs.recommended` | Jest rules       | [View rules →](/rules/?preset=jestPlugin.configs.recommended) | [`eslint-plugin-jest` `configs["flat/recommended"]`](https://github.com/jest-community/eslint-plugin-jest/tree/v29.16.0#recommended) |
| `jestPlugin.configs.style`       | Jest style rules | [View rules →](/rules/?preset=jestPlugin.configs.style)       | [`eslint-plugin-jest` `configs["flat/style"]`](https://github.com/jest-community/eslint-plugin-jest/tree/v29.16.0#style)             |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
