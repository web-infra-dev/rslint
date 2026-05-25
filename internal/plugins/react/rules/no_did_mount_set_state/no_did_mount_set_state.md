# no-did-mount-set-state

Disallow `this.setState` inside `componentDidMount`.

Updating the state immediately after the initial mount triggers a second
`render()` call and can lead to property/layout thrashing.

## Rule Details

This rule flags any `this.setState` call whose enclosing class method, class
field initializer, or object-literal property is keyed `componentDidMount`.
By default, calls inside a nested function (regular or arrow) are allowed —
enable `disallow-in-func` to forbid them too.

Examples of **incorrect** code for this rule:

```javascript
var Hello = createReactClass({
  componentDidMount: function () {
    this.setState({
      name: this.props.name.toUpperCase(),
    });
  },
});
```

```javascript
class Hello extends React.Component {
  componentDidMount() {
    this.setState({
      name: this.props.name.toUpperCase(),
    });
  }
}
```

Examples of **correct** code for this rule:

```javascript
var Hello = createReactClass({
  componentDidMount: function () {
    someNonMemberFunction(arg);
    this.someHandler = this.setState;
  },
});
```

```javascript
var Hello = createReactClass({
  componentDidMount: function () {
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
{ "react/no-did-mount-set-state": ["error", "disallow-in-func"] }
```

With `disallow-in-func` set, the rule also flags `this.setState` calls inside
nested functions:

```javascript
var Hello = createReactClass({
  componentDidMount: function () {
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
`componentDidMount`.

## Differences from ESLint

- Pre-release version strings such as `"16.3.0-rc.1"` are treated as
  `16.3.0` (rule becomes a no-op). `eslint-plugin-react` follows semver,
  where `16.3.0-rc.1` ranks below `16.3.0`, so it keeps the rule active.
  If you pin a pre-release version and want the rule to flag
  `this.setState` inside `componentDidMount`, set
  `settings.react.version` to a release version such as `"16.2.0"`.

## Original Documentation

- [eslint-plugin-react / no-did-mount-set-state](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-did-mount-set-state.md)
