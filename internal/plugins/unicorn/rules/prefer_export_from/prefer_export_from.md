# prefer-export-from

## Rule details

Prefer `export…from` when forwarding bindings from another module. The rule
checks named, default and namespace imports, including single `export const`
aliases whose local name is unused. An attached JSDoc comment keeps such an
alias in place.

Example of **incorrect** code for this rule.

```javascript
import { parse } from './parser.js';
export { parse };
```

Example of **correct** code for this rule.

```javascript
export { parse } from './parser.js';
```

## Fixes and suggestions

Fixes keep imports with other references and value imports that are re-exported
only as types. Comments inside the declarations limit the rule to a diagnostic.

An `export…from` declaration exposes a live binding. When a fix replaces
`export default value` or `export const alias = value`, updates to the source
binding become visible to consumers of the re-export.

## Options

### checkUsedVariables

Type `boolean`. Default `true`.

Set this option to `false` to skip an entire import declaration when any of its
bindings is used outside a re-export. When the declaration contains both
re-exported and unused bindings, the rule offers a suggestion for the re-exports.

```json
{
  "unicorn/prefer-export-from": ["error", { "checkUsedVariables": false }]
}
```

## Differences from upstream

- Fixes preserve dependency evaluation order. If a file imports `./setup.js`
  before `./consumer.js`, the direct re-export of `./setup.js` stays before
  `./consumer.js`. Re-exports with different `with` attributes stay separate.
- Re-exporting `ns` from `import main, * as ns from 'lib'` keeps the default
  import valid. Imports used as JSX factories also stay in place, including
  custom factories and imports used after a local binding shadows the factory
  elsewhere in the file.
- With `verbatimModuleSyntax` enabled in the linted project, `import {type T}`
  stays in place when it is the only binding in the declaration. The import
  can still load the dependency at runtime.
- Re-exports inside `declare module` and re-exports of `import defer` bindings
  receive diagnostics only. Direct `eval()` also blocks edits that would remove
  a local binding it could read.
- For files using legacy `assert { … }` import assertions, switch to
  `with { … }` import attributes before applying the rule.

## Original documentation

- [eslint-plugin-unicorn: prefer-export-from](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-export-from.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-export-from.js)
