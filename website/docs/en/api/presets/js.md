# js

`js` exposes presets for Rslint's built-in implementation of supported [ESLint 10.x core rules](https://eslint.org/docs/v10.x/rules/). Its preset is based on the configuration published by `@eslint/js` 10.x.

```ts
import { defineConfig, js } from '@rslint/core';

export default defineConfig([js.configs.recommended]);
```

## Presets

| Preset                   | Description                  | View rules                                            | Source                                                                                                                             |
| ------------------------ | ---------------------------- | ----------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `js.configs.recommended` | JavaScript recommended rules | [View rules →](/rules/?preset=js.configs.recommended) | [`@eslint/js` `configs.recommended`](https://eslint.org/docs/v10.x/use/configure/migration-guide#predefined-and-shareable-configs) |

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
