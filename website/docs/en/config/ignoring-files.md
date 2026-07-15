# Ignoring Files

## ignores

- **Type:** `string[]`

Glob patterns for files to exclude. An entry containing **only** `ignores` and an optional `name` acts as a global ignore: matching files are removed from the lint target set. An entry-level ignore prevents that entry's `files` selector, rules, and options from contributing. It cannot remove a path selected by the default extension baseline or another entry, so such a path may still receive configuration or a zero-rule syntax pass.

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

### The `globalIgnores` helper

Writing an entry with only `ignores` is common enough that `@rslint/core` exports a `globalIgnores` helper, mirroring ESLint v10. It returns a config entry containing just the given patterns, so the global-ignore intent is explicit:

```ts
import { defineConfig, globalIgnores } from '@rslint/core';

export default defineConfig([
  globalIgnores(['**/dist/**', '**/fixtures/**']),
  // ... other entries
]);
```

This is exactly equivalent to writing the entry by hand:

```ts
{
  ignores: ['**/dist/**', '**/fixtures/**'],
}
```

`globalIgnores` throws a `TypeError` if it receives a non-array or an empty array.

### Pattern types in global ignores

Global ignore patterns affect both file matching and directory traversal (including config discovery in monorepos):

| Pattern    | Effect                                               |
| ---------- | ---------------------------------------------------- |
| `dir/**`   | Ignores directory and all contents, blocks traversal |
| `dir/**/*` | Ignores files inside, but allows directory traversal |
| `dir/*`    | Ignores direct children files only                   |

Use `dir/**` to completely exclude a directory. Use `dir/**/*` when the walker must still enter the directory so later negations can make selected files or config candidates reachable. An automatically discovered `rslint.config.*` that still matches the file-cover ignore is not loaded; explicitly negate that candidate when you want it to become a config boundary.

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

The CLI, JavaScript API, and LSP automatically read `.gitignore` files and treat their patterns as additional global ignores. Collection starts at the directory of the governing rslint config and never searches its parents. In a multi-config repository, a child config starts a new `.gitignore` scope, so put package-specific ignores beside that package's config. In the editor, saved `.gitignore` changes refresh diagnostics for open files.

- **Nested `.gitignore` files** inside one config-owned tree are supported — each one only affects its own directory subtree
- **Parent patterns cascade** to child directories within that tree (e.g., root `dist/` also ignores `packages/app/dist/` when both use the root config)
- **Child `.gitignore` can override** parent patterns with `!` negation
- **Child configs are boundaries** — they do not inherit `.gitignore` files from a parent config directory
- Config `!` negation can also override `.gitignore` patterns (they are evaluated sequentially in the same global ignores list)

```text
# .gitignore
dist/
coverage/
*.log

# packages/app/.gitignore
!dist/          # re-include dist/ under packages/app/
```

To make a lint target reachable, re-include it in the applicable `.gitignore` policy or with a later global ignore entry in `rslint.config.*`:

```text
# .gitignore
dist/*
!dist/important.ts
```

Configuration discovery is independent of `.gitignore`: an automatically discovered or explicitly selected `rslint.config.*` is still loaded when its path matches an ignore rule. `.gitignore` is applied later, when Go selects lint targets inside that config's ownership scope.
