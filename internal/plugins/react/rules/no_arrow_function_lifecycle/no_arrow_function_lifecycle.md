# no-arrow-function-lifecycle

Disallow arrow functions for React lifecycle methods.

Using arrow functions for lifecycle methods makes components harder to test,
can interfere with hot reloading, and provides no practical performance
benefit.

## Rule Details

Examples of **incorrect** code for this rule:

```jsx
class Hello extends React.Component {
  render = () => {
    return <div />;
  }
}
```

```jsx
var AnotherHello = createReactClass({
  render: () => {
    return <div />;
  },
});
```

Examples of **correct** code for this rule:

```jsx
class Hello extends React.Component {
  render() {
    return <div />;
  }
}
```

```jsx
var AnotherHello = createReactClass({
  render() {
    return <div />;
  },
});
```

The rule also checks the static `getDerivedStateFromProps` lifecycle method
when it is declared as an arrow-function class field.

The rule is automatically fixable by the `--fix` CLI option.

## When Not To Use It

If you do not want to enforce prototype placement for React lifecycle methods,
you can disable this rule.

## Original Documentation

- [eslint-plugin-react: no-arrow-function-lifecycle](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/no-arrow-function-lifecycle.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-arrow-function-lifecycle.js)
