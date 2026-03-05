# no-for-in-array

## Rule Details

Disallows iterating over arrays with a `for-in` loop. Using `for-in` on arrays is problematic because it skips holes, returns indices as strings rather than numbers, and may visit inherited enumerable properties from the prototype chain. Use `for-of`, `Array.prototype.forEach`, or a standard `for` loop instead.

Examples of **incorrect** code for this rule:

```typescript
const arr = [1, 2, 3];
for (const index in arr) {
  console.log(index); // "0", "1", "2" (strings, not numbers)
}

for (const key in ['a', 'b', 'c']) {
  console.log(key);
}
```

Examples of **correct** code for this rule:

```typescript
const arr = [1, 2, 3];
for (const value of arr) {
  console.log(value);
}

arr.forEach((value, index) => {
  console.log(index, value);
});

for (let i = 0; i < arr.length; i++) {
  console.log(arr[i]);
}
```

## Original Documentation

- [typescript-eslint no-for-in-array](https://typescript-eslint.io/rules/no-for-in-array)
