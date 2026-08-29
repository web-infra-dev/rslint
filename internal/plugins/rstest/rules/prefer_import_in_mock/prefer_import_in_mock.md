# prefer-import-in-mock

## Rule Details

Requires the module passed to `rs.mock()` or `rs.doMock()` to be written as a dynamic `import()` rather than as a bare path. Both forms mock the same module, but the `import()` form ties the call to the module's real type, so a factory that returns the wrong shape, a stale export name, and a path that no longer resolves all become type errors instead of a mock that silently does nothing. It also gives the factory's return value completion in the editor.

The rule covers the calls Rstest hoists and rewrites: `mock` and `doMock` read off a receiver spelled `rs` or `rstest`. That spelling is all the rule asks for, because it is all Rstest asks for — the call is rewritten whether the name came from an `import` of `@rstest/core`, a `require` of it, the [globals](https://rstest.rs/config/test/globals) Rstest installs, a local variable, or a function parameter. A receiver spelled anything else is never rewritten, so `vi.mock('./cart')` after `import { rstest as vi } from '@rstest/core'` is not reported here; neither is a receiver reached through a namespace, a computed member such as `rs['mock']`, or an optional chain. `mockRequire` and `doMockRequire` mock the CommonJS entry and accept a module name only, so a path given to either of them is left as written, as is one given to `unmock`, `importActual`, `importMock`, and the rest of the module registry APIs.

Because Rstest lifts the call above the module's imports, only a call that stands on its own as a statement is reported. One whose value is consumed — an argument to another call, a variable initializer, an operand of a comma expression, an awaited expression, the body of an arrow function — is either left untransformed, and throws, or is lifted out of an expression that no longer parses without it. The argument list has to be one the transform can read positionally too: a spread argument makes it give up, and a third argument fails the build. Wrapping the path repairs none of those, so none of them is reported.

Parentheses and TypeScript's type-only syntax are transparent throughout, around the call, the callee, the receiver, or the path: `(rs as any).mock('./cart')`, `rs!.mock('./cart')` and `rs.mock(('./cart'))` are reported exactly like the bare form.

Only a path written as a quoted string is reported; a path built at runtime has no type to recover, and a template literal path fails Rstest's build whether or not this rule rewrites it. A call that already carries an explicit type argument, `rs.mock<{ total: number }>('./cart')`, is left alone as well, since it has already stated the mocked module's shape.

## Incorrect

```ts
import { expect, rs, test } from '@rstest/core';
import { fetchUser } from './user-service';

rs.mock('./user-service', () => ({
  fetchUser: rs.fn().mockResolvedValue({ id: 1, name: 'Ada' }),
}));

test('loads the current user', async () => {
  await expect(fetchUser(1)).resolves.toEqual({ id: 1, name: 'Ada' });
});
```

## Correct

```ts
import { expect, rs, test } from '@rstest/core';
import { fetchUser } from './user-service';

rs.mock(import('./user-service'), () => ({
  fetchUser: rs.fn().mockResolvedValue({ id: 1, name: 'Ada' }),
}));

test('loads the current user', async () => {
  await expect(fetchUser(1)).resolves.toEqual({ id: 1, name: 'Ada' });
});
```

## Options

```json
{
  "rstest/prefer-import-in-mock": [
    "error",
    {
      "fixable": false
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `fixable` | `boolean` | `true` | Attach the autofix to the report. |

## Autofix

Wraps the path in `import()`. The path is reused exactly as it was written, so its quote characters and any escape sequences inside it are preserved, and every other argument, the mock factory and the `{ spy: true }` / `{ mock: true }` options object included, is left untouched.

The whole first argument is replaced, so parentheses around the path go, and so does an assertion such as `'./cart' as string`, which no longer describes what the argument holds. A comment written in the space that goes with them would be deleted, so the call is then reported without a fix. Setting `fixable` to `false` reports every call without a fix.
