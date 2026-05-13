# no-static-element-interactions

## Rule Details

Static HTML elements such as `<div>` and `<span>` carry no semantic meaning.
Attaching mouse / keyboard / focus event handlers to them without an
interactive ARIA `role` makes the element invisible to assistive technology
— screen-reader and keyboard-only users cannot tell that the element is
supposed to respond to interaction. Prefer a semantic element (`<button>`,
`<a href>`, `<input>`, …); if that is not possible, give the element an
appropriate `role` and the keyboard support the role implies.

The rule reports a JSX opening element when **all** of the following hold:

- The resolved element name is in the HTML DOM set (custom React components
  are skipped — the rule does not know what low-level element they render).
- The element declares at least one interactive event handler from the
  configured `handlers` list (default: every focus, keyboard, and mouse
  handler exposed by `jsx-ast-utils`), and that handler's value statically
  resolves to something other than `null` / `undefined`.
- The element is not hidden from screen readers (no `aria-hidden={true}` /
  `aria-hidden="true"`, not an `<input type="hidden">`).
- The `role` attribute, when statically a literal string, is not
  `presentation` or `none`.
- The element is neither inherently interactive (e.g. `<button>`,
  `<a href>`, `<input type="button">`) nor inherently non-interactive
  (e.g. `<article>`, `<li>`, `<p>`); its `role` does not resolve to an
  interactive, non-interactive, or abstract ARIA role.

Examples of **incorrect** code for this rule:

```jsx
<div onClick={() => {}} />
<span onKeyDown={() => {}} />
<a onClick={() => {}} />
<section onClick={() => {}} />
```

Examples of **correct** code for this rule:

```jsx
<button onClick={() => {}} />
<a href="/profile" onClick={() => {}} />
<input type="button" onClick={() => {}} />
<div role="button" tabIndex={0} onClick={() => {}} />
<div role="presentation" onClick={() => {}} />
<div onClick={null} />
<article onClick={() => {}} />
```

## Rule Options

### `handlers`

Type: `string[]`. Default: `["onFocus", "onBlur", "onKeyDown", "onKeyPress", "onKeyUp", "onClick", "onContextMenu", "onDblClick", "onDoubleClick", "onDrag", "onDragEnd", "onDragEnter", "onDragExit", "onDragLeave", "onDragOver", "onDragStart", "onDrop", "onMouseDown", "onMouseEnter", "onMouseLeave", "onMouseMove", "onMouseOut", "onMouseOver", "onMouseUp"]`. The upstream `recommended` preset narrows this to `["onClick", "onMouseDown", "onMouseUp", "onKeyPress", "onKeyDown", "onKeyUp"]`.

Overrides the list of event-handler prop names the rule treats as
"interactive". Only attributes whose name (case-insensitive) is in this
list participate in the check. An empty array (`handlers: []`) disables
the rule — no handler ever matches.

Examples of **incorrect** code with `{ "handlers": ["onCustomClick"] }`:

```json
{ "jsx-a11y/no-static-element-interactions": ["error", { "handlers": ["onCustomClick"] }] }
```

```jsx
<div onCustomClick={() => {}} />
```

Examples of **correct** code with `{ "handlers": ["onCustomClick"] }`:

```json
{ "jsx-a11y/no-static-element-interactions": ["error", { "handlers": ["onCustomClick"] }] }
```

```jsx
<div onClick={() => {}} />
```

### `allowExpressionValues`

Type: `boolean`. Default: `false`. The upstream `recommended` preset sets
this to `true`.

When `true`, an element whose `role` attribute is a non-literal expression
(e.g. `role={ROLE_BUTTON}`, `role={isButton ? "button" : "link"}`) is
exempt — the rule cannot statically determine whether the role is
interactive, and the option opts into trusting the developer. Note that
`role={undefined}` is treated as a literal-equivalent and is **not**
exempted by this option.

Examples of **correct** code with `{ "allowExpressionValues": true }`:

```json
{ "jsx-a11y/no-static-element-interactions": ["error", { "allowExpressionValues": true }] }
```

```jsx
<div role={ROLE_BUTTON} onClick={() => {}} />
<div role={isButton ? "button" : "link"} onClick={() => {}} />
```

Examples of **incorrect** code with `{ "allowExpressionValues": false }`:

```json
{ "jsx-a11y/no-static-element-interactions": ["error", { "allowExpressionValues": false }] }
```

```jsx
<div role={ROLE_BUTTON} onClick={() => {}} />
```

## Resources

- [WCAG 4.1.2 — Name, Role, Value](https://www.w3.org/WAI/WCAG21/Understanding/name-role-value)
- [WAI-ARIA — Widget Roles](https://www.w3.org/TR/wai-aria-1.2/#widget_roles)
- [MDN — ARIA: button role](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Roles/button_role)

## Original Documentation

- [eslint-plugin-jsx-a11y/no-static-element-interactions](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/no-static-element-interactions.md)
