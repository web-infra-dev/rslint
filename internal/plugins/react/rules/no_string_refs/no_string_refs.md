# no-string-refs

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

- `settings.react.version` — when set to a version `>= 18.3.0`, `this.refs` accesses are not reported (they are writable on modern React). When unset, `settings.react.defaultVersion` is used, then latest.
- `settings.react.pragma` — used to recognize `<pragma>.createClass(...)` and classes extending `<pragma>.Component` / `<pragma>.PureComponent`. Defaults to `React`; a file-level `@jsx` annotation takes precedence.
- `settings.react.createClass` — the identifier used for ES5 component factories. Defaults to `createReactClass`.

## Differences from ESLint

- **JSDoc `@extends` / `@augments` tags are not honored**. eslint-plugin-react has an `isExplicitComponent` path that treats a class as a React component when its JSDoc contains `@extends React.Component` or `@augments React.Component`, even without an `extends` clause. rslint only recognizes components through the actual `extends` clause; JSDoc-only component declarations are not flagged.
- **`settings.react.version = "detect"` is not resolved from `node_modules`**. rslint uses `settings.react.defaultVersion` when provided and otherwise treats the version as latest. Set an explicit version string (e.g. `"18.2.0"`) for exact version-aware behavior.
- **Semver range strings (`^18.0.0`, `~18.0.0`, `>=17 <19`, etc.) are interpreted differently**. eslint-plugin-react coerces the setting to one version before comparison, while rslint extracts the first numeric triple and compares it directly. Prefer an exact version string for predictable behavior.
- **Computed identifier member names are not treated as literal property names**. eslint-plugin-react reads ESTree's `property.name` without checking `computed`, so it reports `this[refs]` and classifies `React[createClass](...)` when those identifiers happen to have the configured names. rslint does not assume their runtime values. Private `this.#refs` is likewise not the public legacy `this.refs` API and is not reported.
- **Invalid `settings.react.createClass` values do not terminate linting**. eslint-plugin-react throws while initializing the rule; rslint treats the configured string as a non-matching factory name because the native rule API has no recoverable settings-error channel.

## Original Documentation

- [eslint-plugin-react: no-string-refs](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/no-string-refs.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-string-refs.js)
