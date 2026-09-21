# no-unsupported-features/node-builtins

Disallow static Node.js APIs that are unavailable in the configured Node version range.

## Rule details

The rule checks Node globals, CommonJS imports, ESM imports and re-exports, `process.getBuiltinModule()`, and `import.meta` properties. It follows aliases and object destructuring. An API must be available throughout the configured range, including gaps between releases that received backports.

With `version: '18.0.0'`, these examples are **incorrect**:

```js
fetch('https://example.com');
import.meta.dirname;
```

These examples are **correct** for that version:

```js
import { readFile } from 'node:fs/promises';
const contents = await readFile('data.txt', 'utf8');
```

This rule does not provide automatic fixes or suggestions.

## Options

```js
export default [{
  plugins: ['node'],
  rules: {
    'node/no-unsupported-features/node-builtins': ['error', {
      version: '>=16.0.0',
      allowExperimental: false,
      ignores: [],
    }],
  },
}];
```

- `version`: a Node version range. Resolution checks this option, `settings.n.version`, `settings.node.version`, the nearest `package.json`'s `engines.node`, then its `devEngines.runtime` entry named `node`. Invalid ranges fall through. The fallback is `>=16.0.0`.
- `allowExperimental`: defaults to `false`. When enabled, an API with an experimental release is checked against those release versions. Otherwise, it must have a stable release available throughout the configured range.
- `ignores`: exact API names from diagnostics, such as `fetch`, `fs.rm`, or `import.meta.dirname`. Names omit the `node:` prefix for both module spellings. Ignoring a module name does not ignore its members.

For example, `allowExperimental: true` permits `fetch()` on Node 18, while the default requires its stable release in Node 21. `node:` module names have their own introduction versions.

## Limitations

Only static APIs are checked. The rule does not infer instance methods, function parameters, option properties, or events. Dynamic imports and unknown property names are not tracked, matching upstream behavior.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-unsupported-features/node-builtins.md)
- [Upstream implementation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unsupported-features/node-builtins.js)
