# basePath

- **Type:** `string`

`basePath` sets the directory from which one flat-config entry is matched. It is a literal directory path, not a glob. Its per-file configuration is inactive outside that directory; explicit TypeScript projects retain the [owner-wide behavior](#typescript-projects) described below.

```ts
{
  basePath: 'packages/app',
  files: ['src/**/*.ts'],
  ignores: ['src/generated/**'],
  languageOptions: {
    parserOptions: { project: ['./tsconfig.json'] },
  },
  rules: {
    'no-debugger': 'error',
  },
}
```

In this example, `files`, `ignores`, and the explicit `project` path all start from `packages/app`. `basePath` changes only their starting directory; `files` and `ignores` keep the same glob syntax and matching behavior as entries without `basePath`.

An entry containing only `basePath`, `ignores`, and an optional `name` is still a [global ignore entry](/config/ignoring-files#global-and-entry-level-ignores). Its patterns are scoped to the effective base directory and participate in normal lint-target and config-candidate traversal.

## Resolution base

A relative `basePath` resolves from the ConfigArray base. That base depends on how the config was selected:

| Config entry source                                                | `basePath` resolves from    | Relative paths when `basePath` is absent           |
| ------------------------------------------------------------------ | --------------------------- | -------------------------------------------------- |
| Automatically discovered config module                             | Config module directory     | Config module directory                            |
| Explicit `--config`, API `overrideConfigFile`, or fixed LSP config | Invocation/workspace cwd    | Config module directory (existing Rslint behavior) |
| API inline `overrideConfig` in automatic-discovery mode            | Discovered config directory | API `cwd`                                          |
| API inline `overrideConfig` with an explicit config or no config   | API `cwd`                   | API `cwd`                                          |

An inline override therefore uses the discovered module directory for `basePath` in automatic mode, and API `cwd` with an explicit config or `overrideConfigFile: true`. This matches how ESLint appends `overrideConfig` to the selected ConfigArray. Relative fields in an inline entry without `basePath` keep Rslint's existing API-`cwd` behavior.

For example, if `/project` invokes an external config:

```bash
cd /project
rslint --config /configs/rslint.config.ts
```

Then `basePath: 'app'` means `/project/app`, not `/configs/app`.

Absolute paths are used as written. An empty string is also valid and still counts as an authored `basePath`; this matters when an inline override's ordinary path origin differs from its ConfigArray base. Glob characters are literal in `basePath`, so `basePath: 'packages/*'` names a directory containing `*` rather than selecting every package.

## TypeScript projects

`basePath` moves explicit `languageOptions.parserOptions.project` literals and globs. It does not move the governing config's implicit `tsconfig.json` fallback or change Rslint's owner-wide project collection. An explicit `project: []` still disables the fallback.

Because explicit projects are collected for the governing config, a missing project can still report an error even when the entry's `files` patterns select no lint target. See [`languageOptions.parserOptions.project`](/config/language-options#languageoptionsparseroptionsproject) for the complete project behavior.

## Directory and ignore boundaries

For directory-ignore decisions, a scoped global ignore does not match the effective base directory itself. When `basePath` points above the selected ConfigArray base, a scoped global ignore cannot prune that base itself. Descendants still match normally relative to `basePath`.

`basePath` does not move:

- The config owner or requested scan root
- The root used to collect `.gitignore` files
- The existing `files` or `ignores` matcher and glob grammar

See [`.gitignore` integration](/config/ignoring-files#gitignore-integration) for its independent collection rules.

The directory named by `basePath` does not need to exist when the config loads. Non-string values are rejected immediately; a missing explicit TypeScript project is validated separately.
