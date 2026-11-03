# no-double-comparison

## Rule Details

Disallow two comparisons of the same operands that can be combined into one.
Operands may appear in either order; the suggestion retains the first
comparison's operand order.

Examples of **incorrect** code:

```javascript
x === y || x < y;
x === y || x > y;
x < y || x > y;
x <= y && x >= y;
x <= y && x !== y;
x >= y && x !== y;
x === y || y < x;
```

Examples of **correct** code:

```javascript
x <= y;
x >= y;
x !== y;
x === y;
x < y;
x > y;
```

## Suggestions

The rule offers editor suggestions. Combining comparisons can change behavior
for mixed types, `NaN`, `undefined`, symbols, or objects with getters or custom
coercion. Review the suggestion before applying it. For example,
`undefined === undefined || undefined < undefined` is `true`, while
`undefined <= undefined` is `false`.

Loose equality, optional chains, known regular-expression operands and
comparisons with different operands are ignored. Longer logical chains are
checked pairwise. A diagnostic has no suggestion when replacing it would
discard a comment; comments inside the retained operands are preserved.

## Options

This rule has no options.

## Original Documentation

- [eslint-plugin-unicorn: no-double-comparison](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-double-comparison.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-double-comparison.js)
