# avoid-new

Disallow creating `new` promises outside of utility libs (use `util.promisify` instead).

## Rule Details

This rule discourages using the `new Promise` constructor directly. When promisifying callback-based functions, prefer Node's `util.promisify`. When wrapping a plain value or error, prefer `Promise.resolve()` or `Promise.reject()`.

Examples of **incorrect** code for this rule:

```javascript
var x = new Promise(function (resolve, reject) {});
new Promise();
Thing(new Promise(() => {}));
```

Examples of **correct** code for this rule:

```javascript
Promise.resolve();
Promise.reject();
Promise.all();
new Horse();
new PromiseLikeThing();
new Promise.resolve();
```

## Original Documentation

- [eslint-plugin-promise: avoid-new](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/avoid-new.md)
