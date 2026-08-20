# unicornPlugin

`unicornPlugin` exposes Rslint's built-in implementation of supported rules from [`eslint-plugin-unicorn` 73.x](https://github.com/sindresorhus/eslint-plugin-unicorn/tree/v73.0.0). Its preset follows the upstream recommended configuration for the rules Rslint currently supports.

```ts
import { defineConfig, unicornPlugin } from '@rslint/core';

export default defineConfig([unicornPlugin.configs.recommended]);
```

## Presets

| Preset                              | Description   | View rules                                                       | Source                                                                                                                                 |
| ----------------------------------- | ------------- | ---------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `unicornPlugin.configs.recommended` | Unicorn rules | [View rules →](/rules/?preset=unicornPlugin.configs.recommended) | [`eslint-plugin-unicorn` `configs.recommended`](https://github.com/sindresorhus/eslint-plugin-unicorn/tree/v73.0.0#recommended-config) |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
