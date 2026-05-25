# no-did-update-set-state

Disallow `this.setState` inside `componentDidUpdate`.

Updating the state after a component update will trigger a second `render()`
call and can lead to property/layout thrashing.

## Rule Details

This rule flags any `this.setState` call whose enclosing class method, class
field initializer, or object-literal property is keyed `componentDidUpdate`.
By default, calls inside a nested function (regular or arrow) are allowed —
enable `disallow-in-func` to forbid them too.

Examples of **incorrect** code for this rule:

```javascript
var Hello = createReactClass({
  componentDidUpdate: function () {
    this.setState({
      name: this.props.name.toUpperCase(),
    });
  },
});
```

```javascript
class Hello extends React.Component {
  componentDidUpdate() {
    this.setState({
      name: this.props.name.toUpperCase(),
    });
  }
}
```

Examples of **correct** code for this rule:

```javascript
var Hello = createReactClass({
  componentDidUpdate: function () {
    someNonMemberFunction(arg);
    this.someHandler = this.setState;
  },
});
```

```javascript
var Hello = createReactClass({
  componentDidUpdate: function () {
    someClass.onSomeEvent(function (data) {
      this.setState({
        data: data,
      });
    });
  },
});
```

## Rule Options

```json
{ "react/no-did-update-set-state": ["error", "disallow-in-func"] }
```

With `disallow-in-func` set, the rule also flags `this.setState` calls inside
nested functions:

```javascript
var Hello = createReactClass({
  componentDidUpdate: function () {
    someClass.onSomeEvent(function (data) {
      this.setState({
        data: data,
      });
    });
  },
});
```

## React Version

The rule is a no-op when `settings.react.version` is explicitly set to
`>= 16.3.0`, matching `eslint-plugin-react`'s `shouldBeNoop` gate for
`componentDidUpdate`.

## Original Documentation

- [eslint-plugin-react / no-did-update-set-state](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-did-update-set-state.md)
