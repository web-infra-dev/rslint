# require-default-props

## Rule Details

Require a default value for every optional component prop. The rule checks runtime
`propTypes` declarations and locally declared TypeScript prop types.

Examples of **incorrect** code:

```jsx
function Greeting({ name }) {
  return <h1>Hello, {name}</h1>;
}
Greeting.propTypes = { name: PropTypes.string };
```

Examples of **correct** code:

```jsx
function Greeting({ name }) {
  return <h1>Hello, {name}</h1>;
}
Greeting.propTypes = { name: PropTypes.string };
Greeting.defaultProps = { name: 'world' };
```

Required props do not need defaults. An external `defaultProps` assignment that
cannot be resolved, or a spread in the defaults object, skips the component to
avoid reporting defaults supplied elsewhere.

## Options

| Option | Default | Description |
| --- | --- | --- |
| `forbidDefaultForRequired` | `false` | Report defaults supplied for required props. |
| `classes` | `"defaultProps"` | Use `defaultProps`, or set to `"ignore"` to skip classes. |
| `functions` | `"defaultProps"` | Use `defaultProps`, `"defaultArguments"`, or `"ignore"`. |
| `ignoreFunctionalComponents` | `false` | Deprecated alias for `functions: "ignore"`; takes precedence when true. |

### Default arguments

For function components using default parameters, configure:

```json
{ "react/require-default-props": ["error", { "functions": "defaultArguments" }] }
```

Examples of **incorrect** code with this option:

```tsx
function Greeting({ name }: { name?: string }) {
  return <h1>Hello, {name}</h1>;
}
```

Examples of **correct** code with this option:

```tsx
function Greeting({ name = 'world' }: { name?: string }) {
  return <h1>Hello, {name}</h1>;
}
```

This mode reports `defaultProps` on function components and requires optional
props to have defaults in the destructured parameter. It also reports defaults
for required destructured props, independently of `forbidDefaultForRequired`.

### Required props

```json
{ "react/require-default-props": ["error", { "forbidDefaultForRequired": true }] }
```

With this option, a prop declared as `PropTypes.string.isRequired` must not also
appear in `defaultProps`.

## When Not To Use It

Disable this rule when optional props may intentionally remain `undefined`.

## Differences from ESLint

- `React.memo` and `React.forwardRef` components honor `functions: "ignore"` and
  `"defaultArguments"`; eslint-plugin-react v7.37.5 treats their wrappers separately.
- Optional props named `toString`, `constructor`, or other inherited object names
  need explicit defaults. Upstream can mistake inherited properties for defaults.
- Inline class props types and props types declared inside functions are checked;
  upstream can skip these optional props.
- Reading `Component.defaultProps` does not suppress missing-default diagnostics;
  upstream can treat that read as an unresolved default declaration.
- Multiple component values assigned to the same binding produce one diagnostic
  per prop contract. A later component is still checked after a non-component value.

## Original Documentation

- [eslint-plugin-react: require-default-props](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/require-default-props.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/require-default-props.js)
