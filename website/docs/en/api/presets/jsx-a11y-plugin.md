# jsxA11yPlugin

`jsxA11yPlugin` exposes Rslint's built-in implementation of supported rules from [`eslint-plugin-jsx-a11y` 6.x](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/tree/v6.10.2). Its preset follows the upstream flat recommended configuration for the rules Rslint currently supports.

```ts
import { defineConfig, jsxA11yPlugin } from '@rslint/core';

export default defineConfig([jsxA11yPlugin.configs.recommended]);
```

## Presets

| Preset                              | Description    | View rules                                                       | Source                                                                                                                                    |
| ----------------------------------- | -------------- | ---------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `jsxA11yPlugin.configs.recommended` | JSX a11y rules | [View rules →](/rules/?preset=jsxA11yPlugin.configs.recommended) | [`eslint-plugin-jsx-a11y` `flatConfigs.recommended`](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/tree/v6.10.2#shareable-configs) |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
