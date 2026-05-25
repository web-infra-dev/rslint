# no-multi-comp

Disallow multiple component definition per file.

Declaring only one component per file improves readability and reusability of components.

## Rule Details

Examples of **incorrect** code for this rule:

```jsx
var Hello = createReactClass({
  render: function () {
    return <div>Hello {this.props.name}</div>;
  },
});

var HelloJohn = createReactClass({
  render: function () {
    return <Hello name="John" />;
  },
});
```

Examples of **correct** code for this rule:

```jsx
var Hello = require('./components/Hello');

var HelloJohn = createReactClass({
  render: function () {
    return <Hello name="John" />;
  },
});
```

## Rule Options

```json
{ "react/no-multi-comp": ["error", { "ignoreStateless": false }] }
```

### `ignoreStateless`

When `true` the rule will ignore stateless components and will allow you to have multiple stateless components, or one stateful component and some stateless components in the same file.

Examples of **correct** code for this rule with `{ "ignoreStateless": true }`:

```json
{ "react/no-multi-comp": ["error", { "ignoreStateless": true }] }
```

```jsx
function Hello(props) {
  return <div>Hello {props.name}</div>;
}
function HelloAgain(props) {
  return <div>Hello again {props.name}</div>;
}
```

```json
{ "react/no-multi-comp": ["error", { "ignoreStateless": true }] }
```

```jsx
function Hello(props) {
  return <div>Hello {props.name}</div>;
}
class HelloJohn extends React.Component {
  render() {
    return <Hello name="John" />;
  }
}
module.exports = HelloJohn;
```

## When Not To Use It

If you prefer to declare multiple components per file you can disable this rule.

## Differences from ESLint

- Classes recognized only via JSDoc tags (`@extends React.Component` /
  `@augments React.PureComponent`) without an `extends` clause are not treated
  as React components. Add an explicit `extends` clause to align both linters.

## Original Documentation

- [eslint-plugin-react / no-multi-comp](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-multi-comp.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/lib/rules/no-multi-comp.js)
