# exports-style

Enforces a consistent CommonJS export style using either `module.exports` or `exports`.

## Rule details

`exports` initially refers to the same object as `module.exports`. Replacing either value breaks that connection, so mixing the two styles can leave properties out of the exported object.

By default, the rule disallows references to the global `exports` variable. With the `"exports"` option, it disallows references to `module.exports` and direct assignments to `exports`. Locally shadowed variables are ignored.

Examples of **incorrect** code with the default option:

```javascript
exports.foo = 1;
exports.bar = 2;
```

Examples of **correct** code with the default option:

```javascript
module.exports = { foo: 1, bar: 2 };
module.exports.baz = 3;
```

## Options

The first option selects the style:

- `"module.exports"` (default): require `module.exports`.
- `"exports"`: require `exports` property access and disallow assigning to `exports` itself.

The second option is an object with `allowBatchAssign`, which defaults to `false`. Setting it to `true` allows both styles in the same chained assignment.

```javascript
export default [
  {
    files: ["**/*.cjs"],
    plugins: ["node"],
    rules: {
      "node/exports-style": ["error", "exports", { allowBatchAssign: true }],
    },
  },
];
```

Examples of **correct** code with this configuration:

```javascript
module.exports = exports = function foo() {};
exports.bar = 1;
```

With `"exports"` and the default `allowBatchAssign: false`, use property assignments:

```javascript
exports.foo = 1;
exports.bar = 2;
```

## Automatic fixes

This rule reports inconsistent export styles without automatic fixes or suggestions. Update the export code manually so you can preserve which object is exported and which references still point to it.

## Differences from upstream

Unlike eslint-plugin-n, rslint does not automatically rewrite `module.exports` to `exports`. Replacing the exported object and mutating the existing object can produce different results:

```javascript
const old = exports;
module.exports = { a: 1 }; // Reported without an automatic fix.
console.log(old === module.exports); // false
```

Changing the assignment to `exports.a = 1` would make the comparison return `true`. Even replacing `module.exports.foo` with `exports.foo` can change behavior when `exports` is shadowed or the two values no longer refer to the same object.

## Original documentation

- [eslint-plugin-n: exports-style](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/exports-style.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/exports-style.js)
