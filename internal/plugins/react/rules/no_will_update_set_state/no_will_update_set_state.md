# no-will-update-set-state

Disallow `this.setState` inside `componentWillUpdate`.

Updating the state during a component update (before render) is a no-op — the
update is already in flight, so calling `setState` here either triggers an
infinite loop or is silently dropped.

## Rule Details

This rule flags any `this.setState` call whose enclosing class method, class
field initializer, or object-literal property is keyed `componentWillUpdate`.
When `settings.react.version` is `>= 16.3.0` (or unset — treated as latest),
the rule additionally flags the same pattern keyed `UNSAFE_componentWillUpdate`,
matching the renamed alias shipped in React 16.3.

By default, calls inside a nested function (regular or arrow) are allowed —
enable `disallow-in-func` to forbid them too.

Examples of **incorrect** code for this rule:

```javascript
var Hello = createReactClass({
  componentWillUpdate: function () {
    this.setState({
      name: this.props.name.toUpperCase(),
    });
  },
});
```

```javascript
class Hello extends React.Component {
  componentWillUpdate() {
    this.setState({
      name: this.props.name.toUpperCase(),
    });
  }
}
```

Examples of **correct** code for this rule:

```javascript
var Hello = createReactClass({
  componentWillUpdate: function () {
    someNonMemberFunction(arg);
    this.someHandler = this.setState;
  },
});
```

```javascript
var Hello = createReactClass({
  componentWillUpdate: function () {
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
{ "react/no-will-update-set-state": ["error", "disallow-in-func"] }
```

With `disallow-in-func` set, the rule also flags `this.setState` calls inside
nested functions:

```javascript
var Hello = createReactClass({
  componentWillUpdate: function () {
    someClass.onSomeEvent(function (data) {
      this.setState({
        data: data,
      });
    });
  },
});
```

## React Version

Unlike `no-did-update-set-state` / `no-did-mount-set-state`, this rule is not a
no-op at any React version — `componentWillUpdate` is unsafe on every release.
The React version only controls whether `UNSAFE_componentWillUpdate` (the 16.3+
renamed alias) is also flagged, matching `eslint-plugin-react`'s
`shouldCheckUnsafeCb`.

## Original Documentation

- [eslint-plugin-react / no-will-update-set-state](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-will-update-set-state.md)
