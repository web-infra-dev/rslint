# no-useless-return

## Rule Details

A `return;` at the end of a function does the same thing as reaching the end of
the function, so it can be dropped. This rule reports every `return` with no
value that nothing runs after.

A `return` that carries a value is left alone, and so is one inside a loop —
which leaves the loop early — or inside a `finally` block, which replaces the
value the statement was leaving with.

Examples of **incorrect** code for this rule:

```javascript
function foo() {
  return;
}

function bar() {
  doSomething();
  return;
}

function baz() {
  if (condition) {
    doSomething();
    return;
  } else {
    doSomethingElse();
  }
}

function qux() {
  switch (value) {
    case 1:
      doSomething();
      return;
  }
}
```

Examples of **correct** code for this rule:

```javascript
function foo() {
  return 5;
}

function bar() {
  if (condition) {
    return;
  }
  doSomething();
}

function baz() {
  for (const item of items) {
    if (item.done) {
      return;
    }
    process(item);
  }
}

function qux() {
  try {
    return computeValue();
  } finally {
    return fallback;
  }
}
```

## Original Documentation

- [ESLint: no-useless-return](https://eslint.org/docs/latest/rules/no-useless-return)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-useless-return.js)
