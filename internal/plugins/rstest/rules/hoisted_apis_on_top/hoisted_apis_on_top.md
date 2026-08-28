# hoisted-apis-on-top

## Rule Details

Requires the module-mock APIs that Rstest lifts to the top of a module to be written there. [`rs.mock`, `rs.mockRequire`, `rs.unmock`, `rs.unmockRequire` and `rs.hoisted`](https://rstest.rs/api/runtime-api/rstest/mock-modules) all run before the rest of the file, ahead of its own imports, wherever they appear. A call written inside an `if`, a loop, a function, or a test body therefore reads as conditional or scoped when it is neither: the mock is registered for the whole file either way. `rs.doMock`, `rs.doMockRequire`, `rs.doUnmock` and `rs.doUnmockRequire` are the same four module operations without the lift, so they belong in a runtime location and are never reported.

The receiver has to be written as `rs` or `rstest`, because that is how the build recognizes these calls. Parentheses and TypeScript wrappers around it are transparent, so `(rs).mock()`, `rs!.mock()` and `(rs as any).mock()` are covered too. No import or scope analysis is involved, which matches how the calls are rewritten: a `rs` imported from another package, or shadowed by a local declaration, still has its call lifted and is still reported. Conversely, a call written through a renamed binding (`import { rs as r }`, then `r.mock()`), through `import.meta.rstest`, through a computed property (`rs['mock']()`), or on an optional chain (`rs?.mock()`) is not lifted at all — it runs and fails where it is written — so this rule leaves it alone.

`rs.hoisted` may also be written as the initializer of a top-level variable, with or without `await`, since sharing its return value with the rest of the file is what it exists for. The other four evaluate to `undefined`, so a statement of the file is the only shape accepted for them; `const done = rs.mock('./payment-gateway')` is reported even at the top of the file.

## Incorrect

```ts
describe('checkout', () => {
  it('reports a declined card', async () => {
    rs.mock('./payment-gateway', () => ({ charge: rs.fn() }));

    const { charge } = await import('./payment-gateway');

    expect(charge).not.toHaveBeenCalled();
  });
});
```

## Correct

```ts
rs.mock('./payment-gateway', () => ({ charge: rs.fn() }));

describe('checkout', () => {
  it('reports a declined card', async () => {
    const { charge } = await import('./payment-gateway');

    expect(charge).not.toHaveBeenCalled();
  });
});
```

## Suggestions

Moving the call to the top places it after the file's last import, or at the very start when the file has none. A call written as a statement of its own leaves an empty statement behind, so an `if` written without braces keeps a body. A `rs.hoisted` bound by a single-declarator variable statement travels together with its binding, since lifting only the call would assign the binding when its block runs while the factory had already executed. The other four APIs leave `undefined` where their value was read, which is what they evaluated to. A `rs.hoisted` whose value is read from anywhere else — an argument, an object property, a declaration that also binds something else, a `for` header — is reported without this suggestion, because nothing can stand in for the value it produced.

The second suggestion rewrites `mock`, `mockRequire`, `unmock` and `unmockRequire` to `doMock`, `doMockRequire`, `doUnmock` and `doUnmockRequire`, keeping the call where it is and giving it the timing the surrounding code implies. `rs.hoisted` has no non-hoisted counterpart, so it offers only the move.
