# array-callback-return

## Rule Details

Enforces `return` statements in callbacks of array methods such as `map`, `filter`, `find`, `every`, `some`, `reduce`, `flatMap`, `sort`, `toSorted`, and `Array.from`. Callbacks for these methods must return a value; otherwise, the code likely contains a mistake. When the `checkForEach` option is enabled, it also disallows returning values from `forEach` callbacks.

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

[1, 2, 3].forEach(x => x * x); // with checkForEach: true
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
```

## Original Documentation

- [ESLint array-callback-return](https://eslint.org/docs/latest/rules/array-callback-return)
