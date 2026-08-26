# no-undefined

## Rule Details

Disallow the use of `undefined` as an identifier.

The `undefined` variable in JavaScript is actually a property of the global object. As such, in ECMAScript 3 it was possible to overwrite the value of `undefined`. While ECMAScript 5 disallows overwriting `undefined`, it's still possible to shadow `undefined`, such as:

```javascript
function doSomething(data) {
  const undefined = "hi";

  // doesn't do what you think it does
  if (data === undefined) {
    // ...
  }
}
```

Because `undefined` can be overwritten or shadowed, reading `undefined` can give an unexpected value. (This is not the case for `null`, which is a keyword that always produces the same value.) To guard against this, you can avoid all uses of `undefined`, which is what some style guides recommend and what this rule enforces. Those style guides then also recommend:

- Variables that should be `undefined` are simply left uninitialized. (All uninitialized variables automatically get the value of `undefined` in JavaScript.)
- Checking if a value is `undefined` should be done with `typeof`.
- Using the `void` operator to generate the value of `undefined` if necessary.

Examples of **incorrect** code for this rule:

```javascript
const foo = undefined;

const undefined = "foo";

if (foo === undefined) {
  // ...
}

function baz(undefined) {
  // ...
}

bar(undefined, "lorem");
```

Examples of **correct** code for this rule:

```javascript
const foo = void 0;

const Undefined = "foo";

if (typeof foo === "undefined") {
  // ...
}

global.undefined = "foo";

bar(void 0, "lorem");
```

## Options

This rule has no options.

## Differences from ESLint

- A named class declaration whose name is `undefined` (`class undefined {}`) is reported once. ESLint reports the same position twice.
- An assignment-destructuring pattern with a shorthand default whose name is `undefined` (e.g. `({ undefined = 1 } = target)`) is reported once. ESLint reports the same position twice.

## Original Documentation

- [ESLint: no-undefined](https://eslint.org/docs/latest/rules/no-undefined)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-undefined.js)
