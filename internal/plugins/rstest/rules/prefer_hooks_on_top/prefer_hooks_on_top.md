# prefer-hooks-on-top

## Rule Details

Requires `beforeAll`, `beforeEach`, `afterEach` and `afterAll` to be declared before the first test case of the suite they belong to. A hook is attached to the whole suite wherever it is written, so one declared halfway down a `describe` body still wraps every test above it: `beforeAll` runs before the first case, and `beforeEach` and `afterEach` surround each of them. Reading the file top to bottom suggests the opposite, that the hook only covers the cases that follow it.

Each suite body is judged on its own, so a nested `describe` starts over and the cases already registered in its parent do not matter. The order of the hooks among themselves is not this rule's concern, and statements that are not test cases — a helper call, a constant, an `import` — can sit anywhere.

A suite body is either the callback written inline at the `describe` call or a function passed to it by name, and every function body is judged on its own regardless of which it is. A hook written in a function that is never called registers nothing at all, so it is never reported, and a test case in such a function does not count against the hooks that follow it. The other side of that boundary is that a hook registered through a helper the suite calls is judged against the helper rather than the suite, and so goes unreported: the rule covers what is written directly in a suite body, and registrations reached indirectly are outside what it reads.

Hooks and test cases are recognized through every way Rstest exposes them: globals, named, renamed, namespace and `require` bindings from `@rstest/core` and `rstack/test`, `import.meta.rstest`, and the hooks `@rstest/playwright` exposes as members of its test object. A test case counts whether it is registered plainly, through a modifier such as `.only` or `.skip`, through `.each` or `.for` with an array or a tagged template, or through a test API extended with fixtures. Calling `test.extend({ ... })` on its own registers nothing — it builds another test function — so hooks may follow it. `onTestFinished` and `onTestFailed` are registered from inside a running test rather than from a suite body and are never reported. The rule needs no type information. See the [Rstest hooks documentation](https://rstest.rs/api/runtime-api/test-api/hooks) for lifecycle execution details.

## Incorrect

```ts
describe('checkout', () => {
  test('charges the card', async () => {
    await checkout(cart);
    expect(gateway.charge).toHaveBeenCalled();
  });

  beforeEach(() => {
    gateway.charge.mockClear();
  });
});
```

## Correct

```ts
describe('checkout', () => {
  beforeEach(() => {
    gateway.charge.mockClear();
  });

  test('charges the card', async () => {
    await checkout(cart);
    expect(gateway.charge).toHaveBeenCalled();
  });
});
```
