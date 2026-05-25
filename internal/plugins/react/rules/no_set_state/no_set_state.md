# no-set-state

## Rule Details

Disallow any use of `this.setState(...)` inside a React class component or
ES5 `createReactClass(...)` component. When using an architecture that
separates application state from UI components (e.g. Flux / Redux), local
component state is rarely needed and `setState` calls should be replaced
with explicit dispatches to the external store.

The rule fires on every `this.setState(...)` call whose lexically enclosing
scope is a detected React component:

- An ES6 class extending `Component` / `PureComponent` (bare or
  pragma-qualified, e.g. `React.Component`).
- An object literal passed to `createReactClass(...)` (or
  `<pragma>.<createClass>(...)`).
- A stateless functional component (capital-cased name returning JSX or
  `null`).

A bare property reference such as `this.someHandler = this.setState;`
does NOT fire — the rule listens for CallExpressions only.

Examples of **incorrect** code for this rule:

```jsx
var Hello = createReactClass({
  getInitialState: function () {
    return { name: this.props.name };
  },
  handleClick: function () {
    this.setState({
      name: this.props.name.toUpperCase(),
    });
  },
  render: function () {
    return (
      <div onClick={this.handleClick.bind(this)}>Hello {this.state.name}</div>
    );
  },
});
```

```jsx
class Hello extends React.Component {
  someMethod = () => {
    this.setState({ name: this.props.name.toUpperCase() });
  };
  render() {
    return <div onClick={this.someMethod}>Hello {this.state.name}</div>;
  }
}
```

Examples of **correct** code for this rule:

```jsx
var Hello = createReactClass({
  render: function () {
    return (
      <div onClick={this.props.handleClick}>Hello {this.props.name}</div>
    );
  },
});
```

```jsx
class Hello extends React.Component {
  someMethod() {
    const fn = this.setState;
    this.someHandler = this.setState;
  }
  render() {
    return <div>{this.props.name}</div>;
  }
}
```

## Original Documentation

- [eslint-plugin-react `no-set-state`](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-set-state.md)
