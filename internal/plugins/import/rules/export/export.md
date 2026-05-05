# export

## Rule Details

Reports any invalid exports — that is, re-exports of the same name. Multiple
default exports and duplicate named exports are usually copy/paste mistakes
and at best leave the consumer ambiguous about which value they will receive.

The rule also handles TypeScript's declaration-merging semantics, so the
following pairs do not conflict:

- A value and a type with the same name (`export const Foo` plus
  `export type Foo` / `export interface Foo`).
- A namespace merged with a single class, single enum, or any number of
  function overloads.

Each TypeScript namespace (and each ambient `declare module 'name' { ... }`
block) forms its own scope, so identical names in sibling namespaces are not
treated as duplicates.

`export * from './mod'` is expanded against the upstream module's named
exports, so a duplicate introduced by mixing a re-export-all with a local
export of the same name is detected. When the upstream module exposes only a
default export (no named ones), the re-export-all line itself is reported. If
the upstream module fails to parse, the re-export-all is reported with the
underlying parser diagnostics.

Examples of **incorrect** code for this rule:

```javascript
export default class MyClass {}
export default function makeClass() {}
```

```javascript
export const foo = function () {};
export { bar as foo };
```

```typescript
export type Foo = string;
export type Foo = number;
```

```typescript
export class Foo {}
export class Foo {}
export namespace Foo {}
```

```javascript
// upstream './other' only exports `default`
export * from './other';
// → No named exports found in module './other'.
```

Examples of **correct** code for this rule:

```javascript
export default class MyClass {}
export const helper = () => {};
```

```typescript
export const Foo = 1;
export type Foo = number;
```

```typescript
export class Foo {}
export namespace Foo {
  export class Bar {}
}
```

```typescript
export function fff(a: string): void;
export function fff(a: number): void;
export function fff(a: string | number): void {}
```

## Differences from ESLint

The `Parse errors in imported module 'X': ...` diagnostic may differ from
ESLint in:

- The wording of each parse error.
- The number of parse errors reported for the same file (you may see every
  parse error instead of only the first).
- The reported `(line:col)` for the same parse error.

All other diagnostic messages and report positions match the upstream rule.

## Original Documentation

- [import/export](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/export.md)
