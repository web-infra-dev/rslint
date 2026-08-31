# hoisted-apis-on-top

## Rule Details

Requires the module-mock APIs that Rstest lifts to the top of a module to be written there. [`rs.mock`, `rs.mockRequire`, `rs.unmock`, `rs.unmockRequire` and `rs.hoisted`](https://rstest.rs/api/runtime-api/rstest/mock-modules) all run before the rest of the file, ahead of its own imports, wherever they appear. A call written inside an `if`, a loop, a function, or a test body therefore reads as conditional or scoped when it is neither: the mock is registered for the whole file either way. `rs.doMock`, `rs.doMockRequire`, `rs.doUnmock` and `rs.doUnmockRequire` are the same four module operations without the lift, so they belong in a runtime location and are never reported.

A call only counts when Rstest actually lifts it, which takes two things. The call has to be written as the whole of a statement, or as the whole initializer of a variable declaration — the latter optionally behind an `await`. Written anywhere else — as an argument, an object property, an arrow's expression body, an operand of `&&`, the right-hand side of an assignment, or with an `await` in statement position — it is left where it is and fails when it runs, which is a different problem from the one this rule describes. And the call has to name the API as a plain dotted member off `rs` or `rstest`: a renamed binding (`import { rs as r }`, then `r.mock()`), `import.meta.rstest`, a computed property (`rs['mock']()`) and an optional chain (`rs?.mock()`) are all left alone as well. Parentheses and TypeScript wrappers are transparent everywhere — around the receiver, around the callee, and around the complete call — so `(rs).mock()`, `rs!.mock()`, `(rs.mock as any)()` and `rs.mock('./m') as unknown` are all covered.

No import or scope analysis is involved, which matches how the calls are rewritten: a `rs` imported from another package, or shadowed by a local declaration, still has its call lifted and is still reported.

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

Moving the call to the top writes its statement after the file's shebang, its directive prologue and its last import, whichever comes last, so nothing that has to stay first is displaced. A statement of its own leaves an empty statement behind, which keeps a body under an `if` written without braces. A declaration moves whole, since the lift takes only the call and would otherwise leave the binding assigned when its block runs while the factory had already executed. The move is withheld when the declaration cannot travel alone — a second declarator in the same statement, or a `for` header — and when a name it binds already exists at the top of the file, where re-declaring it would not parse.

The second suggestion rewrites `mock`, `mockRequire`, `unmock` and `unmockRequire` to `doMock`, `doMockRequire`, `doUnmock` and `doUnmockRequire`, keeping the call where it is and giving it the timing the surrounding code implies. `rs.hoisted` has no non-hoisted counterpart, so it offers only the move.
