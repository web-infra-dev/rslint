# sort-comp

## Rule Details

Enforce a consistent order for methods and properties in React components.

The default order is static methods, lifecycle members, other members, and
then `render`. Lifecycle members follow React's conventional order, beginning
with `displayName` and ending with `componentWillUnmount`.

## Options

The rule accepts an object with `order` and `groups` properties. `order` is an
array of method names, special groups, or JavaScript regular-expression
strings such as `"/^on.+$/"`. `groups` defines named arrays that can be used
from `order`; it extends the built-in `lifecycle` group.

Special groups include:

- `static-variables`, `static-methods`
- `instance-variables`, `instance-methods`
- `type-annotations`, `getters`, `setters`
- `everything-else` and `render`

A member that matches no configured entry belongs to `everything-else` when
that group is present. Members that match multiple entries may be valid in any
compatible position.

```json
{
  "react/sort-comp": ["error", {
    "order": [
      "static-methods",
      "lifecycle",
      "/^on.+$/",
      "render",
      "everything-else"
    ]
  }]
}
```

## Differences from ESLint

The upstream documentation describes automatic fixing through the separate
`react-codemod` `sort-comp` transform. rslint reports the same ordering
diagnostics but does not invoke that external codemod.

## Original Documentation

- [eslint-plugin-react: sort-comp](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/sort-comp.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/sort-comp.js)
