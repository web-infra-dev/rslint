# no-new-func

## Rule Details

Disallows creating functions from strings using the `Function` constructor. Passing a string to the `Function` constructor requires the engine to parse that string, similar to `eval`.

Examples of **incorrect** code for this rule:

```javascript
var a = new Function('a', 'b', 'return a + b');
var b = Function('a', 'b', 'return a + b');
var c = Function.call(null, 'a', 'b', 'return a + b');
var d = Function.apply(null, ['a', 'b', 'return a + b']);
var e = Function.bind(null, 'a', 'b', 'return a + b')();
var f = Function.bind(null, 'a', 'b', 'return a + b');
```

Examples of **correct** code for this rule:

```javascript
var x = function (a, b) {
  return a + b;
};
```

## Original Documentation

- [ESLint: no-new-func](https://eslint.org/docs/latest/rules/no-new-func)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-new-func.js)
