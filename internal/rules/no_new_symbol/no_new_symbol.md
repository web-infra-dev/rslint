# no-new-symbol

## Rule Details

Disallows `new Symbol()`. `Symbol` is not intended to be used with the `new` operator, but to be called as a function. Calling `new Symbol()` throws a `TypeError` at runtime because `Symbol` is not a constructor.

Examples of **incorrect** code for this rule:

```javascript
var foo = new Symbol('foo');
new Symbol();
```

Examples of **correct** code for this rule:

```javascript
var foo = Symbol('foo');
```

## Original Documentation

- [ESLint: no-new-symbol](https://eslint.org/docs/latest/rules/no-new-symbol)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-new-symbol.js)
