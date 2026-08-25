# warn-todo

## Rule Details

Disallow `.todo` on Rstest `test`, `it`, and `describe` registrations.

A todo registration is a placeholder: it names work that has not been written
yet, and the runner reports it without running anything. That is useful while
drafting, but a `.todo` left in a committed suite is a test that can never fail,
so this rule flags every one of them.

Examples of **incorrect** code:

```ts
test.todo('handles an empty cart');
it.todo('rejects an expired token');
describe.todo('checkout flow');
test.only.todo('rounds the total');
test.todo.each([1, 2])('case %i');
```

Examples of **correct** code:

```ts
test('handles an empty cart', () => {
  expect(total([])).toBe(0);
});

describe('checkout flow', () => {
  it('rejects an expired token', () => {
    expect(() => charge(expiredToken)).toThrow();
  });
});

test.skip('rounds the total', () => {
  expect(round(1.005)).toBe(1.01);
});
```

This rule only reports. It does not autofix and offers no suggestion, because
the replacement for a todo is the test body, which the rule cannot write.

This rule is the opposite of `prefer-todo`, which steers empty tests toward
`.todo`. The two cannot both be satisfied, so enable at most one of them.

## What counts as a todo registration

The rule follows a registration through every way Rstest's APIs can be reached:

- globals
- `@rstest/core` named imports and aliases
- `require('@rstest/core')`
- namespace imports and namespace members
- `import.meta.rstest`
- same-file `const` aliases
- `@rstest/playwright` (`test` and `describe`; that package does not export `it`)

A `test` or `describe` that comes from anywhere else keeps its own meaning and is
never reported.

The report anchors on the `todo` accessor written at the call site, in whatever
spelling it was written — `.todo`, `['todo']`, or `` [`todo`] ``. When `.todo`
was applied to an alias rather than at the call site, the call site has no such
accessor, so the report anchors on the identifier that resolves to the todo
registration:

```ts
const pending = test.todo;
pending('handles an empty cart'); // reported on `pending`
```

Only the `.todo` modifier registers a todo. A test options object accepts
`timeout`, `retry`, `repeats`, and `meta`, so a `todo` property written there
has no effect on the runner and is not reported:

```ts
test('handles an empty cart', { todo: true }, () => {}); // runs normally
```
