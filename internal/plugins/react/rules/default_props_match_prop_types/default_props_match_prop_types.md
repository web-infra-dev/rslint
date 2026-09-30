# default-props-match-prop-types

Ensures that each default prop has a corresponding, non-required prop declaration.

## Rule Details

A default for an undeclared prop can indicate a misspelling or an incomplete refactor. A default for a required prop can indicate that the prop should be optional.

The rule checks `defaultProps` assignments, static class fields and getters, and `getDefaultProps` in `createReactClass` components. Prop declarations can use runtime `propTypes` or TypeScript annotations.

Examples of **incorrect** code for this rule:

```jsx
function Greeting({ name }) {
  return <h1>Hello {name}</h1>;
}
Greeting.propTypes = { name: PropTypes.string.isRequired };
Greeting.defaultProps = { name: 'World', title: 'Hello' };
```

Examples of **correct** code for this rule:

```jsx
function Greeting({ name }) {
  return <h1>Hello {name}</h1>;
}
Greeting.propTypes = { name: PropTypes.string };
Greeting.defaultProps = { name: 'World' };
```

```tsx
function Greeting({ name }: { name?: string }) {
  return <h1>Hello {name}</h1>;
}
Greeting.defaultProps = { name: 'World' };
```

Components without any known prop declarations are ignored. Default objects containing spreads or unresolved external defaults are also ignored.

## Options

### `allowRequiredDefaults`

A boolean, defaulting to `false`. When `true`, required props may have defaults; defaults for undeclared props are still reported.

```json
{ "react/default-props-match-prop-types": ["error", { "allowRequiredDefaults": true }] }
```

```jsx
function Greeting({ name }) {
  return <h1>Hello {name}</h1>;
}
Greeting.propTypes = { name: PropTypes.string.isRequired };
Greeting.defaultProps = { name: 'World' };
```

## Differences from ESLint

- Quoted, numeric, and static computed keys compare by their value, including `C['defaultProps']` and `C.defaultProps['name']`; ESLint may skip them or report `undefined`.
- Parenthesized and TypeScript-wrapped receivers are recognized, including `(C as any).defaultProps`; ESLint ignores TypeScript-wrapped receivers.
- Typed props parameters may have any name, and type aliases follow lexical scope; ESLint limits parameter detection and searches top-level type declarations.
- Ambient `React.FC` declarations are recognized consistently with and without a typed Program; ESLint ignores them without a local React import.
- Configured React pragma aliases contribute `forwardRef` generic props, while ESLint only recognizes the literal `React.forwardRef` receiver.
- Class fields follow React's instance/static contract: typed `props` must be an instance field, while runtime `propTypes` must be static.
- Unknown prop declarations, including spreads and imported types, suppress missing-type reports. Known required props remain checked; ESLint can report props hidden by an unknown declaration as missing.
- Dynamic default keys are ignored when their names cannot be determined, instead of being compared as source text.

## Original Documentation

- [eslint-plugin-react: default-props-match-prop-types](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/default-props-match-prop-types.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/default-props-match-prop-types.js)
