# guard-for-in

## Rule Details

Require `for-in` loops to include an `if` statement. Iterating a `for-in` loop over an object exposes inherited prototype properties in addition to the object's own keys, so the body should typically be guarded (e.g. with `Object.hasOwn`, `Object.prototype.hasOwnProperty.call`, or a short-circuit `continue`) to filter unwanted properties.

Examples of **incorrect** code for this rule:

```javascript
for (key in foo) {
  doSomething(key);
}
```

Examples of **correct** code for this rule:

```javascript
for (key in foo) {
  if (Object.hasOwn(foo, key)) {
    doSomething(key);
  }
}

for (key in foo) {
  if (Object.prototype.hasOwnProperty.call(foo, key)) {
    doSomething(key);
  }
}

for (key in foo) {
  if (!Object.hasOwn(foo, key)) continue;
  doSomething(key);
}
```

## Original Documentation

- https://eslint.org/docs/latest/rules/guard-for-in
