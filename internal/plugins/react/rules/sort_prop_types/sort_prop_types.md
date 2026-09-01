# sort-prop-types

## Rule Details

Enforce a consistent order for React `propTypes` declarations. By default,
property names are compared case-sensitively in alphabetical order. A spread
property starts a new ordering group, so properties on either side of a spread
are not compared with each other.

Examples of **incorrect** code for this rule:

```jsx
Component.propTypes = {
  zebra: PropTypes.string,
  alpha: PropTypes.string,
};
```

Examples of **correct** code for this rule:

```jsx
Component.propTypes = {
  alpha: PropTypes.string,
  zebra: PropTypes.string,
};
```

The rule checks assignments, class properties, object-literal `propTypes`
properties, and configured prop-wrapper functions. It can also inspect
`PropTypes.shape` declarations and TypeScript props declared on the first
parameter of a function component.

## Options

All options are `false` by default.

### `requiredFirst`

Require declarations whose validator ends in `.isRequired` to appear before
optional declarations.

Examples of **incorrect** code with `{ "requiredFirst": true }`:

```json
{ "react/sort-prop-types": ["error", { "requiredFirst": true }] }
```

```jsx
Component.propTypes = {
  label: PropTypes.string,
  id: PropTypes.string.isRequired,
};
```

Examples of **correct** code with `{ "requiredFirst": true }`:

```jsx
Component.propTypes = {
  id: PropTypes.string.isRequired,
  label: PropTypes.string,
};
```

### `callbacksLast`

Require callback props whose names begin with `on` followed by an uppercase
ASCII letter to appear after other props. Callback props remain alphabetically
sorted among themselves unless `noSortAlphabetically` is enabled.

Examples of **incorrect** code with `{ "callbacksLast": true }`:

```json
{ "react/sort-prop-types": ["error", { "callbacksLast": true }] }
```

```jsx
Component.propTypes = {
  onChange: PropTypes.func,
  name: PropTypes.string,
};
```

Examples of **correct** code with `{ "callbacksLast": true }`:

```jsx
Component.propTypes = {
  name: PropTypes.string,
  onChange: PropTypes.func,
};
```

### `ignoreCase`

Compare property names without regard to case.

Examples of **incorrect** code with `{ "ignoreCase": true }`:

```json
{ "react/sort-prop-types": ["error", { "ignoreCase": true }] }
```

```jsx
Component.propTypes = {
  Zeta: PropTypes.string,
  alpha: PropTypes.string,
};
```

### `noSortAlphabetically`

Disable alphabetical ordering. This is useful when only `requiredFirst` or
`callbacksLast` should determine the order.

```json
{
  "react/sort-prop-types": [
    "error",
    { "callbacksLast": true, "noSortAlphabetically": true }
  ]
}
```

```jsx
Component.propTypes = {
  title: PropTypes.string,
  id: PropTypes.string,
  onSubmit: PropTypes.func,
  onChange: PropTypes.func,
};
```

### `sortShapeProp`

Apply the same ordering rules to the object passed to a `shape` call. This also
works when the shape object is first assigned to an identifier.

Examples of **incorrect** code with `{ "sortShapeProp": true }`:

```json
{ "react/sort-prop-types": ["error", { "sortShapeProp": true }] }
```

```jsx
Component.propTypes = {
  address: PropTypes.shape({
    zip: PropTypes.string,
    city: PropTypes.string,
  }),
};
```

Examples of **correct** code with `{ "sortShapeProp": true }`:

```jsx
Component.propTypes = {
  address: PropTypes.shape({
    city: PropTypes.string,
    zip: PropTypes.string,
  }),
};
```

### `checkTypes`

Check TypeScript props declared as an inline type literal or a type alias on a
function declaration or arrow function's first parameter.

Examples of **incorrect** code with `{ "checkTypes": true }`:

```json
{ "react/sort-prop-types": ["error", { "checkTypes": true }] }
```

```tsx
type Props = {
  zIndex: number;
  ariaLabel: string;
};

const Component = (props: Props) => null;
```

Examples of **correct** code with `{ "checkTypes": true }`:

```tsx
const Component = (props: { ariaLabel: string; zIndex: number }) => null;
```

## Differences from ESLint

- `eslint-plugin-react` can automatically reorder declarations with `--fix`.
  rslint reports the same ordering problems but does not currently apply a fix.

## Original Documentation

- [eslint-plugin-react: sort-prop-types](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/sort-prop-types.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/sort-prop-types.js)
