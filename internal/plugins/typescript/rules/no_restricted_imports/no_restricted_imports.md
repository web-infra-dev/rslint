# no-restricted-imports

Disallow specified modules when loaded by `import`.

## Rule Details

This rule extends the base [`no-restricted-imports`](https://eslint.org/docs/latest/rules/no-restricted-imports) rule. It adds first-class support for TypeScript-only forms:

- Type-only imports: `import type Foo from 'foo'`
- Inline type specifiers: `import { type Foo } from 'foo'`
- CommonJS `import = require(...)` (including `import type x = require(...)`)
- Type-only re-exports: `export type { Foo } from 'foo'`

The schema, message ids, and reporting positions are all identical to the base rule. The only addition is the `allowTypeImports` option on each `paths` entry and `patterns` entry — when set, type-only imports/exports of the matching source are exempted from the restriction.

## Options

The rule accepts the same option shapes as the base `no-restricted-imports` rule. Each entry in `paths` (object form) and `patterns` (object form) may also set:

- `allowTypeImports: boolean` — when `true`, do not report a type-only import or export of this path/pattern. Default `false`.

Examples of **incorrect** code with `["error", { "paths": [{ "name": "import-foo", "message": "Use import-bar instead.", "allowTypeImports": true }] }]`:

```json
{
  "@typescript-eslint/no-restricted-imports": [
    "error",
    {
      "paths": [
        {
          "name": "import-foo",
          "message": "Use import-bar instead.",
          "allowTypeImports": true
        }
      ]
    }
  ]
}
```

```ts
import foo from 'import-foo';
export { foo } from 'import-foo';
```

Examples of **correct** code with the same options:

```ts
import type foo from 'import-foo';
import type _ = require('import-foo');
export type { foo } from 'import-foo';
```

Examples of **incorrect** code with `["error", { "patterns": [{ "group": ["import1/private/*"], "message": "private modules are not allowed.", "allowTypeImports": true }] }]`:

```json
{
  "@typescript-eslint/no-restricted-imports": [
    "error",
    {
      "patterns": [
        {
          "group": ["import1/private/*"],
          "message": "private modules are not allowed.",
          "allowTypeImports": true
        }
      ]
    }
  ]
}
```

```ts
import foo from 'import1/private/bar';
```

Examples of **correct** code with the same options:

```ts
import type foo from 'import1/private/bar';
export type { foo } from 'import1/private/bar';
```

## When Not To Use It

If you do not need to restrict imports of any modules, do not enable this rule.

## Original Documentation

- [typescript-eslint rule documentation](https://typescript-eslint.io/rules/no-restricted-imports)
- [Base ESLint rule documentation](https://eslint.org/docs/latest/rules/no-restricted-imports)
