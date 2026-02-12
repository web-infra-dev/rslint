# no-compare-neg-zero

## Rule Details

Disallows comparing against `-0` using equality and relational operators (`==`, `===`, `!=`, `!==`, `>`, `>=`, `<`, `<=`). Comparing directly to `-0` does not work as intended because `+0 === -0` is `true`. To check whether a value is `-0`, use `Object.is(x, -0)` instead.

Examples of **incorrect** code for this rule:

```javascript
if (x === -0) {
}

if (x == -0) {
}

if (x > -0) {
}

if (x !== -0) {
}
```

Examples of **correct** code for this rule:

```javascript
if (x === 0) {
}

if (Object.is(x, -0)) {
}

if (x > 0) {
}
```

## Original Documentation

- [ESLint no-compare-neg-zero](https://eslint.org/docs/latest/rules/no-compare-neg-zero)
