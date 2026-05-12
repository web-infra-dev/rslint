# interactive-supports-focus

## Rule Details

Enforce that elements with interactive ARIA roles and at least one mouse or
keyboard event handler are also keyboard-reachable — either inherently
focusable, or via an explicit `tabIndex`.

When a non-interactive DOM element such as `<div>` or `<span>` is given an
interactive role (e.g. `role="button"`) and a `onClick` (or any other mouse /
keyboard) handler, screen reader and keyboard-only users still need to be
able to bring focus to it. Either set `tabIndex="0"` (standalone control) or
`tabIndex="-1"` (programmatically focusable element inside a composite
widget), or — better — use a semantic element like `<button>` or `<a href>`
that is already in the tab order.

The rule fires on a JSX opening element when **all** of the following hold:

- The resolved element name is in the HTML DOM set (custom React components
  are skipped — the rule does not know what low-level element they render).
- The element declares at least one mouse or keyboard event handler
  (`onClick`, `onMouseDown`, `onKeyDown`, …).
- The element is not disabled (no HTML5 `disabled` attribute with a value
  other than `undefined`, no `aria-disabled={true}` / `aria-disabled="true"`).
- The element is not hidden from screen readers (no `aria-hidden={true}` /
  `aria-hidden="true"`, not an `<input type="hidden">`).
- The `role` attribute, when statically a literal string, is not
  `presentation` or `none`.
- The `role` attribute resolves to a literal interactive (widget) role.
- The element is neither inherently interactive (e.g. `<button>`, `<a href>`)
  nor inherently non-interactive (e.g. `<article>`, `<li>`, `<p>`), and its
  `role` does not resolve to a non-interactive role.
- The element does not already declare a `tabIndex` (any value upstream
  `getTabIndex` resolves to a non-`undefined` value).

Examples of **incorrect** code for this rule:

```jsx
<div role="button" onClick={() => {}} />
<span role="checkbox" onMouseDown={check} />
<div role="slider" onKeyDown={onKey} />
```

Examples of **correct** code for this rule:

```jsx
<button onClick={() => {}} />
<a href="/" onClick={() => {}} />
<div role="button" tabIndex="0" onClick={() => {}} />
<div role="menuitem" tabIndex="-1" onClick={() => {}} />
<div role="presentation" onClick={() => {}}>
  <button>Save</button>
</div>
```

## Rule Options

### `tabbable`

Type: `string[]`. Default: `[]`. The upstream `recommended` preset sets this
to `["button", "checkbox", "link", "searchbox", "spinbutton", "switch",
"textbox"]`; the `strict` preset adds `progressbar` and `slider`.

A list of ARIA roles that must be sequentially tabbable (`tabIndex="0"`).
When the offending element's role is in this list, the diagnostic reads
`"…must be tabbable."` and the only suggested fix is `tabIndex={0}`. Roles
that are NOT in this list use the `"…must be focusable."` diagnostic and
offer both `tabIndex={0}` and `tabIndex={-1}` as suggestions.

Examples of **incorrect** code with `{ "tabbable": ["button"] }`:

```json
{ "jsx-a11y/interactive-supports-focus": ["error", { "tabbable": ["button"] }] }
```

```jsx
<div role="button" onClick={() => {}} />
```

Examples of **correct** code with `{ "tabbable": ["button"] }`:

```json
{ "jsx-a11y/interactive-supports-focus": ["error", { "tabbable": ["button"] }] }
```

```jsx
<div role="button" tabIndex="0" onClick={() => {}} />
```

## Resources

- [WCAG 2.1.1 — Keyboard](https://www.w3.org/WAI/WCAG21/Understanding/keyboard)
- [WAI-ARIA Authoring Practices — Keyboard Interaction](https://www.w3.org/WAI/ARIA/apg/practices/keyboard-interface/)
- [MDN — Keyboard-navigable JavaScript widgets](https://developer.mozilla.org/en-US/docs/Web/Accessibility/Keyboard-navigable_JavaScript_widgets)
- [MDN — `tabindex`](https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/tabindex)

## Original Documentation

- [eslint-plugin-jsx-a11y/interactive-supports-focus](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/interactive-supports-focus.md)
