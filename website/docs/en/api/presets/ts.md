# ts

`ts` exposes Rslint's built-in implementation of supported [typescript-eslint 8.x](https://v8--typescript-eslint.netlify.app/rules/) rules and its baseline, type-checked, strict, and stylistic presets.

```ts
import { defineConfig, ts } from '@rslint/core';

export default defineConfig([ts.configs.recommended]);
```

`ts.configs.base` is setup-only, enables no rules, and is already included by every other TypeScript preset. Like upstream, these presets do not set `projectService` or `project`. Enable [`projectService: true`](/config/language-options#languageoptionsparseroptionsprojectservice) explicitly for automatic project discovery, or keep using explicit project paths.

## Presets

| Preset                              | Description                                          | View rules                                                       | Source                                                                                                                                    |
| ----------------------------------- | ---------------------------------------------------- | ---------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `ts.configs.base`                   | Declares the TypeScript plugin                       | —                                                                | [`typescript-eslint` `configs.base`](https://v8--typescript-eslint.netlify.app/users/configs/#base)                                       |
| `ts.configs.recommended`            | TypeScript recommended rules                         | [View rules →](/rules/?preset=ts.configs.recommended)            | [`typescript-eslint` `configs.recommended`](https://v8--typescript-eslint.netlify.app/users/configs/#recommended)                         |
| `ts.configs.recommendedTypeChecked` | TypeScript recommended rules, including typed ones   | [View rules →](/rules/?preset=ts.configs.recommendedTypeChecked) | [`typescript-eslint` `configs.recommendedTypeChecked`](https://v8--typescript-eslint.netlify.app/users/configs/#recommended-type-checked) |
| `ts.configs.strict`                 | TypeScript recommended rules plus opinionated extras | [View rules →](/rules/?preset=ts.configs.strict)                 | [`typescript-eslint` `configs.strict`](https://v8--typescript-eslint.netlify.app/users/configs/#strict)                                   |
| `ts.configs.strictTypeChecked`      | TypeScript strict rules, including typed ones        | [View rules →](/rules/?preset=ts.configs.strictTypeChecked)      | [`typescript-eslint` `configs.strictTypeChecked`](https://v8--typescript-eslint.netlify.app/users/configs/#strict-type-checked)           |
| `ts.configs.stylistic`              | TypeScript consistency rules                         | [View rules →](/rules/?preset=ts.configs.stylistic)              | [`typescript-eslint` `configs.stylistic`](https://v8--typescript-eslint.netlify.app/users/configs/#stylistic)                             |
| `ts.configs.stylisticTypeChecked`   | TypeScript consistency rules, including typed ones   | [View rules →](/rules/?preset=ts.configs.stylisticTypeChecked)   | [`typescript-eslint` `configs.stylisticTypeChecked`](https://v8--typescript-eslint.netlify.app/users/configs/#stylistic-type-checked)     |

`ts.configs.base` has no filtered rules view. Add one baseline and optionally one stylistic layer; you normally do not need to add `ts.configs.base` separately.

See [Rules & Presets](/config/rules-and-presets) for guidance on choosing and layering presets.
