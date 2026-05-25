# no-unsafe

## Rule Details

Disallow the legacy `UNSAFE_`-prefixed React lifecycle methods that are unsafe under async rendering (React 16.3+):

- `UNSAFE_componentWillMount` — replace with `componentDidMount`.
- `UNSAFE_componentWillReceiveProps` — replace with `getDerivedStateFromProps`.
- `UNSAFE_componentWillUpdate` — replace with `componentDidUpdate`.

The rule fires only when the method is declared inside a detected React component:

- An ES6 class extending `Component` / `PureComponent` (bare or pragma-qualified, e.g. `React.Component`).
- An object literal passed to `createReactClass(...)` (or `<pragma>.<createClass>(...)`).

When `settings.react.version` is below `16.3.0` the rule disables itself entirely (the `UNSAFE_` aliases were introduced in that release).

Examples of **incorrect** code for this rule:

```jsx
class Foo extends React.Component {
  UNSAFE_componentWillMount() {}
  UNSAFE_componentWillReceiveProps() {}
  UNSAFE_componentWillUpdate() {}
}
```

```jsx
const Foo = createReactClass({
  UNSAFE_componentWillMount: function () {},
  UNSAFE_componentWillReceiveProps: function () {},
  UNSAFE_componentWillUpdate: function () {},
});
```

Examples of **correct** code for this rule:

```jsx
class Foo extends React.Component {
  componentDidMount() {}
  componentDidUpdate() {}
}
```

```jsx
class Foo extends Bar {
  // Foo is not a React component — `UNSAFE_*` is just a method name here.
  UNSAFE_componentWillMount() {}
}
```

## Rule Options

```json
{ "react/no-unsafe": ["error", { "checkAliases": true }] }
```

### `checkAliases` (default: `false`)

When `true`, the rule additionally flags the unprefixed legacy aliases (`componentWillMount`, `componentWillReceiveProps`, `componentWillUpdate`) inside detected React components. The replacement suggestions are the same as for the `UNSAFE_`-prefixed forms.

Examples of **incorrect** code with `{ "checkAliases": true }`:

```json
{ "react/no-unsafe": ["error", { "checkAliases": true }] }
```

```jsx
class Foo extends React.Component {
  componentWillMount() {}
  componentWillReceiveProps() {}
  componentWillUpdate() {}
}
```

## Settings

This rule is influenced by the shared React settings:

- `settings.react.version` — when below `16.3.0`, the rule is a no-op (default: latest).
- `settings.react.pragma` — the `React` object name used for `<pragma>.Component` / `<pragma>.PureComponent` extends matching (default: `"React"`).
- `settings.react.createClass` — the `createReactClass` factory name used for ES5 component detection (default: `"createReactClass"`).

## Original Documentation

- [eslint-plugin-react / no-unsafe](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-unsafe.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/lib/rules/no-unsafe.js)
