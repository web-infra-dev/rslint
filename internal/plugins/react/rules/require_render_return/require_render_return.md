# react/require-render-return

Enforce ES5 or ES6 class for returning value in render function.

When writing the `render` method in a component it is easy to forget to return
the JSX content. This rule will warn if the `return` statement is missing.

## Rule Details

This rule applies to components written as either:

- An ES5-style `createReactClass({ ... })` (or `<pragma>.createReactClass({ ... })`) object.
- An ES6 class extending `Component` / `PureComponent` (bare or qualified by the React pragma).

For each such component it inspects every member whose *Identifier* key is
`render`:

- A class method `render() { ... }` or an object shorthand `render() { ... }`.
- A property assignment `render: function () { ... }` or `render: () => { ... }`.
- A class field `render = function () { ... }` or `render = () => { ... }`.
- A getter or setter `get render() { ... }` / `set render(v) { ... }`.

Non-Identifier keys are *not* considered `render`:

- String-literal keys (`"render"() { ... }` / `"render": function () { ... }`)
- Computed keys (`['render']() { ... }`, `` [`render`]() { ... } ``, `[tag`render`]() { ... }`)
- Numeric / BigInt keys

This mirrors upstream `astUtil.getPropertyName`, which reads only
`nameNode.name` — so anything that isn't a plain identifier returns `""` and
fails the `render` match.

A `render` whose value is anything other than a FunctionExpression /
ArrowFunction (or a shorthand method / accessor, which are themselves
function-like) is ignored.

The component is considered to return when *any* `render`-named member
satisfies one of:

1. It is an arrow function with an expression body (implicit return), e.g. `render = () => <div/>`.
2. Its block body contains a `return` statement that is not nested inside
   another function-like boundary (FunctionExpression, FunctionDeclaration,
   ArrowFunction, MethodDeclaration, accessor, constructor).

Upstream's depth regex `/Function(Expression|Declaration)$/` is unanchored, so
`ArrowFunctionExpression` is also captured — meaning a `return` inside a
nested arrow (IIFE or otherwise) inside `render()` does *not* count as
render's own return. This rule follows that behavior exactly.

When multiple members are named `render` (e.g. `static render` alongside an
instance `render`), the component passes as soon as any one of them has a
qualifying return. If none do, the first matching member is reported.

Examples of **incorrect** code for this rule:

```jsx
var Hello = createReactClass({
  render() {
    <div>Hello</div>;
  },
});
```

```jsx
class Hello extends React.Component {
  render() {
    <div>Hello</div>;
  }
}
```

Examples of **correct** code for this rule:

```jsx
var Hello = createReactClass({
  render() {
    return <div>Hello</div>;
  },
});
```

```jsx
class Hello extends React.Component {
  render() {
    return <div>Hello</div>;
  }
}
```

## Original Documentation

- [eslint-plugin-react / require-render-return](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/require-render-return.md)
