# no-namespace

## Rule Details

Enforce that namespaces are not used in React elements. React does not support
namespace-qualified element names such as `<svg:circle />`.

Examples of **incorrect** code for this rule:

```jsx
<ns:TestComponent />
React.createElement('ns:TestComponent');
```

Examples of **correct** code for this rule:

```jsx
<TestComponent />
<testComponent />
React.createElement('TestComponent');
```

## Original Documentation

- [eslint-plugin-react: no-namespace](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/no-namespace.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-namespace.js)
