# Rslint

The `Rslint` class is the ESLint-style programmatic API for linting files and in-memory source from a JavaScript host.

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

Both `lintFiles` and `lintText` return ESLint-shaped `LintResult[]` values.

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

With automatic discovery, Go selects each file's nearest config and owns ignore and target-admission semantics. The JavaScript host evaluates and normalizes the JS or TS config modules selected for that run.

See the [`basePath` configuration reference](/config/base-path) for automatic, explicit-file, and inline-override behavior.

:::warning
Object-form community plugins are not supported in `overrideConfig`, because a plugin worker cannot re-import an in-memory plugin object. Put community plugin declarations in a JS or TS config file. Array-form built-in plugins work in `overrideConfig`.
:::

See [Configuration File](/config/configuration-file) for config discovery and flat-config behavior.

## lintFiles

```ts
lintFiles(patterns: string | string[]): Promise<LintResult[]>
```

Lints files matched by one or more glob patterns resolved against `cwd`.

```ts
const results = await rslint.lintFiles([
  'src/**/*.ts',
  'test/**/*.ts',
  '!test/fixtures/**',
]);
```

Supported files excluded by global config ignores or `.gitignore` are omitted. With automatic discovery, files in different monorepo packages can use different nearest configs. Results are ordered by file path rather than glob-walk order.

If no file matches, `lintFiles` returns an empty array. This differs from ESLint v10's default `errorOnUnmatchedPattern` behavior, which throws for an unmatched glob.

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

Line, column, and fix offsets match ESLint's byte-order-mark behavior: a leading byte order mark is not included in indexed source text, while fixed whole-file output preserves it unless a fix removes it.

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

`virtualFiles` is an overlay, not a filesystem sandbox. Rslint can still read from disk for `.gitignore`, module resolution, and files absent from the map.

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
