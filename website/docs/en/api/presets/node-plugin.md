# nodePlugin

`nodePlugin` exposes Rslint's built-in implementation of supported rules from [`eslint-plugin-n` 18.x](https://github.com/eslint-community/eslint-plugin-n/tree/v18.3.0). Rules use Rslint's native `node/*` prefix instead of upstream's `n/*`.

```ts
import { defineConfig, nodePlugin } from '@rslint/core';

export default defineConfig([nodePlugin.configs.recommended]);
```

## Presets

The camelCase keys below map to upstream's preset names:

| Preset                                 | Description                                     | View rules                                                          | Source                                                                                                                                                  |
| -------------------------------------- | ----------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `nodePlugin.configs.recommended`       | Node.js rules for mixed CommonJS and ES modules | [View rules →](/rules/?preset=nodePlugin.configs.recommended)       | [`eslint-plugin-n` `configs.recommended`](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/configs/recommended.js)                  |
| `nodePlugin.configs.recommendedModule` | Node.js rules for ES modules                    | [View rules →](/rules/?preset=nodePlugin.configs.recommendedModule) | [`eslint-plugin-n` `configs['recommended-module']`](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/configs/recommended-module.js) |
| `nodePlugin.configs.recommendedScript` | Node.js rules for CommonJS                      | [View rules →](/rules/?preset=nodePlugin.configs.recommendedScript) | [`eslint-plugin-n` `configs['recommended-script']`](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/configs/recommended-script.js) |

`recommended` selects the nearest valid `package.json` from `process.cwd()` and applies `.mjs`/`.cjs` overrides. The API's `cwd` does not change this lookup; use an explicit preset when the process directory differs from the lint project.

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
