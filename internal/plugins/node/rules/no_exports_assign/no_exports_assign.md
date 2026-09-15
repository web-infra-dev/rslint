# no-exports-assign

Disallow assignments to the global `exports` variable.

The `node` plugin ports this rule from `eslint-plugin-n`.

## Rule Details

Assigning a new value to `exports` does not replace the value exported by a
CommonJS module. Assign to `module.exports` instead.

Examples of **incorrect** code:

```javascript
exports = {};
exports = { foo: 1 };
```

Examples of **correct** code:

```javascript
module.exports = { foo: 1 };
module.exports.foo = 1;
exports.bar = 2;

module.exports = exports = {};
exports = module.exports = {};

function update(exports) {
  exports = {};
}
```

The rule checks assignment expressions, including compound assignments. It
allows a chained assignment when the immediately enclosing or right-hand
assignment targets the global `module.exports`. Computed access such as
`module['exports']` does not qualify for this exception.

Only variables belonging to the global scope are checked. Local parameters,
imports, and other local bindings named `exports` are allowed.

## Options

This rule has no options and provides no automatic fixes or suggestions.

Enable the bundled `node` plugin. CommonJS files supply `exports` and `module`
globals automatically; other source types can configure them explicitly:

```javascript
export default [
  {
    plugins: ['node'],
    languageOptions: {
      globals: { exports: 'writable', module: 'readonly' },
    },
    rules: { 'node/no-exports-assign': 'error' },
  },
];
```

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-exports-assign.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-exports-assign.js)
