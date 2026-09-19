# vuePlugin

`vuePlugin` exposes Rslint's built-in implementation of supported rules from [`eslint-plugin-vue` 10.x](https://github.com/vuejs/eslint-plugin-vue). Its preset follows the upstream essential configuration for the rules Rslint currently supports.

```ts
import { defineConfig, vuePlugin } from '@rslint/core';

export default defineConfig([vuePlugin.configs.essential]);
```

Unlike every other preset, this one selects `**/*.vue`. A Vue single file component is a supported file type but is not part of Rslint's discovery baseline, so components are linted only where a config asks for them, either through this preset or through your own `files` entry.

## Presets

| Preset                        | Description                 | View rules                                                 | Source                                                                                             |
| ----------------------------- | --------------------------- | ---------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| `vuePlugin.configs.essential` | Vue rules that catch errors | [View rules →](/rules/?preset=vuePlugin.configs.essential) | [`eslint-plugin-vue` `flat/essential`](https://eslint.vuejs.org/user-guide/#bundle-configurations) |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
