# prefer-es6-class

## Rule Details

Enforce consistency between the two React component declaration styles: ES6
classes extending `React.Component` / `React.PureComponent` and the legacy
`createReactClass({...})` factory. The rule reports whichever style does not
match the configured option.

- `"always"` (default) — flag `createReactClass({...})` calls; prefer ES6
  classes.
- `"never"` — flag ES6 class declarations extending `Component` /
  `PureComponent` (bare or pragma-qualified); prefer `createReactClass`.

## Options

```json
{ "react/prefer-es6-class": ["error", "always"] }
```

Accepted values: `"always"` (default), `"never"`.

Examples of **incorrect** code for this rule (default `"always"`):

```javascript
var Hello = createReactClass({
  render: function () {
    return <div>Hello {this.props.name}</div>;
  },
});
```

Examples of **correct** code for this rule (default `"always"`):

```javascript
class Hello extends React.Component {
  render() {
    return <div>Hello {this.props.name}</div>;
  }
}
```

Examples of **incorrect** code for this rule with `"never"`:

```json
{ "react/prefer-es6-class": ["error", "never"] }
```

```javascript
class Hello extends React.Component {
  render() {
    return <div>Hello {this.props.name}</div>;
  }
}
```

Examples of **correct** code for this rule with `"never"`:

```json
{ "react/prefer-es6-class": ["error", "never"] }
```

```javascript
var Hello = createReactClass({
  render: function () {
    return <div>Hello {this.props.name}</div>;
  },
});
```

## Settings

The rule honors the shared React settings:

- `settings.react.pragma` — controls the namespace used for qualified
  references (default `"React"`, so `React.Component` and `React.createClass`
  are recognized).
- `settings.react.createClass` — controls the factory name (default
  `"createReactClass"`).

## Original Documentation

- [eslint-plugin-react `prefer-es6-class`](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/prefer-es6-class.md)
