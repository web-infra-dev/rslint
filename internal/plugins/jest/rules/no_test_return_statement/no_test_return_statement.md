# no-test-return-statement

## Rule Details

Disallow explicitly returning from tests. Tests in Jest should be void and not return values. If you are returning Promises then you should update the test to use `async/await`.

This rule reports the first `return` statement written directly in a test callback's block body. Returns nested in other statements or inside nested functions, and arrow functions without braces, are not reported. Hooks and `describe` callbacks are not checked.

Examples of **incorrect** code for this rule:

```javascript
test('return an expect', () => {
  return expect(1).toBe(1);
});

it('returning a promise', function () {
  return new Promise(res => setTimeout(res, 100)).then(() => expect(1).toBe(1));
});
```

Examples of **correct** code for this rule:

```javascript
it('noop', function () {});

test('noop', () => {});

test('one arrow', () => expect(1).toBe(1));

test('empty');

test('one', () => {
  expect(1).toBe(1);
});

it('one', function () {
  expect(1).toBe(1);
});

it('returning a promise', async () => {
  await new Promise(res => setTimeout(res, 100));
  expect(1).toBe(1);
});
```

## Differences from upstream

- A callback passed by name is reported only while that name is used for nothing but test callbacks. `it('one', check); function check() { return expect(1).toBe(1); }` is reported, but adding `const value = check();`, exporting `check`, or reassigning it leaves it unreported, because the function's return value may be needed by those other uses. Upstream reports the function in all of these cases.
- A callback bound with `const check = () => { ... }` is checked like a function declaration. Upstream checks only function declarations passed by name.

## Original Documentation

- [eslint-plugin-jest: no-test-return-statement](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/no-test-return-statement.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/no-test-return-statement.ts)
