# no-misrefactored-assignment

## Rule Details

Disallow arithmetic or bitwise compound assignments that repeat the target in
an immediate operand of the matching operator. The assignment `a += a + b`
uses the previous value of `a` twice. This can happen when rewriting
`a = a + b` as a compound assignment.

For `+`, `*`, `&`, `|`, and `^`, the target can appear on either side of the
binary expression. For other operators, the rule checks only the left operand.
Logical assignments (`&&=`, `||=`, and `??=`) are ignored.

The rule reports these assignments.

```javascript
a += a + b;
a *= b * a;
```

These assignments are allowed.

```javascript
a += b;
a *= b;
a += b + c;
a += a - b;
a -= b - a;
```

## Suggestions

Removing the duplicate changes runtime behavior, so the rule offers an editor
suggestion. The suggestion preserves comments in the retained operand. If the
edit would discard a comment, the rule reports the assignment without a
suggestion.

## Options

This rule has no options. Enable it explicitly with
`'unicorn/no-misrefactored-assignment': 'error'` in a config that enables the
`unicorn` plugin.

## Differences from upstream

ESLint Unicorn reports `counters["1"] += counters[(1).toString()] + delta`
because both expressions access the same property. Rslint leaves this
assignment unreported. Rslint reports the assignment when the right-hand
side uses `counters[1]` directly.

## Original Documentation

- [eslint-plugin-unicorn: no-misrefactored-assignment](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-misrefactored-assignment.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-misrefactored-assignment.js)
