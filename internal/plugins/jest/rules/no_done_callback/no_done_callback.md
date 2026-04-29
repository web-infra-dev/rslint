# jest/no-done-callback

## Rule Details

Disallow using a `done`-style callback in Jest tests and hooks. Returning a Promise (or using `async`/`await`) is more reliable than relying on `done`, which can silently pass or time out when not invoked correctly.

For non-`async` functions the rule reports `noDoneCallback` and suggests wrapping the body in `new Promise(done => ...)`. For `async` functions it reports `useAwaitInsteadOfCallback`.

Examples of **incorrect** code for this rule:

```javascript
beforeEach(done => {
  done();
});

test('myFunction()', done => {
  done();
});

test('myFunction()', async done => {
  await fetchData();
  done();
});
```

Examples of **correct** code for this rule:

```javascript
beforeEach(() => {
  return setupUsTheBomb();
});

test('myFunction()', () => {
  expect(myFunction()).toBeTruthy();
});

test('myFunction()', async () => {
  const data = await fetchData();
  expect(data).toBe('peanut butter');
});
```

## Original Documentation

- [jest/no-done-callback](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-done-callback.md)
