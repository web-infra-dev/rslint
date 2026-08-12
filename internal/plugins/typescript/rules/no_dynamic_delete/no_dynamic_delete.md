# no-dynamic-delete

## Rule Details

Disallow the `delete` operator on computed property keys unless the key is a literal.

Examples of **incorrect** code for this rule:

```typescript
const container: { [i: string]: 0 } = {};
delete container[name];
delete container['aa' + 'b'];
delete container[`name`];
```

Examples of **correct** code for this rule:

```typescript
const container: { [i: string]: 0 } = {};
delete container['name'];
delete container[7];
```

## Original Documentation

- [typescript-eslint: no-dynamic-delete](https://typescript-eslint.io/rules/no-dynamic-delete)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/no-dynamic-delete.ts)
