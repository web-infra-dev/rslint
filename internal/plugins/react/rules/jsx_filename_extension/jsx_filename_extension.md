# react/jsx-filename-extension

## Rule Details

Restrict file extensions that may contain JSX. By default, only `.jsx` files may contain JSX.

Examples of **incorrect** code for this rule:

```javascript
// In file: Component.js
var Hello = <div>Hello</div>;
```

Examples of **correct** code for this rule:

```jsx
// In file: Component.jsx
var Hello = <div>Hello</div>;
```

## Options

- `extensions`: An array of allowed file extensions (default: `[".jsx"]`).
- `allow`: `"always"` (default) or `"as-needed"`.
  - `"always"`: Allow the specified extensions whether or not they contain JSX.
  - `"as-needed"`: Only allow the specified extensions when the file actually contains JSX.
- `ignoreFilesWithoutCode`: When `true`, files with no code statements are ignored in `"as-needed"` mode.

## Original Documentation

- [react/jsx-filename-extension](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-filename-extension.md)
