# languageOptions

- **Type:** `object`

Configures the JavaScript language environment and TypeScript project information for matching files. Nested language options from matching entries merge recursively; later arrays and scalar values replace earlier values.

## languageOptions.ecmaVersion

- **Type:** `number | 'latest'`
- **Default:** `'latest'`

Selects the standard ECMAScript globals exposed to native rules. Accepted numbers match ESLint and Espree: `3`, `5`, edition aliases `6` through `17`, or years `2015` through `2026`. Edition aliases are normalized to their year (`6` is ES2015 and `17` is ES2026).

The `'latest'` value remains semantic rather than being frozen into the config, so it follows the ESLint version targeted by Rslint. This option currently selects globals; it does not change TypeScript's parser target.

```ts
{
  languageOptions: {
    ecmaVersion: 'latest',
  },
}
```

## languageOptions.sourceType

- **Type:** `'module' | 'script' | 'commonjs'`
- **Default:** exact lowercase `.cjs` extension → `'commonjs'`; every other filename → `'module'`

Selects the module kind used by the per-file language context, including CommonJS globals (`require`, `module`, `exports`, `global`) and whether the top-level scope is the global object. When omitted, a filename with an exact lowercase `.cjs` extension resolves to `'commonjs'`; every other filename, including unknown extensions and extension-less filenames, resolves to `'module'` before rules see the value, matching ESLint. An authored value applies on every extension.

This option does not change TypeScript parsing or compiler module resolution. Support in an individual native rule depends on that rule consulting the configured value; rules that still document syntax-based module detection continue to use that behavior.

Set `sourceType` directly on `languageOptions`; the legacy `languageOptions.parserOptions.sourceType` location is not supported.

```ts
{
  files: ['scripts/**/*.js'],
  languageOptions: {
    sourceType: 'commonjs',
  },
}
```

## languageOptions.parserOptions.projectService

- **Type:** `boolean`

Automatically finds a TypeScript project for each linted file so type-aware rules can run. Enable this option explicitly; TypeScript presets do not enable it.

Rslint checks `tsconfig.json`, then `jsconfig.json`, beside the file and in parent directories until it finds a project that includes the file through its `files` or `include` settings. It also follows project references, including references to custom config names such as `tsconfig.app.json`. The file's nearest matching project takes precedence over a different tsconfig beside the Rslint config or in the current working directory.

JavaScript files use the same discovery. A JS file explicitly listed in `files` can receive type information even with `allowJs: false`; `allowJs` still controls whether `include` patterns select JS files.

```ts
{
  languageOptions: {
    parserOptions: {
      projectService: true,
    },
  },
}
```

`projectService: true` cannot be combined with a `project` string or array, including `[]`. This also applies when the options come from different matching entries. To replace an earlier `project` setting with automatic discovery, set `project: false` alongside `projectService: true`.

Set `projectService: false` to disable automatic discovery and use an explicit `project` instead. To disable type-aware linting for matching files, set both `projectService: false` and `project: false`. JavaScript configurations also accept `projectService: null` as a reset, although the public TypeScript type is `boolean`.

