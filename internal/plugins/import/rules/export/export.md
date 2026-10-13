# export

## Rule Details

Reports duplicate named or default exports, including conflicts introduced by `export *`. Every conflicting occurrence is reported because the rule cannot determine which export was intended.

Explicit value exports take precedence over names from `export *`, regardless of statement order. If a name is explicitly exported more than once, only those explicit exports are reported. Star conflicts for other names are still reported.

This rule is enabled in `importPlugin.configs.recommended`.

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

```javascript
// Override value while forwarding the other exports from base.js.
const value = 2;
export { value };
export * from "./base.js";
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

Unlike eslint-plugin-import v2.32.0, Rslint permits explicit value exports to shadow star exports, following [ECMAScript ResolveExport](https://tc39.es/ecma262/multipage/ecmascript-language-scripts-and-modules.html#sec-resolveexport). The example above is valid even when `base.js` exports `value` or several star exports provide that name.

Namespace re-exports such as `export * as ns from "./base.js"` and `export * as "ns" from "./base.js"` declare the single explicit name `ns`. They can shadow a star export named `ns`, and duplicate explicit uses of that name are reported. Upstream skips identifier aliases and treats string aliases as star exports.

Export classification is syntactic: explicit `export type` declarations and `type` specifiers cannot resolve runtime star conflicts. Type-only namespace aliases use the type namespace; the existing upstream grouping of type and value names is otherwise preserved.

Diagnostics for invalid dependency syntax can differ from ESLint. For example, if `dependency.js` contains `return; export {};`, then `export * from "./dependency.js"` reports "No named exports found" in Rslint, while ESLint's default configuration reports a parse error. Check and fix the dependency's syntax before investigating missing exports.

## Original Documentation

- [eslint-plugin-import: export](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/export.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/export.js)
