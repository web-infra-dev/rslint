# Ignoring Files

## ignores

- **Type:** `string[]`

Glob patterns for files to exclude. An entry with **only** `ignores` (no other fields) acts as a global ignore — matching files are excluded from all rules.

```ts
// Global ignore entry
{
  ignores: ['**/dist/**', '**/fixtures/**'],
}

// Entry-level ignore (only applies to this entry)
{
  files: ['**/*.ts'],
  ignores: ['**/*.test.ts'],
  rules: { /* ... */ },
}
```

### Pattern types in global ignores

Global ignore patterns affect both file matching and directory traversal (including config discovery in monorepos):

| Pattern    | Effect                                               |
| ---------- | ---------------------------------------------------- |
| `dir/**`   | Ignores directory and all contents, blocks traversal |
| `dir/**/*` | Ignores files inside, but allows directory traversal |
| `dir/*`    | Ignores direct children files only                   |

Use `dir/**` to completely exclude a directory (including any nested configs inside it). Use `dir/**/*` if you only want to ignore files but still allow nested configs to be discovered.

You can use `!` negation patterns to re-include specific files. Patterns are evaluated sequentially — later patterns override earlier ones:

```ts
// Global ignore: re-include specific file
{
  ignores: ['build/**/*', '!build/test.js'],
}

// Entry-level ignore: re-include a subdirectory
{
  files: ['**/*.ts'],
  ignores: ['vendor/**/*', '!vendor/keep/**/*'],
  rules: { /* ... */ },
}

// Across separate global ignore entries
{ ignores: ['build/**/*'] },
{ ignores: ['!build/test.js'] },
```

:::warning
For directory-level patterns (`dir/**`), `!` negation cannot re-include files because the directory traversal is blocked entirely. Use `dir/**/*` instead if you need negation:

```ts
// ✅ dir/**/* allows traversal — negation works
{
  ignores: ['build/**/*', '!build/test.js'];
}

// ❌ dir/** blocks traversal — negation has no effect
{
  ignores: ['build/**', '!build/test.js'];
}
```

:::

:::tip
`node_modules` and `.git` are automatically excluded by rslint — you don't need to add them to ignores.
:::

## .gitignore integration

Rslint automatically reads `.gitignore` files and treats their patterns as additional global ignores. This means files ignored by git (build outputs, coverage reports, etc.) are also ignored by the linter without extra configuration.

- **Nested `.gitignore` files** are supported — each one only affects its own directory subtree
- **Parent patterns cascade** to child directories (e.g., root `dist/` also ignores `packages/app/dist/`)
- **Child `.gitignore` can override** parent patterns with `!` negation
- Config `!` negation can also override `.gitignore` patterns (they are evaluated sequentially in the same global ignores list)

```text
# .gitignore
dist/
coverage/
*.log

# packages/app/.gitignore
!dist/          # re-include dist/ under packages/app/
```

If you need to lint a file that is in `.gitignore`, add a `!` negation pattern in your config's global ignores:

```ts
export default [
  {
    ignores: ['!dist/important.ts'], // override .gitignore for this file
  },
  // ...
];
```
