# ban-types

## Rule Details

Disallow certain types. Some built-in TypeScript types like `String`, `Boolean`, `Number`, `Symbol`, `BigInt`, `Object`, `Function`, and `{}` are almost always mistakes when used as types. This rule bans these problematic types and suggests preferred alternatives.

The default banned types include the uppercase wrapper types (`String` -> `string`, `Boolean` -> `boolean`, etc.) as well as `Function`, `Object`, and `{}`. Custom types can be added or defaults can be overridden.

Examples of **incorrect** code for this rule:

```typescript
const str: String = 'hello';
const num: Number = 42;
const obj: Object = {};
const fn: Function = () => {};
```

Examples of **correct** code for this rule:

```typescript
const str: string = 'hello';
const num: number = 42;
const obj: object = {};
const fn: () => void = () => {};
```

## Original Documentation

- [typescript-eslint ban-types](https://typescript-eslint.io/rules/ban-types)
