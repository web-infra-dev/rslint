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

## Differences from ESLint

TypeScript and Flow type declarations are not yet used as prop declarations by
the native rule. External `propTypes` expressions are treated as accepting any
property, matching the rule's purpose of avoiding false positives when the
declaration is maintained elsewhere.

- [eslint-plugin-react: prop-types](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/prop-types.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/prop-types.js)
