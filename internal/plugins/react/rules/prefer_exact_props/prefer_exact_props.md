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

- Standalone and optional-chain reads of `.propTypes` are ignored. Version
  7.37.5 of eslint-plugin-react throws while evaluating those expressions.

## Original Documentation

- [eslint-plugin-react: prefer-exact-props](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/prefer-exact-props.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/prefer-exact-props.js)
