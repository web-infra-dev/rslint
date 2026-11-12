# array-callback-return

## Rule Details

Enforces `return` statements in callbacks of array methods such as `map`, `filter`, `find`, `findIndex`, `findLast`, `findLastIndex`, `every`, `some`, `reduce`, `reduceRight`, `flatMap`, `sort`, `toSorted`, and `Array.from`. Callbacks for these methods must return a value; otherwise, the code likely contains a mistake. The same checks apply to typed array methods. Async and generator callbacks are exempt.

The rule recognizes static computed method names, such as `items["map"]`, and callbacks selected by logical or conditional expressions or returned from immediately invoked functions. It checks methods with these names even when the receiver is not an array.

### Options

- `allowImplicit` (default: `false`): When set to `true`, allows callbacks to implicitly return `undefined` by using `return;` without a value.
- `checkForEach` (default: `false`): When set to `true`, also checks that `forEach` callbacks do not return a value.
- `allowVoid` (default: `false`): When set to `true` along with `checkForEach`, allows `forEach` callbacks to return `void` expressions (e.g., `void bar(x)`).

Examples of **incorrect** code for this rule:

```javascript
var squares = [1, 2, 3].map(function (x) {
  x * x;
});

var bools = [1, 2, 3].filter(function (x) {
  if (x > 2) {
    return true;
  }
  // missing return in else path
});

// with checkForEach: true
[1, 2, 3].forEach((x) => x * x);
```

Examples of **correct** code for this rule:

```javascript
var squares = [1, 2, 3].map(function (x) {
  return x * x;
});

var bools = [1, 2, 3].filter(function (x) {
  return x > 2;
});

[1, 2, 3].forEach((x) => {
  console.log(x);
});

// with checkForEach: true and allowVoid: true
[1, 2, 3].forEach((x) => void bar(x));
```

## Suggestions

With `checkForEach: true`, an expression-bodied arrow callback offers a suggestion to wrap its expression in braces. With `allowVoid: true`, returned expressions also offer a suggestion to prepend `void`. Suggestions are applied individually; the rule does not automatically fix code.

## Differences from upstream

ESLint v9.39.1 can produce invalid code when `wrapBraces` turns a function or class expression into an anonymous declaration, or a `"use strict"` string into a directive with a non-simple parameter list. Rslint preserves expression context by adding parentheses for these bodies and reuses existing parentheses. This also covers composite expressions whose leftmost expression is a function or class, such as `() => function () {}()` and `() => class {}.name`. For example, `() => function () {}()` becomes `() => {(function () {}())}`, and `(x = 0) => "use strict"` becomes `(x = 0) => {("use strict")}`.

## Original Documentation

- [ESLint: array-callback-return](https://eslint.org/docs/latest/rules/array-callback-return)
- [Source code](https://github.com/eslint/eslint/blob/v9.39.1/lib/rules/array-callback-return.js)
