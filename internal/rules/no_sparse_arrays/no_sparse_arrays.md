# no-sparse-arrays

## Rule Details

Disallows sparse arrays, which are array literals that contain empty slots created by extra commas. Sparse arrays can be confusing because the empty slots are `undefined` but behave differently from explicitly setting an element to `undefined` (for example, `Array.prototype.forEach` skips sparse entries). Extra commas are usually a typo.

Examples of **incorrect** code for this rule:

```javascript
var items = [1, , 3];

var colors = ['red', , 'blue'];
```

Examples of **correct** code for this rule:

```javascript
var items = [1, 2, 3];

var colors = ['red', 'blue'];

var arr = [1, undefined, 3]; // explicit undefined is fine
```

## Original Documentation

- [ESLint no-sparse-arrays](https://eslint.org/docs/latest/rules/no-sparse-arrays)
