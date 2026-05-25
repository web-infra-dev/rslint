# no-aria-hidden-on-focusable

## Rule Details

Disallow `aria-hidden="true"` from being set on focusable elements. An
element with `aria-hidden="true"` is removed from the accessibility tree, so
a screen reader will skip it; if the element is still in the focus order,
keyboard and screen-reader users can land on a control that announces
nothing, causing confusion.

The rule fires on a JSX element when **both** of the following hold:

- The `aria-hidden` prop resolves to JS boolean `true`. The boolean
  attribute form (`<div aria-hidden />`), the case-insensitive strings
  `"true"` / `"True"` / `"TRUE"`, the explicit `{true}`, and any expression
  that statically evaluates to `true` all count.
- The element is focusable per `aria-query`'s element-role map:
  - **Inherently interactive** elements (`<button>`, `<input>`,
    `<textarea>`, `<select>`, `<a href>`, `<area href>`, ...) are focusable
    unless `tabIndex` resolves to a negative integer.
  - **Non-interactive** elements (`<div>`, `<span>`, `<p>`, ...) and
    **custom components** become focusable only when `tabIndex` resolves to
    a non-negative integer.

This rule takes no options.

Examples of **incorrect** code for this rule:

```jsx
<div aria-hidden="true" tabIndex="0" />
<input aria-hidden="true" />
<a href="/" aria-hidden="true" />
<button aria-hidden="true" />
<textarea aria-hidden="true" />
<p tabindex="0" aria-hidden="true">text</p>
```

Examples of **correct** code for this rule:

```jsx
<div aria-hidden="true" />
<div onClick={() => void 0} aria-hidden="true" />
<img aria-hidden="true" />
<a aria-hidden="false" href="#" />
<button aria-hidden="true" tabIndex="-1" />
<button />
<a href="/" />
```

## Resources

- [Deque University — `aria-hidden` elements do not contain focusable elements](https://dequeuniversity.com/rules/axe/html/4.4/aria-hidden-focus)
- [W3C ACT — Element with `aria-hidden` has no content in sequential focus navigation](https://www.w3.org/WAI/standards-guidelines/act/rules/6cfa84/proposed/)
- [MDN — `aria-hidden`](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Attributes/aria-hidden)

## Original Documentation

- [eslint-plugin-jsx-a11y/no-aria-hidden-on-focusable](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/no-aria-hidden-on-focusable.md)
