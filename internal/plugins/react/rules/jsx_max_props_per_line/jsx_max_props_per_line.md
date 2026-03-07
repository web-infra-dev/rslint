# react/jsx-max-props-per-line

## Rule Details

Limit the maximum number of props on a single line in JSX. Useful for readability when elements have many props.

Examples of **incorrect** code with the default `{ "maximum": 1, "when": "always" }`:

```jsx
<Foo bar="a" baz="b" />
```

Examples of **correct** code with the default `{ "maximum": 1, "when": "always" }`:

```jsx
<Foo bar="a" />
<Foo
  bar="a"
  baz="b"
/>
```

## Options

- `maximum`: Maximum number of props per line. Can be a number or `{ single: N, multi: N }`.
- `when`: `"always"` (default) or `"multiline"`. If `"multiline"`, only enforces the limit on multiline JSX elements.

## Original Documentation

- [react/jsx-max-props-per-line](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-max-props-per-line.md)
