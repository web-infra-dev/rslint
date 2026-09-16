# no-deprecated-api

Disallow deprecated static Node.js APIs.

## Rule details

The rule checks configured Node globals, CommonJS imports, ESM imports and re-exports, and `process.getBuiltinModule()`. It follows aliases and object destructuring, including `node:` module names.

Examples of **incorrect** code:

```js
const fs = require('node:fs');
fs.exists('file.txt', () => {});

const { SlowBuffer } = require('buffer');
const data = new Buffer(10);
```

Examples of **correct** code:

```js
const fs = require('node:fs');
fs.access('file.txt', () => {});

const data = Buffer.alloc(10);
```

The rule reports deprecated APIs even when the configured Node version predates their deprecation. The version range controls which replacement APIs appear in the message. This rule does not provide automatic fixes or suggestions.

## Options

```js
export default [{
  plugins: ['node'],
  rules: {
    'node/no-deprecated-api': ['error', {
      version: '>=16.0.0',
      ignoreModuleItems: [],
      ignoreGlobalItems: [],
    }],
  },
}];
```

- `version`: a Node version range. Resolution checks this option, `settings.n.version`, `settings.node.version`, the nearest `package.json`'s `engines.node`, then its `devEngines.runtime` entry named `node`. Invalid ranges fall through. The fallback is `>=16.0.0`.
- `ignoreModuleItems`: exact module API names to ignore, such as `fs.exists`, `buffer.Buffer()` or `new buffer.Buffer()`. Use names without `node:` for both import spellings.
- `ignoreGlobalItems`: exact global API names to ignore, such as `Buffer()`, `new Buffer()` or `process.binding`. Global and module ignore lists are independent.
- `ignoreIndirectDependencies`: accepted for compatibility with the upstream deprecated option; it has no effect.

For example, `ignoreModuleItems: ['new buffer.Buffer()']` allows `new (require('buffer').Buffer)(10)` while still reporting calls without `new`.

## Limitations

Only static APIs are checked. The rule does not infer instance types, dynamic property names, values passed as arguments, or aliases stored on object properties. Reassigning a local alias does not cancel its earlier tracked origin. A user-installed package does not hide a Node builtin: use `require('punycode/')` to select that package rather than `require('punycode')`.

## Differences from upstream

Valid ranges containing a major, minor, or patch number of `4294967295` or
greater conservatively omit version-dependent replacement recommendations.
They do not fall through to another configured range. For example, with
`version: ">4294967295"`, `Buffer()` still produces a deprecation diagnostic,
but without recommending `Buffer.alloc()` or `Buffer.from()`; upstream
includes those replacements.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-deprecated-api.md)
- [Upstream implementation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-deprecated-api.js)
