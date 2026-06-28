# promise/no-nesting

## Rule Details

Disallow nesting `.then()` or `.catch()` statements inside promise callbacks when the
inner call does not depend on variables introduced by the enclosing callback.

Deeply nested promise chains are harder to read and can usually be rewritten as a flat
chain. This rule flags the inner `.then()` / `.catch()` when its arguments do not
reference any binding (parameter or local variable) that belongs to the immediately
enclosing promise callback, because in that case the nesting is unnecessary.

Examples of **incorrect** code for this rule:

```javascript
doThing().then(function () {
  return a.then();
});

doThing().then(() => b.catch());

doThing().then(function () {
  return a.then(function () {
    return b.catch();
  });
});
```

Examples of **correct** code for this rule:

```javascript
// Flat chain — no nesting
doThing().then(function () {
  return 4;
});

// Inner call uses a closure variable from the enclosing callback — cannot be flattened
doThing().then((a) => getB(a).then((b) => getC(a, b)));

// Promise.resolve / Promise.all inside a callback is fine
doThing().then(function () {
  return Promise.resolve(4);
});
```

## Original Documentation

https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/no-nesting.md
