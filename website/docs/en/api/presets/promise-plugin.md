# promisePlugin

`promisePlugin` exposes Rslint's built-in implementation of supported rules from [`eslint-plugin-promise` 7.x](https://github.com/eslint-community/eslint-plugin-promise/tree/v7.3.0). Its preset follows the upstream flat recommended configuration for the rules Rslint currently supports.

```ts
import { defineConfig, promisePlugin } from '@rslint/core';

export default defineConfig([promisePlugin.configs.recommended]);
```

## Presets

| Preset                              | Description   | View rules                                                       | Source                                                                                                                               |
| ----------------------------------- | ------------- | ---------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `promisePlugin.configs.recommended` | Promise rules | [View rules →](/rules/?preset=promisePlugin.configs.recommended) | [`eslint-plugin-promise` `configs["flat/recommended"]`](https://github.com/eslint-community/eslint-plugin-promise/tree/v7.3.0#usage) |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
