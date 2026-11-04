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

The rule removes a comparison automatically when preceding conditions already
establish it, or it is statically known to be true. The supporting conditions
must evaluate without user-defined hooks. Safe removals of ordering comparisons
are offered as editor suggestions.

An edit must also preserve the expression's returned value. In a condition such
as `if (...)` only truthiness matters, but an assigned or returned `&&` expression
can produce a string, number or object. A diagnostic has no edit when removing
the comparison could change that value or discard a comment.

## Options

This rule has no options.

## Differences from upstream

Rslint withholds edits that Unicorn v79.0.0 can offer unsafely:

- A later condition is not used to justify deleting an earlier guard. Otherwise
  the edit could expose a previously skipped expression or remove an exception.
- Property reads, implicit conversions and other evaluations that may invoke
  user code cannot support an edit. For example, a getter between equality
  checks could change the variables being compared. Global names without a
  local binding are also left unchanged because they may be accessor-backed.
- Removing the final comparison must preserve the returned value. For example,
  `a === b && a === 1 && label && b === 1` is reported without an edit when used
  as a value: dropping the final comparison could return `label` instead of
  `true`.

These cases still receive diagnostics for manual review.

## Original Documentation

- [eslint-plugin-unicorn: no-redundant-comparison](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-redundant-comparison.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-redundant-comparison.js)
