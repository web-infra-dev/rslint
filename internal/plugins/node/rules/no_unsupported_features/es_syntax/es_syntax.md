# no-unsupported-features/es-syntax

Disallows ECMAScript features that are unavailable in part of the configured
Node.js version range.

## Rule details

The rule checks syntax such as optional chaining, dynamic imports, class fields,
static blocks, regular expressions, and top-level `await`. It also checks the
built-in APIs included in the upstream feature table, such as `Object.hasOwn`
and `Array.prototype.toSorted`. Local bindings that shadow built-in globals
are excluded.

With `version: ">=16.0.0"`, these examples are **incorrect**:

```javascript
class Cache {
  static {
    this.entries = new Map();
  }
}

Object.hasOwn(object, key);
[3, 1, 2].toSorted();
```

These features require Node.js 16.11, 16.9, and 20 respectively. Set a version
range that supports the features you use, or use older syntax and APIs:

```javascript
class Cache {}
Cache.entries = new Map();

Object.prototype.hasOwnProperty.call(object, key);
[3, 1, 2].slice().sort();
```

The rule does not provide automatic fixes or suggestions.

## Options

```javascript
export default [
  {
    plugins: ["node"],
    rules: {
      "node/no-unsupported-features/es-syntax": [
        "error",
        {
          version: ">=16.0.0",
          ignores: [],
        },
      ],
    },
  },
];
```

### version

The first valid Node.js version range is used, in this order:

1. The rule's `version` option.
2. `settings.n.version`, then `settings.node.version`.
3. The nearest valid `package.json`'s `engines.node`.
4. Its `devEngines.runtime` entry named `node`.
5. The default, `>=16.0.0`.

Every version in the range must support a feature. For example, dynamic imports
support `^12.17.0 || >=13.2.0`, but `>=12.17.0` also includes unsupported Node.js
13.0 and 13.1. Strict mode is taken into account for features whose support in
older Node.js releases depends on it.

### ignores

An array of features to skip, for example when a transpiler or polyfill supplies
them. Each feature accepts its upstream rule name, the name without `no-`, and
the camelCase spelling: `no-optional-chaining`, `optional-chaining`, and
`optionalChaining` are equivalent. Legacy aliases such as `new.target`,
`regexpLookbehind`, and `trailingCommasInFunctions` are also supported.

See the [upstream feature table](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unsupported-features/es-syntax.json)
for the complete list and version ranges.

### Shared settings

`settings.node.version` sets the target for multiple Node rules. The upstream
`settings.n.version` spelling remains supported and takes precedence when both
are set.

`settings["es-x"].aggressive: true` also reports prototype properties on receivers
whose type cannot be determined. By default, only recognized receivers are
reported. When a TypeScript project is configured, available type information
is used to recognize receivers, including unions and constrained generics.

## Differences from upstream

- With `version: ">=15 || >20 <15"`, dynamic imports and nullish coalescing
  are accepted: the impossible `>20 <15` alternative matches no Node.js version.
  Upstream reports both features. Use `>=15` to get the same result in both tools.
- Legacy octal literals such as `0755` produce a syntax error. Upstream accepts
  this spelling in non-strict code. Use the decimal spelling `493` if
  compatibility with Node.js versions before modern octal syntax is required.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-unsupported-features/es-syntax.md)
- [Upstream rule](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unsupported-features/es-syntax.js)
