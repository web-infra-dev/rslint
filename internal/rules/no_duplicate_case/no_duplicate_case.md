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

https://eslint.org/docs/latest/rules/no-duplicate-case
