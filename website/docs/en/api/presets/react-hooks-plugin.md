# reactHooksPlugin

`reactHooksPlugin` exposes Rslint's built-in implementation of supported rules from [`eslint-plugin-react-hooks` 7.x](https://react.dev/reference/eslint-plugin-react-hooks). Its preset follows the upstream flat recommended configuration for the rules Rslint currently supports.

```ts
import { defineConfig, reactHooksPlugin } from '@rslint/core';

export default defineConfig([reactHooksPlugin.configs.recommended]);
```

## Presets

| Preset                                 | Description       | View rules                                                          | Source                                                                                                                                                                                             |
| -------------------------------------- | ----------------- | ------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `reactHooksPlugin.configs.recommended` | React Hooks rules | [View rules →](/rules/?preset=reactHooksPlugin.configs.recommended) | [`eslint-plugin-react-hooks` `configs.flat.recommended`](https://github.com/facebook/react/tree/eslint-plugin-react-hooks%407.1.1/packages/eslint-plugin-react-hooks#flat-config-eslintconfigjsts) |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
