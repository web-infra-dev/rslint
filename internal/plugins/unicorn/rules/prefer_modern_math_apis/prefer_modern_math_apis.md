# prefer-modern-math-apis

## Rule details

Prefer modern `Math` APIs over legacy patterns.

Standard `Math` methods express mathematical intent directly. This rule detects
logarithm conversions that can use `Math.log10()` or `Math.log2()`, and square roots
of squares that can use `Math.abs()` or `Math.hypot()`.

Replacements can change rounding, overflow, underflow, or coercion behavior. The
rule automatically fixes cases where it can establish that the behavior is
preserved, and offers suggestions for other cases. It withholds edits that would
remove comments.

Examples of **incorrect** code for this rule:

```javascript
Math.log(x) * Math.LOG10E;
Math.LOG10E * Math.log(x);
Math.log(x) / Math.LN10;
Math.log(x) * Math.LOG2E;
Math.LOG2E * Math.log(x);
Math.log(x) / Math.LN2;
Math.sqrt(a * a + b * b);
Math.sqrt(a ** 2 + b ** 2);
Math.sqrt(x ** 2);
```

Examples of **correct** code for this rule:

```javascript
Math.log10(x);
Math.log2(x);
Math.hypot(a, b);
Math.abs(x);
```

The rule ignores optional calls, computed methods, spread arguments, and other
arithmetic patterns, including `Math.pow()` calls.

## Options

This rule has no options.

## Differences from upstream

Calls through a locally declared `Math` variable receive suggestions instead of
automatic fixes, because that object can implement different operations.

Locally declared `Number`, `Infinity`, `NaN`, and `undefined` are not assumed to
hold their usual builtin values. When an input's value cannot be established,
the rule offers a suggestion instead of automatically changing the expression:

```javascript
function magnitude(Infinity) {
  // Reported with a suggestion, without an automatic fix.
  return Math.sqrt(Infinity ** 2);
}
```

The same restriction applies when the corresponding global is disabled with
`languageOptions.globals`. Calls using the standard globals retain the upstream
automatic fixes.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-modern-math-apis.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-modern-math-apis.js)
