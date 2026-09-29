# prefer-exact-props

Require React `propTypes` definitions to use an exact wrapper.

## Rule Details

This rule reports non-empty `propTypes` objects that are not wrapped by a
function marked with `exact: true` in `settings.propWrapperFunctions`. Empty
objects are allowed.

Configure an exact wrapper in the shared React settings:

```json
{
  "settings": {
    "propWrapperFunctions": [{ "property": "exact", "exact": true }]
  }
}
```

Examples of **incorrect** code for this rule:

```jsx
function Component(props) {
  return <div>{props.name}</div>;
}

Component.propTypes = {
  name: PropTypes.string,
};
```

```jsx
class Component extends React.Component {
  static propTypes = otherWrapper({
    name: PropTypes.string,
  });
}
```

Examples of **correct** code for this rule:

```jsx
function Component(props) {
  return <div>{props.name}</div>;
}

Component.propTypes = exact({
  name: PropTypes.string,
});
```

```jsx
class Component extends React.Component {
  static propTypes = {};
}
```

## Differences from ESLint

- Only assignments to `.propTypes` are checked. Other reads are ignored;
  eslint-plugin-react 7.37.5 may report them or throw while evaluating them.
- Class fields are checked only when they are static and their runtime name is
  `propTypes`. Instance fields, `#private` fields, and TypeScript `props`
  fields are ignored.
- Computed member names are checked only when written as a string or static
  template literal, such as `Component['propTypes']`. Variables used as keys
  are ignored.
- A `propTypes` value referenced through a variable is checked only when the
  variable is a directly initialized `const`. Mutable and destructured
  bindings are treated as unknown. Later property mutations of the referenced
  `const` object are not tracked.

## Original Documentation

- [eslint-plugin-react: prefer-exact-props](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/prefer-exact-props.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/prefer-exact-props.js)
