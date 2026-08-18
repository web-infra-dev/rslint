# require-local-test-context-for-concurrent-snapshots

## Rule Details

Concurrent Rstest tests must use the `expect` provided by their local Test Context for snapshot assertions. A global or imported `expect` can share snapshot state between concurrently running tests.

Examples of **incorrect** code for this rule:

```typescript
test.concurrent('renders', () => {
  expect(render()).toMatchSnapshot();
});
```

Examples of **correct** code for this rule:

```typescript
test.concurrent('renders', ({ expect }) => {
  expect(render()).toMatchSnapshot();
});
```

An explicit inner `sequential` or `concurrent` registration overrides the mode inherited from an enclosing describe block. The mode is read only from the chained modifiers: Rstest takes just `timeout`, `retry`, `repeats` and `meta` out of the options object of a `(name, options, fn)` registration, so `concurrent` and `sequential` written there have no effect at runtime and none here either.

The rule recognizes Rstest aliases, named callbacks, `.each`, `.for`, `import.meta.rstest`, and `@rstest/playwright` registrations.

Assertions written inside a closure, hook or helper declared within a concurrent body are checked too, because such a function runs as part of the concurrent test whenever it runs at all:

```typescript
test.concurrent('renders', async () => {
  await Promise.all(items.map(async item => {
    expect(item).toMatchSnapshot();
  }));
});
```

A helper declared outside every registration callback is not checked, since the rule cannot tell which tests call it. A callback passed by a name the type checker cannot resolve is only matched against a function declared at the top level of the file under that exact name, so a same-named function nested inside another callback is never blamed for a test that does not run it.

Rstest's `matchSnapshot` Chai alias is also checked because it uses the same snapshot state as `toMatchSnapshot`.
