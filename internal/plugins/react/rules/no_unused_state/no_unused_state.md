# no-unused-state

## Rule Details

Warns when a React component defines state fields that are never read anywhere in the component. State can be defined via constructor assignments (`this.state = {}`), class property declarations (`state = {}`), `this.setState()` calls, or `getInitialState()` returns in ES5 components.

Examples of **incorrect** code for this rule:

```javascript
class MyComponent extends React.Component {
  state = { foo: 0 };
  render() {
    return <SomeComponent />;
  }
}
```

```javascript
var MyComponent = createReactClass({
  getInitialState: function() {
    return { foo: 0 };
  },
  render: function() {
    return <SomeComponent />;
  }
});
```

Examples of **correct** code for this rule:

```javascript
class MyComponent extends React.Component {
  state = { foo: 0 };
  render() {
    return <SomeComponent foo={this.state.foo} />;
  }
}
```

```javascript
class MyComponent extends React.Component {
  state = { foo: 0 };
  render() {
    const { foo } = this.state;
    return <SomeComponent foo={foo} />;
  }
}
```

## Original Documentation

- [eslint-plugin-react/no-unused-state](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-unused-state.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/lib/rules/no-unused-state.js)
