# require-local-test-context-for-concurrent-snapshots

## Rule Details

Requires concurrent tests to use the `expect` from their local Test Context for snapshot assertions. A shared global or imported `expect` can mix snapshot state between tests that run at the same time.

The rule checks snapshot matchers in concurrent tests, including concurrent suites and parameterized tests. It requires type information to identify the local Test Context.

## Incorrect

```ts
test.concurrent('renders a user', () => {
  expect(renderUser()).toMatchSnapshot();
});
```

## Correct

```ts
test.concurrent('renders a user', ({ expect }) => {
  expect(renderUser()).toMatchSnapshot();
});
```
