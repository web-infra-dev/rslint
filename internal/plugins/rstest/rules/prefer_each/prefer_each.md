# prefer-each

## Rule Details

Prefers parameterized Rstest registrations over manually registering tests, suites, or hooks inside a loop. `.each` keeps related cases together in test output and gives each case a clear title.

Loops used as ordinary test logic are allowed. When a loop registers one test, the diagnostic recommends `test.each` or `it.each`; loops that register suites, hooks, or several entries recommend `describe.each`.

## Incorrect

```ts
for (const input of [1, 2, 3]) {
  test(`doubles ${input}`, () => {
    expect(double(input)).toBe(input * 2);
  });
}
```

## Correct

```ts
test.each([1, 2, 3])('doubles %i', (input) => {
  expect(double(input)).toBe(input * 2);
});
```
