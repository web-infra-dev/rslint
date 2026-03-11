# array-callback-return

## Rule Details

Enforces `return` statements in callbacks of array methods such as `map`, `filter`, `find`, `findIndex`, `findLast`, `findLastIndex`, `every`, `some`, `reduce`, `reduceRight`, `flatMap`, `sort`, `toSorted`, and `Array.from`. Callbacks for these methods must return a value; otherwise, the code likely contains a mistake.

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
[1, 2, 3].forEach(x => x * x);
```

Examples of **correct** code for this rule:

```javascript
var squares = [1, 2, 3].map(function (x) {
  return x * x;
});

var bools = [1, 2, 3].filter(function (x) {
  return x > 2;
});

[1, 2, 3].forEach(x => {
  console.log(x);
});

// with checkForEach: true and allowVoid: true
[1, 2, 3].forEach(x => void bar(x));
```

## Original Documentation

- [ESLint array-callback-return](https://eslint.org/docs/latest/rules/array-callback-return)
