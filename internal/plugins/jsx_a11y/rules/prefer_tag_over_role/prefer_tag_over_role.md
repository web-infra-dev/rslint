# prefer-tag-over-role

## Rule Details

This rule enforces that an explicit ARIA
[`role`](https://www.w3.org/TR/wai-aria-1.2/#role_definitions) is not used in
place of a semantic HTML element that already provides that role natively.
Native elements expose their semantics to assistive technology automatically
and come with built-in keyboard, focus, and form behavior — re-declaring the
role on a generic container loses those affordances and adds another way for
the implementation to diverge from the announced role.

The rule fires on a JSX opening element when **all** of the following hold:

- The element carries a `role` attribute whose value resolves to a non-empty
  string.
- The LAST whitespace-separated token of that string is a non-abstract ARIA
  role with one or more semantic HTML element mappings (`heading` → `<h1>` …
  `<h6>`, `checkbox` → `<input type="checkbox">`, `link` → `<a href=...>` /
  `<area href=...>`, etc.).
- The element's effective tag name (after resolving
  `settings['jsx-a11y'].polymorphicPropName` and
  `settings['jsx-a11y'].components`) is NOT one of those semantic HTML
  elements.

Non-literal role values (Identifiers, CallExpressions, template literals
with substitutions, etc.) cannot be statically determined to a string and
never trigger the rule. The `role` attribute name is matched
case-insensitively, matching `getProp`'s default. Multi-token role values
(`role="button checkbox"`) are matched on the LAST token only —
`role="button checkbox"` is treated as `role="checkbox"` for the purpose of
this rule. Whitespace inside a role string is interpreted as the ASCII space
character `U+0020`; tabs and other Unicode whitespace are part of the token.

Examples of **incorrect** code for this rule:

```jsx
<div role="checkbox" />
<div role="heading" />
<div role="link" />
<div role="rowgroup" />
<div role="banner" />
<span role="checkbox" />
<other role="checkbox" />
<div role="button checkbox" />
```

Examples of **correct** code for this rule:

```jsx
<div />
<div role="unknown" />
<other />
<img role="img" />
<input role="checkbox" />
```

## Accessibility guidelines

- [WAI-ARIA Roles model](https://www.w3.org/TR/wai-aria-1.2/#role_definitions)
- [Using ARIA — Rule 1: If you can use a native HTML element](https://www.w3.org/TR/using-aria/#rule1)

### Resources

- [MDN — WAI-ARIA Roles](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Roles)
- [aria-query `roleElements` map](https://github.com/A11yance/aria-query)

## Original Documentation

- [eslint-plugin-jsx-a11y/prefer-tag-over-role](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/prefer-tag-over-role.md)
