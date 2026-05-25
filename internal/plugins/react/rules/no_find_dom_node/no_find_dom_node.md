# no-find-dom-node

Disallow usage of `findDOMNode`.

Facebook will eventually deprecate `findDOMNode` as it blocks certain
improvements in React in the future. It is recommended to use callback refs
instead.

## Rule Details

This rule flags any call whose callee is the identifier `findDOMNode` (bare
call) or a member-access whose property name is `findDOMNode`
(`React.findDOMNode(...)`, `ReactDOM.findDOMNode(...)`, etc.). Computed
(bracket) access such as `React['findDOMNode'](...)` is intentionally not
flagged — this mirrors the upstream rule's AST check.

Examples of **incorrect** code for this rule:

```jsx
class MyComponent extends Component {
  componentDidMount() {
    findDOMNode(this).scrollIntoView();
  }
  render() {
    return <div />;
  }
}
```

```jsx
class MyComponent extends Component {
  componentDidMount() {
    React.findDOMNode(this).scrollIntoView();
  }
  render() {
    return <div />;
  }
}
```

Examples of **correct** code for this rule:

```jsx
class MyComponent extends Component {
  componentDidMount() {
    this.node.scrollIntoView();
  }
  render() {
    return <div ref={(node) => (this.node = node)} />;
  }
}
```

## Original Documentation

- [eslint-plugin-react / no-find-dom-node](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-find-dom-node.md)
