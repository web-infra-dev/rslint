# reactPlugin

`reactPlugin` exposes Rslint's built-in implementation of supported rules from [`eslint-plugin-react` 7.x](https://github.com/jsx-eslint/eslint-plugin-react/tree/v7.37.5). Its preset follows the upstream flat recommended configuration for the rules Rslint currently supports.

```ts
import { defineConfig, reactPlugin } from '@rslint/core';

export default defineConfig([reactPlugin.configs.recommended]);
```

## Presets

| Preset                            | Description | View rules                                                     | Source                                                                                                                          |
| --------------------------------- | ----------- | -------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `reactPlugin.configs.recommended` | React rules | [View rules →](/rules/?preset=reactPlugin.configs.recommended) | [`eslint-plugin-react` `configs.flat.recommended`](https://github.com/jsx-eslint/eslint-plugin-react/tree/v7.37.5#flat-configs) |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
