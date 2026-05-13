# no-noninteractive-element-interactions

## Rule Details

Non-interactive HTML elements and non-interactive ARIA roles indicate
_content_ and _containers_ in the user interface. A non-interactive element
does not support event handlers (mouse and key handlers). Non-interactive
elements include `<main>`, `<area>`, `<h1>` (and `<h2>` …), `<p>`, `<img>`,
`<li>`, `<ul>` and `<ol>`. Non-interactive WAI-ARIA roles include `article`,
`banner`, `complementary`, `img`, `listitem`, `main`, `region` and `tooltip`.

When this rule fires, move the event handler to a semantically interactive
element (`<button>`, `<a href>` …) or to an inner element with
`role="presentation"`.

The rule fires on a JSX opening element when **all** of the following hold:

- The resolved element name is in the HTML DOM set (custom React components
  are skipped — the rule does not know what low-level element they render).
- After the per-element allow-list filter (see Rule Options below), the
  element declares at least one of the configured `handlers` whose value
  resolves to a non-nullish expression (`prop={null}` / `prop={undefined}`
  do not count).
- The element does not have `contentEditable="true"` (compared against the
  raw attribute source, so `contentEditable={true}` and
  `contentEditable={"true"}` do NOT exempt).
- The element is not hidden from screen readers (no `aria-hidden={true}` /
  `aria-hidden="true"`, not an `<input type="hidden">`).
- The `role` attribute, when statically a literal string, is not
  `presentation` or `none`.
- The element is neither inherently interactive (e.g. `<button>`,
  `<a href>`) nor carrying an interactive `role`.
- The element IS inherently non-interactive (e.g. `<article>`, `<li>`,
  `<p>`) OR has a non-interactive `role` (e.g. `role="article"`).
- The `role` attribute, when statically a literal string, is not in the
  ARIA abstract role set (`command`, `composite`, `input`, `landmark`,
  `range`, `roletype`, `section`, `sectionhead`, `select`, `structure`,
  `widget`, `window`).

Examples of **incorrect** code for this rule:

```jsx
<li onClick={() => {}} />
<div onClick={() => {}} role="listitem" />
<article onClick={() => {}} />
<h1 onClick={() => {}} />
```

Examples of **correct** code for this rule:

```jsx
<div onClick={() => {}} role="button" />
<div onClick={() => {}} role="presentation" />
<input type="text" onClick={() => {}} />
<button onClick={() => {}} className="foo" />
<div onClick={() => {}} role="button" aria-hidden />
<Input onClick={() => {}} type="hidden" />
```

## Rule Options

### `handlers`

Type: `string[]`. Default: the union of `focus`, `image`, `keyboard`, and
`mouse` event handlers — `onFocus`, `onBlur`, `onLoad`, `onError`,
`onKeyDown`, `onKeyPress`, `onKeyUp`, `onClick`, `onContextMenu`,
`onDblClick`, `onDoubleClick`, `onDrag`, `onDragEnd`, `onDragEnter`,
`onDragExit`, `onDragLeave`, `onDragOver`, `onDragStart`, `onDrop`,
`onMouseDown`, `onMouseEnter`, `onMouseLeave`, `onMouseMove`, `onMouseOut`,
`onMouseOver`, `onMouseUp`.

The upstream `recommended` preset overrides this with `["onClick",
"onError", "onLoad", "onMouseDown", "onMouseUp", "onKeyPress", "onKeyDown",
"onKeyUp"]`; the `strict` preset omits the override and uses the default.

A list of handler prop names that the rule considers an "interactive
listener". Adjust to expand or shrink the rule's coverage surface — an
explicit empty array disables the rule entirely.

Examples of **incorrect** code with `{ "handlers": ["onClick"] }`:

```json
{ "jsx-a11y/no-noninteractive-element-interactions": ["error", { "handlers": ["onClick"] }] }
```

```jsx
<article onClick={() => {}} />
```

Examples of **correct** code with `{ "handlers": ["onClick"] }`:

```json
{ "jsx-a11y/no-noninteractive-element-interactions": ["error", { "handlers": ["onClick"] }] }
```

```jsx
<article onMouseDown={() => {}} />
```

### Per-element allow-list

Type: `string[]`, keyed by HTML element name. Default: not set. The
upstream `recommended` and `strict` presets set `body`, `iframe`, and
`img` to `["onError", "onLoad"]`.

Any option key other than `handlers` is interpreted as an HTML element
name → list of handler-prop names that are permitted on that element. The
rule filters out matching non-spread attributes before the
interactive-handler scan, so `<iframe onLoad={…} />` with `iframe:
["onError", "onLoad"]` short-circuits, while `<iframe onClick={…} />`
still reports.

Examples of **correct** code with `{ "iframe": ["onError", "onLoad"] }`:

```json
{ "jsx-a11y/no-noninteractive-element-interactions": ["error", { "iframe": ["onError", "onLoad"] }] }
```

```jsx
<iframe onLoad={() => {}} />
<iframe onError={() => {}} />
```

Examples of **incorrect** code with `{ "iframe": ["onError", "onLoad"] }`:

```json
{ "jsx-a11y/no-noninteractive-element-interactions": ["error", { "iframe": ["onError", "onLoad"] }] }
```

```jsx
<iframe onClick={() => {}} />
```

## Resources

- [WCAG 4.1.2 — Name, Role, Value](https://www.w3.org/WAI/WCAG21/Understanding/name-role-value)
- [WAI-ARIA — Usage Intro](https://www.w3.org/TR/wai-aria-1.1/#usage_intro)
- [WAI-ARIA Authoring Practices Guide — Design Patterns and Widgets](https://www.w3.org/TR/wai-aria-practices-1.1/#aria_ex)
- [WAI-ARIA Authoring Practices — Keyboard Navigation Conventions](https://www.w3.org/TR/wai-aria-practices-1.1/#kbd_generalnav)
- [MDN — ARIA Techniques: Using the `button` role](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/ARIA_Techniques/Using_the_button_role#Keyboard_and_focus)

## Original Documentation

- [eslint-plugin-jsx-a11y/no-noninteractive-element-interactions](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/no-noninteractive-element-interactions.md)
