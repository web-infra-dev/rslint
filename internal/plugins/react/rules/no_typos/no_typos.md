# no-typos

Disallow common typos in React component declarations.

## Rule Details

This rule catches casing mistakes in React-specific identifiers that are easy to
miss because the surrounding code still runs — just not the way the author
expected.

It flags four classes of typo:

1. **Static class property names** on React components: `propTypes`,
   `contextTypes`, `childContextTypes`, `defaultProps`. Any identifier whose
   lowercased form equals one of these but whose casing differs is reported —
   whether declared as a static class field (`static PropTypes = {}`) or
   assigned externally (`Component.PropTypes = {}`).
2. **React lifecycle method names** on class components or inside
   `createReactClass({ ... })`:
   - Instance: `getDefaultProps`, `getInitialState`, `getChildContext`,
     `componentWillMount`, `UNSAFE_componentWillMount`, `componentDidMount`,
     `componentWillReceiveProps`, `UNSAFE_componentWillReceiveProps`,
     `shouldComponentUpdate`, `componentWillUpdate`, `UNSAFE_componentWillUpdate`,
     `getSnapshotBeforeUpdate`, `componentDidUpdate`, `componentDidCatch`,
     `componentWillUnmount`, `render`.
   - Static: `getDerivedStateFromProps`.
3. **Static lifecycle declared non-static**: `getDerivedStateFromProps` (or a
   casing variant of it) declared as an instance method triggers
   `staticLifecycleMethod` in addition to any casing typo.
4. **`PropTypes` usage errors** — when a `prop-types` or `react` import is
   present:
   - A property name in `PropTypes.<name>` / `<alias>.<name>` that is not one
     of the keys exported by `prop-types` (e.g. `PropTypes.Number`,
     `PropTypes.bools`) emits `typoPropType`.
   - A chain qualifier that is not `isRequired` (e.g. `.isrequired`) emits
     `typoPropTypeChain`.
   - `import 'react'` with no binding emits `noReactBinding`.
   - `import 'prop-types'` with no binding emits `noPropTypesBinding`.

### Component detection

A class is considered a React component when it `extends` `Component` or
`PureComponent` — either as a bare identifier or qualified by the React
pragma (`React.Component` by default, configurable via
`settings.react.pragma`). Classes with a JSDoc `@extends React.Component` /
`@augments React.Component` tag are also recognized.

For `Component.PropTypes = ...` / `Component.propTypes = ...` assignments the
rule resolves `Component` file-wide to:

- a class (declared with either `class Foo` or `const Foo = class`) matching
  the above extends rule, or
- a function declaration / expression / arrow function whose body directly
  contains a `return <jsx />`.

Classes and functions that don't satisfy those conditions — including
`React.forwardRef` / `React.memo` / `styled.xxx` results — are not treated as
components, matching `eslint-plugin-react`'s behavior.

## Differences from ESLint

- **Computed / bracket-notation access is not tracked.** Upstream documents
  `Component['PropTypes']` as "currently not supported" and this rule follows
  suit — only `PropertyAccessExpression` (`a.b`) is examined. Code like
  `Foo['prop' + 'Types'] = {}` silently passes, matching ESLint.

Examples of **incorrect** code for this rule:

```jsx
class MyComponent extends React.Component {
  static PropTypes = {};
}
```

```jsx
class MyComponent extends React.Component {
  componentwillMount() {}
}
```

```jsx
import PropTypes from 'prop-types';
class MyComponent extends React.Component {}
MyComponent.propTypes = {
  a: PropTypes.Number,
  b: PropTypes.string.isrequired,
};
```

```jsx
class MyComponent extends React.Component {
  getDerivedStateFromProps() {}
}
```

Examples of **correct** code for this rule:

```jsx
class MyComponent extends React.Component {
  static propTypes = {};
  componentWillMount() {}
  static getDerivedStateFromProps() {}
  render() {
    return null;
  }
}
```

```jsx
import PropTypes from 'prop-types';
class MyComponent extends React.Component {}
MyComponent.propTypes = {
  a: PropTypes.number,
  b: PropTypes.string.isRequired,
};
```

## Original Documentation

- [eslint-plugin-react / no-typos](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-typos.md)
