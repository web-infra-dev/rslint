# consistent-type-specifier-style

## Rule Details

Enforces a consistent position for `type` markers in named imports. This rule is automatically fixable and does not require type information.

Default imports, namespace imports, side-effect imports, and empty imports do not need named type markers. The fixer can leave extra spaces. It does not merge duplicate import declarations; use `import/no-duplicates` to enforce that separately.

## Options

The rule accepts one string:

- `"prefer-inline"`: put `type` on each named type import.
- `"prefer-top-level"`: use a separate `import type` declaration.

When the option is omitted, the rule behaves as `"prefer-top-level"`, matching eslint-plugin-import v2.32.0's runtime. Its upstream documentation and schema advertise `"prefer-inline"` as the default; specify the option explicitly to choose that style.

### `prefer-inline`

Examples of **incorrect** code for this rule with the `"prefer-inline"` option:

```typescript
import type { Foo, Bar as Baz } from 'module';
```

Examples of **correct** code for this rule with the `"prefer-inline"` option:

```typescript
import { type Foo, type Bar as Baz } from 'module';
```

### `prefer-top-level`

Examples of **incorrect** code for this rule with the `"prefer-top-level"` option:

```typescript
import { value, type Foo } from 'module';
```

Examples of **correct** code for this rule with the `"prefer-top-level"` option:

```typescript
import { value } from 'module';
import type { Foo } from 'module';
```

## Differences from ESLint

- **Flow `typeof` imports are unsupported.** This rule supports TypeScript type imports, but not Flow's `import typeof Foo from 'foo'` or `import { typeof Foo } from 'foo'` forms.
- **Imports with attributes are not autofixed.** Imports with `with { ... }`, including `import type` declarations with a `"resolution-mode"` attribute, are reported without an automatic fix to preserve their module resolution and avoid invalid TypeScript.
- **Autofixes do not delete comments.** For example, `import { type /*A*/ Foo /*B*/ as /*C*/ Bar } from 'm';` is reported without a fix under `"prefer-top-level"`. Fixes remain available when the comments can stay in place.
- **Autofixes preserve quoted import names.** With `"prefer-top-level"`, rslint keeps the original name and its quotes. eslint-plugin-import v2.32.0 replaces `'a-b'` with `undefined` in this case, which changes the imported name:

  ```typescript
  // Before
  import { type 'a-b' as A } from 'module';

  // After
  import type {'a-b' as A} from 'module';
  ```

## Original Documentation

- [eslint-plugin-import: consistent-type-specifier-style](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/consistent-type-specifier-style.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/consistent-type-specifier-style.js)
