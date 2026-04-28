# no-unused-class-component-methods

## Rule Details

Warns you if you have defined a method or property but it is never being used anywhere.

Examples of **incorrect** code for this rule:

```jsx
class Foo extends React.Component {
  handleClick() {}
  render() {
    return null;
  }
}

class Foo extends React.Component {
  action = () => {};
  render() {
    return null;
  }
}

var Foo = createReactClass({
  a: 3,
  render() {
    return null;
  },
});
```

Examples of **correct** code for this rule:

```jsx
class Foo extends React.Component {
  handleClick() {}
  render() {
    return <button onClick={this.handleClick}>Text</button>;
  }
}

class Foo extends React.Component {
  action = () => {};
  anotherAction = () => this.action();
  render() {
    return <button onClick={this.anotherAction}>Example</button>;
  }
}

var Foo = createReactClass({
  getInitialState() {
    return { value: 0 };
  },
  render() {
    return <div>{this.state.value}</div>;
  },
});
```

Canonical React lifecycle methods (e.g. `componentDidMount`, `render`,
`shouldComponentUpdate`) and reserved property names (`state` on ES6 classes;
`getInitialState`, `getDefaultProps`, `mixins` on `createReactClass`) are
always ignored, even when unused.

## Original Documentation

- [react/no-unused-class-component-methods](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-unused-class-component-methods.md)
