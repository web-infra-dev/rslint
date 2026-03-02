# restrict-plus-operands

## Rule Details

Require both operands of addition to be the same type and be `bigint`, `number`, or `string`. The `+` operator in TypeScript can be used for both addition and string concatenation. This rule ensures that operands of `+` are both numbers, both bigints, or both strings, preventing accidental implicit type coercions that can lead to unexpected results (e.g., `"1" + 2` becoming `"12"` instead of `3`).

Examples of **incorrect** code for this rule:

```typescript
const result = 1 + '2';
const bad = 1n + 2;
const invalid = {} + [];
```

Examples of **correct** code for this rule:

```typescript
const sum = 1 + 2;
const concat = 'a' + 'b';
const bigSum = 1n + 2n;
const explicit = String(1) + '2';
```

## Original Documentation

- [typescript-eslint restrict-plus-operands](https://typescript-eslint.io/rules/restrict-plus-operands)
