# rslim/require-artifact-dependency-entry

Warn when an artifact-built module has a runtime dependency on a source-built TypeScript module that Rslim may otherwise prune.

The rule covers two cases from the Rslim integration notes:

- A TypeScript source file belongs to a project with `noCheck`. Its static runtime imports and re-exports are checked for dependencies on source-built modules in other referenced projects.
- A `.js` file has a same-name `.d.ts` declaration file. Its runtime dependencies on source-built TypeScript modules are checked.

Imports within the same `noCheck` project are ignored because those modules are built as artifacts together. Imports into another `noCheck` project are also ignored. The rule needs a resolved source target and, for project references, the target project's compiler options; unresolved imports and declaration files without a known source are ignored. A missing TypeChecker by itself is not treated as proof of an artifact build.

Each reported dependency crosses from artifact-built code into source-built code. For named imports and re-exports, the diagnostic lists candidate export names from the syntax; namespace imports need a broader review. Check the emitted JavaScript and the imported module's runtime uses, including dependencies reached through other artifact modules. Add `@entry` to declarations that Rslim must retain. In Lib Mode, adding the target source file to Rslim `entries` can preserve its exports. If the module has required top-level side effects, preserve those through an entry file too; `@entry` on an individual declaration does not preserve the entire module.

The rule reports direct dependency edges for each artifact module that Rslint checks. A chain of artifact modules is covered only when each module's source or runtime `.js` file is included in linting; inspect any unchecked emitted artifacts separately.

Whole `import type` and `export type` declarations are ignored. Inline `type` specifiers still produce diagnostics because `verbatimModuleSyntax` can preserve their module side effects. Dynamic `import(...)`, `require(...)`, and TypeScript `import x = require(...)` are handled by `rslim/require-dynamic-import-entry`. This rule does not infer which exports the emitted artifact actually uses, traverse its emitted dependency graph, read Rspack or Rslim configuration, or automatically add annotations. A diagnostic can be suppressed with `rslint-disable-next-line rslim/require-artifact-dependency-entry` after the target has been checked, with a reason recorded in the comment.

To cover the declaration-pair case, include the JavaScript runtime files in Rslint's file selection. Third-party artifacts outside that selection need a separate review. A `noCheck` option is a review signal, not proof that the deployed build uses emitted JavaScript; verify the actual build route before suppressing a warning.
