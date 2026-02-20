# no-dynamic-delete

## Rule Details

Disallow the `delete` operator on computed property keys unless the key is a literal.

Examples of **incorrect** code for this rule:

```typescript
const container: { [i: string]: 0 } = {};
delete container[name];
delete container['aa' + 'b'];
```

Examples of **correct** code for this rule:

```typescript
const container: { [i: string]: 0 } = {};
delete container['name'];
delete container[7];
```

## Original Documentation

https://typescript-eslint.io/rules/no-dynamic-delete
