# importPlugin

`importPlugin` contains Rslint's bundled import/export preset.

```ts
import { defineConfig, importPlugin } from '@rslint/core';

export default defineConfig([importPlugin.configs.recommended]);
```

## Presets

| Preset                             | Description         | View rules                                                      |
| ---------------------------------- | ------------------- | --------------------------------------------------------------- |
| `importPlugin.configs.recommended` | Import/export rules | [View rules →](/rules/?preset=importPlugin.configs.recommended) |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
