# prefer-global-number-constants

## Rule Details

Prefer the global `NaN`, `Infinity`, and `-Infinity` constants over the corresponding properties of `Number`.

References to shadowed globals and assignment targets are excluded. Positive constants are automatically fixed only when no internal comment would be lost. `Number.NEGATIVE_INFINITY` is reported without an automatic fix to preserve expression precedence. This rule is the opposite of the `checkNaN` and `checkInfinity` options of `prefer-number-properties`.

Examples of **incorrect** code:

```javascript
const value = Number.POSITIVE_INFINITY;
```

Examples of **correct** code:

```javascript
const value = Infinity;
```

## Options

This rule has no options. It is enabled at `error` severity in `unicornPlugin.configs.recommended`.

The recommended preset explicitly disables `checkNaN` and `checkInfinity` on
`prefer-number-properties`, retaining its other checks while preventing the
rules from undoing each other's numeric-constant fixes.

## Differences from upstream

Unicorn v75.0.0 offers a fix for `const {NaN: value} = Number` that removes the
`value` binding. This rule reports the destructuring use without a fix instead.
Loop assignment targets, such as `for (Number.NaN of values) {}`, are left
unchanged, consistently with other assignment targets.

Unlike upstream's flow-insensitive tracking, this rule follows only aliases
initialized in a variable declaration and never reassigned. Conditional and
logical receivers, default parameters, assignment expressions, and comma
expressions are skipped. For example, after `let N = Number; N = {NaN: 42};`,
`N.NaN` must not be replaced with `NaN`. This also preserves side effects in
expressions such as `(sideEffect(), Number).NaN`. Stable aliases such as
`const N = Number; N.NaN` still receive fixes.

## Original Documentation

- [eslint-plugin-unicorn: prefer-global-number-constants](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/prefer-global-number-constants.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-global-number-constants.js)
