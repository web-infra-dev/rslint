# react/no-string-refs

Disallow using deprecated string refs.

## Rule Details

React used to support string refs (e.g. `ref="name"`, then accessed via `this.refs.name`), but string refs are deprecated — they tie the ref to the component that rendered it (making composition surprising), interact poorly with `<StrictMode>`, and cannot be cleaned up automatically. Callback refs (`ref={node => ...}`) and `React.createRef()` / `useRef()` should be used instead.

This rule reports two things:

- A string literal (or, optionally, a template literal) used as a `ref` prop value: `ref="hello"`, `ref={'hello'}`, `ref={\`hello\`}`.
- Access of `this.refs` inside an ES5 `createReactClass({...})` component or an ES6 class extending `React.Component` / `React.PureComponent`. React 18.3.0 made `this.refs` writable, so the check is skipped when `settings.react.version` is set to 18.3.0 or later.

Examples of **incorrect** code for this rule:

```jsx
var Hello = createReactClass({
  componentDidMount: function () {
    var component = this.refs.hello;
  },
  render: function () {
    return <div ref="hello">Hello</div>;
  },
});
```

```jsx
var Hello = createReactClass({
  render: function () {
    return <div ref={'hello'}>Hello</div>;
  },
});
```

Examples of **correct** code for this rule:

```jsx
var Hello = createReactClass({
  componentDidMount: function () {
    var component = this.hello;
  },
  render: function () {
    return <div ref={(c) => (this.hello = c)}>Hello</div>;
  },
});
```

## Options

### `noTemplateLiterals`

When `true`, template literals used as a `ref` value are also reported (by default only plain string literals are flagged).

```json
{ "react/no-string-refs": ["error", { "noTemplateLiterals": true }] }
```

```jsx
var Hello = createReactClass({
  render: function () {
    return <div ref={`hello`}>Hello</div>;
  },
});
```

## Settings

- `settings.react.version` — when set to a version `>= 18.3.0`, `this.refs` accesses are not reported (they are writable on modern React). When unset, defaults to latest.
- `settings.react.pragma` — used to recognize `<pragma>.createClass(...)` and classes extending `<pragma>.Component` / `<pragma>.PureComponent`. Defaults to `React`.
- `settings.react.createClass` — the identifier used for ES5 component factories. Defaults to `createReactClass`.

## Differences from ESLint

- **JSDoc `@extends` / `@augments` tags are not honored**. eslint-plugin-react has an `isExplicitComponent` path that treats a class as a React component when its JSDoc contains `@extends React.Component` or `@augments React.Component`, even without an `extends` clause. rslint only recognizes components through the actual `extends` clause; JSDoc-only component declarations are not flagged.
- **`settings.react.version = "detect"` is not resolved**. eslint-plugin-react reads `react` from `node_modules` to auto-detect the version. rslint treats `"detect"` (and any non-numeric version string) as "latest" (999.999.999), which means `this.refs` is not reported in that mode. Set an explicit version string (e.g. `"18.2.0"`) to exercise the `< 18.3.0` gate.
- **Semver range strings (`^18.0.0`, `~18.0.0`, `>=17 <19`, etc.) are interpreted loosely**. eslint-plugin-react passes the value to `semver.satisfies`, which throws on non-exact versions. rslint extracts the leading numeric triple (`^18.0.0` → `18.0.0`) and compares it directly. Prefer an exact version string for predictable behavior.

## Original Documentation

- https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-string-refs.md