Files outside the discovered projects still receive syntax diagnostics and lint rules and fixes that do not require types. Type-aware rules are skipped. This also applies to files that a project only imports or references without including them through `files` or `include`; see [gap files](/guide/type-checking#gap-files). Invalid project configuration still reports an error.

Unlike typescript-eslint, Rslint skips type-aware rules for files outside configured projects instead of rejecting those files. The typescript-eslint options `allowDefaultProject`, `defaultProject`, `loadTypeScriptPlugins`, and `extraFileExtensions` are not supported. `project: true` is also unsupported; use `projectService: true` for automatic discovery.

In tsconfig, `disableReferencedProjectLoad` prevents discovery through project references, and `disableSolutionSearching` stops further searches in parent directories.

## languageOptions.parserOptions.tsconfigRootDir

- **Type:** `string`
- **Default:** the directory containing the Rslint config file; API `cwd` for inline-only configuration

Sets the directory where `projectService` stops searching parent directories. It must be an absolute path. It also sets the base directory for relative [`project`](#languageoptionsparseroptionsproject) paths.

This option does not select a tsconfig by itself or limit which files a project can import. Project references and `extends` can point outside this directory. Files outside this directory can still find projects in their own parent directories.

The default comes from the Rslint config used for the file, including a custom config selected with `--config` or API `overrideConfigFile`. Imported presets and `basePath` do not change that default.

| Configuration                              | Default discovery boundary       |
| ------------------------------------------ | -------------------------------- |
| Root config used from a package directory  | Root config directory            |
| Nested config used for the file            | Nested config directory          |
| Explicit `--config` / `overrideConfigFile` | Selected config file's directory |
| Inline-only API config                     | API `cwd`                        |
| Loaded config plus inline overrides        | Loaded config file's directory   |

Set an explicit package directory if discovery should stop there instead of continuing to the repository's root tsconfig.

JavaScript configurations also accept `null` to reset an inherited value to the default, although the public TypeScript type is `string`. This also restores the usual base for relative `project` paths, including any `basePath`. A later `undefined` leaves the inherited value unchanged.

## languageOptions.parserOptions.project

- **Type:** `string | string[] | false | null`

Specifies explicit `tsconfig.json` paths. Glob patterns are supported for monorepos. Files included by these tsconfigs receive full type information, enabling type-aware rules such as `@typescript-eslint/no-floating-promises` and `@typescript-eslint/await-thenable`.

Files outside all tsconfigs are still linted, but only rules that do not require type information run.

```ts
{
  languageOptions: {
    parserOptions: {
      projectService: false,
      project: ['./tsconfig.json', './packages/*/tsconfig.json'],
    },
  },
}
```

For linting, only configuration entries matching the file contribute settings. A later `project` value replaces the earlier list. Rslint uses the first tsconfig in that list that includes the file through its `files` or `include` settings. A file that is only imported by a project still skips type-aware rules; add it to the tsconfig's `files` or `include` to enable them.

| Matching settings in order                    | Result for lint rules                                              |
| --------------------------------------------- | ------------------------------------------------------------------ |
| `project: 'a.json'`, then `project: 'b.json'` | Use `b.json` if it includes the file                               |
| `project: 'a.json'`, then `project: []`       | Skip type-aware rules                                              |
| Final `project: false` or `null`              | Clear explicit projects; `projectService: true` can still find one |

Relative project paths and glob patterns resolve from the configuration entry's [`basePath`](/config/base-path), when set, or its usual configuration directory. An explicit `tsconfigRootDir` overrides that base. The `basePath` reference also covers API inline configuration and custom config path rules.

When neither `project` nor `projectService` is enabled, lint still runs rules that do not require types. It does not automatically use a nearby `tsconfig.json`.

The `--type-check` and `--type-check-only` commands check entire configured projects, including imported files. Disabling type-aware lint rules for a file does not exclude it from these TypeScript checks. See [what gets type-checked](/guide/type-checking#what-gets-type-checked).

## languageOptions.globals

- **Type:** `Record<string, boolean | null | 'true' | 'false' | 'readonly' | 'readable' | 'writable' | 'writeable' | 'off'>`

Declares globals available to matching files. Values are normalized before rules or third-party plugins receive the scope:

- Writable: `true`, `'true'`, `'writable'`, `'writeable'`
- Read-only: `false`, `null`, `'false'`, `'readonly'`, `'readable'`
- Disabled: `'off'`

A disabled value removes a declaration inherited from an earlier matching entry, including an ECMAScript built-in. The read-only and writable levels are distinct wherever a rule acts on assignment: `no-global-assign` reports writes to a read-only global and allows them on a writable one.

```ts
{
  languageOptions: {
    globals: {
      BUILD_ID: 'readonly',
      testRuntime: 'writable',
    },
  },
}
```

ECMAScript built-ins are declared according to `languageOptions.ecmaVersion` (`Array` from ES3, `Promise` from ES2015, and so on). Globals added by a runtime — `window` and `document` in browsers, or `process` and `__dirname` in Node.js — are not enabled by default.

`@rslint/core` includes the [`globals`](https://www.npmjs.com/package/globals) catalog and exports its environment maps directly, so no extra dependency is required:

```ts
import { defineConfig, globals } from '@rslint/core';

export default defineConfig([
  {
    files: ['**/*.js'],
    languageOptions: {
      globals: {
        ...globals.browser,
        BUILD_ID: 'readonly',
      },
    },
  },
]);
```

The export has the same set names, global names, and boolean access values as importing the npm package directly: `false` means read-only and `true` means writable. In the published package, each set is synchronously loaded and cached the first time its property is read, so importing `@rslint/core` does not parse the complete catalog.

Compose multiple environments with ordinary object spreads; later spreads and explicit properties take precedence:

```ts
languageOptions: {
  globals: {
    ...globals.browser,
    ...globals.worker,
    location: 'off',
  },
}
```

`globals.node` includes the CommonJS globals (`require`, `module`, `exports`, `__dirname`, and `__filename`); use `globals.nodeBuiltin` for Node.js ESM files that should not receive them. The included catalog also exposes the upstream `builtin`, `es3`, `es5`, and `es20xx` maps for API parity, but `languageOptions.ecmaVersion` is the preferred way to select standard-language globals because it keeps parsing and both rule runtimes on the same edition.

Every map is an explicit globals declaration. It does not change the parser edition, and it can intentionally override the edition-derived set. For example, an upstream host map containing `Temporal` declares that name even when `ecmaVersion` is `2025`.

Loaded maps are shared and cached, so compose and override them with object spreads instead of mutating `globals.browser` or another map in place. Enumerating only `Object.keys(globals)` remains lazy; reading or spreading the complete `globals` object necessarily loads every map.

Flat config continues to merge individual global names in matching-entry order. Scope environment maps with `files`, and use a later explicit `{ process: 'off' }` when one inherited global must be removed.

:::tip
TypeScript's compiler and type-aware rules can resolve declarations from `lib.dom.d.ts`, `@types/node`, and project `.d.ts` files. ESLint-compatible global rules such as `no-undef` and `no-global-assign` intentionally use the flat config's globals instead of TypeScript ambient declarations. The TypeScript presets disable `no-undef`; if you enable such a global rule for TypeScript files, configure their runtime environments too.
:::
