# no-object-type-as-default-prop

Disallow newly created reference values as default props in function components.

## Rule Details

An inline object, array, function, class, constructed value, regular expression,
JSX element, or `Symbol()` default creates a different value on each render.
Passing it to a memoized child or using it as a hook dependency can cause
unnecessary renders or a render loop. Use a stable reference instead.

The rule checks direct property defaults in the first destructured parameter of
a detected function component. It does not check nested properties, defaults for
the entire parameter, or `defaultProps` assignments. As in upstream, functions
passed directly to component wrappers such as `React.memo` and `React.forwardRef`
are not checked. JSX fragments and calls other than a direct `Symbol()` call are
outside the rule's checks.

Examples of **incorrect** code for this rule:

```jsx
function List({ items = [] }) {
  return <Widget items={items} />;
}
```

```jsx
const Panel = ({ options = {}, onChange = () => {} }) => (
  <Widget options={options} onChange={onChange} />
);
```

Examples of **correct** code for this rule:

```jsx
const emptyItems = [];

function List({ items = emptyItems, limit = 10 }) {
  return <Widget items={items} limit={limit} />;
}
```

```jsx
const emptyOptions = {};
const noop = () => {};

const Panel = ({ options = emptyOptions, onChange = noop }) => (
  <Widget options={options} onChange={onChange} />
);
```

## Options

This rule has no options.

## Original Documentation

- [eslint-plugin-react: no-object-type-as-default-prop](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/no-object-type-as-default-prop.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-object-type-as-default-prop.js)
