# role-supports-aria-props

## Rule Details

Elements with ARIA roles must only use `aria-*` properties that the role supports. Many DOM elements carry an implicit ARIA role (e.g. `<a href="#" />` → `link`, `<input type="checkbox" />` → `checkbox`); elements with an explicit `role` attribute carry that role directly. The supported-properties table is sourced from [`aria-query`](https://github.com/A11yance/aria-query)'s `rolesMap`.

The rule fires on a JSX opening element. It resolves the effective element name (through `settings['jsx-a11y'].polymorphicPropName` and `settings['jsx-a11y'].components`), looks up the explicit `role` attribute, and falls back to the implicit role table when no explicit `role` is set. It then walks every JSX attribute (excluding spreads and attributes whose value is `null`/`undefined`) and reports any `aria-*` attribute that is not in the role's supported-properties set.

Notable behavior locked in to mirror upstream:

- Membership against `rolesMap` is case-sensitive — `role="BUTTON"` is silently not validated because `aria-query`'s keys are lowercase.
- Mixed-case ARIA prop names (e.g. `aria-Checked`) are silently not validated for the same reason.
- Spread attributes are opaque, even when the spread argument is a literal object containing `aria-*` keys.
- Some elements have a context-sensitive implicit role: `<a>` / `<area>` / `<link>` only acquire `link` when an `href` attribute is present; `<img>` loses its `img` role when `alt=""` or the literal `src` contains `.svg`; `<input>` / `<menu>` / `<menuitem>` / `<select>` depend on `type` / `multiple` / `size` values.

Examples of **incorrect** code for this rule:

```jsx
<a href="#" aria-checked />
<area href="#" aria-checked />
<link href="#" aria-checked />
<menu type="toolbar" aria-haspopup />
<input type="radio" aria-invalid />
<input type="checkbox" aria-haspopup />
<aside aria-checked />
<ul aria-expanded />
<details aria-expanded />
<dialog aria-expanded />
<article aria-expanded />
<body aria-expanded />
<li aria-expanded />
<nav aria-expanded />
<ol aria-expanded />
<output aria-expanded />
<section aria-expanded />
<tbody aria-expanded />
<tfoot aria-expanded />
<thead aria-expanded />
<div role="link" aria-checked />
```

Examples of **correct** code for this rule:

```jsx
<Foo bar />
<div />
<div role="presentation" {...props} />
<a href="#" aria-expanded />
<area href="#" aria-expanded />
<link href="#" aria-expanded />
<menu type="toolbar" aria-activedescendant />
<input type="checkbox" aria-checked />
<input type="radio" aria-checked />
<input type="range" aria-valuemax />
<button aria-pressed />
<form aria-hidden />
<h1 aria-hidden />
<select aria-expanded />
<datalist aria-expanded />
<div role="heading" aria-level />
<h2 role="presentation" aria-level={null} />
<h2 role="presentation" aria-level={undefined} />
```

## Rule Options

This rule takes no options.

## Accessibility guidelines

- [WCAG 4.1.2](https://www.w3.org/WAI/WCAG21/Understanding/name-role-value)

### Resources

- [ARIA Spec — States and Properties](https://www.w3.org/TR/wai-aria/#states_and_properties)
- [ARIA Spec — Supported States and Properties](https://www.w3.org/TR/wai-aria/#supportedState)
- [Chrome Audit Rules, AX_ARIA_10](https://github.com/GoogleChrome/accessibility-developer-tools/wiki/Audit-Rules#ax_aria_10)
- [aria-query — `rolesMap`](https://github.com/A11yance/aria-query/tree/HEAD/src/etc/roles)

## Original Documentation

- [eslint-plugin-jsx-a11y/role-supports-aria-props](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/role-supports-aria-props.md)
