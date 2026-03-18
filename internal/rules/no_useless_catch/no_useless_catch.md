# no-useless-catch

## Rule Details

Disallows catch clauses that only rethrow the caught error. A catch clause that only rethrows the original error is redundant, and has no effect on the runtime behavior of the program. These redundant clauses can be a source of confusion and code bloat, so it is better to disallow them.

Examples of **incorrect** code for this rule:

```javascript
try {
  doSomethingThatMightThrow();
} catch (e) {
  throw e;
}

try {
  doSomethingThatMightThrow();
} catch (e) {
  throw e;
} finally {
  cleanUp();
}
```

Examples of **correct** code for this rule:

```javascript
try {
  doSomethingThatMightThrow();
} catch (e) {
  doSomethingBeforeRethrow();
  throw e;
}

try {
  doSomethingThatMightThrow();
} catch (e) {
  handleError(e);
}

try {
  doSomethingThatMightThrow();
} catch ({ message }) {
  throw message;
}
```

## Original Documentation

- [ESLint no-useless-catch](https://eslint.org/docs/latest/rules/no-useless-catch)
