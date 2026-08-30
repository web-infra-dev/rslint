# require-mock-type-parameters

## Rule Details

Requires `rs.fn()` to be given a type parameter. Without one the mock is typed as a function that accepts any arguments and returns `any`, so nothing about how the test drives it is checked: a call with the wrong arguments, a `mockResolvedValue` of the wrong shape, and an assertion against a property the real function never returns all pass the compiler. Writing the signature — `rs.fn<(id: number) => Promise<User>>()` — makes the mock stand in for the function it replaces, in the editor as well as at build time.

`fn` is an ordinary function on the utilities object, so the rule follows the binding: it is reported whether the receiver came from an `import` of `@rstest/core` under either of its names, a `require` of it, a further rename such as `import { rs as mocker }`, a namespace as in `core.rs.fn()`, or the [globals](https://rstest.rs/config/test/globals) Rstest installs. A receiver the file declares itself — a local object, a function parameter, a binding imported from somewhere else — is a different object and is left alone. The call has to go through a plain dotted member, so `rs['fn']()` is not reported, and `rs.fn` read as a value without being called carries no call to write a type parameter on.

With `checkImportFunctions` enabled the rule also covers the four APIs that load a module: `importActual`, `importMock`, `requireActual` and `requireMock`. Each returns `Record<string, unknown>` unless told what the module exports, so the result has no named exports and nothing callable on it. These four are rewritten by Rstest's build rather than called, which is matched on the name written at the call site: the receiver must be spelled `rs` or `rstest`, so a renamed binding and a namespace are left as they are, because Rstest leaves them alone too and they throw where they stand. The member itself may be written either way — `rs.importActual('./dep')`, `rs['importActual']('./dep')` and `rs.importActual?.('./dep')` are all rewritten and all reported — while an optional receiver, `rs?.importActual('./dep')`, is not rewritten and is left alone. For the same reason a receiver the file declares as a local variable or a parameter is not reported, and neither is a call the build cannot rewrite — a path that is not a quoted string, or any argument count other than one.

Parentheses and TypeScript's type-only syntax are transparent throughout: `(rs as any).fn()`, `rs!.fn()` and `(rs.fn)()` are reported exactly like the bare form. A call that already carries a type argument is what the rule asks for and is never reported.

## Incorrect

```ts
import { expect, rs, test } from '@rstest/core';
import { checkout } from './checkout';

test('charges the card once', async () => {
  const charge = rs.fn();

  await checkout({ items: [{ sku: 'A1', price: 10 }] }, charge);

  expect(charge).toHaveBeenCalledExactlyOnceWith(10);
});
```

## Correct

```ts
import { expect, rs, test } from '@rstest/core';
import { checkout } from './checkout';

test('charges the card once', async () => {
  const charge = rs.fn<(amount: number) => Promise<void>>();

  await checkout({ items: [{ sku: 'A1', price: 10 }] }, charge);

  expect(charge).toHaveBeenCalledExactlyOnceWith(10);
});
```

## Options

```json
{
  "rstest/require-mock-type-parameters": [
    "error",
    {
      "checkImportFunctions": true
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `checkImportFunctions` | `boolean` | `false` | Also require type parameters for `importActual`, `importMock`, `requireActual` and `requireMock`. |
