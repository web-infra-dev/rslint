# no-unused-prop-types

Warns when a React component defines a prop type that it never uses.

The rule checks runtime `propTypes` declarations and TypeScript object-shaped props. Nested `shape` and `exact` declarations are skipped by default because static analysis cannot always determine whether their members are used.

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

`ignore` accepts prop names that should not be checked. `customValidators` lists custom validator namespaces whose arguments remain opaque. Set `skipShapeProps` to `false` to check nested `shape` and `exact` members.

```json
{ "react/no-unused-prop-types": ["error", { "skipShapeProps": false }] }
```

## Differences from ESLint

- TypeScript prop annotations are reported on the property name rather than including the trailing type annotation in the diagnostic range.
- Aliases follow lexical bindings. For example, a `shared` variable inside a component does not replace the outer `shared` in `Foo.propTypes = shared`, and an unrelated local variable named `props` does not consume component props.
- Explicitly declared props named `toString`, `constructor`, or other `Object.prototype` members count as used when accessed.
- Static computed keys such as `static ['propTypes']` and `{ ['Foo']: component }` are recognized like their non-computed equivalents.
- React generic annotations are checked even when the component has no props parameter, for example `const Foo: React.FC<Props> = () => <div />`.
- Configured custom validators remain opaque with or without `.isRequired`.

## Original Documentation

- [eslint-plugin-react: no-unused-prop-types](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/no-unused-prop-types.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-unused-prop-types.js)
