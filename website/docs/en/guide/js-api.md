# JavaScript API

The JavaScript API lets you run rslint programmatically — lint files or in-memory source from a JavaScript runtime script, an editor integration, or a build tool. It is designed for JavaScript runtime hosts such as Node.js, Bun, or Deno when they can load npm packages and provide the Node-compatible filesystem and process APIs that `@rslint/core` uses. Its surface is aligned with [ESLint](https://eslint.org/docs/latest/integrate/nodejs-api)'s v10 programmatic API shape, so most ESLint API code ports over with minimal changes.

All config resolution (override config, config-file selection, discovery, normalization) happens in JavaScript; rslint's engine receives only the final resolved config object and never reads config from disk.

## Getting started

```ts
import { Rslint } from '@rslint/core';

const rslint = new Rslint();
const results = await rslint.lintFiles(['src/**/*.ts']);

for (const result of results) {
  console.log(result.filePath, result.errorCount, result.warningCount);
}
```

`new Rslint(options)` creates a linter instance. Both `lintFiles` and `lintText` are async and return an ESLint-shaped `LintResult[]`.

## Linting files

`lintFiles` takes one or more glob patterns (resolved against `cwd`) and lints every matching file. By default the nearest config is auto-discovered from `cwd`.

```ts
const results = await rslint.lintFiles(['src/**/*.ts', 'test/**/*.ts']);
```

Results are ordered by the linted file's path (deterministic), not by glob-walk order.

If no file matches the patterns, `lintFiles` returns an empty array rather than throwing — unlike ESLint v10, whose default `errorOnUnmatchedPattern` throws on an unmatched glob.

## Linting a string

`lintText` lints an in-memory string as if it lived at `filePath`:

```ts
const [result] = await rslint.lintText('const x = 1', {
  filePath: 'example.ts',
});
```

`lintText` always returns exactly one result — for the linted buffer. If you omit `filePath`, the result's `filePath` is the `"<text>"` sentinel (matching ESLint).

## Fully in-memory linting

By default `lintText` still reads the config and tsconfig from disk. To lint with **no disk access at all** — source, config, and tsconfig all in memory — combine `overrideConfigFile: true` (use only the inline config), an inline `overrideConfig`, and a `virtualFiles` overlay:

```ts
const rslint = new Rslint({
  cwd: '/', // virtual root: doesn't touch the host cwd or disk
  overrideConfigFile: true, // use only overrideConfig — skip config discovery
  overrideConfig: [
    {
      files: ['**/*.ts'],
      // The tsconfig + parserOptions.project below are needed ONLY for
      // type-aware rules (like no-for-in-array). Syntax-only rules need neither.
      languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
      plugins: ['@typescript-eslint'],
      rules: { '@typescript-eslint/no-for-in-array': 'error' },
    },
  ],
  virtualFiles: {
    'tsconfig.json': JSON.stringify({
      compilerOptions: { strict: true },
      files: ['./a.ts'],
    }),
  },
});

const [result] = await rslint.lintText(
  'const a = [1];\nfor (const k in a) {}\n',
  { filePath: 'a.ts' },
);
```

`virtualFiles` is an in-memory file overlay (path → content) — an rslint extension; ESLint has no in-memory file map. Type-aware rules run against it with no disk access: put the `tsconfig.json` that `parserOptions.project` names, plus any dependency files, in the overlay.

**Declaring plugins.** A rule from a plugin (`@typescript-eslint/*`, `unicorn/*`, and so on) runs only when that plugin is listed in `plugins` — rslint enforces this exactly like ESLint. Core rules (no `/` prefix) need no declaration.

**Type-aware vs syntax-only rules.** The `tsconfig.json` and `parserOptions.project` matter only for **type-aware** rules, which need a real TypeScript program (see [Type Checking](/guide/type-checking)). If your config has only syntax-only rules, you can drop both — no tsconfig, no `parserOptions.project`.

**Use relative paths** in `virtualFiles` keys and inside the tsconfig:

- **`virtualFiles` keys**: prefer relative paths (`'tsconfig.json'`). Keys are resolved against `cwd`, so a relative key always lines up with `parserOptions.project` (also resolved against `cwd`). An absolute key like `'/tsconfig.json'` happens to match only when `cwd` is `/`; with any other `cwd` it lands at the filesystem root and won't match the project path.
- **Inside the tsconfig** (`files`) and in **`parserOptions.project`**: use relative paths. The TypeScript compiler resolves these, and a bare POSIX-absolute path (such as `/a.ts`) has no drive letter on Windows, so it won't match the overlay.

**Pin the tsconfig to explicit `files`** — a broad `include` glob is expanded against the real filesystem and would scan `cwd` on disk.

## Auto-fixing

Pass `fix: true`. A result whose file a fix changed then carries an `output` string — the full fixed source; results with no applied fix have no `output`.

**Write fixes to disk** with the static `Rslint.outputFixes`:

```ts
const rslint = new Rslint({ fix: true });
const results = await rslint.lintFiles(['src/**/*.ts']);
await Rslint.outputFixes(results); // writes fixed files back to disk
```

`Rslint.outputFixes` writes back only results whose `filePath` is absolute. A `lintText` result is absolute — and so will be written — when you pass a `filePath`; only a result with no `filePath` (the non-absolute `"<text>"` sentinel) is skipped.

**Apply fixes in memory** — to fix without touching disk, read `output` directly and don't call `outputFixes`:

```ts
const rslint = new Rslint({ fix: true /* + your in-memory config */ });
const [result] = await rslint.lintText('let x = foo!!.bar', {
  filePath: 'a.ts',
});
const fixed = result.output ?? 'let x = foo!!.bar'; // fixed source, or the original if nothing changed
```

`lintText` with `fix: true` never writes to disk — the fixed source comes back as `result.output`. For edit-level control, each `result.messages[].fix` is a `{ range: [start, end], text }` edit (UTF-16 offsets) you can splice into the source yourself; for more than one fix prefer `output`, which is already the safely merged whole-file result.

## Lifecycle

Each `Rslint` instance owns a long-lived rslint engine child process. You **don't** need to call `close()` — like ESLint, a one-off script exits cleanly on its own (the idle child is unref'd, so it never blocks the event loop).

