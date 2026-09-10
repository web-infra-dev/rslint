# react/no-unused-prop-types

Warns when a React component defines a prop type that it never uses.

The rule checks runtime `propTypes` declarations and TypeScript object-shaped
props. Nested `shape` and `exact` declarations are skipped by default because
static analysis cannot always determine whether their members are used.

## Rule Details

Examples of **incorrect** code for this rule:

```jsx
class Hello extends React.Component {
  render() {
    return <div>Hello</div>;
  }
}

Hello.propTypes = {
  name: PropTypes.string,
};
```

Examples of **correct** code for this rule:

```jsx
class Hello extends React.Component {
  render() {
    return <div>Hello {this.props.name}</div>;
  }
}

Hello.propTypes = {
  name: PropTypes.string,
};
```

## Options

`ignore` accepts prop names that should not be checked. `customValidators`
lists validator namespaces that should be treated like `PropTypes`. Set
`skipShapeProps` to `false` to check nested `shape` and `exact` members.

```json
{ "react/no-unused-prop-types": ["error", { "skipShapeProps": false }] }
```

## Differences from ESLint

- TypeScript prop annotations are reported on the property name rather than
  including the trailing type annotation in the diagnostic range.

## Original Documentation

- [eslint-plugin-react: no-unused-prop-types](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/no-unused-prop-types.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-unused-prop-types.js)
