# no-with

## Rule Details

Disallow `with` statements.

The `with` statement is potentially problematic because it adds members of an object to the current scope, making it impossible to tell what a variable inside the block actually refers to. In strict mode, `with` statements are not allowed at all.

Examples of **incorrect** code for this rule:

```javascript
with (point) {
  r = Math.sqrt(x * x + y * y); // is r a member of point?
}
```

Examples of **correct** code for this rule:

```javascript
const r = Math.sqrt(point.x * point.x + point.y * point.y);
```

## Original Documentation

- [ESLint no-with](https://eslint.org/docs/latest/rules/no-with)
