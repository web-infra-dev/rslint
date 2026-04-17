# prefer-template

## Rule Details

This rule is aimed to flag usage of `+` operators with strings. It encourages
the use of template literals instead of string concatenation.

Examples of **incorrect** code for this rule:

```javascript
var str = 'Hello, ' + name + '!';
var str = 'Time: ' + 12 * 60 * 60 * 1000;
```

Examples of **correct** code for this rule:

```javascript
var str = 'Hello World!';
var str = `Hello, ${name}!`;
var str = `Time: ${12 * 60 * 60 * 1000}`;
```

This rule does not report two string literals concatenated together (for
example, `"Hello, " + "World!"`), which is reported by the
[`no-useless-concat`](https://eslint.org/docs/latest/rules/no-useless-concat)
rule instead.

The rule provides an autofix that rewrites the concatenation as a single
template literal, preserving comments around the `+` operators. Autofix is
skipped when any operand contains an octal or non-octal-decimal escape
sequence, because those cannot be represented in a template literal.

## Original Documentation

- [ESLint rule: prefer-template](https://eslint.org/docs/latest/rules/prefer-template)
