# no-deprecated

## Rule Details

Disallow usage of declarations marked with `@deprecated`.

## Differences from upstream

### Computed property keys

Rslint checks the key expression's own deprecation independently of the
accessed property. typescript-eslint v8.72.2 misses the deprecated variable
in both accesses below:

```ts
/** @deprecated Use currentKey. */
const key = 'ok';
const obj = { ok: 1 };

obj[key]; // Reports key.
obj[(key)]; // Reports key.
```

If both the key variable and the accessed property are deprecated, each
reference is reported. The `allow` option applies independently to them.

### JavaScript prototype definitions

Rslint checks deprecated prototype members of JavaScript constructor
functions, including references through import aliases and instances held in
bindings that are not reassigned. It does not infer a constructor from a method's name.

The assignment carrying the annotation defines the deprecated member and is
not reported. Subsequent references, including ordinary writes, are checked.
typescript-eslint v8.72.2 also reports the defining assignment below:

```js
function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy = function () {};

new Box().legacy(); // Reports legacy.
```

### Dotted namespaces

Rslint does not report the identifiers in a dotted namespace declaration.
typescript-eslint v8.72.2 also reports `Legacy` and `API` in the declaration
below. Rslint reports them when the namespace is referenced:

```ts
/** @deprecated Use the replacement API. */
namespace Legacy.API {
  export const value = 1;
}

Legacy.API.value; // Reports Legacy and API.
```

The annotation applies to the entire dotted declaration and is also checked
through import aliases and destructuring. A separately declared namespace
inside a namespace block does not inherit the outer namespace's annotation.

## Original Documentation

- [typescript-eslint: no-deprecated](https://typescript-eslint.io/rules/no-deprecated)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.72.2/packages/eslint-plugin/src/rules/no-deprecated.ts)
