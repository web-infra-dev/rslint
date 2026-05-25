# jsx-no-leaked-render

## Rule Details

In React, conditionally rendering content with `&&` is a common pattern, but it can leak unintended values into the DOM when the left-hand side is falsy but not boolean. For example, `{count && <Something/>}` renders `0` as a literal `"0"` in React DOM and crashes React Native when `count` is `0`. The same applies to `NaN`, and in React 17 also to `''`.

Examples of **incorrect** code for this rule:

```jsx
const Component = ({ count, title }) => {
  return <div>{count && title}</div>;
};
```

```jsx
const Component = ({ elements }) => {
  return <div>{elements.length && <List elements={elements} />}</div>;
};
```

Examples of **correct** code for this rule:

```jsx
const Component = ({ elements }) => {
  return <div>{!!elements.length && <List elements={elements} />}</div>;
};
```

```jsx
const Component = ({ elements }) => {
  return <div>{elements.length > 0 && <List elements={elements} />}</div>;
};
```

```jsx
const Component = ({ elements }) => {
  return (
    <div>
      {elements.length ? <List elements={elements} /> : null}
    </div>
  );
};
```

## Options

This rule accepts an options object as its second argument.

### `validStrategies`

**Type:** `Array<"ternary" | "coerce">` &nbsp; **Default:** `["ternary", "coerce"]`

Which conversion strategies count as valid. The first entry of the array is the strategy used for autofix.

```json
{ "react/jsx-no-leaked-render": ["error", { "validStrategies": ["ternary"] }] }
```

```jsx
const Component = ({ count, title }) => {
  return <div>{count ? title : null}</div>;
};
```

```json
{ "react/jsx-no-leaked-render": ["error", { "validStrategies": ["coerce"] }] }
```

```jsx
const Component = ({ count, title }) => {
  return <div>{!!count && title}</div>;
};
```

### `ignoreAttributes`

**Type:** `boolean` &nbsp; **Default:** `false`

When `true`, JSX attribute values are exempt from the check; only the children of JSX elements are inspected. Useful when a downstream component accepts truthy-coerced props.

```json
{ "react/jsx-no-leaked-render": ["error", { "ignoreAttributes": true }] }
```

```jsx
const Component = ({ enabled, checked }) => {
  return <CheckBox checked={enabled && checked} />;
};
```

## Original Documentation

- ESLint plugin: [`react/jsx-no-leaked-render`](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-no-leaked-render.md)
