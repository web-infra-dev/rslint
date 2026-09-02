# prop-types

Disallow missing props validation in a React component definition. The rule
checks props read by React class components, `createReactClass` components, and
stateless function components against their declared `propTypes`.

```jsx
function Hello({ name }) {
  return <div>{name}</div>;
}

Hello.propTypes = { name: PropTypes.string.isRequired };
```

The `ignore` option excludes named props. `skipUndeclared` limits checking to
components with a `propTypes` declaration. `customValidators` is accepted for
configuration compatibility; custom validator calls are treated as declared
prop types.

## TypeScript support

TypeScript prop declarations are supported for function parameters, React
component generics, class `props` declarations, type aliases, interfaces,
intersections, unions, and `ReturnType` declarations. React component wrappers
such as `FC`, `FunctionComponent`, `VFC`, `VoidFunctionComponent`,
`ForwardRefRenderFunction`, and `PropsWithChildren` are recognized when they
resolve to the React import.

Validation follows the native rule's shallow TypeScript behavior: a declared
top-level property accepts deeper reads. Imported or otherwise unresolved types
and external `propTypes` expressions are treated as accepting any property.
Flow-specific declarations and unsupported TypeScript utility types retain the
native rule's opaque behavior.

- [eslint-plugin-react: prop-types](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/prop-types.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/prop-types.js)
