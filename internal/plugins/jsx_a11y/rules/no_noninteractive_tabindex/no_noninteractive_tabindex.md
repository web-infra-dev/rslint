# no-noninteractive-tabindex

## Rule Details

Tab key navigation should be limited to elements on the page that can be
interacted with. A `tabIndex` of `0` (or any positive value) on a
non-interactive HTML element forces it into the keyboard tab order even
though there is nothing for the user to do once focus arrives — disrupting
expectations for both sighted keyboard users and assistive-technology users.

The rule fires on a JSX opening element when **all** of the following hold:

- The element has a `tabIndex` prop whose value resolves to a usable
  non-negative integer (`tabIndex={0}`, `tabIndex="0"`, `tabIndex={5}`, …).
- The resolved element name is in the HTML DOM set (custom React components
  are skipped — the rule does not know what low-level element they render).
- The element is not inherently interactive (e.g. `<button>`, `<a href>`,
  `<input>`, `<textarea>`, `<select>`).
- The `role` attribute, when present and statically a literal string, is not
  in the interactive (widget) ARIA role set.

A `tabIndex` of `-1` (negative) is always allowed — that is the standard
"focusable via JavaScript but not via Tab" pattern.

Examples of **incorrect** code for this rule:

```jsx
<div tabIndex="0" />
<article tabIndex="0" />
<article tabIndex={0} />
<div role="article" tabIndex="0" />
```

Examples of **correct** code for this rule:

```jsx
<div />
<MyButton tabIndex={0} />
<button />
<button tabIndex="0" />
<button tabIndex={0} />
<div tabIndex="-1" />
<div role="button" tabIndex="0" />
<article tabIndex="-1" />
```

## Rule Options

### `tags`

Type: `string[]`. Default: not set.

A list of element names that should be exempt from the rule. Useful when a
codebase has a vetted convention for putting `tabIndex` on a specific
non-interactive tag.

Examples of **correct** code with `{ "tags": ["div"] }`:

```json
{ "jsx-a11y/no-noninteractive-tabindex": ["error", { "tags": ["div"] }] }
```

```jsx
<div tabIndex="0" />
```

### `roles`

Type: `string[]`. Default: not set. The upstream `recommended` preset sets
this to `["tabpanel"]`.

A list of literal ARIA `role` values that should be exempt — the rule
short-circuits when the element's `role` attribute resolves to one of these
strings.

Examples of **correct** code with `{ "roles": ["tabpanel"] }`:

```json
{ "jsx-a11y/no-noninteractive-tabindex": ["error", { "roles": ["tabpanel"] }] }
```

```jsx
<div role="tabpanel" tabIndex="0" />
```

### `allowExpressionValues`

Type: `boolean`. Default: `false`. The upstream `recommended` preset sets
this to `true`.

When `true`, an element whose `role` attribute is a non-literal expression
(e.g. `role={ROLE_BUTTON}`, `role={isButton ? "button" : "link"}`) is
exempt — the rule cannot statically determine whether the role is
interactive, and the option opts into trusting the developer.

Examples of **correct** code with `{ "allowExpressionValues": true }`:

```json
{ "jsx-a11y/no-noninteractive-tabindex": ["error", { "allowExpressionValues": true }] }
```

```jsx
<div role={ROLE_BUTTON} onClick={() => {}} tabIndex="0" />
<div role={isButton ? "button" : "link"} onClick={() => {}} tabIndex="0" />
```

## Resources

- [WCAG 2.1.1 — Keyboard](https://www.w3.org/WAI/WCAG21/Understanding/keyboard)
- [MDN — `tabindex`](https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/tabindex)
- [WAI-ARIA — Widget Roles](https://www.w3.org/TR/wai-aria-1.2/#widget_roles)

## Original Documentation

- [eslint-plugin-jsx-a11y/no-noninteractive-tabindex](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/no-noninteractive-tabindex.md)
