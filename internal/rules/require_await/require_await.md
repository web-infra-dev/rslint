# require-await

## Rule Details

This rule warns on async functions that have no `await` expression.

Examples of **incorrect** code for this rule:

```javascript
async function foo() {
  doSomething();
}

bar(async () => {
  doSomething();
});
```

Examples of **correct** code for this rule:

```javascript
async function foo() {
  await doSomething();
}

bar(async () => {
  await doSomething();
});

function baz() {
  doSomething();
}

bar(() => {
  doSomething();
});

async function noop() {}
```

Async generator functions are ignored by this rule.

## When Not To Use It

If you don't want to warn on async functions that have no `await` expression, then it's safe to disable this rule.

## Original Documentation

- [ESLint require-await](https://eslint.org/docs/latest/rules/require-await)
