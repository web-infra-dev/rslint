# no-eq-null

## Rule Details

Comparing to `null` without a type-checking operator (`==` or `!=`) can have unintended results as the comparison will evaluate to `true` when comparing not just to `null`, but also to `undefined`.

The `no-eq-null` rule aims to reduce potential bugs and unwanted behavior by ensuring that comparisons to `null` only match `null`, and not also `undefined`. As such, it will flag comparisons to `null` when using `==` and `!=`.

Examples of **incorrect** code for this rule:

```javascript
if (foo == null) {
  bar();
}

while (qux != null) {
  baz();
}
```

Examples of **correct** code for this rule:

```javascript
if (foo === null) {
  bar();
}

while (qux !== null) {
  baz();
}
```

## Original Documentation

- [ESLint: no-eq-null](https://eslint.org/docs/latest/rules/no-eq-null)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-eq-null.js)
