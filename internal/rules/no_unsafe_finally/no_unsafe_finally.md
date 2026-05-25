# no-unsafe-finally

## Rule Details

Disallows control flow statements (`return`, `throw`, `break`, `continue`) inside `finally` blocks. When control flow statements are used inside `finally` blocks, they override the control flow of `try` and `catch` blocks, which can lead to unexpected behavior and make code harder to understand.

Examples of **incorrect** code for this rule:

```javascript
function foo() {
  try {
    return 1;
  } catch (err) {
    return 2;
  } finally {
    return 3; // overrides the return in try/catch
  }
}

function bar() {
  try {
    doSomething();
  } finally {
    throw new Error(); // overrides any error thrown in try
  }
}

label: try {
  return 0;
} finally {
  break label; // overrides the return in try
}
```

Examples of **correct** code for this rule:

```javascript
function foo() {
  try {
    return 1;
  } catch (err) {
    return 2;
  } finally {
    console.log('done');
  }
}

function bar() {
  try {
    doSomething();
  } finally {
    // control flow inside nested functions is fine
    function cleanup(x) {
      return x;
    }
    cleanup();
  }
}

function baz() {
  try {
    doSomething();
  } finally {
    // break/continue inside loops within finally is fine
    while (condition) {
      break;
    }
  }
}
```

## Original Documentation

- [ESLint no-unsafe-finally](https://eslint.org/docs/latest/rules/no-unsafe-finally)
