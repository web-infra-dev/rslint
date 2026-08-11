# consistent-return

## Rule Details

This rule requires `return` statements to either always or never specify values.

A function is inconsistent when one `return` in it hands back a value and
another does not, or when it returns a value on one path and runs off the end of
its body on another — because falling off the end hands back `undefined`.

The check applies to each function on its own. A nested function is judged
separately from the function that contains it, and the first `return` reached in
a function sets the expectation for the rest of it.

Two shapes are exempt from the "runs off the end" half of the check, because
returning a value from them is how a constructor overrides the object being
built: a class `constructor`, and a function whose own name starts with an
uppercase letter.

Examples of **incorrect** code for this rule:

```javascript
function doSomething(condition) {
  if (condition) {
    return true;
  } else {
    return;
  }
}

function doSomethingElse(condition) {
  if (condition) {
    return true;
  }
}
```

Examples of **correct** code for this rule:

```javascript
function doSomething(condition) {
  if (condition) {
    return true;
  } else {
    return false;
  }
}

function Foo() {
  if (!(this instanceof Foo)) {
    return new Foo();
  }
}
```

## Options

### `treatUndefinedAsUnspecified`

When `true`, `return undefined;` and `return void 0;` are read as returning
nothing, so they pair with a bare `return;`. It defaults to `false`.

Examples of **correct** code for this rule with `{ "treatUndefinedAsUnspecified": true }`:

```json
{ "consistent-return": ["error", { "treatUndefinedAsUnspecified": true }] }
```

```javascript
function doSomething(condition) {
  if (condition) {
    return undefined;
  } else {
    return;
  }
}
```

Examples of **incorrect** code for this rule with `{ "treatUndefinedAsUnspecified": true }`:

```json
{ "consistent-return": ["error", { "treatUndefinedAsUnspecified": true }] }
```

```javascript
function doSomething(condition) {
  if (condition) {
    return true;
  }
  return undefined;
}
```

## Original Documentation

- [ESLint: consistent-return](https://eslint.org/docs/latest/rules/consistent-return)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/consistent-return.js)

https://eslint.org/docs/latest/rules/consistent-return
