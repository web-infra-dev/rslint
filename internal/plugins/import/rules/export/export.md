# export

## Rule Details

Reports duplicate named or default exports, including conflicts introduced by `export *`. Every conflicting occurrence is reported because the rule cannot determine which export was intended.

Examples of **incorrect** code for this rule:

```javascript
export default class MyClass {}
export default makeClass;
```

```javascript
export const foo = function () {};

function bar() {}
export { bar as foo };
```

Examples of **correct** code for this rule:

```javascript
export const foo = "foo";
export const bar = "bar";
```

```typescript
export const Foo = 1;
export type Foo = number;

export function parse(value: string): string;
export function parse(value: number): string;
export function parse(value: string | number): string {
  return String(value);
}
```

The rule checks each TypeScript namespace or ambient module declaration separately. It permits function overload signatures and namespace merging with functions, classes, and enums, following upstream behavior.

A star re-export also reports when its target has no named exports. Default exports are excluded from `export *`. Unresolved, ignored, and non-ES modules are skipped.

## Options

This rule has no options. It does not provide automatic fixes or suggestions.

## Differences from upstream

Diagnostics for invalid dependency syntax can differ from ESLint. For example, if `dependency.js` contains `return; export {};`, then `export * from "./dependency.js"` reports "No named exports found" in Rslint, while ESLint's default configuration reports a parse error. Check and fix the dependency's syntax before investigating missing exports.

## Original Documentation

- [eslint-plugin-import: export](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/export.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/export.js)
