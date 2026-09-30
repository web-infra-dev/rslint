# prefer-each

## Rule Details

Prefer `.each` over wrapping `describe`/`test`/`it` in native `for` loops, for
clearer output and easier filtering. A loop that only runs logic inside a test
function is ignored; a loop that registers tests, suites, or hooks is reported
wherever it sits.

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

// we don't warn on loops _in_ test functions that register nothing because
// those typically involve complex setup that is better done in the test
// function itself
it('returns numbers that are greater than five', () => {
  for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
  }
});
```

## Original Documentation

- [eslint-plugin-jest: prefer-each](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/prefer-each.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/prefer-each.ts)
