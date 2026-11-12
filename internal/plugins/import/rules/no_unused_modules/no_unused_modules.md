# no-unused-modules

Report modules without any exports, and exports that no other module in the project imports.

## Rule Details

The rule does nothing until at least one of `missingExports` or `unusedExports` is enabled.

With `missingExports`, a module is reported when it exports nothing. The diagnostic covers the module's first statement.

```js
// Incorrect
class MyClass {
  /*...*/
}

function makeClass() {
  return new MyClass(...arguments);
}
```

```js
// Correct
export default function () {
  /*...*/
}
```

```js
// Correct
export const foo = function () {
  /*...*/
};
```

```js
// Correct
export { foo, bar };
```

```js
// Correct
export { foo as bar };
```

With `unusedExports`, every file matched by `src` is read and the rule records which exports other files use. An export is reported when no analyzed file uses it. The following count as uses:

- `import { a as alias } from './a'` uses `a`. A default import, and `import { default as x }`, use `default`.
- `import * as ns from './a'`, `import ns = require('./a')`, and `import('./a')` with a string literal use every export, including `default`. The namespace does not need to be referenced.
- `export { a } from './a'` uses `a` of `./a`. The re-exporting file's own `a` still needs a consumer of its own.
- `export * from './a'` and `export * as ns from './a'` use every named export of `./a`, but not `default`.
- `import type` and `export type` count like value imports.
- A static `require('./a')` uses what the surrounding code reads: `const { a } = require('./a')` and `require('./a').a` use `a`. A plain `const lib = require('./a')` or any other use of the result counts as every export. A bare `require('./a');` statement uses nothing. Calls are ignored in files that declare their own `require`.

Side-effect imports (`import './a'`), dynamic imports of non-literal paths, and template literals use nothing. The analysis is static: it does not check whether an imported binding is referenced, or whether the importing file is reachable from an entry point.

Given these files:

```js
// file-f.js
import { e } from './file-a';
import { f } from './file-b';
import * as fileC from './file-c';
export { default, i0 } from './file-d'; // both will be reported

export const j = 99; // will be reported
```

```js
// file-d.js
export const i0 = 9; // will not be reported
export const i1 = 9; // will be reported
export default () => {}; // will not be reported
```

```js
// file-c.js
export const h = 8; // will not be reported
export default () => {}; // will not be reported: file-f imports the namespace
```

```js
// file-b.js
import two, { b, c, doAnything } from './file-a';

export const f = 6; // will not be reported
```

```js
// file-a.js
const b = 2;
const c = 3;
const d = 4;

export const a = 1; // will be reported

export { b, c }; // will not be reported

export { d as e }; // will not be reported

export function doAnything() {
  // some code
} // will not be reported

export default 5; // will not be reported
```

Exports are not reported in a file that the nearest `package.json` publishes as an entry point: a file named by `main`, `bin` (a string or any value of an object), `browser` (likewise), or any target of `exports`, under any subpath or condition. A `*` in an `exports` target matches any path, as a subpath pattern does. A `package.json` with `"private": true` publishes nothing.

Module specifiers are resolved with `import/resolver`, which defaults to Node resolution followed by TypeScript resolution. An analyzed file that belongs to a TypeScript project loaded for this lint run is resolved with that project's compiler options, so `paths` aliases defined in different tsconfig files are all followed. Other files are resolved without `paths`. Rslint loads only the projects that contain a linted file: when you lint part of a workspace, an importer in another project that uses a `paths` alias is not counted. Lint the projects together, or use relative imports between them. Imports written in files matching `import/ignore`, and requests that resolve into `node_modules`, do not count.

`unusedExports` is reported only by the CLI and the JavaScript API, which lint one fixed view of the project. Editor integrations do not report unused exports, because a single open project cannot see every importer. `missingExports` works everywhere.

## Options

An object with at least one of `missingExports` or `unusedExports` set to `true`:

