# jsx-closing-tag-location

## Rule Details

Enforce the closing tag location for multiline JSX elements. The closing tag should match the indentation of the opening tag.

Examples of **incorrect** code for this rule:

```jsx
<Foo>
  bar
    </Foo>
```

Examples of **correct** code for this rule:

```jsx
<Foo>bar</Foo>
<Foo>
  bar
</Foo>
```

## Options

- `"tag-aligned"` (default): Closing tag must be aligned with the opening tag's column position.
- `"line-aligned"`: Closing tag must be aligned with the indentation of the line containing the opening tag (useful when the opening tag is not at the start of the line, e.g., `var x = <div>`).

## Original Documentation

- [eslint-plugin-react: jsx-closing-tag-location](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-closing-tag-location.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-closing-tag-location.js)
