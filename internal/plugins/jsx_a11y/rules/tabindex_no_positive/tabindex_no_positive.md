# tabindex-no-positive

## Rule Details

Avoid `tabIndex` values greater than zero. A positive `tabIndex` forces the
element to the front of the keyboard tab order regardless of where it sits
in the document — disrupting the natural top-to-bottom flow that sighted
keyboard users and assistive-technology users rely on (see WCAG 2.4.3,
Focus Order).

The rule fires on a JSX attribute when **all** of the following hold:

- The attribute name resolves (case-insensitively) to `tabIndex` —
  `tabIndex`, `tabindex`, `TABINDEX`, etc. all qualify.
- The attribute's value, after `Number(...)` coercion of the
  statically-extracted literal, is a real number greater than zero.

The rule does **not** consider the element type, role, or any plugin
settings — it fires on `<div tabIndex={1} />` and `<MyButton tabIndex={5} />`
alike. Use `jsx-a11y/no-noninteractive-tabindex` if you also need
element-interactivity gating.

A `tabIndex` of `0` is the standard "focusable via Tab in document order"
value and is allowed. A negative `tabIndex` (`-1`, `-100`, ...) is the
standard "focusable only via JavaScript" pattern and is allowed.

Non-integer values are **not** exempt — `tabIndex={1.5}` and `tabIndex="0.5"`
both report. Values that statically resolve to NaN (non-numeric strings,
function calls, identifiers, conditional / logical / binary expressions,
etc.) are silently skipped, since the rule can't statically determine the
final value.

Examples of **incorrect** code for this rule:

```jsx
<span tabIndex="1">foo</span>
<span tabIndex="3">bar</span>
<span tabIndex={5}>baz</span>
<span tabIndex={1.589}>qux</span>
<span tabIndex="0x10">hex</span>
<span tabIndex />               {/* boolean form → Number(true) = 1 */}
<span tabIndex={true} />
```

Examples of **correct** code for this rule:

```jsx
<span tabIndex="0">foo</span>
<span tabIndex="-1">bar</span>
<span tabIndex={0}>baz</span>
<span tabIndex={-1}>qux</span>
<span tabIndex={undefined} />
<span tabIndex={null} />
<span tabIndex={cond ? 1 : 2} />  {/* conditionals aren't statically extracted */}
```

## Rule Options

This rule takes no options.

## Resources

- [WCAG 2.4.3 — Focus Order](https://www.w3.org/WAI/WCAG21/Understanding/focus-order)
- [MDN — `tabindex`](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Global_attributes/tabindex)
- [Chrome Accessibility Developer Tools — AX_FOCUS_03](https://github.com/GoogleChrome/accessibility-developer-tools/wiki/Audit-Rules#ax_focus_03)

## Original Documentation

- [eslint-plugin-jsx-a11y/tabindex-no-positive](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/tabindex-no-positive.md)
