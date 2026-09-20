# no-negation-in-equality-check

## Rule Details

Report a single logical negation on the left side of `===`, `!==`, `==`, or
`!=`. Such an expression compares the negated boolean value, rather than
negating the result of the comparison.

Examples of **incorrect** code:

```javascript
if (!value === expected) {}
if (!value != expected) {}
```

Examples of **correct** code:

```javascript
if (value !== expected) {}
if (!(value === expected)) {}
```

Double negation and negation on the right side are not reported. Other binary
operators are outside this rule's scope.

## Suggestions

The rule suggests removing the left-side `!` and inverting the comparison
operator. This is an editor suggestion, not an automatic fix: the two
expressions can have different behavior, so the intended comparison must be
chosen explicitly.

Suggestions preserve surrounding comments and keyword spacing. They also add
parentheses or a leading semicolon when removing `!` would otherwise change a
return, throw, or statement boundary.

## Options

This rule has no options.

## Original Documentation

- [eslint-plugin-unicorn: no-negation-in-equality-check](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/no-negation-in-equality-check.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-negation-in-equality-check.js)
