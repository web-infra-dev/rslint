# no-extra-non-null-assertion

## Rule Details

Disallow extra non-null assertions.

The `!` non-null assertion operator in TypeScript is used to assert that a value's type does not include `null` or `undefined`. Using the operator any more than once on a single value does nothing.

Examples of **incorrect** code for this rule:

```typescript
const bar = foo!!.bar;
function foo(bar?: { n: number }) {
  return bar!?.n;
}
```

Examples of **correct** code for this rule:

```typescript
const bar = foo!.bar;
function foo(bar?: { n: number }) {
  return bar?.n;
}
```

## Original Documentation

https://typescript-eslint.io/rules/no-extra-non-null-assertion
