# no-negated-condition

## Rule Details

Disallow negated conditions in `if` statements with an `else` branch and in
conditional expressions. The rule reports `!`, `!=`, and `!==` conditions.
An `if` without an `else`, or followed directly by `else if`, is allowed.

Examples of **incorrect** code for this rule:

```javascript
if (!ready) {
    wait();
} else {
    start();
}

value !== expected ? fallback : result;
```

Examples of **correct** code for this rule:

```javascript
if (ready) {
    start();
} else {
    wait();
}

value === expected ? result : fallback;

if (!ready) {
    wait();
}
```

## Fixes

The fix inverts the condition and swaps the branches. Leading comments move
with their branch. Comments directly after either branch prevent automatic
fixing because their association is ambiguous. Unbraced `if` branches gain
braces when swapped.

This rule adds automatic fixes to ESLint's `no-negated-condition` rule.
Disable the core rule when enabling this rule to avoid duplicate diagnostics.

## Options

This rule has no options. It supports automatic fixes.

## Differences from upstream

Compared with eslint-plugin-unicorn v77.0.0, automatic fixes keep code valid when
removing `!` would expose an object, function, or class expression at the start of
a statement, an object at the start of an arrow function body, or `let[...]` at
the start of a statement in a script. Parentheses preserve the expression, and a
semicolon is added when needed to keep the preceding statement separate.

For example, `!{} ? a : b` becomes `({}) ? b : a`, and
`() => !{} ? a : b` becomes `() => ({}) ? b : a`.

## Original Documentation

- [eslint-plugin-unicorn: no-negated-condition](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/no-negated-condition.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/no-negated-condition.js)
