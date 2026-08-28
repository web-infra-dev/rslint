# prefer-import-in-mock

## Rule Details

Requires the module passed to `rs.mock()` or `rs.doMock()` to be written as a dynamic `import()` rather than as a bare path. Both forms mock the same module, but the `import()` form ties the call to the module's real type, so a factory that returns the wrong shape, a stale export name, and a path that no longer resolves all become type errors instead of a mock that silently does nothing. It also gives the factory's return value completion in the editor.

The rule covers the calls Rstest hoists and rewrites: `mock` and `doMock` read off `rs` or `rstest`, whether those come from an `import` of `@rstest/core`, a `require` of it, or the [globals](https://rstest.rs/config/test/globals) Rstest installs. `mockRequire` and `doMockRequire` mock the CommonJS entry and accept a module name only, so a path given to either of them is left as written, as is one given to `unmock`, `importActual`, `importMock`, and the rest of the module registry APIs.

Only a path written as a quoted string is reported — a path built at runtime has no type to recover. A call that already carries an explicit type argument, `rs.mock<{ total: number }>('./cart')`, is left alone as well: it has already stated the mocked module's shape. So is a call that Rstest's mock transform does not rewrite in the first place, including one made through a renamed binding such as `import { rstest as vi } from '@rstest/core'`, through a namespace import, or with optional chaining; those calls fail at run time with a message from Rstest, and reporting them here would only suggest a change that leaves them failing.

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

Wraps the path in `import()`. The path is reused exactly as it was written, so its quote characters and any escape sequences inside it are preserved, and every other argument, the mock factory and the `{ spy: true }` / `{ mock: true }` options object included, is left untouched. Setting `fixable` to `false` reports the call without a fix.
