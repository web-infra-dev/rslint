# jsx-curly-newline

Enforce consistent line breaks immediately inside curly braces in JSX
attributes and expressions.

## Rule Details

This rule checks both braces of every JSX expression container. It can require,
forbid, or keep consistent the newline directly after `{` and directly before
`}`. Comments are preserved: a diagnostic adjacent to a comment may not be
autofixable.

Examples of **incorrect** code with the default `"consistent"` option:

```jsx
<div>{ foo
}</div>

<div>{
  foo }</div>
```

Examples of **correct** code:

```jsx
<div>{ foo }</div>

<div>{
  foo
}</div>
```

## Rule Options

The default is `"consistent"`, an alias for:

```json
{ "singleline": "consistent", "multiline": "consistent" }
```

`"never"` is an alias for `{ "singleline": "forbid", "multiline":
"forbid" }`. Alternatively, use an object with independently configured
`singleline` and `multiline` values: `"consistent"`, `"require"`, or
`"forbid"`.

```json
{ "react/jsx-curly-newline": ["error", { "singleline": "require", "multiline": "forbid" }] }
```

With `"require"`, both immediate brace boundaries must have newlines. With
`"forbid"`, neither boundary may have one. With `"consistent"`, the right
boundary follows the existing state of the left boundary.

## Original Documentation

- [eslint-plugin-react: jsx-curly-newline](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-curly-newline.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-curly-newline.js)
