# no-unsupported-features/es-builtins

Reports ECMAScript globals and static properties that are unavailable in some
Node.js versions allowed by your configured version range.

## Configuration

```javascript
export default [
  {
    plugins: ["node"],
    rules: {
      "node/no-unsupported-features/es-builtins": ["error", {
        version: ">=18.0.0",
        ignores: [],
      }],
    },
  },
];
```

## Rule details

With `version: ">=18.0.0"`, this is **incorrect** because `Map.groupBy` needs
Node.js 21 or later:

```javascript
const groups = Map.groupBy(items, item => item.category);
```

These APIs are **correct** for the same range:

```javascript
const entries = Object.fromEntries(pairs);
const calendars = Intl.supportedValuesOf("calendar");
```

The rule follows static property access, aliases and destructuring. It respects
local bindings and configured globals. It offers no automatic fixes or suggestions.

## Options

### version

A Node.js version range such as `">=18.0.0"` or `"^18 || >=20"`. Every version
allowed by the range must support the API.

The rule selects the first valid range from:

1. The rule's `version` option.
2. `settings.n.version`, then `settings.node.version`.
3. `engines.node` in the nearest `package.json`.
4. `devEngines.runtime` for the Node.js runtime in that package.
5. The default `">=16.0.0"`.

For example, share a target version through
`settings: { node: { version: ">=18.0.0" } }`.

### ignores

An array of exact API names, for example `["Promise.any", "Map.groupBy"]`.
Use this when a polyfill provides an API. Ignoring a global does not ignore its
properties: `"Atomics"` and `"Atomics.add"` are separate entries.

Accepted names and support versions follow the pinned upstream API table.

## Limitations

As in upstream, this rule checks known globals and static properties. It does
not detect instance methods such as `array.at()`, function parameters, options
passed to APIs, or events. Computed properties must have a statically known name.

## Differences from upstream

Contradictory alternatives do not affect a version range. For example, both
`>=16 || >20 <16` and `>20 <16 || >=16` allow `Promise.any`, since `>20 <16`
contains no versions. Upstream reports it for the first ordering. Removing the
contradictory alternative gives consistent results in both tools.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-unsupported-features/es-builtins.md)
- [Upstream implementation and API table](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unsupported-features/es-builtins.js)
