# prefer-for-of

## Rule Details

Enforce the use of `for-of` loop over the standard `for` loop where possible.

Many developers default to writing `for (let i = 0; i < ...; i++)` loops to iterate over arrays. However, in many cases the loop iterator variable is only used to access the respective element of the array. In such cases, a `for-of` loop is simpler and more readable.

This rule will report when a `for` loop can be replaced with a `for-of` loop.

Examples of **incorrect** code for this rule:

```javascript
declare const array: string[];

for (let i = 0; i < array.length; i++) {
  console.log(array[i]);
}
```

Examples of **correct** code for this rule:

```javascript
// for-of loop
for (const x of array) {
  console.log(x);
}

// Index variable is used for more than just array access
for (let i = 0; i < array.length; i++) {
  console.log(i, array[i]);
}

// Array element is being assigned
for (let i = 0; i < array.length; i++) {
  array[i] = 0;
}
```

## Original Documentation

[typescript-eslint: prefer-for-of](https://typescript-eslint.io/rules/prefer-for-of)
