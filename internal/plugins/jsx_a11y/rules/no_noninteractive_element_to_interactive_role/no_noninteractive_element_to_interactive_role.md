# no-noninteractive-element-to-interactive-role

## Rule Details

Inherently non-interactive HTML elements such as `<address>`, `<article>`,
`<table>`, `<ul>`, headings, list items, and structural landmarks are
content containers — the browser supplies them no keyboard focus, no
activation behavior, and no assistive-technology widget semantics.
Promoting them to a widget by assigning an interactive ARIA role
(`button`, `checkbox`, `link`, `menuitem`, `tab`, `slider`, …) without
also wiring up the keyboard and focus contract those roles imply produces
a control that screen-reader and keyboard users cannot meaningfully
operate.

If the surrounding context calls for an interactive widget, use the
genuine HTML element (`<button>`, `<a href>`, `<input>`, …) so the
browser provides the affordance, or wrap the non-interactive content
inside a separate interactive container.

The rule fires on every `role` JSX attribute when **all** of the
following hold:

- The resolved element name is in the HTML DOM set (custom React
  components are skipped — the rule does not know what low-level element
  they render).
- The attribute name is literally `role` (case-sensitive); namespaced
  attributes such as `mynamespace:role` are not checked.
- The element + role pair does not match an entry in the per-element
  allow-list (see Rule Options below).
- The element is inherently non-interactive (e.g. `<article>`,
  `<address>`, `<h1>`, `<li>`, `<ul>`, `<table>`).
- The `role` attribute, when statically a literal string, resolves to an
  interactive (widget) role such as `button`, `checkbox`, `link`,
  `menuitem`, `tab`, or `radio`.

Examples of **incorrect** code for this rule:

```jsx
<article role="button" />
<h1 role="menuitem">Save</h1>
<ul role="menu" />
```

Examples of **correct** code for this rule:

```jsx
<article role="article" />
<h1 role="heading">Save</h1>
<button>Save</button>
<div role="button" />
```

## Rule Options

### Per-element allow-list

Type: `{ [tagName: string]: string[] }`. Default: not set. The upstream
`recommended` preset enables list / table allowances — `ul` and `ol`
accept `listbox` / `menu` / `menubar` / `radiogroup` / `tablist` /
`tree` / `treegrid`; `li` accepts `menuitem` / `menuitemcheckbox` /
`menuitemradio` / `option` / `row` / `tab` / `treeitem`; `table`
accepts `grid`; `td` accepts `gridcell`; `fieldset` accepts
`radiogroup` / `presentation`. The `strict` preset omits the allow-list
entirely.

Each key is an HTML element name and each value is an array of role
strings exempt from the rule for that element. Non-string entries are
silently ignored; non-array values cause the entire entry to be
dropped — only `string[]` allow-lists are honored.

Examples of **correct** code with `{ "ul": ["menu", "menubar"] }`:

```json
{ "jsx-a11y/no-noninteractive-element-to-interactive-role": ["error", { "ul": ["menu", "menubar"] }] }
```

```jsx
<ul role="menu" />
<ul role="menubar" />
```

Examples of **incorrect** code with `{ "ul": ["menu", "menubar"] }`:

```json
{ "jsx-a11y/no-noninteractive-element-to-interactive-role": ["error", { "ul": ["menu", "menubar"] }] }
```

```jsx
<ul role="tablist" />
```

## Resources

- [WCAG 4.1.2 — Name, Role, Value](https://www.w3.org/WAI/WCAG21/Understanding/name-role-value)
- [WAI-ARIA — Widget Roles](https://www.w3.org/TR/wai-aria-1.2/#widget_roles)
- [WAI-ARIA — Non-interactive Content Roles](https://www.w3.org/TR/wai-aria-1.2/#document_structure_roles)
- [MDN — Using ARIA: Roles, states, and properties](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Attributes)

## Original Documentation

- [eslint-plugin-jsx-a11y/no-noninteractive-element-to-interactive-role](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/no-noninteractive-element-to-interactive-role.md)