Call `close()` only in a long-running host (an editor server, a watch process) that creates many instances, to free each child promptly:

```ts
const rslint = new Rslint();
try {
  await rslint.lintFiles(['src/**/*.ts']);
} finally {
  await rslint.close();
}
```

Or use `await using` for automatic disposal at the end of scope:

```ts
await using rslint = new Rslint();
await rslint.lintFiles(['src/**/*.ts']);
```

Native `await using` needs a runtime with explicit resource management support, which Node.js 22 lacks (a bare `.mjs` throws a SyntaxError). Compile with a `using`-aware toolchain such as TypeScript 5.2+, or use the `try` / `finally` form above, which does not rely on native `using` syntax.

## Result shape

Both methods resolve to `LintResult[]`:

| Field                 | Type            | Description                                                         |
| --------------------- | --------------- | ------------------------------------------------------------------- |
| `filePath`            | `string`        | Absolute path, or `"<text>"` for `lintText` called with no filePath |
| `messages`            | `LintMessage[]` | Diagnostics for this file                                           |
| `errorCount`          | `number`        | Number of error-severity messages                                   |
| `warningCount`        | `number`        | Number of warning-severity messages                                 |
| `fixableErrorCount`   | `number`        | Errors that have an auto-fix                                        |
| `fixableWarningCount` | `number`        | Warnings that have an auto-fix                                      |
| `output`              | `string?`       | Fixed source — present only when `fix: true` changed the file       |

Each `LintMessage`:

| Field                   | Type                                         | Description                                                                  |
| ----------------------- | -------------------------------------------- | ---------------------------------------------------------------------------- |
| `ruleId`                | `string \| null`                             | The rule that produced the message (`null` if none)                          |
| `severity`              | `1 \| 2`                                     | `2` = error, `1` = warning                                                   |
| `message`               | `string`                                     | Human-readable message                                                       |
| `messageId`             | `string?`                                    | Stable message id, when the rule provides one                                |
| `line` / `column`       | `number`                                     | 1-based position (`column` counts UTF-16 code units)                         |
| `endLine` / `endColumn` | `number?`                                    | 1-based end position, when available                                         |
| `fix`                   | `{ range: [number, number]; text: string }?` | Flat UTF-16 offset range + replacement text                                  |
| `suggestions`           | `LintSuggestion[]?`                          | Suggested fixes (each with `desc`, `fix`, and optional `messageId` / `data`) |

## Options

`new Rslint(options)` accepts:

| Option               | Type                                               | Default     | Description                                                                                                                            |
| -------------------- | -------------------------------------------------- | ----------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `cwd`                | `string`                                           | current cwd | Base directory for config discovery and relative path resolution                                                                       |
| `overrideConfig`     | `RslintConfigEntry \| RslintConfigEntry[] \| null` | —           | Extra config appended after the resolved/discovered config (ESLint's `overrideConfig`)                                                 |
| `overrideConfigFile` | `string \| true \| null`                           | `null`      | `string`: use this config file (no discovery); `true`: use only `overrideConfig` (no file, no discovery); `null`/absent: auto-discover |
| `fix`                | `boolean`                                          | `false`     | Apply rule auto-fixes; results carry `output`                                                                                          |
| `virtualFiles`       | `Record<string, string>`                           | —           | In-memory file overlay (path → content) for fully in-memory linting                                                                    |
