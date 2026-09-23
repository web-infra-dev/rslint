# basePath

- **Type:** `string`

`basePath` sets the directory from which one flat-config entry is matched. It is a literal directory path, not a glob. The entry's lint settings apply only to matching files within that directory.

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

An entry containing only `basePath`, `ignores`, and an optional `name` is still a [global ignore entry](/config/ignoring-files#global-and-entry-level-ignores). Its ignore patterns start from the `basePath` directory.

## Resolution base

A relative `basePath` resolves from a directory determined by how the config was selected:

| Config entry source                                                | `basePath` resolves from    | Relative paths when `basePath` is absent           |
| ------------------------------------------------------------------ | --------------------------- | -------------------------------------------------- |
| Automatically discovered config module                             | Config module directory     | Config module directory                            |
| Explicit `--config`, API `overrideConfigFile`, or fixed LSP config | Invocation/workspace cwd    | Config module directory (existing Rslint behavior) |
| API inline `overrideConfig` in automatic-discovery mode            | Discovered config directory | API `cwd`                                          |
| API inline `overrideConfig` with an explicit config or no config   | API `cwd`                   | API `cwd`                                          |

An inline override therefore uses the discovered config directory for `basePath` in automatic mode, and API `cwd` with an explicit config or `overrideConfigFile: true`. Without `basePath`, relative paths in an inline override resolve from API `cwd`.

For example, if `/project` invokes an external config:

```bash
cd /project
rslint --config /configs/rslint.config.ts
```

Then `basePath: 'app'` means `/project/app`, not `/configs/app`.

Absolute paths are used as written. An empty string is also valid and uses the directory in the table's `basePath` column. Glob characters are literal in `basePath`, so `basePath: 'packages/*'` names a directory containing `*` rather than selecting every package.

## TypeScript projects

Relative `languageOptions.parserOptions.project` paths and glob patterns resolve from `basePath`. An explicit `tsconfigRootDir` overrides that base; resetting it to `null` in JavaScript configuration restores the entry's `basePath`.

For linting, a later matching `project` value replaces an earlier list and uses the `basePath` of the entry that supplies it. `project: []`, `false`, or `null` clears explicit projects for matching files. Automatic discovery can still run when `projectService: true` is combined with `project: false` or `null`; combining it with `project: []` is an error. See [parser options](/config/language-options#languageoptionsparseroptionsproject).

`basePath` does not change where `projectService` stops searching parent directories; use `tsconfigRootDir` for that. It also does not move the default `tsconfig.json` lookup beside the Rslint config used by the type-check commands. Ordinary lint does not use that default.

`--type-check` and `--type-check-only` check all listed projects, including those in entries whose `files` patterns select no lint files. A missing project path can therefore cause these commands to fail even if no lint files match that entry. See [what gets type-checked](/guide/type-checking#what-gets-type-checked).

## Directory and ignore boundaries

Global ignore patterns scoped with `basePath` can ignore files and subdirectories, but not the `basePath` directory itself. If `basePath` points above the config's resolution base shown in the table, the patterns cannot ignore that resolution base itself either.

`basePath` does not move:

- Which Rslint config is used or which directory the command scans
- The root used to collect `.gitignore` files
- The pattern syntax for `files` or `ignores`

See [`.gitignore` integration](/config/ignoring-files#gitignore-integration) for its independent collection rules.

The directory named by `basePath` does not need to exist when the config loads. Non-string values are rejected immediately; a missing explicit TypeScript project is validated separately.
