# jsx-tag-spacing

Enforce whitespace around JSX tag brackets and slashes. This rule supports automatic fixes.

## Rule details

This rule checks the opening and closing tags of JSX elements, including self-closing tags. It does not check fragment brackets (`<>` and `</>`).

Examples of **incorrect** code with the default options:

```jsx
< App/>
<App / >
```

Examples of **correct** code with the default options:

```jsx
<App />
<App value={value}></App>
<App
  value={value}
/>
```

## Options

The rule accepts one object. Set any option to `"allow"` to disable that check.

| Option              | Default    | Accepted values                                          |
| ------------------- | ---------- | -------------------------------------------------------- |
| `closingSlash`      | `"never"`  | `"always"`, `"never"`, `"allow"`                            |
| `beforeSelfClosing` | `"always"` | `"always"`, `"never"`, `"proportional-always"`, `"allow"`     |
| `afterOpening`      | `"never"`  | `"always"`, `"never"`, `"allow-multiline"`, `"allow"`         |
| `beforeClosing`     | `"allow"`  | `"always"`, `"never"`, `"proportional-always"`, `"allow"`     |

- `closingSlash` controls whitespace inside `</` and `/>`.
- `beforeSelfClosing` controls whitespace before `/>`. A bracket on a separate line is allowed. `"proportional-always"` requires a space for single-line tags and a newline before the slash for multiline tags.
- `afterOpening` controls whitespace after `<` or `</`. `"allow-multiline"` permits a line break while forbidding spaces on the same line.
- `beforeClosing` controls whitespace before `>` in opening and closing tags. It does not check self-closing tags. `"proportional-always"` requires the bracket on a separate line for multiline tags.

To require a separate closing line for multiline tags:

```json
{
  "react/jsx-tag-spacing": [
    "error",
    {
      "beforeSelfClosing": "proportional-always",
      "beforeClosing": "proportional-always"
    }
  ]
}
```

## Differences from upstream

Rslint does not support closing tags written with whitespace between `<` and `/`, such as `<App>< /App>`. Use `</App>` instead. Avoid `closingSlash: "always"` for paired tags: its automatic fix produces `< /App>`, which rslint cannot lint.

## Original documentation

- [eslint-plugin-react: jsx-tag-spacing](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-tag-spacing.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-tag-spacing.js)
