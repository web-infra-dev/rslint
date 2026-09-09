# languageOptions

- **Type:** `object`

Configures the JavaScript language environment and TypeScript project information for matching files. Nested language options from matching entries merge recursively; later arrays and scalar values replace earlier values. Ordinary explicit-project loading retains the separate [owner declaration list](#languageoptionsparseroptionsproject) described below.

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

Discovers a TypeScript config that directly includes each selected file through `files` or `include`. Enable it explicitly; TypeScript presets do not set this option, matching typescript-eslint presets. Discovery starts beside the source file, checking `tsconfig.json`, then `jsconfig.json`, and continues through ancestors when a config does not own the file. Project references can lead to custom config names such as `tsconfig.app.json`.

The nearest owning project wins over a different tsconfig beside the Rslint config or in the current working directory. Reference ownership follows TypeScript's source redirects and reference order. Each selected project keeps its complete root files and dependencies; selecting one lint file limits lint execution, not the type context.

JavaScript files use the same discovery. A JS file explicitly listed in `files` can receive types even with `allowJs: false`; that option still controls JS glob inclusion. Files reached only through imports or triple-slash references are not configured lint roots and use gap linting, whether they are JS or TS.

```ts
{
  languageOptions: {
    parserOptions: {
      projectService: true,
    },
  },
}
```

`projectService: true` cannot be combined with an effective `project` string, array or `[]`, including values inherited from different matching entries. A final matching `project: false` or `project: null` clears that conflict and disables explicit binding for the target; service can still select a project. A final matching `projectService: false` disables automatic discovery and the implicit default project, but preserves the owner's explicit declarations described below. JavaScript configurations and legacy JSON migration also accept `projectService: null` as a runtime reset, outside the public TypeScript type.

A selected file that does not belong to a discovered project uses Rslint's existing [source-only gap fallback](/guide/type-checking#gap-files). Syntax diagnostics and rules that do not require types still run; type-aware rules are skipped. Other files in the same lint request keep their own project context. Config and Program failures are still errors.

This differs from typescript-eslint, which can admit imported-only files and rejects unowned files unless `allowDefaultProject` permits them. Rslint does not create a typed default project. To force source-only linting even when a file has an owning project, set both `projectService: false` and `project: false` for that file scope.

Object options such as `allowDefaultProject`, `defaultProject`, and `loadTypeScriptPlugins`, as well as `extraFileExtensions`, are not implemented. `project: true` is also unsupported; use `projectService: true` for automatic discovery.

When a tsconfig sets `disableReferencedProjectLoad`, Rslint stops discovering projects through those references. This is independent of earlier linted files. Upstream can still use previously loaded referenced projects; Rslint does not reproduce that history-dependent exception. `disableSolutionSearching` stops further ancestor search.

## languageOptions.parserOptions.tsconfigRootDir

- **Type:** `string`
- **Default:** the directory of the governing Rslint config file; API `cwd` for inline-only configuration

An absolute directory for the host operating system that stops upward project discovery when the search reaches it. Trailing separators and dot segments are normalized before comparing the boundary. It does not select a tsconfig by itself. If the target is outside this directory's ancestor chain, it can still discover its own ancestors. Project references and `extends` may point outside the boundary. The final value after matching and merging must be absolute; a relative or empty string in an unmatched or overridden entry does not fail the request.

JavaScript configurations and legacy JSON migration also accept `null` to reset an inherited boundary to the default. This is runtime compatibility, outside the public TypeScript type. JavaScript configurations checked with `checkJs` and `strictNullChecks` are still subject to that type. A later `undefined` preserves an inherited value.

Go resolves the default from the config file selected for each target. An explicitly selected config uses that file's directory, including custom filenames. Imported presets, helper modules, object or rules spread, and `basePath` do not change it. With API `overrideConfigFile: true`, the inline configuration uses API `cwd`; an inline override appended to a loaded config retains that config's default directory.

This follows typescript-eslint's documented config-directory default. Its implementation instead infers candidates from preset access on the JavaScript call stack: missing candidates fall back to process cwd, and multiple candidates can produce an ambiguity error. Rslint uses its known config owner directly. These heuristic edge cases differ; no preset access or process-global candidate state is involved.

Validation applies to the final matched value. Invalid types, relative paths and empty strings in unmatched entries or replaced by a later value do not fail a lint request.

An explicitly set `tsconfigRootDir` also anchors every relative `project` declaration for that target, preserving declaration order. Without it, or after a null reset, each declaration retains its own authored path origin described below. It does not move Rslint's implicit governing-directory `tsconfig.json` fallback when no project paths are declared.

| Configuration                                          | Default discovery boundary                       |
| ------------------------------------------------------ | ------------------------------------------------ |
| Root config used from a package cwd                    | Root config directory                            |
| Nested config owns the target                          | Nested config directory                          |
| Explicit `--config` / `overrideConfigFile`             | Selected config file's directory                 |
| Imported or transformed presets, including JSON copies | Governing config directory                       |
| Inline-only API config                                 | API `cwd`                                        |
| Loaded config plus inline overrides or `basePath`      | Governing config directory                       |
| Explicit absolute `tsconfigRootDir`                    | That directory                                   |
| Later `null` / `undefined`                             | Restore the default / retain the inherited value |

A root config can therefore allow ancestor discovery into a large root TypeScript project even when linting a single package. Set an explicit package boundary if that is the intended scope; selected Programs retain their complete files and dependencies.

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

Rslint collects explicit project strings and arrays from the governing config in declaration order. The list includes entries whose `files`, `ignores` or `basePath` do not match the target; a missing declaration can therefore fail the load. The first project listing the target as a root wins. Only when no project lists it as a root does Rslint try import membership in declaration order. Adding `projectService` or `tsconfigRootDir` does not change this ordinary project order.

This declaration list differs from typescript-eslint's final matching `project` value. Matching and merging still determine the effective service/root options, project/service conflicts and the new `false`/`null` clear values.

| Entries in order                                                                     | Ordinary target binding                                                        |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------ |
| `project: 'a.json'`, then `project: 'b.json'`                                        | Search a, then b, with direct-root priority                                    |
| `project: 'a.json'`, then `project: []`                                              | Keep a; [] does not erase earlier declarations                                 |
| Final matching `project: false` or `null`                                            | Use no explicit/default project for that target; enabled service may still run |
| `project: 'a.json'`, then false, then `project: 'b.json'`                            | Final clear is canceled; search the original a, b list again                   |
| Matching `projectService: false`, with an explicit declaration in an unmatched entry | Disable automatic/default discovery but keep that explicit declaration         |

When an entry has `basePath`, its explicit project literals and globs resolve from that directory unless the target has an explicit `tsconfigRootDir`. The directory remains literal even if its name contains glob characters. A later null root reset restores each declaration's original base. Targets with different roots keep separate eligible project lists even if their Programs contain overlapping files.

When both project settings are omitted, Rslint retains the governing config directory's default `tsconfig.json` fallback; neither `basePath` nor `tsconfigRootDir` moves this implicit lookup. A declaration of `project: []` suppresses fallback when no paths were declared. A final matching false/null or `projectService: false` disables default binding for that target. Unmatched false/null does not disable another target's fallback. See [`basePath`](/config/base-path) for path origins.

Plain whole-directory CLI lint builds explicit projects for active ordinary owners; focused CLI/API lint selects needed projects from target membership. `--type-check` and `--type-check-only` retain [program-wide explicit checking](/guide/type-checking#what-gets-type-checked), including declarations outside the lint target scope. These modes also check complete service-selected Programs. A per-target clear does not erase the owner's program-wide declarations.

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
