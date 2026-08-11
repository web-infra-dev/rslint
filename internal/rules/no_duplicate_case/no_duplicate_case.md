# no-duplicate-case

## Rule Details

Disallow duplicate case labels in `switch` statements. Duplicate case labels indicate a probable mistake.

Examples of **incorrect** code for this rule:

```javascript
switch (a) {
  case 1:
    break;
  case 1:
    break;
}

switch (a) {
  case 'a':
    break;
  case 'a':
    break;
}
```

Examples of **correct** code for this rule:

```javascript
switch (a) {
  case 1:
    break;
  case 2:
    break;
}

switch (a) {
  case 'a':
    break;
  case 'b':
    break;
}
```

## Original Documentation

- [ESLint: no-duplicate-case](https://eslint.org/docs/latest/rules/no-duplicate-case)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-duplicate-case.js)

https://eslint.org/docs/latest/rules/no-duplicate-case
