# no-useless-concat

## Rule Details

Disallow unnecessary concatenation of literals or template literals. This rule flags a `+` that joins two string literals or template literals on the same line, since they could be written as a single literal.

Concatenation that spans multiple source lines is intentionally not reported.

Examples of **incorrect** code for this rule:

```javascript
var a = `some` + `string`;
var b = '1' + '0';
var c = '1' + `0`;
var d = `1` + '0';
var e = `1` + `0`;
```

Examples of **correct** code for this rule:

```javascript
// When the variables could hold non-strings
var a = 1 + 1;
var b = 1 + '1';
var c = foo + bar;
var d = 'foo' + bar;
```

## Original Documentation

- [ESLint: no-useless-concat](https://eslint.org/docs/latest/rules/no-useless-concat)
