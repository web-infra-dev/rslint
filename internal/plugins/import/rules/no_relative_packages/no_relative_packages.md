# no-relative-packages

Forbid importing another package through a relative path. Use its package name instead, especially in workspaces.

## Rule Details

The rule resolves the import and compares the nearest named `package.json` for the importing file and target. Unnamed packages inherit a named ancestor. Unresolved imports, missing package names, and equal package names are allowed.

```js
// packages/foo/index.js
// Incorrect, when ../bar belongs to the package named "bar":
import bar from '../bar';

// Correct:
import bar from 'bar';
```

## Options

An optional object accepts:

- `esmodule` (default `true`): inspect imports, re-exports, and dynamic imports.
- `commonjs` (default `false`): inspect direct `require()` calls.
- `amd` (default `false`): inspect dependency arrays in `require()` and `define()`.
- `ignore`: an array of regular expressions matching module specifiers to skip.

## Autofix

This rule does not provide automatic fixes. Running `--fix` leaves these imports unchanged.

## Suggestions

A manual suggestion can replace the relative path with a package import when both requests resolve to the same JavaScript file under default Node resolution. The check respects package entry points, public `exports` entries, the import or require mode, and default symlink resolution. The package name must resolve to the original package rather than a different installed copy.

The candidate preserves the written request within the package. For example, `../b/sub` can suggest `b/sub`, and `../b` can suggest `b`, even when the entry is `dist/entry.js`. These suggestions appear only after the same-file check succeeds.

Suggestions are limited to `.js`, `.cjs`, and `.mjs` inputs and targets. They are omitted for TypeScript, AMD, shadowed `require`, configured resolvers, import attributes, URL-like paths, conditional exports, or fallback export arrays. Public export aliases are not inferred: if `b/api` exposes `hidden.js` but `b/hidden.js` is private, choose the public entry manually.

Suggestions assume default Node runtime behavior. Projects using custom conditions, symlink flags, or a bundler should review the suggested change against their own resolution settings.

## Differences from upstream

Upstream v2.32.0 offers unconditional autofixes based on the resolved file directory and the written basename. This can duplicate directory names, lose a package's `main` entry semantics, mistake a file named like its package directory for the package root, turn `..` into `.`, or select a private or different exports target. Rslint intentionally offers no autofix and only verified manual suggestions. Its diagnostic names the original import without promising an unverified replacement path.

Native resolution supports the Node and TypeScript resolvers. Custom JavaScript resolver plugins cannot execute in the native runtime and produce a resolve error. Configure `import/resolver` as `node` or `typescript`.

A non-object `package.json`, such as `[]`, stops package lookup. Upstream can continue to a named ancestor for this malformed manifest and report a cross-package import. Use an object-shaped package manifest.

## References

- [Upstream documentation](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-relative-packages.md)
- [Pinned source](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-relative-packages.js)
