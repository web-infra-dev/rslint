# no-wrapper-object-types

## Rule Details

Disallow using the upper-cased built-in primitive class wrappers — `BigInt`, `Boolean`, `Number`, `Object`, `String`, and `Symbol` — as type names. The lower-cased primitive forms (`bigint`, `boolean`, `number`, `object`, `string`, `symbol`) are the safe choice in every case: primitives are compared by value and have predictable truthiness, while the wrapper objects are compared by reference and are always truthy.

A local declaration with the same name (a `type` alias, `interface`, or `class`) shadows the global wrapper, in which case the reference is no longer the unsafe built-in and is not reported.

Examples of **incorrect** code for this rule:

```typescript
let myBigInt: BigInt;
let myBoolean: Boolean;
let myNumber: Number;
let myString: String;
let mySymbol: Symbol;
let myObject: Object;

class MyClass implements Number {}

interface MyInterface extends Number {}
```

Examples of **correct** code for this rule:

```typescript
let myBigint: bigint;
let myBoolean: boolean;
let myNumber: number;
let myString: string;
let mySymbol: symbol;
let myObject: object;

type Number = 0 | 1;
let value: Number;
```

## Original Documentation

- [typescript-eslint no-wrapper-object-types](https://typescript-eslint.io/rules/no-wrapper-object-types)
