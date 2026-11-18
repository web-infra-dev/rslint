# no-deprecated

## Rule Details

Disallow usage of declarations marked with `@deprecated`.

## Differences from upstream

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
