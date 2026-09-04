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

## Differences from ESLint

The autofix preserves valid syntax in two cases where
eslint-plugin-react v7.37.5 produces malformed text:

- For default, rest, or destructured parameters, rslint preserves the
  parameter syntax in the fix, while ESLint emits empty entries such as `render(, )`.
- For computed keys such as `[render]`, rslint preserves the brackets in the
  fix, while ESLint removes them and emits malformed code such as `[render() ...`.

The fixer also preserves async and TypeScript generic arrow-function
signatures, which avoids the upstream fix dropping those semantics. For class
fields with `readonly` or `accessor` modifiers, the rule reports the problem
but does not offer an autofix because those modifiers are incompatible with
methods.

## Original Documentation

- [eslint-plugin-react: no-arrow-function-lifecycle](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/no-arrow-function-lifecycle.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-arrow-function-lifecycle.js)
