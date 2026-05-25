# no-nonoctal-decimal-escape

## Rule Details

Disallows `\8` and `\9` escape sequences in string literals.

Although `"\8"` and `"\9"` evaluate to the same characters as `"8"` and `"9"`, they are non-octal decimal escape sequences kept only for backward compatibility with web JavaScript. Browsers must support them, but Annex B explicitly allows non-web environments to omit them. The recommended fix is to drop the leading backslash, switch the digit to its `\uXXXX` form, or — when the goal really is to include a backslash — escape the backslash itself.

Examples of **incorrect** code for this rule:

```javascript
"\8";
"\9";
const foo = "w\8less";
const bar = "December 1\9";
const baz = "Don't use \8 and \9 escapes.";
const quux = "\0\8";
```

Examples of **correct** code for this rule:

```javascript
"8";
"9";
const foo = "w8less";
const bar = "December 19";
const baz = "Don't use \\8 and \\9 escapes.";
const quux = "\0\u0038";
```

## Options

This rule has no options.

## Original Documentation

- [ESLint no-nonoctal-decimal-escape](https://eslint.org/docs/latest/rules/no-nonoctal-decimal-escape)
