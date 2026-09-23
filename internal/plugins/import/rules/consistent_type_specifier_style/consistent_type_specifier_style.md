# consistent-type-specifier-style

Enforce a consistent position for `type` markers in named imports. This rule is
automatically fixable and does not require type information.

## Options

The rule accepts one string:

- `"prefer-inline"`: put `type` on each named type import.
- `"prefer-top-level"`: use a separate `import type` declaration.

When the option is omitted, the rule behaves as `"prefer-top-level"`, matching
eslint-plugin-import v2.32.0's runtime. Its upstream documentation and schema
advertise `"prefer-inline"` as the default; specify the option explicitly to
choose that style.

### Prefer inline

With `["error", "prefer-inline"]`, this is incorrect:

```ts
import type { Foo, Bar as Baz } from 'module';
```

Use inline markers instead:

```ts
import { type Foo, type Bar as Baz } from 'module';
```

### Prefer top-level

With `["error", "prefer-top-level"]`, this is incorrect:

```ts
import { value, type Foo } from 'module';
```

Use a separate type import instead:

```ts
import { value } from 'module';
import type { Foo } from 'module';
```

Default imports, namespace imports, side-effect imports, and empty imports do
not need named type markers. The fixer preserves upstream formatting behavior
and can leave extra spaces. It does not merge duplicate import declarations;
use `import/no-duplicates` to enforce that separately.

## Differences from upstream

This rule supports TypeScript type imports. Flow's `import typeof Foo from 'foo'`
and `import { typeof Foo } from 'foo'` forms are unsupported.

Autofixes preserve quoted import names. For example, with `"prefer-top-level"`:

```ts
// Before
import { type 'a-b' as A } from 'module';

// After
import type {'a-b' as A} from 'module';
```

eslint-plugin-import v2.32.0 replaces `'a-b'` with `undefined` in this case,
which changes the imported name. rslint keeps the original name and its quotes.

## Original documentation

- [eslint-plugin-import v2.32.0 documentation](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/consistent-type-specifier-style.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/consistent-type-specifier-style.js)
