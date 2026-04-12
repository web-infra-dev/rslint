# no-mutable-exports

## Rule Details

Forbids the use of mutable exports with `var` or `let`.

Mutable exports can lead to hard-to-understand code because importers might not expect the exported value to change after import. Use `const` for exported values, or export functions/classes instead.

Examples of **incorrect** code for this rule:

```javascript
export let count = 1;
export var count = 1;

let count = 1;
export { count };
```

Examples of **correct** code for this rule:

```javascript
export const count = 1;
export function getCount() {}
export class Counter {}

const count = 1;
export { count };
```

## Original Documentation

https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/no-mutable-exports.md
