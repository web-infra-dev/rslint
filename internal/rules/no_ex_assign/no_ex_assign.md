# no-ex-assign

## Rule Details

Disallow assign another value to the exception parameter a catch clause in a try statement accidentally or purposely. Since there is no arguments object to offer alternative access to this data, assignment of the parameter is absolutely destructive.

Examples of **correct** code for this rule:

```javascript
try {
} catch (e) {
  three = 2 + 1;
}
try {
} catch ({ e }) {
  this.something = 2;
}
function foo() {
  try {
  } catch (e) {
    return false;
  }
}
```

Examples of **incorrect** code for this rule:

```javascript
try {
} catch (e) {
  e = 10;
}
try {
} catch (ex) {
  ex = 10;
}
try {
} catch (ex) {
  [ex] = [];
}
try {
} catch (ex) {
  ({ x: ex = 0 } = {});
}
try {
} catch ({ message }) {
  message = 10;
}
```

## Original Documentation

- [ESLint: no-ex-assign](https://eslint.org/docs/latest/rules/no-ex-assign)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-ex-assign.js)
