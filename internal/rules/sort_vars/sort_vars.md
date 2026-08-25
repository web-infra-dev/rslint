# sort-vars

## Rule Details

This rule requires identifier variables within the same declaration block to be sorted alphabetically. Destructuring declarations are ignored. By default, ordering is case-sensitive.

Examples of **incorrect** code for this rule:

```javascript
let b, a;
let c, D, e;
```

Examples of **correct** code for this rule:

```javascript
let a, b, c;
let G, f, h;
let { b, a } = value;
```

With `{ "ignoreCase": true }`, names are compared without case sensitivity:

```json
{ "sort-vars": ["error", { "ignoreCase": true }] }
```

```javascript
let a, A;
let c, D, e;
```

The rule can automatically reorder declarations when every participating initializer is a literal. It reports without a fix when reordering might change evaluation order.

## Original Documentation

- [ESLint: sort-vars](https://eslint.org/docs/latest/rules/sort-vars)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.0/lib/rules/sort-vars.js)
