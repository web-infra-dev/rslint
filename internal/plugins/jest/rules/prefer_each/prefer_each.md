# prefer-each

## Rule Details

Prefer `.each` over wrapping `describe`/`test`/`it` in native `for` loops, for
clearer output and easier filtering. Loops inside a test function are ignored.

Examples of **incorrect** code for this rule:

```js
for (const number of getNumbers()) {
  it('is greater than five', function () {
    expect(number).toBeGreaterThan(5);
  });
}

for (const [input, expected] of data) {
  beforeEach(() => setupSomething(input));

  test(`results in ${expected}`, () => {
    expect(doSomething()).toBe(expected);
  });
}
```

Examples of **correct** code for this rule:

```js
it.each(getNumbers())(
  'only returns numbers that are greater than seven',
  number => {
    expect(number).toBeGreaterThan(7);
  },
);

describe.each(data)('when input is %s', ([input, expected]) => {
  beforeEach(() => setupSomething(input));

  test(`results in ${expected}`, () => {
    expect(doSomething()).toBe(expected);
  });
});

// we don't warn on loops _in_ test functions because those typically involve
// complex setup that is better done in the test function itself
it('returns numbers that are greater than five', () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});
```

## Original Documentation

- [jest/prefer-each](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-each.md)
