# forbid-elements

You may want to forbid usage of certain elements in favor of others, (e.g. forbid all `<div />` and use `<Box />` instead). This rule allows you to configure a list of forbidden elements and to specify their desired replacements.

## Rule Details

This rule checks all JSX elements and `React.createElement` calls and verifies that no forbidden elements are used. This rule is off by default. If on, no elements are forbidden by default.

## Rule Options

```json
{ "react/forbid-elements": ["error", { "forbid": ["button"] }] }
```

### `forbid`

An array of strings and/or objects. An object in this array may have the following properties:

- `element` (required): the name of the forbidden element (e.g. `'button'`, `'Modal'`)
- `message`: additional message that gets reported

A string item in the array is a shorthand for `{ element: string }`.

Examples of **correct** code for this rule:

```json
{ "react/forbid-elements": ["error", { "forbid": ["button"] }] }
```

```jsx
<Button />
```

```json
{ "react/forbid-elements": ["error", { "forbid": [{ "element": "button" }] }] }
```

```jsx
<Button />
```

Examples of **incorrect** code for this rule:

```json
{ "react/forbid-elements": ["error", { "forbid": ["button"] }] }
```

```jsx
<button />;
React.createElement('button');
```

```json
{ "react/forbid-elements": ["error", { "forbid": ["Modal"] }] }
```

```jsx
<Modal />;
React.createElement(Modal);
```

```json
{ "react/forbid-elements": ["error", { "forbid": ["Namespaced.Element"] }] }
```

```jsx
<Namespaced.Element />;
React.createElement(Namespaced.Element);
```

```json
{
  "react/forbid-elements": [
    "error",
    {
      "forbid": [
        { "element": "button", "message": "use <Button> instead" },
        "input"
      ]
    }
  ]
}
```

```jsx
<div>
  <button />
  <input />
</div>;
React.createElement(
  'div',
  {},
  React.createElement('button', {}, React.createElement('input')),
);
```

## When Not To Use It

If you don't want to forbid any elements.

## Original Documentation

- [eslint-plugin-react / forbid-elements](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/forbid-elements.md)
