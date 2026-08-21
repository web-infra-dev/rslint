# no-continue

## Rule Details

Disallows `continue` statements. When used incorrectly, `continue` makes code less testable, less readable, and less maintainable. Structured control flow statements such as `if` should be used instead.

Examples of **incorrect** code for this rule:

```javascript
let sum = 0,
  i;

for (i = 0; i < 10; i++) {
  if (i >= 5) {
    continue;
  }

  sum += i;
}
```

```javascript
let sum = 0,
  i;

labeledLoop: for (i = 0; i < 10; i++) {
  if (i >= 5) {
    continue labeledLoop;
  }

  sum += i;
}
```

Examples of **correct** code for this rule:

```javascript
let sum = 0,
  i;

for (i = 0; i < 10; i++) {
  if (i < 5) {
    sum += i;
  }
}
```

## Original Documentation

- [ESLint: no-continue](https://eslint.org/docs/latest/rules/no-continue)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-continue.js)
