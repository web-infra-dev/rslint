# aria-unsupported-elements

Certain HTML elements do not support ARIA roles, states, or properties. Adding
ARIA attributes (any `aria-*` attribute) or a `role` attribute to these elements
is invalid.

The reserved elements are: `base`, `col`, `colgroup`, `head`, `html`, `link`,
`meta`, `noembed`, `noscript`, `param`, `picture`, `script`, `source`, `style`,
`title`, `track`.

## Rule Details

This rule enforces that the reserved DOM elements do not contain `role` or
ARIA attributes. Attribute names are matched case-insensitively.

Examples of **incorrect** code for this rule:

```jsx
<base aria-hidden />
<meta charset="UTF-8" aria-hidden="false" />
<html aria-required />
<style role="presentation"></style>
<script aria-label="x"></script>
```

Examples of **correct** code for this rule:

```jsx
<base />
<meta charset="UTF-8" />
<html lang="en" />
<style></style>
<div aria-hidden />
<span role="button" />
```

## Original Documentation

- [eslint-plugin-jsx-a11y/aria-unsupported-elements](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/aria-unsupported-elements.md)
