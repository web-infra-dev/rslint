# no-ternary

## Rule Details

This rule disallows ternary operators.

Examples of **incorrect** code for this rule:

```javascript
const foo = isBar ? baz : qux;

function quux() {
  return foo ? bar() : baz();
}
```

Examples of **correct** code for this rule:

```javascript
let foo;

if (isBar) {
    foo = baz;
} else {
    foo = qux;
}

function quux() {
    if (foo) {
        return bar();
    } else {
        return baz();
    }
}
```

## Original Documentation

- [ESLint: no-ternary](https://eslint.org/docs/latest/rules/no-ternary)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-ternary.js)
