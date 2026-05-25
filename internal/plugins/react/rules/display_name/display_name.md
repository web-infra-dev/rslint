# display-name

## Rule Details

Disallow missing `displayName` in a React component definition.

`displayName` allows you to name your component. This name is used by React in debugging messages.

Examples of **incorrect** code for this rule:

```jsx
var Hello = createReactClass({
  render: function () {
    return <div>Hello {this.props.name}</div>;
  },
});

const Hello = React.memo(({ a }) => {
  return <>{a}</>;
});

export default ({ a }) => {
  return <>{a}</>;
};
```

Examples of **correct** code for this rule:

```jsx
var Hello = createReactClass({
  displayName: 'Hello',
  render: function () {
    return <div>Hello {this.props.name}</div>;
  },
});

const Hello = React.memo(function Hello({ a }) {
  return <>{a}</>;
});
```

## Rule Options

### `ignoreTranspilerName` (default: `false`)

When `true` the rule will ignore the name set by the transpiler and require a `displayName` property in this case.

Examples of **correct** code for this rule with `{ ignoreTranspilerName: true }`:

```json
{ "react/display-name": ["error", { "ignoreTranspilerName": true }] }
```

```jsx
var Hello = createReactClass({
  displayName: 'Hello',
  render: function () {
    return <div>Hello {this.props.name}</div>;
  },
});
module.exports = Hello;
```

```jsx
export default class Hello extends React.Component {
  render() {
    return <div>Hello {this.props.name}</div>;
  }
}
Hello.displayName = 'Hello';
```

Examples of **incorrect** code for this rule with `{ ignoreTranspilerName: true }`:

```json
{ "react/display-name": ["error", { "ignoreTranspilerName": true }] }
```

```jsx
var Hello = createReactClass({
  render: function () {
    return <div>Hello {this.props.name}</div>;
  },
});
module.exports = Hello;
```

```jsx
export default class Hello extends React.Component {
  render() {
    return <div>Hello {this.props.name}</div>;
  }
}
```

### `checkContextObjects` (default: `false`)

When `true`, this rule will also warn on context objects (created via `React.createContext()` / `createContext()`) that don't have a `displayName` set.

This option is silently disabled when `settings.react.version` is below `16.3.0`, since `Context.displayName` was introduced in React 16.3.

Examples of **incorrect** code for this rule with `{ checkContextObjects: true }`:

```json
{ "react/display-name": ["error", { "checkContextObjects": true }] }
```

```jsx
const Hello = React.createContext();
```

```jsx
const Hello = createContext();
```

Examples of **correct** code for this rule with `{ checkContextObjects: true }`:

```json
{ "react/display-name": ["error", { "checkContextObjects": true }] }
```

```jsx
const Hello = React.createContext();
Hello.displayName = 'HelloContext';
```

```jsx
const Hello = createContext();
Hello.displayName = 'HelloContext';
```

## About component detection

For this rule to work we need to detect React components, this could be very hard since components could be declared in a lot of ways.

For now we detect components created with:

- `createReactClass()`
- an ES6 class that inherits from `React.Component` / `Component` / `PureComponent`
- a stateless function that returns JSX or the result of a `React.createElement` call
- `React.memo` / `React.forwardRef` (or destructured `memo` / `forwardRef`) wrapping any of the above

## Settings

This rule is influenced by the shared React settings:

- `settings.react.pragma` — the `React` object name used for `<pragma>.Component` / `<pragma>.createContext` / `<pragma>.memo` / `<pragma>.forwardRef` matching (default: `"React"`).
- `settings.react.createClass` — the `createReactClass` factory name used for ES5 component detection (default: `"createReactClass"`).
- `settings.react.version` — gates the version-dependent behaviors:
  - Nested `<pragma>.memo(<pragma>.forwardRef(...))` is allowed for `^0.14.10 || ^15.7.0 || >= 16.12.0` (default: latest).
  - `checkContextObjects` is silently disabled when below `16.3.0`.
- `settings.componentWrapperFunctions` — additional wrapper factories beyond `memo` / `forwardRef`. Accepts strings (`"observer"`) or `{ property, object }` entries. The `<pragma>` placeholder is substituted with the configured pragma.

## When Not To Use It

If you are not interested in giving every component a `displayName` and rely on bundler / transpiler-derived names for debugging, you can disable this rule.

## Original Documentation

- [eslint-plugin-react / display-name](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/display-name.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/lib/rules/display-name.js)
