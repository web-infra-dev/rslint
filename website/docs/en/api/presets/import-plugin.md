# importPlugin

`importPlugin` exposes Rslint's built-in implementation of supported rules from [`eslint-plugin-import` 2.x](https://github.com/import-js/eslint-plugin-import/tree/v2.32.0). Its preset follows the upstream flat recommended configuration for the rules Rslint currently supports.

```ts
import { defineConfig, importPlugin } from '@rslint/core';

export default defineConfig([importPlugin.configs.recommended]);
```

## Presets

| Preset                             | Description         | View rules                                                      | Source                                                                                                                                          |
| ---------------------------------- | ------------------- | --------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `importPlugin.configs.recommended` | Import/export rules | [View rules →](/rules/?preset=importPlugin.configs.recommended) | [`eslint-plugin-import` `flatConfigs.recommended`](https://github.com/import-js/eslint-plugin-import/tree/v2.32.0#config---flat-eslintconfigjs) |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
