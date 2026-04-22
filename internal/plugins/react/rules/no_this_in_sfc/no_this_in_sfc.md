# no-this-in-sfc

## Rule Details

Disallow `this` from being used inside stateless functional components (SFCs).
React supports two component styles: class components access instance state and
props through `this` (e.g. `this.props.foo`), and functional components receive
their props as the first argument. Reaching for `this.props` / `this.state`
inside a functional component is almost always a mistake — typically an
unfinished migration from a class component — because `this` is not bound to
the component instance.

The rule covers `this.x` and `this['x']` access (PropertyAccess and
ElementAccess forms), bracket access included, parens transparent. It targets
the same set of stateless components as eslint-plugin-react: capitalized
function declarations / expressions / arrows that return JSX or `null`,
`React.memo` / `React.forwardRef` (and their bare aliases) wrapped components,
and anonymous `export default function() { ... }` components.

A `this` access inside a class component (an ES6 class extending
`Component` / `PureComponent`) or inside a `createReactClass` ES5 component
is never reported — those bind `this` to the component instance.

A `this` access inside a function that happens to be a property value of an
object literal (e.g. `{ Foo() { return <div>{this.x}</div> } }`) is also not
reported, mirroring upstream's "Property" carve-out.

Examples of **incorrect** code for this rule:

```javascript
function Foo(props) {
  const { foo } = this.props;
  return <div>{foo}</div>;
}
```

```javascript
function Foo(props) {
  return <div>{this.state.foo}</div>;
}
```

```javascript
const Foo = (props) => <span>{this.props.foo}</span>;
```

```javascript
const Foo = React.memo(() => <div>{this.props.foo}</div>);
```

Examples of **correct** code for this rule:

```javascript
function Foo(props) {
  const { foo } = props;
  return <div bar={foo} />;
}
```

```javascript
function Foo({ foo }) {
  return <div bar={foo} />;
}
```

```javascript
class Foo extends React.Component {
  render() {
    const { foo } = this.props;
    return <div bar={foo} />;
  }
}
```

```javascript
const Foo = createReactClass({
  render: function () {
    return <div>{this.props.foo}</div>;
  },
});
```

## Differences from ESLint

- A function wrapped in a custom higher-order component (configured via
  `settings.componentWrapperFunctions` in ESLint) is not treated as a
  stateless component. For example, given `wrap(() => <div>{this.props.x}</div>)`,
  ESLint reports the `this` access while rslint does not. Only `memo` and
  `forwardRef` wrappers (bare, or qualified by the configured React pragma)
  are recognized.

## Original Documentation

- [eslint-plugin-react `no-this-in-sfc`](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-this-in-sfc.md)
