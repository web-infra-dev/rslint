# prefer-const

## Rule Details

Requires `const` declarations for variables that are never reassigned after declared. If a variable is never reassigned, using the `const` declaration is better because it makes the intent clear that the value is not intended to be changed.

Examples of **incorrect** code for this rule:

```javascript
let x = 1;
let obj = { key: 0 };
for (let x in obj) {
  console.log(x);
}
for (let x of [1, 2, 3]) {
  console.log(x);
}
```

Examples of **correct** code for this rule:

```javascript
const x = 1;
const obj = { key: 0 };
let y = 1;
y = 2;
let z;
z = 1;
for (const x in obj) {
  console.log(x);
}
for (const x of [1, 2, 3]) {
  console.log(x);
}
```

## Original Documentation

- [ESLint prefer-const](https://eslint.org/docs/latest/rules/prefer-const)