- `missingExports` (default `false`): report files without any exports.
- `unusedExports` (default `false`): report exports that no analyzed file uses.
- `ignoreUnusedTypeExports` (default `false`): with `unusedExports`, do not report interfaces, type aliases, or enums. This covers `export interface`/`export type`/`export enum` declarations, `export type { T }`, `export { type T }`, and `export { T }` when every local declaration of `T` is an interface, type alias, or enum.
- `src` (default: the working directory): the files analyzed for `unusedExports`. Each entry is a file, a directory, or a glob, resolved against the working directory. A directory includes the files below it with an analyzed extension. A glob selects files by the pattern alone. Dotfiles, dot directories, and `node_modules` are skipped. With `unusedExports`, the array must not be empty.
- `ignoreExports` (default `[]`): files, directories, or globs whose exports are not reported, such as the entry points of a published package. Their imports still count as uses.

The analyzed extensions are `import/extensions` plus any extensions listed in `import/parsers`. Without `import/extensions`, every JavaScript and TypeScript extension (`.js`, `.mjs`, `.cjs`, `.jsx`, `.ts`, `.tsx`, `.mts`, `.cts`) is analyzed.

```js
export default [
  {
    plugins: ['import'],
    rules: {
      'import/no-unused-modules': [
        'warn',
        {
          unusedExports: true,
          src: ['src'],
          ignoreExports: ['src/index.ts'],
        },
      ],
    },
  },
];
```

With `ignoreUnusedTypeExports` set to `true`, the following will not be reported:

```ts
export type Foo = {};
export interface Bar {}
export enum Baz {}
```

## When Not To Use It

If unused files or dead exports in your codebase do not matter to you, you can disable this rule.

## Differences from upstream

- **Analyzed files.** Upstream adds the imports of every linted file to its usage data, including files outside `src`, and keeps that data for the rest of the process. Its result can depend on which files were linted before, and in which order. Rslint reads every file in `src` from the same request, including API `fileContents`, before reporting. Imports in files outside `src` never count. Autofix rounds of other rules are re-analyzed. To count a consumer, include it in `src`.
- **Default extensions.** Without `import/extensions`, upstream analyzes only `.js`, `.mjs`, and `.cjs`, so TypeScript files are skipped unless that setting lists them. Rslint analyzes every JavaScript and TypeScript extension. Set `import/extensions` to restrict it.
- **Ignored files.** Upstream applies `.eslintignore`/`ignorePatterns` when enumerating `src`. Rslint does not apply configuration `ignores` there. Narrow `src`, or list generated files in `ignoreExports`.
- **`ignoreExports` and `missingExports`.** Upstream skips `missingExports` for files matching `ignoreExports`, but only when `unusedExports` is also enabled. Rslint applies `ignoreExports` only to `unusedExports`.
- **Static `require()`.** The upstream documentation says `require` usage counts, but the implementation counts none, so a CommonJS consumer leaves the export reported. Rslint counts static `require()` calls and `import x = require()` as described above.
- **`import { default as x }`.** Upstream does not count this as a use of `default`. Rslint does.
- **Re-exporting modules.** Upstream reports `No exports found` for a module containing only `export * from` or `export * as ns from`, `export namespace`, `export declare function`, `export =`, or `export import`. Rslint counts each of these as an export.
- **Destructured exports.** Upstream crashes on `export const [, a] = list` and reports the name `undefined` for array elements with defaults or nested patterns. Rslint reports each bound name.
- **TypeScript declarations.** Upstream reports names exported inside a namespace (`export namespace N { export const a = 1 }` reports `a`) and never reports the namespace itself. Rslint reports `N` and ignores the namespace body. Upstream does not check `export declare function`. Rslint reports it once, and reports an overloaded function once, on its implementation.
- **Type exports in export lists.** With `ignoreUnusedTypeExports`, upstream exempts only direct `export interface`/`type`/`enum` declarations. Rslint also exempts the type-only export forms listed under the option.
- **Package entry points.** Upstream exempts `main`, `browser`, and `bin` entries only. Rslint also exempts every target of `exports`. The `module` field is not treated as an entry point.
- **Resolvers.** Native resolution supports the Node and TypeScript resolvers. Custom JavaScript resolver plugins cannot run in the native runtime, so requests they would resolve stay unresolved. Configure `import/resolver` as `node` or `typescript`.
- **Editors.** `unusedExports` reports nothing in editor integrations, as explained above. Upstream reports from whatever its long-lived index contains at that moment.
- **Flow.** Flow annotations are not supported.

## References

- [Upstream documentation](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-unused-modules.md)
- [Pinned source](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-unused-modules.js)
