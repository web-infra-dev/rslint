# no-array-delete

## Rule Details

Disallow using the `delete` operator on array values. Using `delete` on an array element sets it to `undefined` and leaves a hole in the array without changing its length. This is almost always a mistake -- `Array.prototype.splice()` should be used instead to remove elements.

The rule uses type information to detect when the `delete` target is an array or tuple type.

Examples of **incorrect** code for this rule:

```typescript
const arr = [1, 2, 3];
delete arr[1]; // arr is now [1, undefined, 3]

const tuple: [string, number] = ['a', 1];
delete tuple[0];
```

Examples of **correct** code for this rule:

```typescript
const arr = [1, 2, 3];
arr.splice(1, 1); // arr is now [1, 3]

const obj: Record<string, number> = { a: 1 };
delete obj['a']; // objects are fine
```

## Original Documentation

- [typescript-eslint no-array-delete](https://typescript-eslint.io/rules/no-array-delete)
