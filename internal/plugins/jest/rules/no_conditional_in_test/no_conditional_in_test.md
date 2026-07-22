# no-conditional-in-test

## Rule Details

Disallow conditional logic in test bodies. A conditional usually indicates that
a test is covering multiple execution paths, which can make it unclear which
behavior the test is intended to verify. Prefer a separate test for each branch.

Examples of **incorrect** code for this rule:

```javascript
it('foo', () => {
  if (true) {
    doTheThing();
  }
});

it('bar', () => {
  switch (mode) {
    case 'none':
      generateNone();
    case 'single':
      generateOne();
    case 'multiple':
      generateMany();
  }

  expect(fixtures.length).toBeGreaterThan(-1);
});

it('qux', async () => {
  const promiseValue = () => {
    return something instanceof Promise
      ? something
      : Promise.resolve(something);
  };

  await expect(promiseValue()).resolves.toBe(1);
});
```

Examples of **correct** code for this rule:

```javascript
describe('my tests', () => {
  if (true) {
    it('foo', () => {
      doTheThing();
    });
  }
});

beforeEach(() => {
  switch (mode) {
    case 'none':
      generateNone();
    case 'single':
      generateOne();
    case 'multiple':
      generateMany();
  }
});

it('bar', () => {
  expect(fixtures.length).toBeGreaterThan(-1);
});

const promiseValue = something => {
  return something instanceof Promise ? something : Promise.resolve(something);
};

it('qux', async () => {
  await expect(promiseValue()).resolves.toBe(1);
});
```

Conditionals outside test bodies, including conditionals in `describe` blocks,
hooks, and helper functions declared outside a test, are not reported.

## Options

- First argument (optional): object with `allowOptionalChaining`
  - `allowOptionalChaining`: whether optional chaining (`?.`) is allowed inside
    test bodies. Default is `true`.

When `allowOptionalChaining` is `false`, optional property access, element
access, and calls are also reported:

```json
{
  "jest/no-conditional-in-test": [
    "error",
    {
      "allowOptionalChaining": false
    }
  ]
}
```

Examples of **incorrect** code with `{ "allowOptionalChaining": false }`:

```javascript
it('foo', () => {
  const value = obj?.bar;
});

it('bar', () => {
  obj?.foo();
});
```

Examples of **correct** code with `{ "allowOptionalChaining": false }`:

```javascript
it('foo', () => {
  const value = obj!.bar;
});
```

## Original Documentation

- [jest/no-conditional-in-test](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-conditional-in-test.md)
