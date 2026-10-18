---
description: 'Exclude files in Rslint with global or entry-level ignores, glob patterns, globalIgnores, and .gitignore integration.'
---

# ignores

- **Type:** `string[]`

Glob patterns for files to exclude. The behavior depends on whether `ignores` appears in a global ignore entry or alongside other configuration fields.

## Global and entry-level ignores

An entry containing `ignores` and no fields other than optional `name` and `basePath` acts as a global ignore: matching files are removed from the lint target set. An entry-level ignore prevents that entry's `files` selector, rules, and options from contributing. It cannot remove a path selected by the default extension baseline or another entry, so such a path may still receive configuration or a zero-rule syntax pass.

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

## The `globalIgnores` helper

Writing a global-ignore entry is common enough that `@rslint/core` exports a `globalIgnores` helper, mirroring ESLint v10. It returns a config entry containing the given patterns, so the global-ignore intent is explicit:

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

## Pattern types in global ignores

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
  ignores: ['build/**/*', '!build/test.js'],
}

// ❌ dir/** blocks traversal — negation has no effect
{
  ignores: ['build/**', '!build/test.js'],
}
```

:::

:::tip
`node_modules` and `.git` are automatically excluded by rslint — you don't need to add them to ignores.
:::

## .gitignore integration

The CLI, JavaScript API, and LSP automatically apply the `.gitignore` rules for each target tree. Starting at the target's physical directory, rslint searches upward for the nearest `.git` directory or file, including Git worktree checkouts. It reads applicable `.gitignore` files from that repository root down to the target. Without a `.git` marker, the ancestor search ends at the filesystem root (the drive or network-share root on Windows).

Selecting the same config and target from a package directory therefore keeps the same Git ignore rules as selecting them from the repository root. Neither the config file's location nor an entry's [`basePath`](/config/base-path) changes the Git boundary. Authored `files` and `ignores` keep their own configuration-relative bases.

- Parent patterns apply to descendants, including across child config files.
- Each nested `.git` starts a separate Git boundary, even if that repository has no `.gitignore`.
- Nested `.gitignore` files affect only their own subtree. Their negations cannot re-include a file while a parent directory remains excluded.
- Directly supplied files also obey Git ignores. Directory aliases use the physical directory's rules; a file symlink keeps the rules for its containing directory and original filename.
- A global config `!` pattern can override the final Git ignore result. It does not make `.gitignore` files inside a Git-excluded directory readable.

For example, a bare directory pattern applies at every depth:

```text
# .gitignore at the repository root
dist/
```

This ignores both `dist/index.js` and `packages/app/dist/index.js`, regardless of where rslint runs or which config owns the file.

To allow selected contents with Git patterns, keep their parent directories reachable:

```text
# .gitignore
dist/*
!dist/important.ts
```

A directory request visits its ancestor chain and reachable descendants. It does not scan unrelated directories to locate repositories, and it does not search beneath a pruned directory for hidden nested repositories. Directly requesting such a nested repository establishes its own entry point.

Git ignores can stop automatic config discovery from entering an excluded directory. They do not filter config filenames in a reachable directory or prevent an explicitly selected config module from loading. In the editor, saved `.gitignore` changes and `.git` marker creation or deletion refresh the config and Git policy together. A newly opened target outside the committed policy is checked after a successful refresh; a failed refresh does not lint or fix it with an incomplete policy.
