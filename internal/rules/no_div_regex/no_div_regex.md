# no-div-regex

## Rule Details

Disallows equal signs explicitly at the beginning of regular expression literals, since `/=` at the start of a regular expression can be visually confused with a division-assignment operator.

Examples of **incorrect** code for this rule:

```javascript
function bar() {
  return /=foo/;
}
```

Examples of **correct** code for this rule:

```javascript
function bar() {
  return /[=]foo/;
}
```

## Original Documentation

- [ESLint: no-div-regex](https://eslint.org/docs/latest/rules/no-div-regex)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-div-regex.js)
