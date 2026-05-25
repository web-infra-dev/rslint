# prefer-stateless-function

## Rule Details

Stateless functional components are simpler than class based components and
will benefit from future React performance optimizations specific to these
components. This rule will check your class based components for missing
state, missing lifecycle methods, missing `this` member usages or other
patterns that suggest the component could be safely written as a stateless
function.

Examples of **incorrect** code for this rule:

```jsx
var Foo = createReactClass({
  render: function () {
    return <div>{this.props.foo}</div>;
  },
});
```

```jsx
class Foo extends React.Component {
  render() {
    return <div>{this.props.foo}</div>;
  }
}
```

Examples of **correct** code for this rule:

```jsx
const Foo = function (props) {
  return <div>{props.foo}</div>;
};
```

```jsx
const Foo = ({ foo }) => <div>{foo}</div>;
```

## Rule Options

```json
{
  "react/prefer-stateless-function": [
    "error",
    { "ignorePureComponents": true }
  ]
}
```

### `ignorePureComponents`

When `true`, classes that extend `React.PureComponent` are exempt — the
rationale being that `PureComponent` already provides a default
`shouldComponentUpdate`, which a plain functional component does not.
Defaults to `false`.

Examples of **correct** code for this rule with `{ "ignorePureComponents": true }`:

```json
{ "react/prefer-stateless-function": ["error", { "ignorePureComponents": true }] }
```

```jsx
class Foo extends React.PureComponent {
  render() {
    return <div>{this.props.foo}</div>;
  }
}
```

## Original Documentation

- [react/prefer-stateless-function](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/prefer-stateless-function.md)
