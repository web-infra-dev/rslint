# jsx-child-element-spacing

Enforce explicit spacing between inline JSX children when a line break could
silently swallow the gap.

React renders adjacent inline children (such as `<a>` or `<code>`) without an
implicit space when the only thing separating them from neighboring text is a
line break in the source. This rule warns about line-break-only spacing
between an inline element and adjacent text so that author intent stays
explicit — wrap with `{' '}` (or `{/* */}`) to keep the gap, or place the text
and element on the same line to remove it.

## Rule Details

The rule applies to text that is adjacent to one of the HTML inline elements
(`a`, `abbr`, `acronym`, `b`, `bdo`, `big`, `button`, `cite`, `code`, `dfn`,
`em`, `i`, `img`, `input`, `kbd`, `label`, `map`, `object`, `q`, `samp`,
`script`, `select`, `small`, `span`, `strong`, `sub`, `sup`, `textarea`, `tt`,
`var`). `br` is intentionally excluded because spacing around it is
inconsequential to the rendered output.

Examples of **incorrect** code for this rule:

```jsx
<App>
  Please take a look at
  <a href="https://js.org">this link</a>.
</App>
```

```jsx
<App>
  <a>bar</a>
  baz
</App>
```

Examples of **correct** code for this rule:

```jsx
<App>
  Please take a look at <a href="https://js.org">this link</a>.
</App>
```

```jsx
<App>
  Please take a look at
  {' '}
  <a href="https://js.org">this link</a>.
</App>
```

```jsx
<App>
  foo<a>bar</a>baz
</App>
```

Block-level elements (e.g. `<p>`, `<div>`) and custom components (e.g.
`<Foo>`, `<Foo.Bar>`) are not flagged — only the inline element list above is.

## Original Documentation

- [react/jsx-child-element-spacing](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-child-element-spacing.md)
