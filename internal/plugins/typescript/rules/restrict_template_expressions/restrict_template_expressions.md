# restrict-template-expressions

## Rule Details

Enforce template literal expressions to be of `string` type. When using template literal expressions (`${expr}`), non-string values are implicitly converted to strings using their `.toString()` method. This can lead to unexpected results such as `"[object Object]"` for objects or `"null"` for null values. This rule ensures that only values of type `string` are used in template expressions by default, though other types can be allowed via options.

Examples of **incorrect** code for this rule:

```typescript
const num = 42;
const str = `value: ${num}`;

const obj = {};
const msg = `result: ${obj}`;
```

Examples of **correct** code for this rule:

```typescript
const name = 'world';
const greeting = `Hello, ${name}`;

const num = 42;
const str = `value: ${String(num)}`;
const str2 = `value: ${num.toString()}`;
```

## Original Documentation

- [typescript-eslint restrict-template-expressions](https://typescript-eslint.io/rules/restrict-template-expressions)
