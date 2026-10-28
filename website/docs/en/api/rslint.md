---
description: 'Use the Rslint class to lint files and in-memory code, write fixes, and dispose of programmatic lint sessions.'
---

# Rslint

The `Rslint` class lets you lint files and in-memory code from JavaScript.

```ts
class Rslint {
  constructor(options?: RslintOptions);

  lintFiles(patterns: string | string[]): Promise<LintResult[]>;
  lintText(
    code: string,
    options?: { filePath?: string },
  ): Promise<LintResult[]>;

  static outputFixes(results: LintResult[]): Promise<void>;
  close(): Promise<void>;
  [Symbol.asyncDispose](): Promise<void>;
}
```

It is designed for Node.js-compatible runtimes that can load npm packages and provide filesystem, process, and child-process APIs.

## Getting started

```ts
import { Rslint } from '@rslint/core';

const rslint = new Rslint();

try {
  const results = await rslint.lintFiles(['src/**/*.ts']);

  for (const result of results) {
    console.log(result.filePath, result.errorCount, result.warningCount);
  }
} finally {
  await rslint.close();
}
```

Both `lintFiles` and `lintText` return a `LintResult[]` containing each file's messages and error and warning counts.

## Constructor

```ts
const rslint = new Rslint(options);
```

| Option               | Type                                        | Default         | Description                                                                                                                                                                      |
| -------------------- | ------------------------------------------- | --------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cwd`                | `string`                                    | `process.cwd()` | Working directory for targets and discovery; also the authored base of relative fields in inline entries without `basePath`                                                      |
| `overrideConfig`     | `RslintConfigEntry \| RslintConfig \| null` | —               | Extra config appended after the selected config; `basePath` inherits the selected ConfigArray base, while entries without it retain Rslint's existing `cwd`-relative behavior    |
| `overrideConfigFile` | `string \| true \| null`                    | `null`          | A module path disables discovery. Its `basePath` resolves from `cwd`; entries without `basePath` retain Rslint's module-directory behavior. `true` uses only the inline override |
| `fix`                | `boolean`                                   | `false`         | Applies auto-fixes and includes changed source in `result.output`                                                                                                                |
| `virtualFiles`       | `Record<string, string>`                    | —               | In-memory path-to-content overlay for project inputs; unresolved reads may still fall back to disk                                                                               |

With automatic discovery, each file uses its nearest rslint config.

See the [`basePath` configuration reference](/config/base-path) for automatic, explicit-file, and inline-override behavior.

:::warning
Object-form community plugins are not supported in `overrideConfig`, because a plugin worker cannot re-import an in-memory plugin object. Put community plugin declarations in a JS or TS config file. Array-form built-in plugins work in `overrideConfig`.
:::

See [Configuration File](/config/configuration-file) for config discovery and flat-config behavior.

## lintFiles

```ts
lintFiles(patterns: string | string[]): Promise<LintResult[]>
```

Accepts file paths, directory paths, glob patterns, or a mix of them. Relative paths and patterns start from `cwd`; absolute paths and patterns also work. Matching follows the same rules as the CLI.

```ts
const results = await rslint.lintFiles([
  'src/**/*.ts',
  'test/',
  'scripts/check.ts',
]);
```

Each checked file produces one result, even when several inputs match it. Results are sorted by file path. Existing file and directory names are treated literally, even if they contain glob characters.

Your config's `files` and `ignores` settings and `.gitignore` still apply. To exclude `test/fixtures`, for example, add `test/fixtures/**` to `ignores` in your config. A `!pattern` input does not remove files selected by other inputs. Files in different packages can use different nearest configs.

`lintFiles` selects files on disk. If a selected file appears in `virtualFiles`, rslint reads its content from memory. Use `lintText` for code that only exists in memory.

If no file matches, `lintFiles` returns an empty array.

## lintText

```ts
lintText(
  code: string,
  options?: { filePath?: string },
): Promise<LintResult[]>
```

Lints an in-memory string as if it lived at `filePath`.

```ts
const [result] = await rslint.lintText('const answer: number = 42;', {
  filePath: 'src/example.ts',
});
```

The method returns the result for the supplied buffer. When `filePath` is omitted, Rslint uses a synthetic TypeScript path for matching and reports the result path as the `"<text>"` sentinel. A supplied path is resolved against `cwd` and returned as an absolute path.

A leading byte order mark is excluded from column and fix offsets. Fixed output preserves it unless a fix removes it.

## In-memory projects

`lintText` normally discovers config and TypeScript project files from disk. To provide the config, tsconfig, and project files from memory, combine `overrideConfigFile: true`, `overrideConfig`, and `virtualFiles`:

```ts
const rslint = new Rslint({
  cwd: '/',
  overrideConfigFile: true,
  overrideConfig: [
    {
      files: ['**/*.ts'],
      languageOptions: {
        parserOptions: { project: ['./tsconfig.json'] },
      },
      plugins: ['@typescript-eslint'],
      rules: {
        '@typescript-eslint/no-for-in-array': 'error',
      },
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
  'const a = [1];\nfor (const key in a) {}\n',
  { filePath: 'a.ts' },
);
```

Rslint reads `virtualFiles` first. It can still read from disk for `.gitignore`, module resolution, and files absent from the map.

Use relative paths for portability. `virtualFiles` keys always resolve against `cwd`; `parserOptions.project` resolves from the config entry's effective base; paths inside a tsconfig resolve from that tsconfig's directory. In the override-only example above all three directories are `cwd`. Prefer an explicit tsconfig `files` list because a broad `include` glob is expanded against the real filesystem from the tsconfig's directory.

TypeScript project data is only necessary for rules that require type information. A configuration containing only syntax-based rules does not need a tsconfig or `parserOptions.project`.

## outputFixes

```ts
static outputFixes(results: LintResult[]): Promise<void>
```

Writes the `output` of fixed results back to disk.

With `fix: true`, rslint runs up to ten lint-fix rounds. Result messages and counts describe the final in-memory source; findings removed by applied fixes do not remain in `messages`.

```ts
const rslint = new Rslint({ fix: true });
const results = await rslint.lintFiles(['src/**/*.ts']);

await Rslint.outputFixes(results);
```

Only results with a string `output` and an absolute `filePath` are written. The non-absolute `"<text>"` result from `lintText` without a path is skipped automatically.

To apply fixes in memory, read `result.output` and do not call `outputFixes`:

```ts
const source = 'let value = input!!.name;';
const [result] = await rslint.lintText(source, { filePath: 'example.ts' });
const fixed = result.output ?? source;
```

Individual auto-fix edits are available through `result.messages[].fix`. Suggestions are exposed separately in `result.messages[].suggestions` and are never applied by `fix: true`.

## close

```ts
close(): Promise<void>
```

Stops the long-lived Rslint engine process owned by this instance. One-off scripts can exit without calling `close()` because the idle child process does not keep the event loop alive. Long-running editors, build tools, and watch processes should close instances they no longer need.

Calling `close()` more than once waits on the same shutdown operation.

## Symbol.asyncDispose

`Rslint` implements the asynchronous disposable protocol, so a compatible toolchain can close it automatically:

```ts
await using rslint = new Rslint();
await rslint.lintFiles(['src/**/*.ts']);
```

Native `await using` requires runtime support for explicit resource management. TypeScript 5.2 or newer can transform the syntax for older runtimes; otherwise use `try` / `finally` and call `close()` directly.
