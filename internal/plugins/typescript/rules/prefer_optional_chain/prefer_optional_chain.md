# prefer-optional-chain

## Rule Details

Enforce using concise optional chain expressions instead of chained logical ANDs, negated logical ORs, or empty object coalescing patterns.

TypeScript 3.7 introduced optional chaining (`?.`) which provides a more concise and readable way to access deeply nested properties that may be null or undefined.

Examples of **incorrect** code for this rule:

```typescript
foo && foo.bar;
foo && foo.bar && foo.bar.baz;
foo && foo.bar();
foo != null && foo.bar;
foo !== undefined && foo.bar;
typeof foo !== 'undefined' && foo.bar;
!foo || !foo.bar;
(foo || {}).bar;
(foo ?? {}).bar;
```

Examples of **correct** code for this rule:

```typescript
foo?.bar;
foo?.bar?.baz;
foo?.bar();
foo?.bar;
foo?.bar;
foo?.bar;
!foo?.bar;
foo?.bar;
foo?.bar;
```

## Options

### `allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing`

Type: `boolean`, default: `false`

When set to `true`, the rule will provide auto-fixes even when the fix may change the return type of the expression. By default, such cases are reported as suggestions only.

### `checkAny`

Type: `boolean`, default: `true`

When set to `true`, the rule will check operands typed as `any` when inspecting boolean expressions.

### `checkUnknown`

Type: `boolean`, default: `true`

When set to `true`, the rule will check operands typed as `unknown` when inspecting boolean expressions.

### `checkString`

Type: `boolean`, default: `true`

When set to `true`, the rule will check operands typed as `string` when inspecting boolean expressions.

### `checkNumber`

Type: `boolean`, default: `true`

When set to `true`, the rule will check operands typed as `number` when inspecting boolean expressions.

### `checkBoolean`

Type: `boolean`, default: `true`

When set to `true`, the rule will check operands typed as `boolean` when inspecting boolean expressions.

### `checkBigInt`

Type: `boolean`, default: `true`

When set to `true`, the rule will check operands typed as `bigint` when inspecting boolean expressions.

### `requireNullish`

Type: `boolean`, default: `false`

When set to `true`, the rule will only report on expressions where at least one operand has a type that includes `null` or `undefined`. This prevents false positives when the expression uses truthy checks for non-nullable types like `string` or `number`.

## Original Documentation

- [typescript-eslint prefer-optional-chain](https://typescript-eslint.io/rules/prefer-optional-chain)
