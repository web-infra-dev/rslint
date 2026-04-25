# react/jsx-curly-brace-presence

## Rule Details

Enforce curly braces or disallow unnecessary curly braces in JSX props and/or
children.

By default, the rule warns about unnecessary curly braces in both JSX props
and children. Prop values that are JSX elements are ignored by default.

Examples of **incorrect** code for this rule:

```jsx
<App prop={'foo'} attr={"bar"}>{'Hello world'}</App>;
```

Examples of **correct** code for this rule:

```jsx
<App prop="foo" attr="bar">Hello world</App>;
```

Examples of **incorrect** code for this rule with `{ "props": "always", "children": "always" }`:

```json
{ "react/jsx-curly-brace-presence": ["error", { "props": "always", "children": "always" }] }
```

```jsx
<App>Hello world</App>;
<App prop='Hello world'>{'Hello world'}</App>;
```

Examples of **incorrect** code for this rule with `{ "props": "always", "children": "always", "propElementValues": "always" }`:

```json
{ "react/jsx-curly-brace-presence": ["error", { "props": "always", "children": "always", "propElementValues": "always" }] }
```

```jsx
<App prop=<div /> />;
```

Examples of **incorrect** code for this rule with `{ "props": "never", "children": "never", "propElementValues": "never" }`:

```json
{ "react/jsx-curly-brace-presence": ["error", { "props": "never", "children": "never", "propElementValues": "never" }] }
```

```jsx
<App prop={<div />} />;
```

## Options

The rule accepts either an options object or a single string shorthand:

```json
{ "react/jsx-curly-brace-presence": ["error", { "props": "never", "children": "never", "propElementValues": "ignore" }] }
```

```json
{ "react/jsx-curly-brace-presence": ["error", "never"] }
```

Each of `props`, `children`, and `propElementValues` accepts:

- `"always"` — enforce curly braces.
- `"never"` — disallow unnecessary curly braces.
- `"ignore"` — disable the check.

The string shorthand sets `props` and `children` to the given value;
`propElementValues` stays at its default of `"ignore"`.

## Original Documentation

- https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-curly-brace-presence.md
