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

Replace the source literal with a double-quoted package import, preserving the written basename and target's subdirectory as upstream does. Scoped package names are supported.

## Differences from upstream

Native resolution supports the Node and TypeScript resolvers. Custom JavaScript resolver plugins cannot execute in the native runtime and produce a resolve error. Configure `import/resolver` as `node` or `typescript`.

A non-object `package.json`, such as `[]`, stops package lookup. Upstream can continue to a named ancestor for this malformed manifest and report a cross-package import. Use an object-shaped package manifest.

## References

- [Upstream documentation](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-relative-packages.md)
- [Pinned source](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-relative-packages.js)
