# require-number-to-fixed-digits-argument

## Rule Details

Requires calls to `Number#toFixed()` to explicitly pass the number of fraction
digits. Writing the argument makes the intended formatting clear instead of
relying on the default value of `0`.

Examples of **incorrect** code for this rule:

```javascript
const value = number.toFixed();
```

Examples of **correct** code for this rule:

```javascript
const integer = number.toFixed(0);
const decimal = number.toFixed(2);
```

## Original Documentation

- [eslint-plugin-unicorn require-number-to-fixed-digits-argument](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/require-number-to-fixed-digits-argument.md)
