# no-redundant-comparison

## Rule Details

Disallow comparisons made redundant by an equality check in the same logical AND.

An equality between references lets the rule compare their conditions, including
transitive equalities and stronger numeric bounds. A strict inequality can also
make a strict comparison redundant.

Examples of **incorrect** code:

```javascript
if (a > 0 && b > 0 && a === b) {}
if (a > 0 && b > 2 && a === b) {}
if (a !== 'test' && b !== 'test' && a === b) {}
if (a === 'test' && b !== 'test' && a !== b) {}
```

Examples of **correct** code:

```javascript
if (a > 0 && a === b) {}
if (b > 2 && a === b) {}
if (a !== 'test' && a === b) {}
if (a === 'test' && a !== b) {}
```

The rule checks `&&` chains. It skips optional chains and chains with explicit
side effects between the first and last analyzed comparisons, including calls,
tagged templates, assignments and updates. Side effects outside that interval
do not prevent the check.

## Fixes and Suggestions

Redundant plain strict comparisons are removed automatically. Removing member
reads or ordering comparisons can skip a getter or an implicit conversion, so
the rule offers an editor suggestion instead. A diagnostic has no edit when
removing the comparison would discard a comment.

## Options

This rule has no options.

## Differences from upstream

No known differences from eslint-plugin-unicorn v79.0.0.

## Original Documentation

- [eslint-plugin-unicorn: no-redundant-comparison](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-redundant-comparison.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-redundant-comparison.js)
