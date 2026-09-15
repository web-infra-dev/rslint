# prefer-await-to-callbacks

Prefer `async`/`await` to callback patterns.

## Rule Details

This rule reports calls to identifiers named `cb` or `callback`, function
parameters with either name in the last position, and calls whose last argument
is a function or arrow function with a first parameter named `err` or `error`.

Examples of **incorrect** code for this rule:

```javascript
cb();
callback();
doSomething(arg, (err) => {});
function doSomethingElse(cb) {}
```

Examples of **correct** code for this rule:

```javascript
await doSomething(arg);
async function doSomethingElse() {}
function* values() {
  yield yieldValue((err) => {});
}
eventEmitter.on('error', (err) => {});
```

The last-argument check allows `on` and `once` listeners and the array methods
`map`, `every`, `forEach`, `some`, `find`, and `filter` with one argument. It also
allows those methods with two arguments on `lodash`, `underscore`, or `_`, and
direct calls to those method names with two arguments.

Callbacks within an `await` or `yield` expression are exempt from the
last-argument check. Direct `cb`/`callback` calls and last parameters with those
names are still reported in those contexts.

This rule has no options and provides no automatic fixes or suggestions.

## Original Documentation

- [eslint-plugin-promise: prefer-await-to-callbacks](https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/docs/rules/prefer-await-to-callbacks.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/rules/prefer-await-to-callbacks.js)
