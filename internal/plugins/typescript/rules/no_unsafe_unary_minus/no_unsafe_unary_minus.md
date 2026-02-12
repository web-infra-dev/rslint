# no-unsafe-unary-minus

## Rule Details

Disallow unary negation of a value that is not a `number` or `bigint`.

Applying the unary minus operator (`-`) to a value that is not a number or bigint type is a likely mistake. JavaScript will coerce the value to a number, often resulting in `NaN`. This rule ensures the operand of unary negation is always `number` or `bigint`.

Examples of **incorrect** code for this rule:

```typescript
declare const str: string;
-str;

declare const bool: boolean;
-bool;

declare const obj: object;
-obj;
```

Examples of **correct** code for this rule:

```typescript
-42;
-someNumber;

declare const big: bigint;
-big;

declare const val: number | bigint;
-val;
```

## Original Documentation

- [typescript-eslint no-unsafe-unary-minus](https://typescript-eslint.io/rules/no-unsafe-unary-minus)
