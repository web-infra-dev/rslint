# ts

`ts` contains Rslint's bundled TypeScript baseline, type-checked, strict, and stylistic presets.

```ts
import { defineConfig, ts } from '@rslint/core';

export default defineConfig([ts.configs.recommended]);
```

`ts.configs.base` is setup-only, enables no rules, and is already included by every other TypeScript preset.

## Presets

| Preset                              | Description                                                | View rules                                                       |
| ----------------------------------- | ---------------------------------------------------------- | ---------------------------------------------------------------- |
| `ts.configs.base`                   | Declares the TypeScript plugin and enables project service | —                                                                |
| `ts.configs.recommended`            | TypeScript recommended rules                               | [View rules →](/rules/?preset=ts.configs.recommended)            |
| `ts.configs.recommendedTypeChecked` | TypeScript recommended rules, including typed ones         | [View rules →](/rules/?preset=ts.configs.recommendedTypeChecked) |
| `ts.configs.strict`                 | TypeScript recommended rules plus opinionated extras       | [View rules →](/rules/?preset=ts.configs.strict)                 |
| `ts.configs.strictTypeChecked`      | TypeScript strict rules, including typed ones              | [View rules →](/rules/?preset=ts.configs.strictTypeChecked)      |
| `ts.configs.stylistic`              | TypeScript consistency rules                               | [View rules →](/rules/?preset=ts.configs.stylistic)              |
| `ts.configs.stylisticTypeChecked`   | TypeScript consistency rules, including typed ones         | [View rules →](/rules/?preset=ts.configs.stylisticTypeChecked)   |

`ts.configs.base` has no filtered rules view. Add one baseline and optionally one stylistic layer; you normally do not need to add `ts.configs.base` separately.

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
