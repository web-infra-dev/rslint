# new-for-builtins

## Rule Details

Many builtin constructors can be called with or without `new`, but constructor
calls are clearer and more consistent when they use `new`. This rule also
reports selected builtin namespace objects that are neither callable nor
constructible.

Optional calls that would need to become optional constructors are ignored,
because `new` cannot be optional.

Examples of **incorrect** code for this rule:

```javascript
const list = Array(10);
const map = Map([["foo", "bar"]]);
const now = Date();
const tag = WebAssembly.JSTag();
const text = new String("value");
```

Examples of **correct** code for this rule:

```javascript
const list = new Array(10);
const map = new Map([["foo", "bar"]]);
const now = String(new Date());
const text = String("value");
```

## Original Documentation

- [eslint-plugin-unicorn: new-for-builtins](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/new-for-builtins.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/rules/new-for-builtins.js)
