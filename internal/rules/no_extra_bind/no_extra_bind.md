# no-extra-bind

## Rule Details

Disallows unnecessary calls to `.bind()`. If a function expression does not use `this`, calling `.bind()` on it is unnecessary. Arrow functions never have their own `this` binding, so `.bind()` on an arrow function is always unnecessary.

Examples of **incorrect** code for this rule:

```javascript
var x = function () {
  foo();
}.bind(bar);

var x = (() => {
  foo();
}).bind(bar);

var x = function () {
  (function () {
    this.bar();
  })();
}.bind(baz);
```

Examples of **correct** code for this rule:

```javascript
var x = function () {
  this.foo();
}.bind(bar);

var x = function (a) {
  return a + 1;
}.bind(foo, bar);

var x = f.bind(bar);
```

## Original Documentation

- [ESLint no-extra-bind](https://eslint.org/docs/latest/rules/no-extra-bind)
