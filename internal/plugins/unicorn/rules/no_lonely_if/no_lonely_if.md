# no-lonely-if

## Rule Details

Disallow an `if` statement as the only statement in another `if` without an `else` branch. The rule combines the conditions with `&&`, reducing indentation while preserving their evaluation order. It also handles statements without braces and preserves comments when fixing code.

Neither `if` statement may have an `else` branch. Other statements in the outer block prevent a report. The core `no-lonely-if` rule handles an `if` inside an `else` block; the two rules can be enabled together.

Examples of **incorrect** code for this rule:

```javascript
if (ready) {
    if (enabled) {
        run();
    }
}
```

Examples of **correct** code for this rule:

```javascript
if (ready && enabled) {
    run();
}
```

## Options

This rule has no options. It supports automatic fixes.

## Differences from upstream

Compared with eslint-plugin-unicorn v76.0.0, automatic fixes preserve the original behavior in these additional cases:

- Arrow functions used as conditions remain grouped, so the second condition is still evaluated. For example, `if (x => x) { if (b) run(); }` becomes `if ((x => x) && b) run();`.
- A statement following the merged `if` stays separate, including when it starts on the same line or immediately after the closing brace. For example, `if (a) { if (b) run() }next()` becomes `if (a && b) run();next()`.

## Original Documentation

- [eslint-plugin-unicorn: no-lonely-if](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/docs/rules/no-lonely-if.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/rules/no-lonely-if.js)
