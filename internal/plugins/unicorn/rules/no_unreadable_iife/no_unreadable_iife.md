# no-unreadable-iife

## Rule Details

Disallow immediately invoked arrow functions with parenthesized expression bodies.
A block body makes the function body easier to distinguish from the surrounding call.

Examples of **incorrect** code for this rule:

```javascript
const foo = (bar => (bar ? bar.baz : baz))(getBar());
const value = (() => ({ ready: true }))();
```

Examples of **correct** code for this rule:

```javascript
const foo = (bar => {
  return bar ? bar.baz : baz;
})(getBar());
const value = { ready: true };
```

## Suggestions

This rule offers an editor suggestion to replace the parenthesized body with a
block containing a return statement. It does not automatically fix code. The
suggestion is withheld when removing the parentheses would discard comments
outside the expression. Comments inside the expression are preserved.

## Options

This rule has no options.

## Original Documentation

- [eslint-plugin-unicorn: no-unreadable-iife](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/no-unreadable-iife.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-unreadable-iife.js)
