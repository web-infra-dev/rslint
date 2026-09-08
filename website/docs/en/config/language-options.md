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

Discovers the configured TypeScript project for each selected file. This is the default in `ts.configs.recommended`. Discovery starts beside the source file, checking `tsconfig.json`, then `jsconfig.json`, and continues through ancestors when a config does not own the file. Project references can lead to custom config names such as `tsconfig.app.json`.

The nearest owning project wins over a different tsconfig beside the Rslint config or in the current working directory. Reference ownership follows TypeScript's source redirects and reference order. Each selected project keeps its complete root files and dependencies; selecting one lint file limits lint execution, not the type context.

```ts
{
  languageOptions: {
    parserOptions: {
      projectService: true,
    },
  },
}
```

`projectService: true` cannot be combined with explicit `project` paths or `project: []`, including values inherited from different matching entries. Remove `project`, or set `projectService: false` to use explicit projects. A later `project: false` or `project: null` clears inherited paths. A later `projectService: false` or `null` disables automatic discovery.

A selected file that does not belong to a discovered project is an error, even if its enabled rules do not require types. It does not enter the [source-only gap fallback](/guide/type-checking#gap-files). In the JavaScript API, this rejects the lint request; results for other files in that request are not returned. To lint such files without type information, override both settings for that file scope:

```ts
{
  files: ['scripts/**/*.js'],
  languageOptions: {
    parserOptions: { projectService: false, project: false },
  },
}
```

Only the boolean form is supported. Object options such as `allowDefaultProject`, `defaultProject`, and `loadTypeScriptPlugins`, as well as `extraFileExtensions`, are not implemented. `project: true` is also unsupported; use `projectService: true` for automatic discovery.

The editor reuses its existing Program store rather than hosting a second TypeScript project service. Normal source/config updates recheck ownership. One long-lived editor difference remains: after all seed documents close, deleting a config, querying a referenced target, and recreating the config may retain its previously loaded identity in Rslint. TypeScript can unload that project in this sequence, making `disableReferencedProjectLoad` reject it until another document loads it. No stale Program is used while the config is missing.

## languageOptions.parserOptions.tsconfigRootDir

- **Type:** `string | null`
- **Default:** the directory of the governing Rslint config

An absolute directory that stops upward project discovery when the search reaches it. It does not select a tsconfig by itself. If the target is outside this directory's ancestor chain, it can still discover its own ancestors. Project references and `extends` may point outside the boundary. Relative paths are rejected; `null` resets an inherited boundary to the default.

API override entries use their authored working-directory base for the default. A config entry's `basePath` scopes matching without moving this boundary. Unlike typescript-eslint's JavaScript call-stack inference, Rslint uses its resolved config origin, so external config modules have a deterministic base.

When set with explicit `project`, relative project paths resolve from `tsconfigRootDir`.

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

When an entry has `basePath`, its explicit project literals and globs resolve from that directory unless `tsconfigRootDir` is set. With `projectService` or `tsconfigRootDir` settings, the matching entries determine each target's effective project policy. Configs using only legacy explicit project paths retain their existing config-wide declaration collection and first-root-before-import selection order.

When both project settings are omitted, Rslint retains the governing config directory's default `tsconfig.json` fallback; `basePath` alone does not move it. An explicit `project: []`, `false`, or `null` disables that fallback. See the [`basePath` configuration reference](/config/base-path) for the path-origin rules.

With automatic discovery, `--type-check` and `--type-check-only` discover projects from the selected file/directory scope, then check the complete selected projects. Legacy explicit project configs, including `projectService: false` with explicit paths, retain program-wide type checking. Configs that mix automatic and explicit discovery, clear inherited projects, or set `tsconfigRootDir` resolve their policies from selected targets.

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
