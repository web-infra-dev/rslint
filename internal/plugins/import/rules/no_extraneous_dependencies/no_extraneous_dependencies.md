# no-extraneous-dependencies

Require imported packages to be declared in the nearest `package.json`. Packages must be resolvable for this rule to check them. It checks imports, re-exports, dynamic imports and CommonJS `require` calls. Builtins and relative imports are ignored.

```js
// Incorrect when lodash is installed but absent from package.json:
import lodash from 'lodash';
```

Declare runtime packages in `dependencies`. By default, packages in `devDependencies`, `optionalDependencies`, `peerDependencies` and `bundledDependencies` are also allowed. `bundleDependencies` is accepted as an alternative spelling. Scoped packages and imports from files within a package are supported.

## Options

Each dependency category accepts a boolean or an array of filename globs. All four default to `true`. An array allows that category only in matching files. Patterns are tested against the filename and against the pattern resolved from the process working directory.

```json
{
  "devDependencies": ["**/*.test.js", "**/*.spec.js"],
  "optionalDependencies": false,
  "peerDependencies": true,
  "bundledDependencies": true
}
```

- `packageDir`: a directory or array of directories containing package manifests. Relative directories resolve from the process working directory. Multiple manifests combine their dependency declarations; missing files are ignored when more than one directory is supplied.
- `includeInternal`: check bare internal aliases, including those matched by `import/internal-regex`. Defaults to `false`.
- `includeTypes`: check type-only imports and re-exports. Defaults to `false`. Mixed value/type imports are always checked.

An empty or missing manifest still permits diagnostics for resolved, undeclared packages, matching the upstream implementation. This rule has no autofix or suggestions.

## Differences from upstream

- Multiple `packageDir` manifests combine bundled dependencies by package name. If one declares `["first"]` and another declares `["second"]`, both packages are allowed, regardless of directory order. Upstream 2.32.0 replaces array entries by index and can incorrectly report the earlier package as missing. This also applies to object declarations and the `bundleDependencies` spelling.

- Invalid `package.json` errors retain the `The package.json file could not be parsed:` prefix but use Go's JSON parser explanation. For example, `{{` produces a different syntax explanation. Manifest errors point to the start of the source file; upstream uses a synthetic line zero location.
- JavaScript resolver plugins such as `import/resolver: "webpack"` cannot run in native rules. Use the supported Node or TypeScript resolver and TypeScript path mappings for aliases. Flow `import typeof` syntax is unsupported; TypeScript type-only imports are supported.

## References

- [Upstream documentation](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-extraneous-dependencies.md)
- [Pinned implementation](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-extraneous-dependencies.js)
